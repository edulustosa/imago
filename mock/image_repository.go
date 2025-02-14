package mock

import (
	"context"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
)

type ImageRepo struct {
	Images []models.Image
}

func NewImageRepo() *ImageRepo {
	return &ImageRepo{}
}

func (r *ImageRepo) Create(_ context.Context, img models.Image) (*models.Image, error) {
	img.ID = len(r.Images) + 1
	img.CreatedAt = time.Now()
	img.UpdatedAt = time.Now()

	r.Images = append(r.Images, img)
	return &img, nil
}
