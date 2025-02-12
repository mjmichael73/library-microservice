package server

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/database"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func FormatValidationErrors(err error) map[string][]string {
	errors := make(map[string][]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()
		message := field + " validation failed on the '" + tag + "' tag"
		errors[field] = append(errors[field], message)
	}
	return errors
}

type Server interface {
	Start() error

	Readiness(ctx echo.Context) error
	Liveness(ctx echo.Context) error

	// Admin Genre CRUD
	GetAllGenres(ctx echo.Context) error
	CreateGenre(ctx echo.Context) error

	// Admin Author CRUD
	GetAllAuthors(ctx echo.Context) error

	// Admin Book CRUD
	GetAllBooks(ctx echo.Context) error
	CreateBook(ctx echo.Context) error
}

type EchoServer struct {
	echo *echo.Echo
	DB   database.DatabaseClient
}

func NewEchoServer(db database.DatabaseClient) Server {
	server := &EchoServer{
		echo: echo.New(),
		DB:   db,
	}
	server.echo.Validator = &CustomValidator{validator: validator.New()}
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	if err := s.echo.Start(":8081"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server shutdown occurred: %s", err)
		return err
	}
	return nil
}

func (s *EchoServer) Readiness(ctx echo.Context) error {
	ready := s.DB.Ready()
	if ready {
		return ctx.JSON(http.StatusOK, models.Health{Status: "OK", Message: "Server is ready."})
	}
	return ctx.JSON(http.StatusInternalServerError, models.Health{Status: "Failure", Message: "Server is not ready."})
}

func (s *EchoServer) Liveness(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, models.Health{Status: "OK", Message: "Server is live"})
}

func (s *EchoServer) registerRoutes() {
	s.echo.GET("/readiness", s.Readiness)
	s.echo.GET("/liveness", s.Liveness)

	adminGroup := s.echo.Group("/admin")

	// Admin Genres
	adminGenreGroup := adminGroup.Group("/genres")
	adminGenreGroup.GET("", s.GetAllGenres)
	adminGenreGroup.POST("", s.CreateGenre)

	// Admin Authors
	adminAuthorGroup := adminGroup.Group("/authors")
	adminAuthorGroup.GET("", s.GetAllAuthors)

	// Admin Books
	adminBookGroup := adminGroup.Group("/books")
	adminBookGroup.GET("", s.GetAllBooks)
	adminBookGroup.POST("", s.CreateBook)

}
