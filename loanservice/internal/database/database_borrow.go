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

func (c Client) CreateBorrow(ctx context.Context, borrow *models.Borrow) (*models.Borrow, error) {
	result := c.DB.WithContext(ctx).Create(&borrow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	return nil, nil
}
