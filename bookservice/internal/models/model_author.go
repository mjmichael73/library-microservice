package models

import "time"

type Author struct {
	AuthorID  string    `json:"author_id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
