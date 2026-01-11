package handlers

import (
	"fmt"
	"log"
	"nimbus-backend/config"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"nimbus-backend/retrieval"
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

		file, err := services.FileServiceInstance.GetFileByID(req.FileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		hasAccess, err := helpers.CanUserAccess(userID, "file", req.FileID, helpers.AccessLevelRead)
		if err != nil || !hasAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyaya erişim yetkiniz yok",
			})
		}

		if file.ProcessingStatus != "completed" {
			return c.Status(409).JSON(fiber.Map{
				"error":  "Dosya henüz işlenmedi. Lütfen işlem tamamlanana kadar bekleyin.",
				"status": file.ProcessingStatus,
			})
		}

		ollamaService := services.NewOllamaService(cfg)
		chromaService := services.NewChromaService(cfg)

		intentClassifier := retrieval.NewIntentClassifier()
		termExtractor := retrieval.NewKeyTermExtractor()

		intentMetadata := intentClassifier.AnalyzeQuery(req.Question)
		log.Printf("Query intent: %s (confidence: %.2f) - %s",
			intentMetadata.Intent, intentMetadata.Confidence, intentMetadata.Explanation)

		keyTerms := termExtractor.ExtractNamedTerms(req.Question)
		log.Printf("Extracted key terms: %v", keyTerms)

		var chunks []services.ChunkResult
		var retrievalErr error

		if intentMetadata.Intent == retrieval.IntentComparison && len(keyTerms) >= 2 {
			log.Printf("Using hybrid search for comparison query with %d terms", len(keyTerms))
			chunks, retrievalErr = performHybridRetrieval(
				ollamaService, chromaService,
				req.Question, keyTerms, req.FileID, intentMetadata.RecommendedTopK)
		} else if intentMetadata.Intent == retrieval.IntentDefinition && len(keyTerms) > 0 {
			log.Printf("Using hybrid search for definition query")
			chunks, retrievalErr = performHybridRetrieval(
				ollamaService, chromaService,
				req.Question, keyTerms, req.FileID, intentMetadata.RecommendedTopK)
		} else if intentMetadata.Intent == retrieval.IntentSummary {
			topK := intentMetadata.RecommendedTopK
			if topK < 10 {
				topK = 10
			}
			log.Printf("Using standard semantic search for summary query with top-k=%d", topK)
			questionEmbedding, embErr := ollamaService.GenerateEmbedding(req.Question)
			if embErr != nil {
				log.Printf("Failed to generate embedding for question: %v", embErr)
				return c.Status(500).JSON(fiber.Map{
					"error": "Soru işlenirken hata oluştu",
				})
			}
			chunks, retrievalErr = chromaService.QuerySimilar(questionEmbedding, req.FileID, topK)
		} else {
			log.Printf("Using standard semantic search with top-k=%d", intentMetadata.RecommendedTopK)
			questionEmbedding, embErr := ollamaService.GenerateEmbedding(req.Question)
			if embErr != nil {
				log.Printf("Failed to generate embedding for question: %v", embErr)
				return c.Status(500).JSON(fiber.Map{
					"error": "Soru işlenirken hata oluştu",
				})
			}

			chunks, retrievalErr = chromaService.QuerySimilar(questionEmbedding, req.FileID, intentMetadata.RecommendedTopK)
		}

		if retrievalErr != nil {
			log.Printf("Failed to retrieve chunks: %v", retrievalErr)
			return c.Status(500).JSON(fiber.Map{
				"error": "İçerik arama işlemi başarısız oldu",
			})
		}

		if len(chunks) == 0 {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosyada ilgili içerik bulunamadı",
			})
		}

		chunkIDs := make([]string, len(chunks))
		for i, chunk := range chunks {
			chunkIDs[i] = chunk.ID
		}
		log.Printf("Retrieved %d relevant chunks for query: %v", len(chunks), chunkIDs)

		if intentMetadata.Intent == "comparison" {
			var comparisonChunks []services.ChunkResult
			var otherChunks []services.ChunkResult
			for _, chunk := range chunks {
				if chunkType, ok := chunk.Metadata["chunk_type"].(string); ok && chunkType == "comparison" {
					comparisonChunks = append(comparisonChunks, chunk)
				} else {
					otherChunks = append(otherChunks, chunk)
				}
			}
			if len(comparisonChunks) > 0 {
				chunks = append(comparisonChunks, otherChunks...)
				log.Printf("Reordered chunks: %d comparison chunks prioritized", len(comparisonChunks))
			}

			var perfectMatchIdx = -1
			for i, chunk := range chunks {
				textLower := strings.ToLower(chunk.Text)
				if strings.Contains(textLower, "comparison table:") || strings.Contains(textLower, "comparison of") {
					allTermsFound := true
					for _, term := range keyTerms {
						if !strings.Contains(textLower, strings.ToLower(term)) {
							allTermsFound = false
							break
						}
					}
					if allTermsFound {
						perfectMatchIdx = i
						break
					}
				}
			}

			if perfectMatchIdx > 0 {
				perfectMatch := chunks[perfectMatchIdx]
				chunks = append(chunks[:perfectMatchIdx], chunks[perfectMatchIdx+1:]...)
				chunks = append([]services.ChunkResult{perfectMatch}, chunks...)
				log.Printf("🎯 Perfect match comparison table moved to position 1 (was at position %d)", perfectMatchIdx+1)
			}

			if perfectMatchIdx >= 0 {
				perfectMatch := chunks[0]
				textLower := strings.ToLower(perfectMatch.Text)
				hasComparisonTableMarker := strings.Contains(textLower, "comparison table:") ||
					strings.Contains(textLower, "comparison of")

				if hasComparisonTableMarker && isSpecificTableQuery(req.Question, keyTerms) {
					chunks = []services.ChunkResult{perfectMatch}
					log.Printf("🔥 Reduced to ONLY perfect match chunk (specific table query)")
				} else {
					log.Printf("✅ Keeping all chunks (multi-chunk question or general comparison)")
				}
			}
		}

		if intentMetadata.Intent == "definition" && len(keyTerms) > 0 {
			primaryTerm := strings.ToLower(keyTerms[0])

			var bestMatchIdx = -1
			var bestScore = 0

			for i, chunk := range chunks {
				textLower := strings.ToLower(chunk.Text)
				score := 0

				if !strings.Contains(textLower, primaryTerm) {
					continue
				}

				if strings.HasPrefix(textLower, primaryTerm) {
					score += 100
				}

				first200 := textLower
				if len(first200) > 200 {
					first200 = first200[:200]
				}
				if strings.Contains(first200, primaryTerm) {
					score += 50
				}

				if strings.Contains(first200, primaryTerm+" ") ||
					strings.Contains(first200, primaryTerm+"\n") ||
					strings.Contains(first200, " "+primaryTerm+" ") {
					score += 30
				}

				if len(chunk.Text) > 200 {
					score += 10
				}

				occurrences := strings.Count(textLower, primaryTerm)
				score += occurrences * 5

				if score > bestScore {
					bestScore = score
					bestMatchIdx = i
				}
			}

			if bestMatchIdx > 0 {
				bestMatch := chunks[bestMatchIdx]
				chunks = append(chunks[:bestMatchIdx], chunks[bestMatchIdx+1:]...)
				chunks = append([]services.ChunkResult{bestMatch}, chunks...)
				log.Printf("📖 Definition term '%s' chunk moved to position 1 (was at position %d, score: %d)", primaryTerm, bestMatchIdx+1, bestScore)
			} else if bestMatchIdx == 0 {
				log.Printf("📖 Definition term '%s' already at position 1 (score: %d)", primaryTerm, bestScore)
			}
		}

		var contextChunks []string
		var sources []string
		for _, chunk := range chunks {
			contextChunks = append(contextChunks, chunk.Text)
			chunkPreview := chunk.Text
			if len(chunkPreview) > 200 {
				chunkPreview = chunkPreview[:200] + "..."
			}
			sources = append(sources, chunkPreview)
		}

		log.Printf("Found %d relevant chunks for question: %s", len(chunks), req.Question)

		answer, err := ollamaService.GenerateRAGResponse(req.Question, contextChunks)
		if err != nil {
			log.Printf("Failed to generate answer: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Cevap oluşturulurken hata oluştu",
			})
		}

		answer = strings.TrimSpace(answer)

		userMessage := models.Message{
			Role:      "user",
			Content:   req.Question,
			Timestamp: time.Now(),
		}
		if err := services.ConversationServiceInstance.AddMessage(userID, req.FileID, userMessage); err != nil {
			log.Printf("Warning: Failed to save user message: %v", err)
		}

		assistantMessage := models.Message{
			Role:      "assistant",
			Content:   answer,
			Sources:   sources,
			Timestamp: time.Now(),
		}
		if err := services.ConversationServiceInstance.AddMessage(userID, req.FileID, assistantMessage); err != nil {
			log.Printf("Warning: Failed to save assistant message: %v", err)
		}

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

		hasAccess, err := helpers.CanUserAccess(userID, "file", fileID, helpers.AccessLevelRead)
		if err != nil || !hasAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyaya erişim yetkiniz yok",
			})
		}

		conversation, err := services.ConversationServiceInstance.GetConversation(userID, fileID)
		if err != nil {
			return c.JSON(models.ConversationResponse{
				ID:        "",
				FileID:    fileID,
				Messages:  []models.Message{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}

		return c.JSON(models.ConversationResponse{
			ID:        conversation.ID.Hex(),
			FileID:    conversation.FileID,
			Messages:  conversation.Messages,
			CreatedAt: conversation.CreatedAt,
			UpdatedAt: conversation.UpdatedAt,
		})
	}
}

// GetUserConversations retrieves all conversations for the current user
func GetUserConversations(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Yetkisiz erişim",
			})
		}

		conversations, err := services.ConversationServiceInstance.GetUserConversations(userID)
		if err != nil {
			log.Printf("Failed to get user conversations: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Sohbet geçmişleri alınamadı",
			})
		}

		return c.JSON(fiber.Map{
			"conversations": conversations,
			"count":         len(conversations),
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

		hasAccess, err := helpers.CanUserAccess(userID, "file", fileID, helpers.AccessLevelRead)
		if err != nil || !hasAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyaya erişim yetkiniz yok",
			})
		}

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

// performHybridRetrieval combines semantic and keyword search
func performHybridRetrieval(
	ollamaService *services.OllamaService,
	chromaService *services.ChromaService,
	query string,
	keywords []string,
	fileID string,
	topK int,
) ([]services.ChunkResult, error) {
	queryEmbedding, err := ollamaService.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	chunks, err := chromaService.HybridSearch(queryEmbedding, keywords, fileID, topK)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	log.Printf("Hybrid retrieval returned %d chunks", len(chunks))
	return chunks, nil
}

// isSpecificTableQuery checks if the query is asking for a specific comparison table
// Returns true if the query is clearly asking for a specific comparison table (e.g., "comparison of X vs Y")
// Returns false for general questions that might need multiple chunks (e.g., "tell me about X and Y")
func isSpecificTableQuery(question string, keyTerms []string) bool {
	qLower := strings.ToLower(question)

	hasComparisonOf := strings.Contains(qLower, "comparison of")
	hasCompareVs := strings.Contains(qLower, "compare") && (strings.Contains(qLower, " vs ") || strings.Contains(qLower, " vs. ") || strings.Contains(qLower, " versus "))
	hasVsComparison := (strings.Contains(qLower, " vs ") && strings.Contains(qLower, "comparison")) ||
		(strings.Contains(qLower, " vs. ") && strings.Contains(qLower, "comparison")) ||
		(strings.Contains(qLower, " versus ") && strings.Contains(qLower, "comparison"))

	hasComparisonIndicators := strings.Contains(qLower, "comparison") ||
		strings.Contains(qLower, " vs ") ||
		strings.Contains(qLower, " vs. ") ||
		strings.Contains(qLower, " versus ")

	if (hasComparisonOf || hasCompareVs || hasVsComparison) && hasComparisonIndicators && len(keyTerms) >= 2 {
		return true
	}

	excludePatterns := []string{
		"tell me about",
		"what is the difference between",
		"what are the differences between",
		"explain",
		"describe",
		"what are the",
		"how are",
		"what about",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(qLower, pattern) && strings.Contains(qLower, " and ") {
			return false
		}
	}

	if strings.Contains(qLower, "page") || strings.Contains(qLower, "section") || strings.Contains(qLower, "chapter") {
		return false
	}

	return false
}
