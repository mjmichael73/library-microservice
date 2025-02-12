package database

import (
	"context"
	"errors"

	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return author, nil
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

func (c Client) GetAuthorById(ctx context.Context, ID string) (*models.Author, error) {
	author := &models.Author{}
	result := c.DB.WithContext(ctx).Where(models.Author{AuthorID: ID}).First(&author)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return author, nil
}

func (c Client) UpdateAuthor(ctx context.Context, author *models.Author) (*models.Author, error) {
	var authors []models.Author
	result := c.DB.WithContext(ctx).
		Model(&authors).
		Clauses(clause.Returning{}).
		Where(&models.Author{AuthorID: author.AuthorID}).
		Updates(models.Author{
			Name: author.Name,
		})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, &dberrors.NotFoundError{
			Entity: "author",
			ID:     author.AuthorID,
		}
	}
	return &authors[0], nil
}

func (c Client) DeleteAuthor(ctx context.Context, ID string) error {
	result := c.DB.WithContext(ctx).Delete(&models.Author{AuthorID: ID})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &dberrors.NotFoundError{}
		}
		return result.Error
	}
	return nil
}
