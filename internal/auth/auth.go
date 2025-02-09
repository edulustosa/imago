package auth

import (
	"context"
	"errors"

	"github.com/edulustosa/imago/internal/database/models"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user models.User) (*models.User, error)
}

type Auth struct {
	repo Repository
}

func New(repo Repository) *Auth {
	return &Auth{repo}
}

type Request struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
