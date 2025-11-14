package routes

import (
	"nimbus-backend/config"
	"nimbus-backend/handlers"
	"nimbus-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, cfg *config.Config) {
	// API v1 routes
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.Get("/google", handlers.GoogleLogin(cfg))
		auth.Get("/google/callback", handlers.GoogleCallback(cfg))
		auth.Post("/logout", handlers.Logout())
	}

	// Legacy route for backward compatibility (without /api/v1 prefix)
	app.Get("/auth/google/callback", handlers.GoogleCallback(cfg))

	// Protected routes
	protected := api.Group("/user")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		protected.Get("/profile", handlers.GetProfile())
	}

	// File routes (protected)
	files := api.Group("/files")
	files.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		files.Get("/upload-url", handlers.GetUploadPresignedURL(cfg))
		files.Post("/", handlers.CreateFile(cfg))
		files.Get("/", handlers.ListUserFiles(cfg))
		files.Delete("/:id", handlers.DeleteFile(cfg))
		files.Get("/download-url", handlers.GetDownloadPresignedURL(cfg))
		files.Get("/preview-url", handlers.GetPreviewPresignedURL(cfg)) // Supports ?file_id=xxx or ?filename=xxx
		files.Post("/:id/move", handlers.MoveFile(cfg))
		files.Get("/onlyoffice-config", handlers.GetOnlyOfficeConfig(cfg)) // OnlyOffice editor config
		files.Get("/content", handlers.GetFileContent(cfg))                // Get file content as text (for code files)
		files.Put("/:id/content", handlers.UpdateFileContent(cfg))         // Update file content (for code files)
		files.Post("/:id/process", handlers.ProcessDocument(cfg))          // Trigger document processing for RAG
	}

	// AI routes (protected)
	ai := api.Group("/ai")
	ai.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		ai.Post("/query", handlers.QueryDocument(cfg))                     // Query processed document
		ai.Get("/conversation", handlers.GetConversationHistory(cfg))      // Get chat history
		ai.Delete("/conversation", handlers.ClearConversationHistory(cfg)) // Clear chat history
	}

	// OnlyOffice callback (public - called by OnlyOffice server)
	api.Post("/files/onlyoffice-callback", handlers.OnlyOfficeCallback(cfg))

	// OnlyOffice document proxy (public - token-protected, called by OnlyOffice server)
	api.Get("/files/onlyoffice-document", handlers.GetOnlyOfficeDocument(cfg))

	// Folder routes (protected)
	folders := api.Group("/folders")
	folders.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		folders.Post("/", handlers.CreateFolder(cfg))
		folders.Get("/", handlers.GetUserFolders(cfg))
		folders.Get("/root", handlers.GetRootContents(cfg))
		folders.Get("/storage", handlers.GetStorageUsage(cfg))
		folders.Get("/:id", handlers.GetFolderContents(cfg))
		folders.Put("/:id", handlers.UpdateFolder(cfg))
		folders.Delete("/:id", handlers.DeleteFolder(cfg))
	}

	// Share routes (access list management)
	shares := api.Group("/shares")
	shares.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		shares.Get("/resource/:resourceId", handlers.GetResourceShares())
		shares.Get("/shared-with-me", handlers.GetSharedWithMe())
		shares.Get("/shared-folder/:folderId", handlers.GetSharedFolderContents())
		shares.Put("/access/:resourceId", handlers.UpdateAccessPermission())
		shares.Delete("/access/:resourceId/:userId", handlers.RemoveUserAccess())
		// Public link access (no JWT required for link generation, but required for access)
		shares.Get("/public/:publicLink", handlers.GetResourceByPublicLink())
	}

	// User search (protected)
	users := api.Group("/users")
	users.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		users.Get("/search", handlers.SearchUsers())
	}
}
