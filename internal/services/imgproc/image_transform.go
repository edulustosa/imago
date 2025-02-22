package imgproc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/google/uuid"
)

type ImageTransformation struct {
	imageRepository img.Repository
	imageStorage    storage.ImageStorage
}

func NewImageTransformation(
	imageRepository img.Repository,
	imageStorage storage.ImageStorage,
) *ImageTransformation {
	return &ImageTransformation{
		imageRepository,
		imageStorage,
	}
}

var ErrImageNotFound = errors.New("image not found: invalid image or user id")

func (it *ImageTransformation) Transform(
	ctx context.Context,
	imageID int,
	userID uuid.UUID,
	t *Transformations,
) (*models.Image, error) {
	imgInfo, err := it.imageRepository.FindByID(ctx, imageID, userID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	imgFile, err := it.imageStorage.DownloadImage(
		ctx,
		fmt.Sprintf("%s/%s", userID.String(), imgInfo.Filename),
	)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	processedImgData, err := processImage(imgFile, t)
	if err != nil {
		return nil, err
	}

	filename := imgInfo.Filename
	if imgInfo.Format != t.Format {
		go it.deleteOriginalImage(userID, imgInfo.Filename)
		filename = changeFileExtension(filename, t.Format)
	}

	imgURL, err := it.imageStorage.Upload(
		ctx,
		processedImgData,
		fmt.Sprintf("%s/%s", userID.String(), filename),
	)
	if err != nil {
		return nil, err
	}

	return it.imageRepository.Update(ctx, imageID, userID, models.Image{
		ImageURL: imgURL,
		Filename: filename,
		Format:   t.Format,
		Alt:      imgInfo.Alt,
	})
}

func processImage(imgFile io.Reader, t *Transformations) ([]byte, error) {
	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("is here")
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	img = Transform(img, t)
	imgBuff := new(bytes.Buffer)
	if err := Encode(imgBuff, img, t.Format); err != nil {
		return nil, err
	}

	return imgBuff.Bytes(), nil
}

func (it *ImageTransformation) deleteOriginalImage(userID uuid.UUID, filename string) {
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := it.imageStorage.Delete(
		ctx,
		fmt.Sprintf("%s/%s", userID.String(), filename),
	)
	if err != nil {
		slog.Error(
			"failed to delete original image",
			"msg", err,
			"user id", userID,
			"image", filename,
		)
	}
}

func changeFileExtension(filename, newExt string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	return fmt.Sprintf("%s.%s", base, newExt)
}
