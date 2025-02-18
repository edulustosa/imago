package auth_test

import (
	"context"
	"testing"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/edulustosa/imago/internal/services/auth"
	"github.com/edulustosa/imago/mock"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()

	repo := user.NewMemoryRepo()
	authService := auth.New(repo)

	t.Run("register", func(t *testing.T) {
		user, err := authService.Register(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("failed to register user: %v", err)
		}

		if user.PasswordHash == "12345678" {
			t.Error("user password is not hashed")
		}
	})

	repo.Users = []models.User{}

	t.Run("user already exists", func(t *testing.T) {
		want := auth.ErrUserAlreadyExists

		_, err := authService.Register(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("failed to register user: %v", err)
		}

		_, got := authService.Register(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	repo := mock.NewUserRepo()
	authService := auth.New(repo)

	t.Run("login", func(t *testing.T) {
		createdUser, err := authService.Register(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("failed to register user: %v", err)
		}

		user, err := authService.Login(ctx, &auth.Request{
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

	t.Run("user not found", func(t *testing.T) {
		want := auth.ErrInvalidCredentials

		_, got := authService.Login(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	repo.Users = []models.User{}

	t.Run("login invalid password", func(t *testing.T) {
		want := auth.ErrInvalidCredentials

		_, err := authService.Register(ctx, &auth.Request{
			Username: "john doe",
			Password: "12345678",
		})
		if err != nil {
			t.Errorf("failed to register user: %v", err)
		}

		_, got := authService.Login(ctx, &auth.Request{
			Username: "john doe",
			Password: "123456",
		})
		if got != want {
			t.Errorf("expected %v, got %v", want, got)
		}
	})
}
