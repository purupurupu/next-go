package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Meta contains pagination metadata
type Meta struct {
	Total       int64 `json:"total"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PerPage     int   `json:"per_page"`
}

// Success sends a successful response with data
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// SuccessWithMessage sends a successful response with data and message
func SuccessWithMessage(c echo.Context, data interface{}, message string) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": message,
		"data":    data,
	})
}

// Created sends a 201 response for newly created resources
func Created(c echo.Context, data interface{}, message string) error {
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": message,
		"data":    data,
	})
}

// NoContent sends a 204 response with no body
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Message sends a response with only a message
func Message(c echo.Context, message string) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": message,
	})
}

// Paginated sends a paginated response with metadata
func Paginated(c echo.Context, data interface{}, total int64, page, perPage int) error {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": data,
		"meta": Meta{
			Total:       total,
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	})
}

// PaginatedWithKey sends a paginated response with a custom key for the data
func PaginatedWithKey(c echo.Context, key string, data interface{}, total int64, page, perPage int) error {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		key: data,
		"meta": Meta{
			Total:       total,
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	})
}
