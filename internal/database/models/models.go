package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Image struct {
	ID        int       `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	ImageURL  string    `json:"imageUrl"`
	Filename  string    `json:"filename"`
	Format    string    `json:"format"`
	Alt       string    `json:"alt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
