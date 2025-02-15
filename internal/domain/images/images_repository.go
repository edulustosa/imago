package images

import (
	"context"
	"fmt"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByID(ctx context.Context, id int, userID uuid.UUID) (*models.Image, error)
	Create(ctx context.Context, img models.Image) (*models.Image, error)
	FindByFilename(ctx context.Context, filename string, userID uuid.UUID) (*models.Image, error)
	Update(ctx context.Context, id int, userID uuid.UUID, imgInfo models.Image) (*models.Image, error)
	FindManyByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.Image, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repo{db}
}

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

const findByID = "SELECT * FROM images WHERE id = $1 AND user_id = $2"

func (r *repo) FindByID(
	ctx context.Context,
	id int,
	userID uuid.UUID,
) (*models.Image, error) {
	row := r.db.QueryRow(ctx, findByID, id, userID)
	return scanImage(row)
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

const findByFilename = "SELECT * FROM images WHERE filename = $1 AND user_id = $2"

func (r *repo) FindByFilename(
	ctx context.Context,
	filename string,
	userID uuid.UUID,
) (*models.Image, error) {
	row := r.db.QueryRow(ctx, findByFilename, filename, userID)
	return scanImage(row)
}

const update = `
	UPDATE images
	SET image_url = $1,
		filename = $2,
		format = $3,
		alt = $4,
		updated_at = NOW()
	WHERE id = $5 AND user_id = $6
	RETURNING *
`

func (r *repo) Update(
	ctx context.Context,
	id int,
	userID uuid.UUID,
	imgInfo models.Image,
) (*models.Image, error) {
	row := r.db.QueryRow(
		ctx,
		update,
		imgInfo.ImageURL,
		imgInfo.Filename,
		imgInfo.Format,
		imgInfo.Alt,
		id,
		userID,
	)
	return scanImage(row)
}

const findManyByUserID = `
	SELECT * FROM images
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
`

func (r *repo) FindManyByUserID(
	ctx context.Context,
	userID uuid.UUID,
	page, limit int,
) ([]models.Image, error) {
	rows, err := r.db.Query(ctx, findManyByUserID, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("could not query images: %w", err)
	}

	images := make([]models.Image, 0, limit)
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, err
		}

		images = append(images, *img)
	}

	return images, nil
}
