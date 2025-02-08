package server

import (
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
	registerResponse := models.RegisterResponse{
		FirstName: resultUser.FirstName,
		LastName:  resultUser.LastName,
		Email:     resultUser.Email,
		Token:     "",
	}
	tokenString, err := token.SignedString(jwtSecret)
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
