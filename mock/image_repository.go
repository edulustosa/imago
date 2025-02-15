package mock

import (
	"context"
	"errors"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/google/uuid"
)

type ImageRepo struct {
	Images []models.Image
}

func NewImageRepo() *ImageRepo {
	return &ImageRepo{}
}

func (r *ImageRepo) FindByFilename(
	_ context.Context,
	filename string,
	userID uuid.UUID,
) (*models.Image, error) {
	for _, img := range r.Images {
		if img.Filename == filename && img.UserID == userID {
			return &img, nil
		}
	}

	return nil, errors.New("image not found")
}

func (r *ImageRepo) Create(_ context.Context, img models.Image) (*models.Image, error) {
	img.ID = len(r.Images) + 1
	img.CreatedAt = time.Now()
	img.UpdatedAt = time.Now()

	r.Images = append(r.Images, img)
	return &img, nil
}

func (r *ImageRepo) FindByID(_ context.Context, id int, userID uuid.UUID) (*models.Image, error) {
	for _, img := range r.Images {
		if img.ID == id && img.UserID == userID {
			return &img, nil
		}
	}

	return nil, errors.New("image not found")
}

func (r *ImageRepo) Update(
	_ context.Context,
	id int,
	usr uuid.UUID,
	imgInfo models.Image,
) (*models.Image, error) {
	for i := range r.Images {
		if r.Images[i].ID == id && r.Images[i].UserID == usr {
			r.Images[i] = imgInfo
			r.Images[i].UpdatedAt = time.Now()
			return &r.Images[i], nil
		}
	}

	return nil, errors.New("image not found")
}

func (r *ImageRepo) FindManyByUserID(
	_ context.Context,
	userID uuid.UUID,
	page, limit int,
) ([]models.Image, error) {
	var images []models.Image
	for _, img := range r.Images {
		if img.UserID == userID {
			images = append(images, img)
		}
	}

	images = images[(page-1)*limit : page*limit]

	return images, nil
}
