package database

import (
	"context"

	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// CreateReview persists a book review. Book existence is validated by the
// caller (handler) so this stays a plain insert.
func (c Client) CreateReview(ctx context.Context, review *models.BookReview) (*models.BookReview, error) {
	result := c.DB.WithContext(ctx).Create(&review)
	if result.Error != nil {
		return nil, result.Error
	}
	return review, nil
}

// GetReviewsByBook returns every review for a book, newest first, as an empty
// (non-nil) slice when there are none.
func (c Client) GetReviewsByBook(ctx context.Context, bookID string) ([]models.BookReview, error) {
	var reviews []models.BookReview
	result := c.DB.WithContext(ctx).
		Where(models.BookReview{BookID: bookID}).
		Order("created_at DESC").
		Find(&reviews)
	if result.Error != nil {
		return nil, result.Error
	}
	if reviews == nil {
		reviews = []models.BookReview{}
	}
	return reviews, nil
}
