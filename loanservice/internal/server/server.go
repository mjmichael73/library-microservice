package server

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mjmichael73/library-microservice/loanservice/internal/database"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	JAEGER_SERVICE_NAME = "loan-service"
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

	CreateBorrow(ctx echo.Context) error
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
	v := validator.New()
	server.echo.Validator = &CustomValidator{validator: v}

	closer, err := server.initJaeger(JAEGER_SERVICE_NAME)
	if err != nil {
		log.Fatalf("Could not initialize tracer: %v", err)
	}
	server.closer = closer
	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			operationName := c.Request().Method + " " + c.Path()
			span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), operationName)
			span.SetTag("http.method", c.Request().Method)
			span.SetTag("http.url", c.Request().RequestURI)
			span.SetTag("component", JAEGER_SERVICE_NAME)
			c.SetRequest(c.Request().WithContext(ctx))
			err = next(c)
			status := c.Response().Status
			span.SetTag("http.status_code", status)
			if status >= 500 {
				span.SetTag("error", true)
			}
			span.Finish()
			return next(c)
		}
	})
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	appPort := os.Getenv("LOANSERVICE_APP_PORT")
	if appPort == "" {
		return errors.New("APP_PORT is not set")
	}
	if err := s.echo.Start(":" + appPort); err != nil && err != http.ErrServerClosed {
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
	s.echo.Use(MetricsMiddleware)
	s.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	s.echo.GET("/readiness", s.Readiness)
	s.echo.GET("/liveness", s.Liveness)

	s.echo.POST("/loan/borrow", s.CreateBorrow)
}
