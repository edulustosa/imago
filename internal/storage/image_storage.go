package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageStorage interface {
	Upload(ctx context.Context, imgData []byte, path string) (string, error)
	GetImage(ctx context.Context, path string) (string, error)
	DownloadImage(ctx context.Context, url string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

type s3ImageStorage struct {
	client *s3.Client
	bucket string
}

func NewS3ImageStorage(client *s3.Client, bucket string) ImageStorage {
	return &s3ImageStorage{
		client,
		bucket,
	}
}

func (s *s3ImageStorage) Upload(
	ctx context.Context,
	imgData []byte,
	path string,
) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Key:    aws.String(path),
		Bucket: aws.String(s.bucket),
		Body:   bytes.NewReader(imgData),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, path), nil
}

func (s *s3ImageStorage) GetImage(ctx context.Context, path string) (string, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get image or does not exist: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, path), nil
}

func (s *s3ImageStorage) DownloadImage(ctx context.Context, url string) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(url),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	return resp.Body, nil
}

func (s *s3ImageStorage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	return err
}

type fsImageStorage struct {
	root string
}

func NewFSImageStorage(root string) ImageStorage {
	return &fsImageStorage{
		root,
	}
}

func (f *fsImageStorage) Upload(
	_ context.Context,
	imgData []byte,
	path string,
) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(filepath.Join(f.root, dir), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(f.root, path), imgData, 0600); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	return filepath.Join(f.root, path), nil
}

func (f *fsImageStorage) GetImage(_ context.Context, path string) (string, error) {
	_, err := os.Open(path)
	return filepath.Join(f.root, path), err
}

func (f *fsImageStorage) DownloadImage(_ context.Context, url string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(f.root, url))
}

func (f *fsImageStorage) Delete(_ context.Context, path string) error {
	return os.Remove(filepath.Join(f.root, path))
}
