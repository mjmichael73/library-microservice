package database

import (
	"context"
	"errors"

	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) GetAllBooks(ctx context.Context, title string) ([]models.Book, error) {
	var books []models.Book
	result := c.DB.WithContext(ctx).Where(models.Book{Title: title}).Find(&books)
	return books, result.Error
}

func (c Client) GetBookByTitle(ctx context.Context, title string) (*models.Book, error) {
	book := &models.Book{}
	result := c.DB.WithContext(ctx).Where(&models.Book{Title: title}).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{
				Entity: "book",
				ID:     title,
			}
		}
		return nil, result.Error
	}
	return book, nil
}

func (c Client) CreateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	result := c.DB.WithContext(ctx).Create(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
	}
	return book, nil
}
