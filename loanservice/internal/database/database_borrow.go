package database

import (
	"context"
	"errors"

	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) GetActiveBorrowByUserIdAndBookId(ctx context.Context, bookId string, userId string) (*models.Borrow, error) {
	borrow := &models.Borrow{}
	result := c.DB.WithContext(ctx).Where(models.Borrow{
		UserID: userId,
		BookID: bookId,
		Status: "active",
	}).First(&borrow)
	return borrow, result.Error
}

// GetActiveBorrowsByBookId returns every active borrow for a book. An empty
// (non-nil) slice means the book has no active borrow, i.e. it is available.
func (c Client) GetActiveBorrowsByBookId(ctx context.Context, bookId string) ([]models.Borrow, error) {
	var borrows []models.Borrow
	result := c.DB.WithContext(ctx).Where(models.Borrow{
		BookID: bookId,
		Status: "active",
	}).Find(&borrows)
	return borrows, result.Error
}

func (c Client) CreateBorrow(ctx context.Context, borrow *models.Borrow) (*models.Borrow, error) {
	result := c.DB.WithContext(ctx).Create(&borrow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	return borrow, nil
}
