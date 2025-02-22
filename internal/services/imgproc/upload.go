package imgproc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/google/uuid"
)

type Upload struct {
	userRepository  user.Repository
	imageRepository img.Repository
	imageStorage    storage.ImageStorage
}

func NewUpload(
	userRepository user.Repository,
	imageRepository img.Repository,
	imageStorage storage.ImageStorage,
) *Upload {
	return &Upload{
		userRepository,
		imageRepository,
		imageStorage,
	}
}

type ImageMetadata struct {
	Filename string
	Format   string
	Alt      string
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidImage = errors.New("failed to decode image: invalid format")
)

func (u *Upload) Do(
	ctx context.Context,
	userID uuid.UUID,
	imgFile []byte,
	metadata *ImageMetadata,
) (*models.Image, error) {
	usr, err := u.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	_, format, err := image.Decode(bytes.NewReader(imgFile))
	if err != nil || !isSameFormat(format, metadata.Format) {
		return nil, ErrInvalidImage
	}

	imgURL, err := u.imageStorage.Upload(
		ctx,
		imgFile,
		fmt.Sprintf("%s/%s", usr.ID.String(), metadata.Filename),
	)
	if err != nil {
		return nil, err
	}

	imgInfo, err := u.imageRepository.FindByFilename(ctx, metadata.Filename, usr.ID)
	if err == nil {
		return imgInfo, nil
	}

	img := models.Image{
		UserID:   usr.ID,
		ImageURL: imgURL,
		Filename: metadata.Filename,
		Format:   metadata.Format,
		Alt:      metadata.Alt,
	}

	return u.imageRepository.Create(ctx, img)
}

var equivFormats = map[string]string{
	"jpeg": "jpg",
	"jpg":  "jpeg",
	"tiff": "tif",
	"tif":  "tiff",
}

func isSameFormat(decodedFormat, metadataFormat string) bool {
	if decodedFormat == metadataFormat {
		return true
	}

	if equiv, ok := equivFormats[decodedFormat]; ok {
		return equiv == metadataFormat
	}

	return false
}
