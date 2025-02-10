package models

type BookAuthor struct {
	BookAuthorID string `json:"book_author_id" gorm:"primaryKey"`
	BookID       string `json:"book_id"`
	AuthorID     string `json:"author_id"`
}
