package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *EchoServer) GetAllAuthors(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	authors, err := s.DB.GetAllAuthors(ctx.Request().Context(), title)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusOK, authors)
}
