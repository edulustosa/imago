package images

import (
	"context"

	"github.com/edulustosa/imago/internal/database/models"
)

type Repository interface {
	Create(ctx context.Context, img models.Image) (*models.Image, error)
}
