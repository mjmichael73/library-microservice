package database

import (
	"context"
	"errors"

	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) GetAllAuthors(ctx context.Context, name string) ([]models.Author, error) {
	var authors []models.Author
	result := c.DB.WithContext(ctx).Where(models.Author{Name: name}).Find(&authors)
	return authors, result.Error
}

func (c Client) GetAuthorByName(ctx context.Context, name string) (*models.Author, error) {
	author := &models.Author{}
	result := c.DB.WithContext(ctx).Where(models.Author{Name: name}).First(&author)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{}
		}
		return nil, result.Error
	}
	return author, result.Error
}

func (c Client) CreateAuthor(ctx context.Context, author *models.Author) (*models.Author, error) {
	result := c.DB.WithContext(ctx).Create(&author)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	return author, nil
}
