package database

import (
	"context"

	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

func (c Client) GetAllGenres(ctx context.Context, title string) ([]models.Genre, error) {
	var genres []models.Genre
	result := c.DB.WithContext(ctx).Where(models.Genre{Title: title}).Find(&genres)
	return genres, result.Error
}
