package database

import (
	"context"
	"fmt"
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

	// Authors
	GetAllAuthors(ctx context.Context, name string) ([]models.Author, error)
	GetAuthorByName(ctx context.Context, name string) (*models.Author, error)
	CreateAuthor(ctx context.Context, author *models.Author) (*models.Author, error)

	// Books
	GetAllBooks(ctx context.Context, title string) ([]models.Book, error)
	GetBookByTitle(ctx context.Context, title string) (*models.Book, error)
	CreateBook(ctx context.Context, book *models.Book) (*models.Book, error)
}

type Client struct {
	DB *gorm.DB
}

func NewDatabaseClient() (DatabaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", "localhost", 54323, "bookservice_db_user", "bookservice_db_password", "bookservice_db", "disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "bookservice.",
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
