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
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mjmichael73/library-microservice/userservice/internal/database"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
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

	// User routes
	RegisterUser(ctx echo.Context) error
	LoginUser(ctx echo.Context) error

	// User Authenticated Routes
	ValidateToken(ctx echo.Context) error
	IsAdmin(ctx echo.Context) error

	// User Profile
	GetProfile(ctx echo.Context) error
	UpdateProfile(ctx echo.Context) error
}

type EchoServer struct {
	echo          *echo.Echo
	DB            database.DatabaseClient
	closer        io.Closer
	KafkaProducer *KafkaProducer
	JWTSecret     []byte
}

func NewEchoServer(db database.DatabaseClient) Server {
	server := &EchoServer{
		echo: echo.New(),
		DB:   db,
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	server.JWTSecret = []byte(jwtSecret)
	server.echo.Validator = &CustomValidator{validator: validator.New()}
	closer, err := server.initJaeger()
	if err != nil {
		log.Fatalf("Could not initialize tracer : %v", err)
	}
	server.closer = closer
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "broker:29092"
	}
	kafkaProducer := NewKafkaProducer(kafkaBroker)
	server.KafkaProducer = kafkaProducer
	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(JaegerTracingMiddleware())
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	appPort := os.Getenv("USERSERVICE_APP_PORT")
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

	authGroup := s.echo.Group("/auth")
	authGroup.POST("/register", s.RegisterUser)
	authGroup.POST("/login", s.LoginUser)

	userGroup := s.echo.Group("/user")
	userGroup.Use(echojwt.JWT(s.JWTSecret))
	userGroup.GET("/validate-token", s.ValidateToken)
	userGroup.GET("/is-admin", s.IsAdmin)
	userGroup.GET("/profile", s.GetProfile)
	userGroup.PUT("/profile", s.UpdateProfile)
}
