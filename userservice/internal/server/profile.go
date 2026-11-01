package server

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
)

// authenticatedUserID extracts the user_id claim from the JWT placed in the
// echo context by the echojwt middleware. It uses safe type assertions so a
// malformed token never panics (see AGENTS.md).
func authenticatedUserID(ctx echo.Context) (string, bool) {
	token, ok := ctx.Get("user").(*jwt.Token)
	if !ok {
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

// GetProfile returns the authenticated user's profile (never the password).
func (s *EchoServer) GetProfile(ctx echo.Context) error {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{
			"status":  "Failed",
			"message": "Unauthorized access",
		})
	}

	user, err := s.DB.GetUserById(ctx.Request().Context(), userID)
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "User not found.",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Profile received successfully.",
		"data":    profileResponse(user),
	})
}

// UpdateProfile updates the authenticated user's first/last name.
func (s *EchoServer) UpdateProfile(ctx echo.Context) error {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{
			"status":  "Failed",
			"message": "Unauthorized access",
		})
	}

	updateRequest := new(models.UserProfileRequest)
	if err := ctx.Bind(updateRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}
	if err := ctx.Validate(updateRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}

	updated, err := s.DB.UpdateUser(ctx.Request().Context(), &models.User{
		UserID:    userID,
		FirstName: updateRequest.FirstName,
		LastName:  updateRequest.LastName,
	})
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "User not found.",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Profile updated successfully.",
		"data":    profileResponse(updated),
	})
}

func profileResponse(user *models.User) models.UserProfileResponse {
	return models.UserProfileResponse{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
	}
}
