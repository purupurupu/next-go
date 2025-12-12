package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps go-playground/validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// New creates a new CustomValidator with custom validations registered
func New() *CustomValidator {
	v := validator.New()

	// Register custom validations
	v.RegisterValidation("hexcolor", validateHexColor)
	v.RegisterValidation("notblank", validateNotBlank)

	return &CustomValidator{validator: v}
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// hexColorRegex matches valid hex color codes (#RGB or #RRGGBB)
var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

// validateHexColor validates that a string is a valid hex color code
func validateHexColor(fl validator.FieldLevel) bool {
	return hexColorRegex.MatchString(fl.Field().String())
}

// validateNotBlank validates that a string is not empty or only whitespace
func validateNotBlank(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

// FormatValidationErrors converts validator.ValidationErrors to a map
func FormatValidationErrors(err error) map[string][]string {
	errors := make(map[string][]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			message := formatErrorMessage(e)
			errors[field] = append(errors[field], message)
		}
	}

	return errors
}

// formatErrorMessage creates a human-readable error message for a validation error
func formatErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + e.Param() + " characters"
	case "max":
		return "Must be at most " + e.Param() + " characters"
	case "hexcolor":
		return "Must be a valid hex color code (e.g., #FF0000)"
	case "notblank":
		return "Cannot be blank"
	case "oneof":
		return "Must be one of: " + e.Param()
	case "gte":
		return "Must be greater than or equal to " + e.Param()
	case "lte":
		return "Must be less than or equal to " + e.Param()
	default:
		return "Invalid value"
	}
}

// SetupValidator configures the validator for an Echo instance
func SetupValidator(e *echo.Echo) {
	e.Validator = New()
}
