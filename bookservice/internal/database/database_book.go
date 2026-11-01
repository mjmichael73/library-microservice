package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"gorm.io/gorm"
)

func (c Client) GetAllBooks(ctx context.Context, title string) ([]models.Book, error) {
	var books []models.Book
	result := c.DB.WithContext(ctx).Where(models.Book{Title: title}).Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	for i := range books {
		authors, err := c.GetBookAuthors(ctx, books[i].BookID)
		if err != nil {
			return nil, err
		}
		genres, err := c.GetBookGenres(ctx, books[i].BookID)
		if err != nil {
			return nil, err
		}
		books[i].Authors = authors
		books[i].Genres = genres
	}
	return books, nil
}

func (c Client) GetBookByTitle(ctx context.Context, title string) (*models.Book, error) {
	book := &models.Book{}
	result := c.DB.WithContext(ctx).Where(&models.Book{Title: title}).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{
				Entity: "book",
				ID:     title,
			}
		}
		return nil, result.Error
	}
	return book, nil
}

func (c Client) CreateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	result := c.DB.WithContext(ctx).Create(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
	}
	return book, nil
}

func (c Client) GetBookById(ctx context.Context, ID string) (*models.Book, error) {
	book := &models.Book{}
	result := c.DB.WithContext(ctx).Where(models.Book{BookID: ID}).First(&book)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &dberrors.NotFoundError{
				Entity: "book",
				ID:     ID,
			}
		}
		return nil, result.Error
	}
	authors, err := c.GetBookAuthors(ctx, book.BookID)
	if err != nil {
		return nil, err
	}
	genres, err := c.GetBookGenres(ctx, book.BookID)
	if err != nil {
		return nil, err
	}
	book.Authors = authors
	book.Genres = genres
	return book, nil
}

func (c Client) UpdateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	result := c.DB.WithContext(ctx).
		Model(&models.Book{}).
		Where(&models.Book{BookID: book.BookID}).
		Updates(models.Book{
			Title:     book.Title,
			Summary:   book.Summary,
			UpdatedAt: time.Now(),
		})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, &dberrors.ConflictError{}
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, &dberrors.NotFoundError{
			Entity: "book",
			ID:     book.BookID,
		}
	}
	// Re-fetch the record rather than relying on a returned slice, which can
	// be empty (or shorter than expected) and panic when indexed.
	return c.GetBookById(ctx, book.BookID)
}

func (c Client) DeleteBook(ctx context.Context, ID string) error {
	result := c.DB.WithContext(ctx).Delete(&models.Book{BookID: ID})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &dberrors.NotFoundError{
				Entity: "book",
				ID:     ID,
			}
		}
		return result.Error
	}
	// GORM does not return an error when Delete matches no rows, so guard
	// against reporting success for a record that does not exist.
	if result.RowsAffected == 0 {
		return &dberrors.NotFoundError{
			Entity: "book",
			ID:     ID,
		}
	}
	return nil
}

func (c Client) GetBookAuthors(ctx context.Context, bookID string) ([]models.Author, error) {
	var links []models.BookAuthor
	if err := c.DB.WithContext(ctx).Where(models.BookAuthor{BookID: bookID}).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []models.Author{}, nil
	}
	ids := make([]string, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.AuthorID)
	}
	var authors []models.Author
	if err := c.DB.WithContext(ctx).Where("author_id IN ?", ids).Find(&authors).Error; err != nil {
		return nil, err
	}
	return authors, nil
}

func (c Client) GetBookGenres(ctx context.Context, bookID string) ([]models.Genre, error) {
	var links []models.BookGenre
	if err := c.DB.WithContext(ctx).Where(models.BookGenre{BookID: bookID}).Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []models.Genre{}, nil
	}
	ids := make([]string, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.GenreID)
	}
	var genres []models.Genre
	if err := c.DB.WithContext(ctx).Where("genre_id IN ?", ids).Find(&genres).Error; err != nil {
		return nil, err
	}
	return genres, nil
}

func (c Client) AttachAuthor(ctx context.Context, bookID, authorID string) error {
	return c.DB.WithContext(ctx).Create(&models.BookAuthor{
		BookAuthorID: uuid.NewString(),
		BookID:       bookID,
		AuthorID:     authorID,
	}).Error
}

func (c Client) DetachAuthor(ctx context.Context, bookID, authorID string) error {
	result := c.DB.WithContext(ctx).Where(models.BookAuthor{BookID: bookID, AuthorID: authorID}).Delete(&models.BookAuthor{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &dberrors.NotFoundError{Entity: "book_author", ID: bookID + ":" + authorID}
	}
	return nil
}

func (c Client) AttachGenre(ctx context.Context, bookID, genreID string) error {
	return c.DB.WithContext(ctx).Create(&models.BookGenre{
		BookGenreID: uuid.NewString(),
		BookID:      bookID,
		GenreID:     genreID,
	}).Error
}

func (c Client) DetachGenre(ctx context.Context, bookID, genreID string) error {
	result := c.DB.WithContext(ctx).Where(models.BookGenre{BookID: bookID, GenreID: genreID}).Delete(&models.BookGenre{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &dberrors.NotFoundError{Entity: "book_genre", ID: bookID + ":" + genreID}
	}
	return nil
}
