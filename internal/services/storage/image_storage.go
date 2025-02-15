package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/images"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/imago"
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

	imgInfo, err := s.imageRepository.FindByFilename(
		ctx,
		metadata.Filename,
		usr.ID,
	)
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

var (
	ErrImageNotFound = errors.New("image not found")
	ErrInvalidFormat = errors.New("invalid format")
)

func (s *ImageStorage) Transform(
	ctx context.Context,
	userID uuid.UUID,
	imgID int,
	t imago.Transformations,
) (*models.Image, error) {
	usr, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	imgInfo, err := s.imageRepository.FindByID(ctx, imgID, usr.ID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	imgData, err := s.uploader.DownloadImage(
		ctx,
		fmt.Sprintf("%s/%s", userID.String(), imgInfo.Filename),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	img = imago.Transform(img, t)
	imgBuffer := new(bytes.Buffer)
	if err := imago.Encode(imgBuffer, img, t.Format); err != nil {
		return nil, ErrInvalidFormat
	}

	filename := updateFilenameExt(imgInfo.Filename, t.Format)
	newImageURL, err := s.uploader.Upload(
		ctx,
		imgBuffer.Bytes(),
		fmt.Sprintf(
			"%s/%s",
			userID.String(),
			filename,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	if t.Format != imgInfo.Format {
		go func() {
			const timeout = 10 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := s.uploader.Delete(
				ctx,
				fmt.Sprintf("%s/%s", userID.String(), imgInfo.Filename),
			); err != nil {
				slog.Error("failed to delete image", "msg", err.Error())
			}
		}()
	}

	imgInfo, err = s.imageRepository.Update(ctx, imgID, usr.ID, models.Image{
		UserID:   usr.ID,
		ImageURL: newImageURL,
		Filename: filename,
		Format:   t.Format,
		Alt:      imgInfo.Alt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update image info: %w", err)
	}

	return imgInfo, nil
}

func updateFilenameExt(filename, newExt string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	return fmt.Sprintf("%s.%s", base, newExt)
}
