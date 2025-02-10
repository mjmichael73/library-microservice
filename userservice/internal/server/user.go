package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key")

func (s *EchoServer) RegisterUser(ctx echo.Context) error {
	registerRequest := new(models.RegisterRequest)
	if err := ctx.Bind(registerRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{"status": "Failed", "message": "Invalid request"})
	}
	if err := ctx.Validate(registerRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}
	// TODO: Create User Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error",
		})
	}

	user := models.User{
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
		Email:     registerRequest.Email,
		Password:  string(hashedPassword),
	}

	resultUser, resultErr := s.DB.RegisterUser(ctx.Request().Context(), &user)
	if resultErr != nil {
		switch resultErr.(type) {
		case *dberrors.ConflictError:
			return ctx.JSON(http.StatusConflict, echo.Map{
				"status":  "Failed",
				"message": "Registeration was not successful",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error",
			})
		}
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	registerResponse := models.RegisterResponse{
		FirstName: resultUser.FirstName,
		LastName:  resultUser.LastName,
		Email:     resultUser.Email,
		Token:     token.Raw,
	}
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Registration was successful, please try to login later on.",
			"data":    registerResponse,
		})
	}
	registerResponse.Token = tokenString
	return ctx.JSON(http.StatusCreated, echo.Map{
		"status":  "Success",
		"message": "Registration was successfull",
		"data":    registerResponse,
	})
}

func (s *EchoServer) LoginUser(ctx echo.Context) error {
	loginRequest := new(models.LoginRequest)
	if err := ctx.Bind(loginRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Invalid request",
		})
	}
	if err := ctx.Validate(loginRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}

	// Check if user exist
	user, err := s.DB.GetUserByEmail(ctx.Request().Context(), loginRequest.Email)
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusUnauthorized, echo.Map{
				"status":  "Failed",
				"message": "Email or password is wrong",
			})
		default:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{
			"status":  "Failed",
			"message": "Email or password is wrong",
		})
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	loginResponse := models.LoginResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Token:     tokenString,
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Login was successfull",
		"data":    loginResponse,
	})

}

func (s *EchoServer) ValidateToken(ctx echo.Context) error {
	fmt.Println("hi")
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"]
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "User is valid",
		"data":    userID,
	})
}
