package database

import (
	"context"

	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)


func (c Client) GetAllBooks(ctx context.Context, title string) ([]models.Book, error) {
	var books []models.Book
	result := c.DB.WithContext(ctx).Where(models.Book{Title: title}).Find(&books)
	return books, result.Error
}