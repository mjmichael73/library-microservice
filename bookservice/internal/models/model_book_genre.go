package models

type BookGenre struct {
	BookGenreID string `json:"book_genre_id" gorm:"primaryKey"`
	BookID      string `json:"book_id"`
	GenreID     string `json:"genre_id"`
}
