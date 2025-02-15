package images

import (
	"context"
	"errors"

	"github.com/edulustosa/imago/internal/database/models"
	"github.com/edulustosa/imago/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	repo           Repository
	userRepository user.Repository
}

func NewService(repo Repository, userRepository user.Repository) *Service {
	return &Service{
		repo,
		userRepository,
	}
}

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrImageNotFound = errors.New("image not found")
)

func (s *Service) GetImage(
	ctx context.Context,
	imgID int,
	userID uuid.UUID,
) (*models.Image, error) {
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	img, err := s.repo.FindByID(ctx, imgID, user.ID)
	if err != nil {
		return nil, ErrImageNotFound
	}

	return img, nil
}
