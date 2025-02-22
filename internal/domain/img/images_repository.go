package img

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByID(ctx context.Context, id int, userID uuid.UUID) (*models.Image, error)
	Create(ctx context.Context, imgInfo models.Image) (*models.Image, error)
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

	return &img, err
}

const findByID = "SELECT * FROM images WHERE id = $1 AND user_id = $2"

func (r *repo) FindByID(
	ctx context.Context,
	id int,
	userID uuid.UUID,
) (*models.Image, error) {
	row := r.db.QueryRow(ctx, findByID, id, userID)

	imgInfo, err := scanImage(row)
	if err != nil {
		return nil, fmt.Errorf("failed to find image: %w", err)
	}

	return imgInfo, nil
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

	imgInfo, err := scanImage(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return imgInfo, nil
}

const findByFilename = "SELECT * FROM images WHERE filename = $1 AND user_id = $2"

func (r *repo) FindByFilename(
	ctx context.Context,
	filename string,
	userID uuid.UUID,
) (*models.Image, error) {
	row := r.db.QueryRow(ctx, findByFilename, filename, userID)

	imgInfo, err := scanImage(row)
	if err != nil {
		return nil, fmt.Errorf("failed to find image by filename: %w", err)
	}

	return imgInfo, nil
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

	img, err := scanImage(row)
	if err != nil {
		return nil, fmt.Errorf("failed to update image: %w", err)
	}

	return img, nil
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

type MemoryRepo struct {
	Images []models.Image
}

var _ Repository = (*MemoryRepo)(nil)

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{}
}

func (r *MemoryRepo) FindByID(
	_ context.Context,
	id int,
	userID uuid.UUID,
) (*models.Image, error) {
	for _, img := range r.Images {
		if img.ID == id && img.UserID == userID {
			return &img, nil
		}
	}

	return nil, fmt.Errorf("failed to find image")
}

func (r *MemoryRepo) Create(
	_ context.Context,
	img models.Image,
) (*models.Image, error) {
	img.ID = len(r.Images) + 1
	img.CreatedAt = time.Now()
	img.UpdatedAt = time.Now()

	r.Images = append(r.Images, img)
	return &img, nil
}

func (r *MemoryRepo) FindByFilename(
	_ context.Context,
	filename string,
	userID uuid.UUID,
) (*models.Image, error) {
	for _, img := range r.Images {
		if img.Filename == filename && img.UserID == userID {
			return &img, nil
		}
	}

	return nil, fmt.Errorf("failed to find image by filename")
}

func (r *MemoryRepo) Update(
	_ context.Context,
	id int,
	userID uuid.UUID,
	imgInfo models.Image,
) (*models.Image, error) {
	for i, img := range r.Images {
		if img.ID == id && img.UserID == userID {
			img.ImageURL = imgInfo.ImageURL
			img.Filename = imgInfo.Filename
			img.Format = imgInfo.Format
			img.Alt = imgInfo.Alt
			img.UpdatedAt = time.Now()

			r.Images[i] = img
			return &img, nil
		}
	}

	return nil, fmt.Errorf("failed to update image")
}

func (r *MemoryRepo) FindManyByUserID(
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

	sort.Slice(images, func(i, j int) bool {
		return images[i].CreatedAt.After(images[j].CreatedAt)
	})

	start := (page - 1) * limit
	if start >= len(images) {
		return []models.Image{}, nil
	}

	end := start + limit
	if end > len(images) {
		end = len(images)
	}

	return images[start:end], nil
}
