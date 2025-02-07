package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseClient interface {
	Ready() bool
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient() (DatabaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", "localhost", 54322, "userservice_db_user", "userservice_db_password", "userservice_db", "disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "wisdom.",
		},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		QueryFields: true,
	})
	if err != nil {
		return nil, err
	}
	client := Client{
		DB: db,
	}
	return client, nil
}

func (c Client) Ready() bool {
	var ready string
	tx := c.DB.Raw("SELECT 1 as ready").Scan(&ready)
	if tx.Error != nil {
		return false
	}
	if ready == "1" {
		return true
	}
	return false
}
