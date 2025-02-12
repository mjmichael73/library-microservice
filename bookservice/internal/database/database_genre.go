package database

import (
	"context"
	"errors"

	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) GetAllGenres(ctx context.Context, title string) ([]models.Genre, error) {
	var genres []models.Genre
	result := c.DB.WithContext(ctx).Where(models.Genre{Title: title}).Find(&genres)
	return genres, result.Error
}

func (c Client) GetGenreByTitle(ctx context.Context, title string) (*models.Genre, error) {
	genre := &models.Genre{}
	result := c.DB.WithContext(ctx).Where(&models.Genre{Title: title}).First(&genre)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{}
		}
		return nil, result.Error
	}
	return genre, nil
}

func (c Client) CreateGenre(ctx context.Context, genre *models.Genre) (*models.Genre, error) {
	result := c.DB.WithContext(ctx).Create(&genre)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	return genre, nil
}
