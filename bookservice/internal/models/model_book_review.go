package models

import "time"

// BookReview is a rating (1–5) and optional comment left by a user on a book.
// Both BookID and UserID are stored as plain UUID strings because users live in
// a separate database (database-per-service), so a real foreign key cannot be
// declared.
type BookReview struct {
	ReviewID  string    `gorm:"primaryKey" json:"review_id"`
	BookID    string    `json:"book_id"`
	UserID    string    `json:"user_id"`
	Rating    int       `gorm:"not null" json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateReviewRequest is the payload accepted by POST /books/:id/reviews. The
// book id comes from the path; the user id is supplied in the body.
//
// NOTE: unlike /loan (R-023), the caller's identity is not yet verified against
// userservice here, so a client could post a review as another user. The
// endpoint is public read/write for the core feature; token validation is
// tracked as follow-up hardening.
type CreateReviewRequest struct {
	UserID  string `json:"user_id" validate:"required"`
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment"`
}
