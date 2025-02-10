package models

type User struct {
	UserID    string `gorm:"primaryKey" json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	IsAdmin   bool   `json:"is_admin"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	IsAdmin   bool   `json:"is_admin"`
}
