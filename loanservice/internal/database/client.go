package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseClient interface {
	Ready() bool

	GetActiveBorrowByUserIdAndBookId(ctx context.Context, bookId string, userId string) (*models.Borrow, error)
	CreateBorrow(ctx context.Context, borrow *models.Borrow) (*models.Borrow, error)
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient() (DatabaseClient, error) {
	dbHost := os.Getenv("LOANSERVICE_DB_HOST")
	dbPort := os.Getenv("LOANSERVICE_DB_INTERNAL_PORT")
	dbName := os.Getenv("LOANSERVICE_DB_NAME")
	dbUser := os.Getenv("LOANSERVICE_DB_USER")
	dbPass := os.Getenv("LOANSERVICE_DB_PASS")
	dbTablePrefix := os.Getenv("LOANSERVICE_DB_TABLE_PREFIX")

	fmt.Println(dbHost, dbPort, dbName, dbUser, dbPass, dbTablePrefix)

	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPass == "" || dbTablePrefix == "" {
		return nil, errors.New("one or more required environment variables are not set")
	}

	dbPortInt, err := strconv.Atoi(dbPort)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT value : %s", dbPort)
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPortInt, dbUser, dbPass, dbName, "disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: dbTablePrefix,
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
