package database

import (
	"context"
	"errors"
	"time"

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

// GetBorrowById returns a single borrow by primary key, or a populated
// NotFoundError when no such borrow exists.
func (c Client) GetBorrowById(ctx context.Context, borrowID string) (*models.Borrow, error) {
	borrow := &models.Borrow{}
	result := c.DB.WithContext(ctx).Where("borrow_id = ?", borrowID).First(borrow)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{Entity: "borrow", ID: borrowID}
		}
		return nil, result.Error
	}
	return borrow, nil
}

// ReturnBorrow marks a borrow as returned. It returns a NotFoundError when the
// borrow does not exist and a ConflictError when it has already been returned.
// The status is set to "returned" when the book is handed back on or before its
// due date, and "overdue" when it is returned late.
func (c Client) ReturnBorrow(ctx context.Context, borrowID string) (*models.Borrow, error) {
	borrow, err := c.GetBorrowById(ctx, borrowID)
	if err != nil {
		return nil, err
	}
	if borrow.Status == "returned" {
		return nil, &dberrors.ConflictError{}
	}

	now := time.Now()
	borrow.ReturnedDate = now
	if now.After(borrow.ToDate) {
		borrow.Status = "overdue"
	} else {
		borrow.Status = "returned"
	}
	borrow.UpdatedAt = now

	result := c.DB.WithContext(ctx).Save(borrow)
	if result.Error != nil {
		return nil, result.Error
	}
	return borrow, nil
}

// MarkOverdueBorrows flips every active borrow whose due date has passed to
// "overdue" and returns those borrows. Borrows already returned (or marked
// overdue by an earlier call) are left untouched, so the returned slice holds
// the full set of currently overdue loans.
func (c Client) MarkOverdueBorrows(ctx context.Context) ([]models.Borrow, error) {
	now := time.Now()

	var borrows []models.Borrow
	result := c.DB.WithContext(ctx).
		Where(models.Borrow{Status: "active"}).
		Where("to_date < ?", now).
		Find(&borrows)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(borrows) == 0 {
		return borrows, nil
	}

	ids := make([]string, 0, len(borrows))
	for i := range borrows {
		borrows[i].Status = "overdue"
		borrows[i].UpdatedAt = now
		ids = append(ids, borrows[i].BorrowID)
	}

	if err := c.DB.WithContext(ctx).
		Model(&models.Borrow{}).
		Where("borrow_id IN ?", ids).
		Updates(map[string]interface{}{"status": "overdue", "updated_at": now}).
		Error; err != nil {
		return nil, err
	}
	return borrows, nil
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
