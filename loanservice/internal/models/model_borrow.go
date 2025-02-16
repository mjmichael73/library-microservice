package models

import "time"

type Borrow struct {
	BorrowID     string    `gorm:"primaryKey" json:"borrow_id"`
	UserID       string    `json:"user_id"`
	BookID       string    `json:"book_id"`
	FromDate     time.Time `json:"from_date"`
	ToDate       time.Time `json:"to_date"`
	ReturnedDate time.Time `json:"returned_date"`
	Status       string    `json:"status"`
	Remarks      string    `json:"remarks"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateBorrowRequest struct {
	UserID string `json:"user_id" validate:"required"`
	BookID string `json:"book_id" validate:"required"`
	// FromDate time.Time `json:"from_date" validate:"required"`
	// ToDate   time.Time `json:"to_date" validate:"required"`
	// Remarks  string    `json:"remarks" validate:"required"`
}

type CreateBorrowResponse struct {
	BorrowID     string    `json:"borrow_id"`
	UserID       string    `json:"user_id"`
	BookID       string    `json:"book_id"`
	FromDate     time.Time `json:"from_date"`
	ToDate       time.Time `json:"to_date"`
	ReturnedDate time.Time `json:"returned_date"`
	Status       string    `json:"status"`
	Remarks      string    `json:"remarks"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
