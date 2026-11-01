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

func (c Client) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	result := c.DB.WithContext(ctx).Where(&models.User{Email: email}).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{}
		}
		return nil, result.Error
	}
	return user, nil
}

func (c Client) GetUserById(ctx context.Context, ID string) (*models.User, error) {
	user := &models.User{}
	result := c.DB.WithContext(ctx).Where(&models.User{UserID: ID}).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{}
		}
		return nil, result.Error
	}
	return user, nil
}

func (c Client) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	user.UserID = uuid.NewString()
	result := c.DB.WithContext(ctx).Create(&user)
	if result.Error != nil {
		if isDuplicateKeyError(result.Error) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, &dberrors.ConflictError{}
	}
	return user, nil
}

// isDuplicateKeyError reports whether err represents a unique-constraint
// violation. It checks gorm.ErrDuplicatedKey first (which the Postgres driver
// translates from SQLSTATE 23505) and then falls back to a safe type assertion
// on *pgconn.PgError so that a non-PgError error never causes a panic.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr != nil && pgErr.Code == "23505" {
		return true
	}
	return false
}
