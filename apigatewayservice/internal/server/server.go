package server

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	return func(c echo.Context) error {
		targetURL, err := url.Parse(target)
		if err != nil {
			return err
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(c.Response().Writer, c.Request())
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
	s.echo.GET("/liveness", s.Liveness)

	authGroup := s.echo.Group("/auth/")
	authGroup.Any("*", reverseProxy("http://userservice-app:8080"))

	s.echo.GET("/user/is-admin", reverseProxy("http://userservice-app:8080"))
	s.echo.GET("/user/validate-token", reverseProxy("http://userservice-app:8080"))

	s.echo.Use(validateTokenMiddleware)
	s.echo.Any("/loan/borrow", reverseProxy("http://loanservice-app:8082"))

	s.echo.Use(isAdminMiddleware)
	s.echo.Any("/admin/*", reverseProxy("http://bookservice-app:8081"))
}
