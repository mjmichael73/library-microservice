package models

import "time"

type Book struct {
	BookID    string    `gorm:"primaryKey" json:"book_id"`
	Title     string    `json:"title"`
	Summary   string    `json:"description"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type CreateBookRequest struct {
	Title   string `json:"title" validate:"required"`
	Summary string `json:"summary" validate:"required"`
}

type CreateBookResponse struct {
	BookID    string    `json:"book_id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
