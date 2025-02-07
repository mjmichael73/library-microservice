package models

type User struct {
	UserID    string `gorm:"primaryKey" json:"customer_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
}
