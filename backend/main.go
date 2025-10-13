package main

import (
	"log"
	"nimbus-backend/config"
	"nimbus-backend/database"
	"nimbus-backend/routes"
	"nimbus-backend/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Config yükleme
	cfg := config.Load()

	// Database bağlantısı
	if err := database.Connect(cfg); err != nil {
		log.Fatal("❌ Database bağlantı hatası:", err)
	}
	defer database.Close()

	// MinIO bağlantısı
	if err := services.InitMinIO(cfg); err != nil {
		log.Fatal("❌ MinIO bağlantı hatası:", err)
	}

	// Fiber uygulaması oluşturma
	app := fiber.New(fiber.Config{
		ServerHeader: "Nimbus",
		AppName:      "Nimbus v1.0",
	})

	// Middleware'ler
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5173",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	// Routes
	routes.SetupRoutes(app, cfg)

	// Sunucu başlatma
	log.Printf("🚀 Nimbus server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
