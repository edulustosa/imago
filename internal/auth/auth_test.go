package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/edulustosa/imago/internal/auth"
	"github.com/edulustosa/imago/internal/database/models"
)

type memoryRepository struct {
	users []models.User
}

func (r *memoryRepository) Create(_ context.Context, user models.User) (*models.User, error) {
	r.users = append(r.users, user)
	return &user, nil
}

func (r *memoryRepository) FindByUsername(_ context.Context, username string) (*models.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

func TestRegister(t *testing.T) {
	repo := new(memoryRepository)
	authService := auth.New(repo)
	ctx := context.Background()

	t.Run("register user", func(t *testing.T) {
		user, err := authService.Register(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("error register user: %v", err)
		}

		if user.PasswordHash == "1234568" {
			t.Error("password is not hashed")
		}
	})

	repo.users = []models.User{}

	t.Run("register user already exists", func(t *testing.T) {
		_, err := authService.Register(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("error register user: %v", err)
		}

		_, err = authService.Register(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != auth.ErrUserAlreadyExists {
			t.Errorf("expected %v, got %v", auth.ErrUserAlreadyExists, err)
		}
	})
}

func TestLogin(t *testing.T) {
	repo := new(memoryRepository)
	authService := auth.New(repo)

	createdUser, err := authService.Register(context.Background(), auth.Request{
		Username: "john doe",
		Password: "12345678",
	})
	if err != nil {
		t.Errorf("error register user: %v", err)
	}

	user, err := authService.Login(context.Background(), auth.Request{
		Username: "john doe",
		Password: "12345678",
	})
	if err != nil {
		t.Errorf("error login user: %v", err)
	}

	if createdUser.Username != user.Username {
		t.Errorf("expected %s, got %s", createdUser.Username, user.Username)
	}
}
