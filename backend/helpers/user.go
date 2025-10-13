package helpers

import (
	"nimbus-backend/models"

	"github.com/gofiber/fiber/v2"
)

// GetCurrentUser - JWT middleware'den gelen user bilgisini döndürür
func GetCurrentUser(c *fiber.Ctx) (*models.Claims, error) {
	user := c.Locals("user")
	if user == nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Kullanıcı bilgisi bulunamadı")
	}

	claims, ok := user.(*models.Claims)
	if !ok {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Geçersiz kullanıcı bilgisi")
	}

	return claims, nil
}

// GetCurrentUserID - Sadece user ID'sini döndürür
func GetCurrentUserID(c *fiber.Ctx) (string, error) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return "", err
	}
	return user.UserID, nil
}
