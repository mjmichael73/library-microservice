package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

func (s *EchoServer) GetAllBooks(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	books, err := s.DB.GetAllBooks(ctx.Request().Context(), title)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Books received successfully.",
		"data":    books,
	})
}

func (s *EchoServer) CreateBook(ctx echo.Context) error {
	createBookRequest := new(models.CreateBookRequest)
	if err := ctx.Bind(createBookRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request.",
		})
	}
	if err := ctx.Validate(createBookRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}
	book, err := s.DB.GetBookByTitle(ctx.Request().Context(), createBookRequest.Title)
	if err == nil && book != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Book with this title already exists.",
		})
	}
	switch err.(type) {
	case *dberrors.NotFoundError:
		newBook := &models.Book{
			BookID:    uuid.NewString(),
			Title:     createBookRequest.Title,
			Summary:   createBookRequest.Summary,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		newBook, err := s.DB.CreateBook(ctx.Request().Context(), newBook)
		if err != nil {
			switch err.(type) {
			case *dberrors.ConflictError:
				return ctx.JSON(http.StatusBadRequest, echo.Map{
					"status":  "Failed",
					"message": "PBook with this title already exists",
				})
			default:
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		}
		return ctx.JSON(http.StatusCreated, echo.Map{
			"status":  "Success",
			"message": "Book created successfully.",
			"data":    newBook,
		})

	default:
		return ctx.JSON(http.StatusInternalServerError, err)
	}
}

func (s *EchoServer) GetBookById(ctx echo.Context) error {
	ID := ctx.Param("id")
	book, err := s.DB.GetBookById(ctx.Request().Context(), ID)
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Book not found",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, err)
		}
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Book received successfully.",
		"data":    book,
	})
}

func (s *EchoServer) UpdateBook(ctx echo.Context) error {
	ID := ctx.Param("id")
	book := new(models.Book)
	if err := ctx.Bind(book); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}
	if ID != book.BookID {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request, the id on path does not match the id in the body",
		})
	}
	book, err := s.DB.UpdateBook(ctx.Request().Context(), book)
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Book not found",
			})
		case *dberrors.ConflictError:
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"status": "Failed",
				"message": "The title of the book already exists.",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status": "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status": "Success",
		"message": "Book updated successfully.",
		"data": book,
	})

}

func (s *EchoServer) DeleteBook(ctx echo.Context) error {
	ID := ctx.Param("id")
	err := s.DB.DeleteBook(ctx.Request().Context(), ID)
	if err != nil {
		 switch err.(type) {
		 case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status": "Failed",
				"message": "Book not found",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, err)
		 }
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status": "Success",
		"message": "Book has been deleted successfully.",
	})
}
