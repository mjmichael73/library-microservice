package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DatabaseClient interface {
	Ready() bool

	// Genres
	GetAllGenres(ctx context.Context, title string) ([]models.Genre, error)
	GetGenreByTitle(ctx context.Context, title string) (*models.Genre, error)
	CreateGenre(ctx context.Context, genre *models.Genre) (*models.Genre, error)
	GetGenreById(ctx context.Context, ID string) (*models.Genre, error)

	// Authors
	GetAllAuthors(ctx context.Context, name string) ([]models.Author, error)
	GetAuthorByName(ctx context.Context, name string) (*models.Author, error)
	CreateAuthor(ctx context.Context, author *models.Author) (*models.Author, error)
	GetAuthorById(ctx context.Context, ID string) (*models.Author, error)
	UpdateAuthor(ctx context.Context, author *models.Author) (*models.Author, error)
	DeleteAuthor(ctx context.Context, ID string) error

	// Books
	GetAllBooks(ctx context.Context, title string) ([]models.Book, error)
	GetBookByTitle(ctx context.Context, title string) (*models.Book, error)
	CreateBook(ctx context.Context, book *models.Book) (*models.Book, error)
	GetBookById(ctx context.Context, ID string) (*models.Book, error)
	UpdateBook(ctx context.Context, book *models.Book) (*models.Book, error)
	DeleteBook(ctx context.Context, ID string) error
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient() (DatabaseClient, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbTablePrefix := os.Getenv("DB_TABLE_PREFIX")

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
