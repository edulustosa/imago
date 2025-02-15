package upload

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Uploader interface {
	Upload(ctx context.Context, imgData []byte, filepath string) (string, error)
	GetImage(ctx context.Context, filepath string) (string, error)
}

type fsUploader struct {
	baseURL string
}

func NewFileSystemUploader(baseURL string) Uploader {
	return &fsUploader{
		baseURL,
	}
}

func (f *fsUploader) Upload(
	_ context.Context,
	imgData []byte,
	path string,
) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, imgData, 0600); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filepath.Join(f.baseURL, path), nil
}

func (f *fsUploader) GetImage(
	_ context.Context,
	path string,
) (string, error) {
	_, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	return filepath.Join(f.baseURL, path), nil
}

type s3Uploader struct {
	client *s3.Client
	bucket string
}

func NewS3Uploader(client *s3.Client, bucket string) Uploader {
	return &s3Uploader{
		client,
		bucket,
	}
}

func (s *s3Uploader) Upload(
	ctx context.Context,
	imgData []byte,
	filepath string,
) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Body:   bytes.NewReader(imgData),
		Key:    aws.String(filepath),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, filepath), nil
}

func (s *s3Uploader) GetImage(ctx context.Context, filepath string) (string, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filepath),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, filepath), nil
}
