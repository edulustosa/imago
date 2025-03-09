package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type TransformationProducer struct {
	writer *kafka.Writer
	redis  *redis.Client
}

func NewTransformationProducer(writer *kafka.Writer, redis *redis.Client) *TransformationProducer {
	return &TransformationProducer{
		writer,
		redis,
	}
}

type TransformationMessage struct {
	ImageID         int                      `json:"imageId"`
	UserID          uuid.UUID                `json:"userId"`
	Transformations *imgproc.Transformations `json:"transformations"`
}

type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type TransformationStatus struct {
	StatusID     uuid.UUID `json:"statusId"`
	ImageID      int       `json:"imageId"`
	Status       Status    `json:"status"`
	ErrorMessage string    `json:"error"`
}

func (p *TransformationProducer) Enqueue(
	ctx context.Context,
	message *TransformationMessage,
) (*TransformationStatus, error) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	callbackID := uuid.New()
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(callbackID.String()),
		Value: msgBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	status := &TransformationStatus{
		StatusID: callbackID,
		ImageID:  message.ImageID,
		Status:   StatusPending,
	}

	statusBytes, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}

	err = p.redis.Set(ctx, callbackID.String(), statusBytes, time.Hour).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to set status in redis: %w", err)
	}

	return status, nil
}

type TransformationConsumer struct {
	reader       *kafka.Reader
	redis        *redis.Client
	db           *pgxpool.Pool
	s3Client     *s3.Client
	bucketName   string
	processingWg sync.WaitGroup
}

func NewTransformationConsumer(
	reader *kafka.Reader,
	redis *redis.Client,
	db *pgxpool.Pool,
	s3Client *s3.Client,
	bucketName string,
) *TransformationConsumer {
	return &TransformationConsumer{
		reader:     reader,
		redis:      redis,
		db:         db,
		s3Client:   s3Client,
		bucketName: bucketName,
	}
}

const numOfWorkers = 5

func (c *TransformationConsumer) Start(ctx context.Context) {
	msgChan := make(chan kafka.Message, numOfWorkers*2)
	for i := range numOfWorkers {
		workerID := i
		c.processingWg.Add(1)
		go c.runWorker(ctx, workerID, msgChan)
	}

	go func() {
		slog.Info("starting transformation consumer")
		defer c.reader.Close()

		for {
			select {
			case <-ctx.Done():
				slog.Info("stopping transformation consumer")
				close(msgChan)
				return
			default:
				msg, err := c.reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != context.Canceled {
						slog.Error("failed to read message", "error", err)
					}
					continue
				}

				msgChan <- msg
			}

		}
	}()
}

func (c *TransformationConsumer) runWorker(ctx context.Context, id int, msgChan <-chan kafka.Message) {
	defer c.processingWg.Done()

	slog.Info("starting transformation worker", "worker_id", id)

	for msg := range msgChan {
		select {
		case <-ctx.Done():
			return
		default:
			callbackID := string(msg.Key)
			slog.Info("processing transformation", "worker_id", id, "callbackID", callbackID)

			if err := c.processMessage(ctx, msg); err != nil {
				slog.Error(
					"failed to process transformation",
					"worker_id", id,
					"callbackID", callbackID,
					"error", err,
				)
			}

			slog.Info("processed transformation", "worker_id", id, "callbackID", callbackID)
		}
	}

	slog.Info("stopping transformation worker", "worker_id", id)
}

func (c *TransformationConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var transformationMessage TransformationMessage
	if err := json.Unmarshal(msg.Value, &transformationMessage); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	callbackID := uuid.MustParse(string(msg.Key))
	status := &TransformationStatus{
		StatusID: callbackID,
		ImageID:  transformationMessage.ImageID,
		Status:   StatusPending,
	}

	if err := c.transformImage(ctx, &transformationMessage); err != nil {
		status.Status = StatusFailed
		status.ErrorMessage = err.Error()
	}

	status.Status = StatusDone
	if err := c.updateStatus(ctx, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (c *TransformationConsumer) transformImage(ctx context.Context, msg *TransformationMessage) error {
	imgRepository := img.NewRepo(c.db)
	imgStorage := storage.NewS3ImageStorage(c.s3Client, c.bucketName)

	transformationService := imgproc.NewImageTransformation(imgRepository, imgStorage)
	_, err := transformationService.Transform(ctx, msg.ImageID, msg.UserID, msg.Transformations)
	if err != nil {
		return err
	}

	return nil
}

func (c *TransformationConsumer) updateStatus(
	ctx context.Context,
	status *TransformationStatus,
) error {
	statusBytes, err := json.Marshal(status)
	if err != nil {
		return err
	}

	err = c.redis.Set(ctx, status.StatusID.String(), statusBytes, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set status in redis: %w", err)
	}

	return nil
}

func (c *TransformationConsumer) Stop() {
	c.processingWg.Wait()
}
