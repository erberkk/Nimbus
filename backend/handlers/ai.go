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

		// Initialize retrieval components
		intentClassifier := retrieval.NewIntentClassifier()
		termExtractor := retrieval.NewKeyTermExtractor()

		// Step 1: Analyze query intent
		intentMetadata := intentClassifier.AnalyzeQuery(req.Question)
		log.Printf("Query intent: %s (confidence: %.2f) - %s",
			intentMetadata.Intent, intentMetadata.Confidence, intentMetadata.Explanation)

		// Step 2: Extract key terms from query
		keyTerms := termExtractor.ExtractNamedTerms(req.Question)
		log.Printf("Extracted key terms: %v", keyTerms)

		// Step 3: Determine retrieval strategy based on intent
		var chunks []services.ChunkResult
		var retrievalErr error

		// Use hybrid search for comparison queries (keyword + semantic) to ensure we find comparison tables
		if intentMetadata.Intent == retrieval.IntentComparison && len(keyTerms) >= 2 {
			log.Printf("Using hybrid search for comparison query with %d terms", len(keyTerms))
			chunks, retrievalErr = performHybridRetrieval(
				ollamaService, chromaService,
				req.Question, keyTerms, req.FileID, intentMetadata.RecommendedTopK)
		} else if intentMetadata.Intent == retrieval.IntentDefinition && len(keyTerms) > 0 {
			// For definition queries, use hybrid search (keyword + semantic)
			log.Printf("Using hybrid search for definition query")
			chunks, retrievalErr = performHybridRetrieval(
				ollamaService, chromaService,
				req.Question, keyTerms, req.FileID, intentMetadata.RecommendedTopK)
		} else if intentMetadata.Intent == retrieval.IntentSummary {
			// For summary queries, retrieve more chunks for comprehensive overview
			topK := intentMetadata.RecommendedTopK
			if topK < 10 {
				topK = 10 // Minimum 10 chunks for summary
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
			// Standard semantic search with dynamic top-k based on intent
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

		// Debug: Log which chunks were retrieved
		chunkIDs := make([]string, len(chunks))
		for i, chunk := range chunks {
			chunkIDs[i] = chunk.ID
		}
		log.Printf("Retrieved %d relevant chunks for query: %v", len(chunks), chunkIDs)

		// If this is a comparison query, prioritize chunks with comparison type
		if intentMetadata.Intent == "comparison" {
			// Separate chunks by type
			var comparisonChunks []services.ChunkResult
			var otherChunks []services.ChunkResult
			for _, chunk := range chunks {
				if chunkType, ok := chunk.Metadata["chunk_type"].(string); ok && chunkType == "comparison" {
					comparisonChunks = append(comparisonChunks, chunk)
				} else {
					otherChunks = append(otherChunks, chunk)
				}
			}
			// Reorder: comparison chunks first, then others
			if len(comparisonChunks) > 0 {
				chunks = append(comparisonChunks, otherChunks...)
				log.Printf("Reordered chunks: %d comparison chunks prioritized", len(comparisonChunks))
			}

			// SECOND-LEVEL REORDERING: Find perfect match (comparison table with ALL query terms)
			// and move it to absolute position 0
			var perfectMatchIdx = -1
			for i, chunk := range chunks {
				textLower := strings.ToLower(chunk.Text)
				// Check if this chunk is a comparison table
				if strings.Contains(textLower, "comparison table:") || strings.Contains(textLower, "comparison of") {
					// Check if it contains ALL key terms
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

			// Move perfect match to position 0
			if perfectMatchIdx > 0 {
				perfectMatch := chunks[perfectMatchIdx]
				// Remove from current position
				chunks = append(chunks[:perfectMatchIdx], chunks[perfectMatchIdx+1:]...)
				// Insert at position 0
				chunks = append([]services.ChunkResult{perfectMatch}, chunks...)
				log.Printf("🎯 Perfect match comparison table moved to position 1 (was at position %d)", perfectMatchIdx+1)
			}

			// SMART OPTIMIZATION: Only reduce to single chunk if:
			// 1. Perfect match found
			// 2. Perfect match contains "COMPARISON TABLE" marker
			// 3. Query is a specific table query (not a general multi-chunk question)
			if perfectMatchIdx >= 0 {
				perfectMatch := chunks[0] // It's now at position 0
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

		// If this is a definition query, prioritize chunks that contain the term prominently
		if intentMetadata.Intent == "definition" && len(keyTerms) > 0 {
			primaryTerm := strings.ToLower(keyTerms[0]) // First key term is usually the term being defined

			// Find the best matching chunk (one that contains the term)
			var bestMatchIdx = -1
			var bestScore = 0

			for i, chunk := range chunks {
				textLower := strings.ToLower(chunk.Text)
				score := 0

				// Check if term appears in chunk
				if !strings.Contains(textLower, primaryTerm) {
					continue
				}

				// Score based on position and prominence
				// Higher score = better match
				if strings.HasPrefix(textLower, primaryTerm) {
					score += 100 // Term at start of chunk
				}

				// Check first 200 chars for prominence
				first200 := textLower
				if len(first200) > 200 {
					first200 = first200[:200]
				}
				if strings.Contains(first200, primaryTerm) {
					score += 50 // Term in first 200 chars
				}

				// Check if term appears as standalone word (not part of another word)
				if strings.Contains(first200, primaryTerm+" ") ||
					strings.Contains(first200, primaryTerm+"\n") ||
					strings.Contains(first200, " "+primaryTerm+" ") {
					score += 30 // Term as standalone word
				}

				// Prefer longer chunks (more complete definitions)
				if len(chunk.Text) > 200 {
					score += 10
				}

				// Count occurrences (more mentions = more relevant)
				occurrences := strings.Count(textLower, primaryTerm)
				score += occurrences * 5

				if score > bestScore {
					bestScore = score
					bestMatchIdx = i
				}
			}

			// Move best match to position 0
			if bestMatchIdx > 0 {
				bestMatch := chunks[bestMatchIdx]
				// Remove from current position
				chunks = append(chunks[:bestMatchIdx], chunks[bestMatchIdx+1:]...)
				// Insert at position 0
				chunks = append([]services.ChunkResult{bestMatch}, chunks...)
				log.Printf("📖 Definition term '%s' chunk moved to position 1 (was at position %d, score: %d)", primaryTerm, bestMatchIdx+1, bestScore)
			} else if bestMatchIdx == 0 {
				log.Printf("📖 Definition term '%s' already at position 1 (score: %d)", primaryTerm, bestScore)
			}
		}

		// Step 4: Extract chunk texts for context
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

// performMultiVectorRetrieval performs separate retrievals for each key term
// and merges the results. This is particularly effective for comparison queries.
func performMultiVectorRetrieval(
	ollamaService *services.OllamaService,
	chromaService *services.ChromaService,
	deduplicator *retrieval.ChunkDeduplicator,
	keyTerms []string,
	fileID string,
	baseTopK int,
) ([]services.ChunkResult, error) {
	var allResultSets [][]retrieval.ChunkResult

	// For each key term, perform a separate embedding and retrieval
	for _, term := range keyTerms {
		log.Printf("Retrieving chunks for term: %s", term)

		// Generate embedding for the term
		termEmbedding, err := ollamaService.GenerateEmbedding(term)
		if err != nil {
			log.Printf("Warning: Failed to generate embedding for term '%s': %v", term, err)
			continue
		}

		// Retrieve chunks for this term (smaller topK per term)
		perTermTopK := baseTopK / len(keyTerms)
		if perTermTopK < 2 {
			perTermTopK = 2
		}

		chunks, err := chromaService.QuerySimilar(termEmbedding, fileID, perTermTopK)
		if err != nil {
			log.Printf("Warning: Failed to retrieve chunks for term '%s': %v", term, err)
			continue
		}

		// Convert services.ChunkResult to retrieval.ChunkResult
		retrievalChunks := make([]retrieval.ChunkResult, len(chunks))
		for i, chunk := range chunks {
			retrievalChunks[i] = retrieval.ChunkResult{
				ID:       chunk.ID,
				Text:     chunk.Text,
				Metadata: chunk.Metadata,
				Distance: chunk.Distance,
			}
		}

		allResultSets = append(allResultSets, retrievalChunks)
		log.Printf("Retrieved %d chunks for term '%s'", len(chunks), term)
	}

	if len(allResultSets) == 0 {
		return nil, fmt.Errorf("failed to retrieve chunks for any term")
	}

	// Merge and deduplicate all results
	mergedChunks := deduplicator.RankAndMerge(allResultSets, baseTopK)

	// Convert back to services.ChunkResult
	finalChunks := make([]services.ChunkResult, len(mergedChunks))
	for i, chunk := range mergedChunks {
		finalChunks[i] = services.ChunkResult{
			ID:       chunk.ID,
			Text:     chunk.Text,
			Metadata: chunk.Metadata,
			Distance: chunk.Distance,
		}
	}

	log.Printf("Multi-vector retrieval: merged %d unique chunks from %d terms",
		len(finalChunks), len(keyTerms))

	return finalChunks, nil
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
	// Generate query embedding
	queryEmbedding, err := ollamaService.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform hybrid search (semantic + keyword)
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
	
	// Check for specific comparison table patterns
	hasComparisonOf := strings.Contains(qLower, "comparison of")
	hasCompareVs := strings.Contains(qLower, "compare") && (strings.Contains(qLower, " vs ") || strings.Contains(qLower, " vs. ") || strings.Contains(qLower, " versus "))
	hasVsComparison := (strings.Contains(qLower, " vs ") && strings.Contains(qLower, "comparison")) || 
	                   (strings.Contains(qLower, " vs. ") && strings.Contains(qLower, "comparison")) ||
	                   (strings.Contains(qLower, " versus ") && strings.Contains(qLower, "comparison"))
	
	// Check if query has comparison indicators with key terms
	hasComparisonIndicators := strings.Contains(qLower, "comparison") || 
	                           strings.Contains(qLower, " vs ") || 
	                           strings.Contains(qLower, " vs. ") || 
	                           strings.Contains(qLower, " versus ")
	
	// If it's a specific comparison pattern with key terms, it's a table query
	if (hasComparisonOf || hasCompareVs || hasVsComparison) && hasComparisonIndicators && len(keyTerms) >= 2 {
		return true
	}
	
	// Exclude general questions that need multiple chunks
	// These patterns indicate the user wants information from multiple sources/chunks
	// Examples: "tell me about X and Y", "what is the difference between X and Y", "explain X and Y"
	// If X is in chunk A (page 10) and Y is in chunk B (page 100), we need BOTH chunks
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
			return false // General question, needs multiple chunks (X in one chunk, Y in another)
		}
	}
	
	// If query mentions page numbers or ranges, it's likely a multi-chunk question
	if strings.Contains(qLower, "page") || strings.Contains(qLower, "section") || strings.Contains(qLower, "chapter") {
		return false
	}
	
	return false // Default: not a specific table query
}

// performQueryExpansion generates multiple query variations and retrieves for each
func performQueryExpansion(
	ollamaService *services.OllamaService,
	chromaService *services.ChromaService,
	deduplicator *retrieval.ChunkDeduplicator,
	originalQuery string,
	keyTerms []string,
	fileID string,
	topK int,
) ([]services.ChunkResult, error) {
	// Expand query into multiple variations
	expandedQueries := retrieval.ExpandQueryForComparison(originalQuery, keyTerms)
	log.Printf("Expanded query into %d variations", len(expandedQueries))

	var allResultSets [][]retrieval.ChunkResult

	// Retrieve for each expanded query
	for i, expQuery := range expandedQueries {
		// Limit to first 5 expansions to avoid too many API calls
		if i >= 5 {
			break
		}

		queryEmbedding, err := ollamaService.GenerateEmbedding(expQuery)
		if err != nil {
			log.Printf("Warning: Failed to generate embedding for expanded query '%s': %v", expQuery, err)
			continue
		}

		chunks, err := chromaService.QuerySimilar(queryEmbedding, fileID, 3) // Smaller per-query topK
		if err != nil {
			log.Printf("Warning: Failed to retrieve for expanded query '%s': %v", expQuery, err)
			continue
		}

		// Convert to retrieval.ChunkResult
		retrievalChunks := make([]retrieval.ChunkResult, len(chunks))
		for j, chunk := range chunks {
			retrievalChunks[j] = retrieval.ChunkResult{
				ID:       chunk.ID,
				Text:     chunk.Text,
				Metadata: chunk.Metadata,
				Distance: chunk.Distance,
			}
		}

		allResultSets = append(allResultSets, retrievalChunks)
	}

	if len(allResultSets) == 0 {
		return nil, fmt.Errorf("failed to retrieve chunks for any expanded query")
	}

	// Merge and deduplicate
	mergedChunks := deduplicator.RankAndMerge(allResultSets, topK)

	// Convert back to services.ChunkResult
	finalChunks := make([]services.ChunkResult, len(mergedChunks))
	for i, chunk := range mergedChunks {
		finalChunks[i] = services.ChunkResult{
			ID:       chunk.ID,
			Text:     chunk.Text,
			Metadata: chunk.Metadata,
			Distance: chunk.Distance,
		}
	}

	log.Printf("Query expansion: merged %d unique chunks from %d queries",
		len(finalChunks), len(allResultSets))

	return finalChunks, nil
}
