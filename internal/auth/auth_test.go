package auth_test

import (
	"context"
	"testing"

	"github.com/edulustosa/imago/internal/auth"
	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/mock"
)

func TestRegister(t *testing.T) {
	repo := mock.NewUserRepo()
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

	repo.Users = []models.User{}

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
	repo := mock.NewUserRepo()
	authService := auth.New(repo)
	ctx := context.Background()

	t.Run("login", func(t *testing.T) {
		createdUser, err := authService.Register(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("error register user: %v", err)
		}

		user, err := authService.Login(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("error login user: %v", err)
		}

		if createdUser.Username != user.Username {
			t.Errorf("expected %s, got %s", createdUser.Username, user.Username)
		}
	})

	repo.Users = []models.User{}
	t.Run("login user not found", func(t *testing.T) {
		_, err := authService.Login(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != auth.ErrInvalidCredentials {
			t.Errorf("expected %v, got %v", auth.ErrInvalidCredentials, err)
		}
	})

	repo.Users = []models.User{}
	t.Run("login invalid password", func(t *testing.T) {
		_, err := authService.Register(ctx, auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("error register user: %v", err)
		}

		_, err = authService.Login(ctx, auth.Request{
			Username: "john doe",
			Password: "123456",
		})
		if err != auth.ErrInvalidCredentials {
			t.Errorf("expected %v, got %v", auth.ErrInvalidCredentials, err)
		}
	})
}
