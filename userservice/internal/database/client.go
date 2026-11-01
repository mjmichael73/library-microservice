package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mjmichael73/library-microservice/userservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	dbConnectAttempts = 10
	dbConnectBaseWait = time.Second
	dbConnectMaxWait  = 10 * time.Second
)

type DatabaseClient interface {
	Ready() bool

	// User
	RegisterUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, ID string) (*models.User, error)
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient() (DatabaseClient, error) {
	dbHost := os.Getenv("USERSERVICE_DB_HOST")
	dbPort := os.Getenv("USERSERVICE_DB_INTERNAL_PORT")
	dbName := os.Getenv("USERSERVICE_DB_NAME")
	dbUser := os.Getenv("USERSERVICE_DB_USER")
	dbPass := os.Getenv("USERSERVICE_DB_PASS")
	dbTablePrefix := os.Getenv("USERSERVICE_DB_TABLE_PREFIX")

	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPass == "" || dbTablePrefix == "" {
		return nil, errors.New("one or more required environment variables are not set")
	}

	dbPortInt, err := strconv.Atoi(dbPort)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT value : %s", dbPort)
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPortInt, dbUser, dbPass, dbName, "disable")
	db, err := openWithRetry(dsn, dbTablePrefix)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	client := Client{
		DB: db,
	}
	return client, nil
}

// openWithRetry opens the database connection and verifies it with a ping,
// retrying with exponential backoff (1s, 2s, 4s, ... capped at 10s) until the
// database becomes available. This guards against the app starting before the
// database container is ready to accept connections.
func openWithRetry(dsn string, tablePrefix string) (*gorm.DB, error) {
	var lastErr error
	wait := dbConnectBaseWait

	for attempt := 1; attempt <= dbConnectAttempts; attempt++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: tablePrefix,
			},
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
			QueryFields: true,
		})
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr != nil {
				err = sqlErr
			} else if pingErr := sqlDB.Ping(); pingErr != nil {
				err = pingErr
			} else {
				return db, nil
			}
		}

		lastErr = err
		if attempt == dbConnectAttempts {
			break
		}
		log.Printf("database not ready (attempt %d/%d): %v; retrying in %s", attempt, dbConnectAttempts, err, wait)
		time.Sleep(wait)
		wait *= 2
		if wait > dbConnectMaxWait {
			wait = dbConnectMaxWait
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", dbConnectAttempts, lastErr)
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
