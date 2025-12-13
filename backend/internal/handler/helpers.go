package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"todo-api/internal/errors"
	"todo-api/internal/middleware"
	"todo-api/internal/validator"
)

// ParseIDParam parses an int64 ID parameter from the URL path.
// Returns the ID or an error if the parameter is invalid.
func ParseIDParam(c echo.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, errors.ValidationFailed(map[string][]string{
			name: {"Invalid " + name},
		})
	}
	return id, nil
}

// BindAndValidate binds the request body to the given struct and validates it.
// Returns an error if binding or validation fails.
func BindAndValidate[T any](c echo.Context, req *T) error {
	if err := c.Bind(req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}
	return nil
}

// GetCurrentUserOrFail retrieves the current user from the context.
// Returns an error if the user is not authenticated.
func GetCurrentUserOrFail(c echo.Context) (*middleware.CurrentUser, error) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return nil, errors.AuthenticationFailed("User not authenticated")
	}
	return currentUser, nil
}
