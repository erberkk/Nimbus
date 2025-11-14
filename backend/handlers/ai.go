package handlers

import (
	"log"
	"nimbus-backend/config"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"nimbus-backend/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type QueryDocumentRequest struct {
	FileID   string `json:"file_id" validate:"required"`
	Question string `json:"question" validate:"required"`
}

type QueryDocumentResponse struct {
	Answer     string   `json:"answer"`
	Sources    []string `json:"sources"`
	ChunkCount int      `json:"chunk_count"`
}

// QueryDocument - Query a processed document using RAG
func QueryDocument(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var req QueryDocumentRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek verisi",
			})
		}

		if req.FileID == "" || req.Question == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "file_id ve question parametreleri gerekli",
			})
		}

		// Get file record
		file, err := services.FileServiceInstance.GetFileByID(req.FileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Check if user has access to the file
		if file.UserID != userID {
			// Check if file is shared with read or write access
			hasAccess := false
			for _, access := range file.AccessList {
				if access.UserID == userID {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				return c.Status(403).JSON(fiber.Map{
					"error": "Bu dosyaya erişim yetkiniz yok",
				})
			}
		}

		// Check if file has been processed
		if file.ProcessingStatus != "completed" {
			return c.Status(409).JSON(fiber.Map{
				"error":  "Dosya henüz işlenmedi. Lütfen işlem tamamlanana kadar bekleyin.",
				"status": file.ProcessingStatus,
			})
		}

		// Initialize services
		ollamaService := services.NewOllamaService(cfg)
		chromaService := services.NewChromaService(cfg)

		// Step 1: Generate embedding for the question
		questionEmbedding, err := ollamaService.GenerateEmbedding(req.Question)
		if err != nil {
			log.Printf("Failed to generate embedding for question: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Soru işlenirken hata oluştu",
			})
		}

		// Step 2: Query Chroma for similar chunks (top-5)
		chunks, err := chromaService.QuerySimilar(questionEmbedding, req.FileID, 5)
		if err != nil {
			log.Printf("Failed to query Chroma: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Benzer içerikler aranırken hata oluştu",
			})
		}

		if len(chunks) == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosyada ilgili içerik bulunamadı",
			})
		}

		// Step 3: Extract chunk texts for context
		var contextChunks []string
		var sources []string
		for _, chunk := range chunks {
			contextChunks = append(contextChunks, chunk.Text)
			// Keep sources short for response
			chunkPreview := chunk.Text
			if len(chunkPreview) > 200 {
				chunkPreview = chunkPreview[:200] + "..."
			}
			sources = append(sources, chunkPreview)
		}

		log.Printf("Found %d relevant chunks for question: %s", len(chunks), req.Question)

		// Step 4: Generate answer using LLM with context
		answer, err := ollamaService.GenerateRAGResponse(req.Question, contextChunks)
		if err != nil {
			log.Printf("Failed to generate answer: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Cevap oluşturulurken hata oluştu",
			})
		}

		// Clean up answer
		answer = strings.TrimSpace(answer)

		// Save user question to conversation history
		userMessage := models.Message{
			Role:      "user",
			Content:   req.Question,
			Timestamp: time.Now(),
		}
		if err := services.ConversationServiceInstance.AddMessage(userID, req.FileID, userMessage); err != nil {
			log.Printf("Warning: Failed to save user message: %v", err)
			// Don't fail the request, just log the error
		}

		// Save assistant answer to conversation history
		assistantMessage := models.Message{
			Role:      "assistant",
			Content:   answer,
			Sources:   sources,
			Timestamp: time.Now(),
		}
		if err := services.ConversationServiceInstance.AddMessage(userID, req.FileID, assistantMessage); err != nil {
			log.Printf("Warning: Failed to save assistant message: %v", err)
			// Don't fail the request, just log the error
		}

		// Return response
		return c.JSON(QueryDocumentResponse{
			Answer:     answer,
			Sources:    sources,
			ChunkCount: len(chunks),
		})
	}
}

// GetConversationHistory retrieves chat history for a specific file
func GetConversationHistory(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Yetkisiz erişim",
			})
		}

		fileID := c.Query("file_id")
		if fileID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "file_id gerekli",
			})
		}

		// Check if user has access to the file
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Verify access
		if file.UserID != userID {
			hasAccess := false
			for _, access := range file.AccessList {
				if access.UserID == userID {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				return c.Status(403).JSON(fiber.Map{
					"error": "Bu dosyaya erişim yetkiniz yok",
				})
			}
		}

		// Get conversation
		conversation, err := services.ConversationServiceInstance.GetConversation(userID, fileID)
		if err != nil {
			// No conversation yet, return empty
			return c.JSON(models.ConversationResponse{
				ID:        "",
				FileID:    fileID,
				Messages:  []models.Message{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}

		// Return conversation
		return c.JSON(models.ConversationResponse{
			ID:        conversation.ID.Hex(),
			FileID:    conversation.FileID,
			Messages:  conversation.Messages,
			CreatedAt: conversation.CreatedAt,
			UpdatedAt: conversation.UpdatedAt,
		})
	}
}

// ClearConversationHistory clears all messages from a conversation
func ClearConversationHistory(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Yetkisiz erişim",
			})
		}

		fileID := c.Query("file_id")
		if fileID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "file_id gerekli",
			})
		}

		// Check if user has access to the file
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Only file owner can clear conversation
		if file.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Sadece dosya sahibi sohbet geçmişini temizleyebilir",
			})
		}

		// Clear conversation
		if err := services.ConversationServiceInstance.ClearConversation(userID, fileID); err != nil {
			log.Printf("Failed to clear conversation: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Sohbet geçmişi temizlenemedi",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Sohbet geçmişi başarıyla temizlendi",
		})
	}
}

