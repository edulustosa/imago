package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/images"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/upload"
	"github.com/google/uuid"
)

type ImageStorage struct {
	uploader        upload.Uploader
	userRepository  user.Repository
	imageRepository images.Repository
}

func NewImageStorage(
	uploader upload.Uploader,
	userRepository user.Repository,
	imageRepository images.Repository,
) *ImageStorage {
	return &ImageStorage{
		uploader,
		userRepository,
		imageRepository,
	}
}

var ErrUserNotFound = errors.New("user not found")

type Metadata struct {
	Filename string
	Format   string
	Alt      string
}

func (s *ImageStorage) Upload(
	ctx context.Context,
	userID uuid.UUID,
	img []byte,
	metadata Metadata,
) (*models.Image, error) {
	usr, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	imgURL, err := s.uploader.Upload(
		ctx,
		img,
		fmt.Sprintf("%s/%s", userID.String(), metadata.Filename),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	imgInfo, err := s.imageRepository.FindByFilename(ctx, metadata.Filename)
	if err == nil {
		return imgInfo, nil
	}

	imgModel := models.Image{
		UserID:   usr.ID,
		ImageURL: imgURL,
		Filename: metadata.Filename,
		Format:   metadata.Format,
		Alt:      metadata.Alt,
	}

	imgInfo, err = s.imageRepository.Create(ctx, imgModel)
	if err != nil {
		return nil, fmt.Errorf("failed to save image info: %w", err)
	}

	return imgInfo, nil
}
