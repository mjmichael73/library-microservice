package models

import "time"

type Author struct {
	AuthorID  string    `json:"author_id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type CreateAuthorRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreateAuthorResponse struct {
	AuthorID  string    `json:"author_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
