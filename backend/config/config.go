package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	MongoURI              string
	MongoDB               string
	JWTSecret             string
	GoogleClientID        string
	GoogleSecret          string
	GoogleRedirect        string
	FrontendURL           string
	MinIOEndpoint         string
	MinIOAccessKey        string
	MinIOSecretKey        string
	MinIOUseSSL           bool
	OnlyOfficeServerURL   string
	OnlyOfficeJWTSecret   string
	BackendURL            string
	MinIOExternalEndpoint string // For OnlyOffice to access MinIO from Docker
	BackendExternalURL    string // For OnlyOffice to access backend from Docker
	OllamaBaseURL         string // Ollama API endpoint
	OllamaEmbedModel      string // Embedding model name
	OllamaLLMModel        string // LLM model name
	ChromaBaseURL         string // Chroma vector DB endpoint
	ChromaTenant          string // Chroma tenant name
	ChromaDatabase        string // Chroma database name
	ChromaCollection      string // Chroma collection name
}

func Load() *Config {
	// .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env dosyası bulunamadı veya yüklenemedi")
	}
	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		MongoURI:              getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:               getEnv("MONGO_DB", "nimbus"),
		JWTSecret:             getEnv("JWT_SECRET", "your-secret-key"),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:          getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirect:        getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:5173"),
		MinIOEndpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:        getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:        getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOUseSSL:           getEnvAsBool("MINIO_USE_SSL", false),
		OnlyOfficeServerURL:   getEnv("ONLYOFFICE_SERVER_URL", "http://localhost:5000"),
		OnlyOfficeJWTSecret:   getEnv("ONLYOFFICE_JWT_SECRET", "your-secret-key"),
		BackendURL:            getEnv("BACKEND_URL", "http://localhost:8080"),
		MinIOExternalEndpoint: getEnv("MINIO_EXTERNAL_ENDPOINT", "host.docker.internal:9000"),
		BackendExternalURL:    getEnv("BACKEND_EXTERNAL_URL", "http://host.docker.internal:8080"),
		OllamaBaseURL:         getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaEmbedModel:      getEnv("OLLAMA_EMBED_MODEL", "all-minilm:l6-v2"),
		OllamaLLMModel:        getEnv("OLLAMA_LLM_MODEL", "llama3:8b"),
		ChromaBaseURL:         getEnv("CHROMA_BASE_URL", "http://localhost:6006"),
		ChromaTenant:          getEnv("CHROMA_TENANT", "default_tenant"),
		ChromaDatabase:        getEnv("CHROMA_DATABASE", "default_database"),
		ChromaCollection:      getEnv("CHROMA_COLLECTION", "nimbus_documents"),
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
