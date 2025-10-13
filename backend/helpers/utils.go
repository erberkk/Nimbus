package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// GeneratePublicLink - Rastgele 16 karakterlik public link oluştur
func GeneratePublicLink() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
