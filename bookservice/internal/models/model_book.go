package models

import "time"

type Book struct {
	BookID    string    `gorm:"primaryKey" json:"book_id"`
	Title     string    `json:"title"`
	Summary   string    `json:"description"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
