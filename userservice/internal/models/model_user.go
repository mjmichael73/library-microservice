package models

type User struct {
	UserID    string `gorm:"primaryKey" json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `gorm:"uniqueIndex;not null" json:"email"`
	Password  string `gorm:"uniqueIndex;not null" json:"-"`
}
