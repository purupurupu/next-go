package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-api/internal/config"
	"todo-api/internal/errors"
	"todo-api/internal/middleware"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/internal/validator"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	userRepo *repository.UserRepository,
	denylistRepo *repository.JwtDenylistRepository,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(userRepo, denylistRepo, cfg),
	}
}

// SignUpRequest represents the request body for user registration
type SignUpRequest struct {
	User struct {
		Email                string `json:"email" validate:"required,email"`
		Password             string `json:"password" validate:"required,min=6"`
		PasswordConfirmation string `json:"password_confirmation" validate:"required"`
		Name                 string `json:"name" validate:"required,min=2,max=50"`
	} `json:"user" validate:"required"`
}

// SignInRequest represents the request body for user login
type SignInRequest struct {
	User struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	} `json:"user" validate:"required"`
}

// AuthResponseData represents the user data in auth responses
type AuthResponseData struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// AuthResponse represents the response for auth endpoints
type AuthResponse struct {
	Status StatusResponse   `json:"status"`
	Data   AuthResponseData `json:"data"`
}

// StatusResponse represents the status part of auth responses
type StatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SignUp handles user registration
// POST /auth/sign_up
func (h *AuthHandler) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}

	// Register user
	user, token, err := h.authService.SignUp(
		req.User.Email,
		req.User.Password,
		req.User.PasswordConfirmation,
		req.User.Name,
	)
	if err != nil {
		return err
	}

	// Set Authorization header
	c.Response().Header().Set("Authorization", "Bearer "+token)

	// Build response
	name := ""
	if user.Name != nil {
		name = *user.Name
	}

	response := AuthResponse{
		Status: StatusResponse{
			Code:    http.StatusCreated,
			Message: "Signed up successfully.",
		},
		Data: AuthResponseData{
			ID:        user.ID,
			Email:     user.Email,
			Name:      name,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	return c.JSON(http.StatusCreated, response)
}

// SignIn handles user login
// POST /auth/sign_in
func (h *AuthHandler) SignIn(c echo.Context) error {
	var req SignInRequest
	if err := c.Bind(&req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}

	// Authenticate user
	user, token, err := h.authService.SignIn(req.User.Email, req.User.Password)
	if err != nil {
		return err
	}

	// Set Authorization header
	c.Response().Header().Set("Authorization", "Bearer "+token)

	// Build response
	name := ""
	if user.Name != nil {
		name = *user.Name
	}

	response := AuthResponse{
		Status: StatusResponse{
			Code:    http.StatusOK,
			Message: "Logged in successfully.",
		},
		Data: AuthResponseData{
			ID:        user.ID,
			Email:     user.Email,
			Name:      name,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	return c.JSON(http.StatusOK, response)
}

// SignOut handles user logout
// DELETE /auth/sign_out
func (h *AuthHandler) SignOut(c echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return errors.AuthenticationFailed("Invalid token")
	}

	// Add token to denylist
	if err := h.authService.SignOut(claims.Jti, claims.ExpiresAt.Time); err != nil {
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": StatusResponse{
			Code:    http.StatusOK,
			Message: "Logged out successfully.",
		},
	})
}
