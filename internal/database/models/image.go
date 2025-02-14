package models

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID        int       `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ImageURL  string    `json:"imageUrl"`
	Format    string    `json:"format"`
	Alt       string    `json:"alt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
