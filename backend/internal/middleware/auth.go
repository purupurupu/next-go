package middleware

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"todo-api/internal/config"
	"todo-api/internal/errors"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/pkg/util"
)

// Context keys for storing user and claims
const (
	CurrentUserKey = "current_user"
	JWTClaimsKey   = "jwt_claims"
)

// CurrentUser represents the authenticated user in the request context
type CurrentUser struct {
	ID    int64
	Email string
	Name  string
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(cfg *config.Config, userRepo *repository.UserRepository, denylistRepo *repository.JwtDenylistRepository) echo.MiddlewareFunc {
	authService := service.NewAuthService(userRepo, denylistRepo, cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errors.AuthenticationFailed("Missing authorization header")
			}

			// Extract Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return errors.AuthenticationFailed("Invalid authorization format")
			}
			tokenString := parts[1]

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "expired") {
					return errors.TokenExpired()
				}
				if strings.Contains(errMsg, "revoked") {
					return errors.TokenRevoked()
				}
				return errors.AuthenticationFailed("Invalid token")
			}

			// Parse user ID from claims
			userID, err := strconv.ParseInt(claims.Sub, 10, 64)
			if err != nil {
				return errors.AuthenticationFailed("Invalid user id in token")
			}

			// Get user from database
			user, err := userRepo.FindByID(userID)
			if err != nil {
				return errors.AuthenticationFailed("User not found")
			}

			// Set current user in context
			c.Set(CurrentUserKey, &CurrentUser{
				ID:    user.ID,
				Email: user.Email,
				Name:  util.DerefString(user.Name, ""),
			})

			// Store claims for later use (e.g., sign out)
			c.Set(JWTClaimsKey, claims)

			return next(c)
		}
	}
}

// GetCurrentUser retrieves the current user from the request context
func GetCurrentUser(c echo.Context) *CurrentUser {
	user, ok := c.Get(CurrentUserKey).(*CurrentUser)
	if !ok {
		return nil
	}
	return user
}

// GetJWTClaims retrieves the JWT claims from the request context
func GetJWTClaims(c echo.Context) *service.JWTClaims {
	claims, ok := c.Get(JWTClaimsKey).(*service.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
