package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	MongoURI       string
	MongoDB        string
	JWTSecret      string
	GoogleClientID string
	GoogleSecret   string
	GoogleRedirect string
	FrontendURL    string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
}

func Load() *Config {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env dosyası bulunamadı veya yüklenemedi")
	}
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:        getEnv("MONGO_DB", "nimbus"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirect: getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:5173"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOUseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
	}

	if cfg.GoogleClientID == "" || cfg.GoogleSecret == "" {
		log.Fatal("❌ GOOGLE_CLIENT_ID ve GOOGLE_CLIENT_SECRET environment variables gerekli!")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}
