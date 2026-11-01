package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mjmichael73/library-microservice/bookservice/internal/database"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	GetGenreById(ctx echo.Context) error

	// Admin Author CRUD
	GetAllAuthors(ctx echo.Context) error
	CreateAuthor(ctx echo.Context) error
	GetAuthorById(ctx echo.Context) error
	UpdateAuthor(ctx echo.Context) error
	DeleteAuthor(ctx echo.Context) error

	// Admin Book CRUD
	GetAllBooks(ctx echo.Context) error
	CreateBook(ctx echo.Context) error
	GetBookById(ctx echo.Context) error
	UpdateBook(ctx echo.Context) error
	DeleteBook(ctx echo.Context) error

	// Book associations
	AttachAuthor(ctx echo.Context) error
	DetachAuthor(ctx echo.Context) error
	AttachGenre(ctx echo.Context) error
	DetachGenre(ctx echo.Context) error

	// Inventory
	DecrementAvailableCopies(ctx echo.Context) error
	IncrementAvailableCopies(ctx echo.Context) error

	// Reviews
	CreateReview(ctx echo.Context) error
	GetReviews(ctx echo.Context) error

	IsBookAvailableToBorrow(ctx echo.Context) error
}

type EchoServer struct {
	echo   *echo.Echo
	DB     database.DatabaseClient
	closer io.Closer
}

func NewEchoServer(db database.DatabaseClient) Server {
	server := &EchoServer{
		echo: echo.New(),
		DB:   db,
	}

	// Initialize Jaeger Tracer
	closer, err := server.initJaeger()
	if err != nil {
		log.Fatalf("Could not initialize tracer: %v", err)
	}
	server.closer = closer
	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(JaegerTracingMiddleware())
	server.echo.Validator = &CustomValidator{validator: validator.New()}
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	appPort := os.Getenv("BOOKSERVICE_APP_PORT")
	if appPort == "" {
		return errors.New("APP_PORT is not set")
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.echo.Start(":" + appPort); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-quit:
		log.Println("Shutting down server gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.echo.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}

	if s.closer != nil {
		if err := s.closer.Close(); err != nil {
			log.Printf("Error closing tracer: %v", err)
		}
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
	s.echo.Use(MetricsMiddleware)
	s.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	s.echo.GET("/readiness", s.Readiness)
	s.echo.GET("/liveness", s.Liveness)

	adminGroup := s.echo.Group("/admin")

	// Admin Genres
	adminGenreGroup := adminGroup.Group("/genres")
	adminGenreGroup.GET("", s.GetAllGenres)
	adminGenreGroup.POST("", s.CreateGenre)
	adminGenreGroup.GET("/:id", s.GetGenreById)

	// Admin Authors
	adminAuthorGroup := adminGroup.Group("/authors")
	adminAuthorGroup.GET("", s.GetAllAuthors)
	adminAuthorGroup.POST("", s.CreateAuthor)
	adminAuthorGroup.GET("/:id", s.GetAuthorById)
	adminAuthorGroup.PUT("/:id", s.UpdateAuthor)
	adminAuthorGroup.DELETE("/:id", s.DeleteAuthor)

	// Admin Books
	adminBookGroup := adminGroup.Group("/books")
	adminBookGroup.GET("", s.GetAllBooks)
	adminBookGroup.POST("", s.CreateBook)
	adminBookGroup.GET("/:id", s.GetBookById)
	adminBookGroup.PUT("/:id", s.UpdateBook)
	adminBookGroup.DELETE("/:id", s.DeleteBook)
	adminBookGroup.POST("/:id/authors/:authorId", s.AttachAuthor)
	adminBookGroup.DELETE("/:id/authors/:authorId", s.DetachAuthor)
	adminBookGroup.POST("/:id/genres/:genreId", s.AttachGenre)
	adminBookGroup.DELETE("/:id/genres/:genreId", s.DetachGenre)
	adminBookGroup.POST("/:id/decrement", s.DecrementAvailableCopies)
	adminBookGroup.POST("/:id/increment", s.IncrementAvailableCopies)

	// Book reviews (public read, any caller may post a review for the core
	// feature). Mirrored under /admin/books so the gateway's admin proxy, which
	// forwards everything under /admin to bookservice, can reach them.
	adminBookGroup.POST("/:id/reviews", s.CreateReview)
	adminBookGroup.GET("/:id/reviews", s.GetReviews)

	bookGroup := s.echo.Group("/books")
	bookGroup.POST("/:id/reviews", s.CreateReview)
	bookGroup.GET("/:id/reviews", s.GetReviews)

	s.echo.GET("/isbookavailabletoboroow/:id", s.IsBookAvailableToBorrow)
}
