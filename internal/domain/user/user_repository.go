package user

import (
	"context"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user models.User) (*models.User, error)
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) Repository {
	return &repo{
		db,
	}
}

func scanUser(row pgx.Row) (*models.User, error) {
	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return &user, err
}

const findByUsername = "SELECT * FROM users WHERE username = $1"

func (r *repo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	row := r.db.QueryRow(ctx, findByUsername, username)
	return scanUser(row)
}

const createUser = "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING *"

func (r *repo) Create(ctx context.Context, user models.User) (*models.User, error) {
	row := r.db.QueryRow(ctx, createUser, user.Username, user.PasswordHash)
	return scanUser(row)
}
