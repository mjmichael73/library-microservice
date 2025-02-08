package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.UserID = uuid.NewString()
	result := c.DB.WithContext(ctx).Create(&user)
	if result.Error != nil {
		// TODO: Fix here later on
		if result.Error.(*pgconn.PgError).Code == "23505" {
			return nil, &dberrors.ConflictError{}
		}
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, &dberrors.ConflictError{}
	}
	return user, nil
}
