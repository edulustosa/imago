package imgproc_test

import (
	"context"
	"os"
	"testing"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/img"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/services/imgproc"
	"github.com/edulustosa/imago/internal/storage"
	"github.com/google/uuid"
)

func TestUploadImage(t *testing.T) {
	ctx := context.Background()

	userRepo := user.NewMemoryRepo()
	imgRepo := img.NewMemoryRepo()
	imageStore := storage.NewFSImageStorage("test_data")

	sut := imgproc.NewUpload(userRepo, imgRepo, imageStore)
	imgData, err := os.ReadFile("./test_data/flowers.jpg")
	if err != nil {
		t.Fatalf("could not read image file: %v", err)
	}

	t.Run("upload", func(t *testing.T) {
		usr, _ := userRepo.Create(ctx, models.User{
			Username:     "test",
			PasswordHash: "test",
		})

		t.Cleanup(func() {
			_ = os.RemoveAll("./test_data/" + usr.ID.String())
		})

		imgInfo, err := sut.Do(ctx, usr.ID, imgData, &imgproc.ImageMetadata{
			Filename: "flowers.jpg",
			Format:   "jpeg",
			Alt:      "flowers",
		})
		if err != nil {
			t.Errorf("could not upload image: %v", err)
		}

		_, err = os.ReadFile(imgInfo.ImageURL)
		if err != nil {
			t.Error("could not read uploaded image")
		}
	})

	reset(userRepo, imgRepo)

	t.Run("invalid user", func(t *testing.T) {
		_, err := sut.Do(ctx, uuid.Nil, imgData, &imgproc.ImageMetadata{
			Filename: "flowers.jpg",
			Format:   "jpeg",
			Alt:      "flowers",
		})
		if err != imgproc.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})

	reset(userRepo, imgRepo)

	t.Run("invalid image", func(t *testing.T) {
		usr, _ := userRepo.Create(ctx, models.User{
			Username:     "test",
			PasswordHash: "test",
		})

		_, err := sut.Do(ctx, usr.ID, imgData, &imgproc.ImageMetadata{
			Filename: "flowers.jpg",
			Format:   "png",
			Alt:      "flowers",
		})
		if err != imgproc.ErrInvalidImage {
			t.Errorf("expected ErrInvalidImage, got: %v", err)
		}
	})

	reset(userRepo, imgRepo)

	t.Run("image already exists", func(t *testing.T) {
		usr, _ := userRepo.Create(ctx, models.User{
			Username:     "test",
			PasswordHash: "test",
		})

		t.Cleanup(func() {
			_ = os.RemoveAll("./test_data/" + usr.ID.String())
		})

		imgInfo, err := sut.Do(ctx, usr.ID, imgData, &imgproc.ImageMetadata{
			Filename: "flowers.jpg",
			Format:   "jpeg",
			Alt:      "flowers",
		})
		if err != nil {
			t.Errorf("could not upload image: %v", err)
		}

		secondImgInfo, err := sut.Do(ctx, usr.ID, imgData, &imgproc.ImageMetadata{
			Filename: "flowers.jpg",
			Format:   "jpeg",
			Alt:      "flowers",
		})
		if err != nil {
			t.Errorf("could not upload second image: %v", err)
		}

		if imgInfo.ID != secondImgInfo.ID {
			t.Error("expected the same image to be returned")
		}
	})
}

func reset(userRepo *user.MemoryRepo, image *img.MemoryRepo) {
	userRepo.Users = []models.User{}
	image.Images = []models.Image{}
}
