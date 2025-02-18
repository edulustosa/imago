package user

import (
	"context"
	"errors"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
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

const findByID = "SELECT * FROM users WHERE id = $1"

func (r *repo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := r.db.QueryRow(ctx, findByID, id)
	return scanUser(row)
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

type MemoryRepo struct {
	Users []models.User
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		Users: []models.User{},
	}
}

func (r *MemoryRepo) FindByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	for _, u := range r.Users {
		if u.ID == id {
			return &u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (r *MemoryRepo) FindByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range r.Users {
		if u.Username == username {
			return &u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (r *MemoryRepo) Create(_ context.Context, user models.User) (*models.User, error) {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.Users = append(r.Users, user)
	return &user, nil
}
