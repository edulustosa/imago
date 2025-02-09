package auth

import (
	"context"
	"errors"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	repo user.Repository
}

func New(repo user.Repository) *Auth {
	return &Auth{repo}
}

type Request struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (a *Auth) Register(ctx context.Context, req Request) (*models.User, error) {
	_, err := a.repo.FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username:     req.Username,
		PasswordHash: string(passwordHashBytes),
	}

	return a.repo.Create(ctx, user)
}

func (a *Auth) Login(ctx context.Context, req Request) (*models.User, error) {
	user, err := a.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
