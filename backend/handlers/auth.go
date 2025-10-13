package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nimbus-backend/config"
	"nimbus-backend/database"
	"nimbus-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

var googleOauthConfig *oauth2.Config

func initGoogleConfig(cfg *config.Config) {
	googleOauthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleSecret,
		RedirectURL:  cfg.GoogleRedirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func GoogleLogin(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if googleOauthConfig == nil {
			initGoogleConfig(cfg)
		}

		// Random state oluşturma
		state, err := generateState()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "State oluşturma hatası",
			})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "oauth_state",
			Value:    state,
			MaxAge:   600, // 10 dakika
			HTTPOnly: true,
			Secure:   false, // Development için
		})

		// OAuth URL'ini oluştur ve kullanıcı seçme ekranını göster
		url := googleOauthConfig.AuthCodeURL(
			state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("prompt", "select_account"),
			oauth2.SetAuthURLParam("include_granted_scopes", "true"),
		)

		return c.Redirect(url)
	}
}

func GoogleCallback(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if googleOauthConfig == nil {
			initGoogleConfig(cfg)
		}

		code := c.Query("code")
		state := c.Query("state")

		// State kontrolü
		cookie := c.Cookies("oauth_state")
		if state != cookie {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz state parametresi",
			})
		}

		// Cookie'yi temizleme
		c.Cookie(&fiber.Cookie{
			Name:     "oauth_state",
			Value:    "",
			MaxAge:   -1,
			HTTPOnly: true,
		})

		if code == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Authorization code eksik",
			})
		}

		// Token exchange
		token, err := googleOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Token exchange error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Token exchange başarısız",
			})
		}

		// Google kullanıcı bilgilerini alma
		googleUser, err := getGoogleUserInfo(token.AccessToken)
		if err != nil {
			log.Printf("Google user info error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Kullanıcı bilgileri alınamadı",
			})
		}

		// Kullanıcıyı veritabanında bul veya oluştur
		user, err := findOrCreateUser(googleUser)
		if err != nil {
			log.Printf("User creation error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Kullanıcı kaydedilemedi",
			})
		}

		// JWT token oluşturma
		jwtToken, err := generateJWT(user, cfg.JWTSecret)
		if err != nil {
			log.Printf("JWT generation error: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Token oluşturma hatası",
			})
		}

		redirectURL := fmt.Sprintf("%s/dashboard?token=%s", cfg.FrontendURL, jwtToken)
		return c.Redirect(redirectURL)
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Client-side logout için basit response
		return c.JSON(fiber.Map{
			"message": "Başarıyla çıkış yapıldı",
		})
	}
}

func GetProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.Claims)

		userResponse := models.UserResponse{
			ID:     user.UserID,
			Email:  user.Email,
			Name:   user.Name,
			Avatar: user.Avatar,
		}

		return c.JSON(fiber.Map{
			"user": userResponse,
		})
	}
}

// Yardımcı fonksiyonlar

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func getGoogleUserInfo(accessToken string) (*GoogleUser, error) {
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser GoogleUser
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &googleUser, nil
}

func findOrCreateUser(googleUser *GoogleUser) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Önce mevcut kullanıcıyı ara
	var user models.User
	err := database.UserCollection.FindOne(ctx, bson.M{"google_id": googleUser.ID}).Decode(&user)

	if err == nil {
		// Kullanıcı mevcut, güncelleme zamanını ayarla
		user.UpdatedAt = time.Now()
		_, err = database.UserCollection.ReplaceOne(ctx, bson.M{"google_id": googleUser.ID}, user)
		return &user, err
	}

	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Yeni kullanıcı oluştur
	newUser := models.User{
		GoogleID:  googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		Avatar:    googleUser.Picture,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = database.UserCollection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func generateJWT(user *models.User, secret string) (string, error) {
	// Token 24 saat geçerli olacak
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := models.Claims{
		UserID:   user.ID.Hex(),
		Email:    user.Email,
		Name:     user.Name,
		Avatar:   user.Avatar,
		GoogleID: user.GoogleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "nimbus-backend",
			Subject:   user.ID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
