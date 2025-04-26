package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mjmichael73/library-microservice/apigatewayservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server interface {
	Start() error

	Liveness(ctx echo.Context) error
}

type EchoServer struct {
	echo   *echo.Echo
	closer io.Closer
}

func NewEchoServer() Server {
	server := &EchoServer{
		echo: echo.New(),
	}

	// Initialize Jaeger Tracer
	closer, err := server.initJaeger()
	if err != nil {
		log.Fatalf("Could not initialize tracer: %v", err)
	}
	server.closer = closer

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://5ae33dfde257f10751bb7f085b115a40@o4504689622384640.ingest.us.sentry.io/4509215776047104",
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	server.echo.Use(middleware.Logger())
	server.echo.Use(middleware.Recover())
	server.echo.Use(sentryecho.New(sentryecho.Options{}))
	server.echo.Use(JaegerTracingMiddleware())
	server.registerRoutes()
	return server
}

func (s *EchoServer) Start() error {
	appPort := os.Getenv("APIGATEWAYSERVICE_APP_PORT")
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

func reverseProxy(target string) echo.HandlerFunc {
	sentry.CaptureMessage("GO HERE " + target)
	return func(c echo.Context) error {
		targetURL, err := url.Parse(target)
		if err != nil {
			return err
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		req := c.Request()
		span := opentracing.SpanFromContext(req.Context())
		if span != nil {
			opentracing.GlobalTracer().Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header),
			)
		}
		proxy.ServeHTTP(c.Response().Writer, req)
		return nil
	}
}

func validateTokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"status":  "Failed",
				"message": "Unauthorized access",
			})
		}
		userServiceURL := "http://userservice-app:8080/user/validate-token"
		valid := validateToken(userServiceURL, authHeader)
		if !valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"status":  "Failed",
				"message": "Unauthorized access",
			})
		}
		return next(c)
	}
}
func validateToken(serviceURL, authHeader string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", serviceURL, nil)
	if err != nil {
		log.Println("Error creating token validation request:", err)
		return false
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error calling user service:", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func isAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"status":  "Failed",
				"message": "Unauthorized access",
			})
		}
		userServiceURL := "http://userservice-app:8080/user/is-admin"
		valid := validateToken(userServiceURL, authHeader)
		if !valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"status":  "Failed",
				"message": "Unauthorized access",
			})
		}
		return next(c)
	}
}
func (s *EchoServer) registerRoutes() {
	s.echo.Use(MetricsMiddleware)
	s.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	s.echo.GET("/liveness", s.Liveness)

	authGroup := s.echo.Group("/auth/")
	authGroup.Any("*", reverseProxy("http://userservice-app:8080"))

	s.echo.GET("/user/is-admin", reverseProxy("http://userservice-app:8080"))
	s.echo.GET("/user/validate-token", reverseProxy("http://userservice-app:8080"))

	protectedRoutes := s.echo.Group("/loan")
	protectedRoutes.Use(validateTokenMiddleware)
	protectedRoutes.Any("/borrow", reverseProxy("http://loanservice-app:8082"))

	adminRoutes := s.echo.Group("/admin")
	adminRoutes.Use(isAdminMiddleware)
	adminRoutes.Any("/*", reverseProxy("http://bookservice-app:8081"))
}
