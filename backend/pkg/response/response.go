package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Success sends a successful response with data
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
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
