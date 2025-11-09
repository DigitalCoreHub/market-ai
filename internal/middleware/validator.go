package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateStruct validates a struct using go-playground/validator
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidateRequest validates the request body and returns errors if invalid
func ValidateRequest(c *fiber.Ctx, req interface{}) error {
	// Parse JSON body
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate struct
	if err := ValidateStruct(req); err != nil {
		errors := make(map[string]string)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				errors[ve.Field()] = getValidationError(ve)
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Validation failed",
			"errors":  errors,
		})
	}

	return nil
}

func getValidationError(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return ve.Field() + " is required"
	case "email":
		return ve.Field() + " must be a valid email"
	case "min":
		return ve.Field() + " must be at least " + ve.Param() + " characters"
	case "max":
		return ve.Field() + " must be at most " + ve.Param() + " characters"
	case "oneof":
		return ve.Field() + " must be one of: " + ve.Param()
	default:
		return ve.Field() + " is invalid"
	}
}

