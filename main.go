package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edulustosa/imago/config"
	"github.com/edulustosa/imago/internal/api/router"
	"github.com/edulustosa/imago/internal/queue"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(
		ctx,
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error(err.Error())

		cancel()
		os.Exit(1)
	}

	slog.Info("To infinity and beyond!")
}

func run(ctx context.Context) error {
	env, err := config.LoadEnv(".")
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, env.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	s3Client, err := loadS3Client(ctx, env)
	if err != nil {
		return fmt.Errorf("failed to load S3 client: %w", err)
	}

	redisClient, err := connectToRedis(ctx, env.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisClient.Close()

	kafkaWriter := &kafka.Writer{
		Addr:                   kafka.TCP(env.KafkaBroker),
		Topic:                  env.KafkaTasksTopic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	r := router.New(router.Server{
		Database:    pool,
		Env:         env,
		S3Client:    s3Client,
		RedisClient: redisClient,
		KafkaWriter: kafkaWriter,
	})
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", env.Addr),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{env.KafkaBroker},
		Topic:       env.KafkaTasksTopic,
		GroupID:     "imago-transformation-processor",
		MaxBytes:    10e6,
		MinBytes:    1e3,
		StartOffset: kafka.LastOffset,
	})

	consumer := queue.NewTransformationConsumer(
		kafkaReader,
		redisClient,
		pool,
		s3Client,
		env.BucketName,
	)
	consumer.Start(ctx)
	defer consumer.Stop()

	defer func() {
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown server", "error", err)
		}
	}()

	errChan := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", env.Addr)
		errChan <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	return nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS("./internal/database/migrations"),
	)
	if err != nil {
		return err
	}

	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	return nil
}

func loadS3Client(ctx context.Context, env *config.Env) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(env.AWSRegion),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(env.AWSAccessKey, env.AWSSecretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

func connectToRedis(ctx context.Context, url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
