package database

import (
	"context"

	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

func (c Client) GetAllAuthors(ctx context.Context, name string) ([]models.Author, error) {
	var authors []models.Author
	result := c.DB.WithContext(ctx).Where(models.Author{Name: name}).Find(&authors)
	return authors, result.Error
}
