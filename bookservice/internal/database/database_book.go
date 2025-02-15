package database

import (
	"context"
	"errors"
	"time"

	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (c Client) GetBookById(ctx context.Context, ID string) (*models.Book, error) {
	book := &models.Book{}
	result := c.DB.WithContext(ctx).Where(models.Book{BookID: ID}).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return book, nil
}

func (c Client) UpdateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	var books []models.Book
	result := c.DB.WithContext(ctx).
		Model(&books).
		Clauses(clause.Returning{}).
		Where(&models.Book{BookID: book.BookID}).
		Updates(models.Book{
			Title:     book.Title,
			Summary:   book.Summary,
			UpdatedAt: time.Now(),
		})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, &dberrors.NotFoundError{
			Entity: "book",
			ID:     book.BookID,
		}
	}
	return &books[0], nil
}

func (c Client) DeleteBook(ctx context.Context, ID string) error {
	result := c.DB.WithContext(ctx).Delete(&models.Book{BookID: ID})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &dberrors.NotFoundError{}
		}
		return result.Error
	}
	return nil
}
