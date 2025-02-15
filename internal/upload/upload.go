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
	DownloadImage(ctx context.Context, url string) ([]byte, error)
	Delete(ctx context.Context, filepath string) error
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
	if err := os.MkdirAll(filepath.Join(f.baseURL, dir), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(f.baseURL, path), imgData, 0600); err != nil {
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

func (f *fsUploader) DownloadImage(
	_ context.Context,
	url string,
) ([]byte, error) {
	return os.ReadFile(url)
}

func (f *fsUploader) Delete(
	_ context.Context,
	path string,
) error {
	return os.Remove(path)
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

func (s *s3Uploader) DownloadImage(ctx context.Context, url string) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(url),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *s3Uploader) Delete(ctx context.Context, filepath string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filepath),
	})

	return err
}
