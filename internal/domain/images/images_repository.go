package images

import (
	"context"
	"fmt"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, img models.Image) (*models.Image, error)
	FindByFilename(ctx context.Context, filename string) (*models.Image, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repo{db}
}

const create = `
	INSERT INTO images (
		user_id,
		image_url,
		filename,
		format,
		alt	
	) VALUES ($1, $2, $3, $4, $5)
	RETURNING *
`

func scanImage(row pgx.Row) (*models.Image, error) {
	var img models.Image
	err := row.Scan(
		&img.ID,
		&img.UserID,
		&img.ImageURL,
		&img.Filename,
		&img.Format,
		&img.Alt,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not scan image: %w", err)
	}
	return &img, nil
}

func (r *repo) Create(ctx context.Context, img models.Image) (*models.Image, error) {
	row := r.db.QueryRow(
		ctx,
		create,
		img.UserID,
		img.ImageURL,
		img.Filename,
		img.Format,
		img.Alt,
	)
	return scanImage(row)
}

const findByFilename = "SELECT * FROM images WHERE filename = $1"

func (r *repo) FindByFilename(ctx context.Context, filename string) (*models.Image, error) {
	row := r.db.QueryRow(ctx, findByFilename, filename)
	return scanImage(row)
}
