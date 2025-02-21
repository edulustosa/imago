package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/edulustosa/imago/internal/services/storage"
	"github.com/edulustosa/imago/internal/upload"
	"github.com/edulustosa/imago/mock"
	"github.com/google/uuid"
)

func TestUploadImage(t *testing.T) {
	ctx := context.Background()

	userRepository := mock.NewUserRepo()
	imageRepository := mock.NewImageRepo()
	uploader := upload.NewFileSystemUploader("./test_data")

	sut := storage.NewImageStorage(uploader, userRepository, imageRepository)

	img, err := getImageData()
	if err != nil {
		t.Errorf("failed to open image: %v", err)
	}

	t.Run("upload image", func(t *testing.T) {
		user, _ := userRepository.Create(ctx, models.User{
			Username:     "test",
			PasswordHash: "test",
		})

		_, err := sut.Upload(ctx, user.ID, img, storage.Metadata{
			Filename: "flowers.png",
			Format:   "png",
			Alt:      "flowers",
		})
		if err != nil {
			t.Errorf("failed to upload image: %v", err)
		}

		t.Cleanup(func() {
			if err := os.RemoveAll(filepath.Join("./test_data", user.ID.String())); err != nil {
				t.Errorf("failed to remove dir: %v", err)
			}
		})

		_, err = os.Open(fmt.Sprintf("./test_data/%s/flowers.png", user.ID))
		if err != nil {
			t.Errorf("flowers.png was not uploaded")
		}
	})

	userRepository.Users = []models.User{}
	imageRepository.Images = []models.Image{}

	t.Run("invalid user", func(t *testing.T) {
		want := storage.ErrUserNotFound

		_, got := sut.Upload(ctx, uuid.New(), img, storage.Metadata{
			Filename: "flowers.png",
			Format:   "png",
			Alt:      "flowers",
		})

		if got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})
}

func TestImageTransformations(t *testing.T) {
	ctx := context.Background()

	userRepository := mock.NewUserRepo()
	imageRepository := mock.NewImageRepo()
	uploader := upload.NewFileSystemUploader("./test_data")

	sut := storage.NewImageStorage(uploader, userRepository, imageRepository)

	img, err := getImageData()
	if err != nil {
		t.Errorf("failed to open image: %v", err)
	}

	t.Run("transform image", func(t *testing.T) {
		user, _ := userRepository.Create(ctx, models.User{
			Username:     "test",
			PasswordHash: "test",
		})

		imgInfo, err := sut.Upload(ctx, user.ID, img, storage.Metadata{
			Filename: "flowers.png",
			Format:   "png",
			Alt:      "flowers",
		})
		if err != nil {
			t.Errorf("failed to upload image: %v", err)
		}

		t.Cleanup(func() {
			if err := os.RemoveAll(filepath.Join("./test_data", user.ID.String())); err != nil {
				t.Errorf("failed to remove dir: %v", err)
			}
		})

		_, err = sut.Transform(ctx, user.ID, imgInfo.ID, &imgproc.Transformations{
			Resize: imgproc.Resize{
				Width:  100,
				Height: 100,
			},
			Format: "jpeg",
		})
		if err != nil {
			t.Errorf("failed to transform image: %v", err)
		}

		img, err := imgio.Open(fmt.Sprintf("./test_data/%s/flowers.jpeg", user.ID.String()))
		if err != nil {
			t.Errorf("failed to open image: %v", err)
		}

		if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
			t.Errorf("image dimensions are incorrect")
		}
	})
}

func getImageData() ([]byte, error) {
	img, err := imgio.Open("./test_data/flowers.jpg")
	if err != nil {
		return nil, err
	}

	imgBuffer := new(bytes.Buffer)
	if err := png.Encode(imgBuffer, img); err != nil {
		return nil, err
	}

	return imgBuffer.Bytes(), nil
}
