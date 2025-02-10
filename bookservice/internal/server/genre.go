package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *EchoServer) GetAllGenres(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	genres, err := s.DB.GetAllGenres(ctx.Request().Context(), title)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusOK, genres)
}
