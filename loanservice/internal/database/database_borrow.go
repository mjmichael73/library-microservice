package database

import (
	"context"

	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

func (c Client) CreateBorrow(ctx context.Context, borrow *models.Borrow) (*models.Borrow, error) {
	return nil, nil
}
