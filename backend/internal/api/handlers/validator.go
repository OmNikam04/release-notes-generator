package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/omnikam04/release-notes-generator/internal/dto"
)

// validate is a singleton validator instance
var validate = validator.New()

// ValidateStruct validates a struct and returns a Fiber error response if validation fails
func ValidateStruct(c *fiber.Ctx, s interface{}) error {
	if err := validate.Struct(s); err != nil {
		// Type assert to validator.ValidationErrors to get detailed error messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Get the first validation error for simplicity
			firstError := validationErrors[0]
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "validation_failed",
				Message: formatValidationError(firstError),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
	}
	return nil
}

// formatValidationError formats a validation error into a user-friendly message
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + err.Param() + " characters"
	case "max":
		return field + " must be at most " + err.Param() + " characters"
	default:
		return field + " is invalid"
	}
}
