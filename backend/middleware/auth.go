package middleware

import (
	"nimbus-backend/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth - Authentication middleware
// Extracts user info from JWT and stores it in locals
func RequireAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization header eksik",
			})
		}

		// Bearer token kontrolü
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{
				"error": "Bearer token formatı hatalı",
			})
		}

		// Token parse etme
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{
				"error": "Geçersiz token",
			})
		}

		// Claims'leri context'e ekleme
		if claims, ok := token.Claims.(*models.Claims); ok {
			c.Locals("user", claims)
		}

		return c.Next()
	}
}
