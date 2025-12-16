package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// OK sends a 200 response with data
func OK(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

// Created sends a 201 response with data
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, data)
}

// NoContent sends a 204 response with no body
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
