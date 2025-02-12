package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
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
