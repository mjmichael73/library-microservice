package server

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/apigatewayservice/internal/models"
)

type Server interface {
	Start() error

	Liveness(ctx echo.Context) error
}

type EchoServer struct {
	echo *echo.Echo
}

func NewEchoServer() Server {
	server := &EchoServer{
		echo: echo.New(),
	}
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	appPort := os.Getenv("USERSERVICE_APP_PORT")
	if appPort == "" {
		return errors.New("APP_PORT is not set")
	}
	if err := s.echo.Start(":" + appPort); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown occurred: %s", err)
		return err
	}
	return nil
}

func (s *EchoServer) Liveness(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, models.Health{Status: "OK", Message: "Server is live"})
}

func (s *EchoServer) registerRoutes() {
	s.echo.GET("/liveness", s.Liveness)
}
