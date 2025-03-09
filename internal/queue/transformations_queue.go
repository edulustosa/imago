package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/google/uuid"
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
