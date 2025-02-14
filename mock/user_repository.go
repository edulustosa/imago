package mock

import (
	"context"
	"errors"
	"time"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/google/uuid"
)

type userMemoryRepo struct {
	Users []models.User
}

func NewUserRepo() *userMemoryRepo {
	return &userMemoryRepo{}
}

func (r *userMemoryRepo) FindByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	for _, u := range r.Users {
		if u.ID == id {
			return &u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (r *userMemoryRepo) FindByUsername(_ context.Context, username string) (*models.User, error) {
	for _, u := range r.Users {
		if u.Username == username {
			return &u, nil
		}
	}

	return nil, errors.New("user not found")
}

func (r *userMemoryRepo) Create(_ context.Context, user models.User) (*models.User, error) {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	r.Users = append(r.Users, user)
	return &user, nil
}
