package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse - Standardized error response
func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

// SuccessResponse - Standardized success response
func SuccessResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"data": data,
	})
}

// UnauthorizedResponse - Standardized unauthorized response
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": message,
	})
}

// BadRequestResponse - Standardized bad request response
func BadRequestResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": message,
	})
}

// NotFoundResponse - Standardized not found response
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": message,
	})
}

// ForbiddenResponse - Standardized forbidden response
func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"error": message,
	})
}

// InternalServerErrorResponse - Standardized internal server error response
func InternalServerErrorResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": message,
	})
}
