package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"nimbus-backend/cache"
	"nimbus-backend/config"
	"nimbus-backend/retrieval"
)

type ChromaService struct {
	baseURL        string
	tenant         string
	database       string
	collectionName string
	collectionID   string // UUID of the collection
	httpClient     *http.Client
	chunkCache     cache.ChunkCache
	fileRouter     *retrieval.FileRouter
	config         *config.Config
}

type ChunkData struct {
	ID        string                 `json:"id"`
	Embedding []float64              `json:"embedding"`
	Text      string                 `json:"text"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ChunkResult struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"document"`
	Metadata map[string]interface{} `json:"metadata"`
	Distance float64                `json:"distance"`
}

type AddDocumentsRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float64              `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
}

type QueryRequest struct {
	QueryEmbeddings [][]float64 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Where           interface{} `json:"where,omitempty"`
}

type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Documents [][]string                 `json:"documents"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Distances [][]float64                `json:"distances"`
}

type CollectionResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewChromaService(cfg *config.Config) *ChromaService {
	var chunkCache cache.ChunkCache
	var fileRouter *retrieval.FileRouter

	// Initialize chunk cache if enabled
	if cfg.EnableChunkCache {
		chunkCache = cache.NewLRUChunkCache(cfg.ChunkCacheSize)
		log.Printf("Chunk cache enabled with size: %d", cfg.ChunkCacheSize)
	}

	// Initialize file router if enabled
	if cfg.EnableFileRouting {
		fileRouter = retrieval.NewFileRouter(24 * time.Hour) // 24 hour TTL
		log.Println("File-level routing enabled")
	}

	return &ChromaService{
		baseURL:        cfg.ChromaBaseURL,
		tenant:         cfg.ChromaTenant,
		database:       cfg.ChromaDatabase,
		collectionName: cfg.ChromaCollection,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		chunkCache: chunkCache,
		fileRouter: fileRouter,
		config:     cfg,
	}
}

// EnsureCollection gets or creates a collection and stores its UUID
// Always verifies the collection exists, even if we have a cached ID
func (s *ChromaService) EnsureCollection() error {
	// Try to get existing collection (even if we have a cached ID, verify it's still valid)
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s",
		s.baseURL, s.tenant, s.database, s.collectionName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create get collection request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	// If collection exists, get its ID
	if resp.StatusCode == http.StatusOK {
		var collResp CollectionResponse
		if err := json.NewDecoder(resp.Body).Decode(&collResp); err != nil {
			return fmt.Errorf("failed to decode collection response: %w", err)
		}
		s.collectionID = collResp.ID
		return nil
	}

	// If collection doesn't exist (404), create it
	if resp.StatusCode == http.StatusNotFound {
		createURL := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
			s.baseURL, s.tenant, s.database)

		createBody := map[string]interface{}{
			"name": s.collectionName,
		}

		jsonData, err := json.Marshal(createBody)
		if err != nil {
			return fmt.Errorf("failed to marshal create collection request: %w", err)
		}

		createReq, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create collection request: %w", err)
		}
		createReq.Header.Set("Content-Type", "application/json")

		createResp, err := s.httpClient.Do(createReq)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(createResp.Body)
			return fmt.Errorf("failed to create collection, status %d: %s", createResp.StatusCode, string(bodyBytes))
		}

		var collResp CollectionResponse
		if err := json.NewDecoder(createResp.Body).Decode(&collResp); err != nil {
			return fmt.Errorf("failed to decode create collection response: %w", err)
		}
		s.collectionID = collResp.ID
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status when getting collection: %d, body: %s", resp.StatusCode, string(bodyBytes))
}

func (s *ChromaService) AddDocuments(chunks []ChunkData) error {
	if len(chunks) == 0 {
		return nil
	}

	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	ids := make([]string, len(chunks))
	embeddings := make([][]float64, len(chunks))
	documents := make([]string, len(chunks))
	metadatas := make([]map[string]interface{}, len(chunks))

	for i, chunk := range chunks {
		ids[i] = chunk.ID
		embeddings[i] = chunk.Embedding
		documents[i] = chunk.Text
		metadatas[i] = chunk.Metadata

		// Warm chunk cache if enabled
		if s.chunkCache != nil && s.config.EnableChunkCache {
			s.chunkCache.Set(chunk.ID, chunk.Embedding)
		}
	}

	reqBody := AddDocumentsRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal add documents request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/upsert",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create add documents request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add documents to chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma add documents api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Sync with file router if enabled and we have file_id metadata
	if s.fileRouter != nil && s.config.EnableFileRouting && len(chunks) > 0 {
		if fileID, ok := chunks[0].Metadata["file_id"].(string); ok {
			go s.syncFileRouter(fileID, chunks)
		}
	}

	return nil
}

// syncFileRouter syncs chunk embeddings with the file router
func (s *ChromaService) syncFileRouter(fileID string, chunks []ChunkData) {
	routerChunks := make([]retrieval.ChunkEmbedding, len(chunks))
	for i, chunk := range chunks {
		routerChunks[i] = retrieval.ChunkEmbedding{
			ChunkID:   chunk.ID,
			Embedding: chunk.Embedding,
			Text:      chunk.Text,
			Metadata:  chunk.Metadata,
		}
	}

	s.fileRouter.SyncWithChroma(fileID, routerChunks)
}

func (s *ChromaService) QuerySimilar(queryEmbedding []float64, fileID string, topK int) ([]ChunkResult, error) {
	// Try file router first if enabled
	if s.fileRouter != nil && s.config.EnableFileRouting {
		if s.fileRouter.HasIndex(fileID) {
			log.Printf("Using file router for fast in-memory search (file: %s)", fileID)

			results := s.fileRouter.SearchInFile(fileID, queryEmbedding, topK)
			chromaResults := make([]ChunkResult, len(results))
			for i, result := range results {
				chromaResults[i] = ChunkResult{
					ID:       result.ChunkID,
					Text:     result.Text,
					Metadata: result.Metadata,
					Distance: result.Distance,
				}

				// Record chunk access for popularity tracking
				if s.chunkCache != nil {
					s.chunkCache.RecordAccess(result.ChunkID)
				}
			}

			return chromaResults, nil
		}
	}

	// Fall back to Chroma query
	return s.queryChromaDB(queryEmbedding, fileID, topK)
}

// queryChromaDB performs the actual Chroma database query
func (s *ChromaService) queryChromaDB(queryEmbedding []float64, fileID string, topK int) ([]ChunkResult, error) {
	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	where := map[string]interface{}{
		"file_id": fileID,
	}

	reqBody := QueryRequest{
		QueryEmbeddings: [][]float64{queryEmbedding},
		NResults:        topK,
		Where:           where,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chroma query api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	var results []ChunkResult
	if len(queryResp.Documents) > 0 && len(queryResp.Documents[0]) > 0 {
		for i := 0; i < len(queryResp.Documents[0]); i++ {
			result := ChunkResult{
				ID:       queryResp.IDs[0][i],
				Text:     queryResp.Documents[0][i],
				Metadata: queryResp.Metadatas[0][i],
				Distance: queryResp.Distances[0][i],
			}
			results = append(results, result)

			// Record chunk access and cache embedding if enabled
			if s.chunkCache != nil {
				s.chunkCache.RecordAccess(result.ID)
			}
		}
	}

	return results, nil
}

func (s *ChromaService) DeleteDocumentChunks(fileID string) error {
	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	where := map[string]interface{}{
		"file_id": fileID,
	}

	deleteReq := map[string]interface{}{
		"where": where,
	}

	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/delete",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete chunks from chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma delete api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Remove from file router if enabled
	if s.fileRouter != nil && s.config.EnableFileRouting {
		s.fileRouter.RemoveFileIndex(fileID)
	}

	return nil
}

// GetCacheStats returns chunk cache statistics if cache is enabled
func (s *ChromaService) GetCacheStats() *cache.CacheStats {
	if s.chunkCache != nil {
		stats := s.chunkCache.Stats()
		return &stats
	}
	return nil
}

// GetRouterStats returns file router statistics if routing is enabled
func (s *ChromaService) GetRouterStats() map[string]interface{} {
	if s.fileRouter != nil {
		return s.fileRouter.GetStats()
	}
	return nil
}

// KeywordSearch performs a keyword-based search
// Optimizes by using FileRouter in-memory index if available
func (s *ChromaService) KeywordSearch(keywords []string, fileID string, topK int) ([]ChunkResult, error) {
	// 1. Try FileRouter first (In-Memory Speed)
	if s.fileRouter != nil && s.config.EnableFileRouting {
		if index, exists := s.fileRouter.GetFileIndex(fileID); exists {
			log.Printf("Using file router for fast in-memory keyword search (file: %s)", fileID)
			return s.searchInMemory(index.Chunks, keywords, topK)
		}
	}

	// 2. Fallback to Chroma (Network Call)
	return s.keywordSearchChroma(keywords, fileID, topK)
}

// searchInMemory performs keyword search on in-memory chunks
func (s *ChromaService) searchInMemory(chunks []retrieval.ChunkEmbedding, keywords []string, topK int) ([]ChunkResult, error) {
	type scoredChunk struct {
		chunk retrieval.ChunkEmbedding
		score int
	}
	var scoredChunks []scoredChunk

	for _, chunk := range chunks {
		text := strings.ToLower(chunk.Text)
		score := 0
		matchedKeywords := 0

		for _, keyword := range keywords {
			lowerKeyword := strings.ToLower(keyword)
			if strings.Contains(text, lowerKeyword) {
				score += 10
				matchedKeywords++
			}
		}

		// Check metadata
		if keyTermsStr, ok := chunk.Metadata["key_terms"].(string); ok {
			keyTermsLower := strings.ToLower(keyTermsStr)
			for _, keyword := range keywords {
				lowerKeyword := strings.ToLower(keyword)
				if strings.Contains(keyTermsLower, lowerKeyword) {
					score += 5
					if !strings.Contains(text, lowerKeyword) {
						matchedKeywords++
					}
				}
			}
		}

		if matchedKeywords > 0 {
			scoredChunks = append(scoredChunks, scoredChunk{chunk: chunk, score: score})
		}
	}

	// Sort by score
	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].score > scoredChunks[j].score
	})

	var results []ChunkResult
	for i := 0; i < len(scoredChunks) && i < topK; i++ {
		results = append(results, ChunkResult{
			ID:       scoredChunks[i].chunk.ChunkID,
			Text:     scoredChunks[i].chunk.Text,
			Metadata: scoredChunks[i].chunk.Metadata,
			Distance: float64(1000 - scoredChunks[i].score), // Mock distance
		})
	}

	return results, nil
}

// keywordSearchChroma performs the actual Chroma database query for keywords
func (s *ChromaService) keywordSearchChroma(keywords []string, fileID string, topK int) ([]ChunkResult, error) {
	// Ensure collection exists
	if err := s.EnsureCollection(); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Get all chunks for this file using GET API (more reliable than zero embedding)
	where := map[string]interface{}{
		"file_id": fileID,
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/get",
		s.baseURL, s.tenant, s.database, s.collectionID)

	// Build request to get all chunks for this file
	getReq := map[string]interface{}{
		"where":   where,
		"include": []string{"documents", "metadatas"},
	}

	jsonData, err := json.Marshal(getReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal get request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chroma get api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read raw response to handle different formats
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read get response: %w", err)
	}

	// Try to decode as flexible format
	var rawResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to decode get response: %w", err)
	}

	// Extract data - handle both []string and [][]string formats
	var ids []string
	var documents []string
	var metadatas []map[string]interface{}

	// Handle IDs
	if idsRaw, ok := rawResp["ids"].(interface{}); ok {
		idsBytes, _ := json.Marshal(idsRaw)
		// Try [][]string first
		var ids2D [][]string
		if err := json.Unmarshal(idsBytes, &ids2D); err == nil && len(ids2D) > 0 {
			ids = ids2D[0]
		} else {
			// Try []string
			json.Unmarshal(idsBytes, &ids)
		}
	}

	// Handle Documents
	if docsRaw, ok := rawResp["documents"].(interface{}); ok {
		docsBytes, _ := json.Marshal(docsRaw)
		// Try [][]string first
		var docs2D [][]string
		if err := json.Unmarshal(docsBytes, &docs2D); err == nil && len(docs2D) > 0 {
			documents = docs2D[0]
		} else {
			// Try []string
			json.Unmarshal(docsBytes, &documents)
		}
	}

	// Handle Metadatas
	if metasRaw, ok := rawResp["metadatas"].(interface{}); ok {
		metasBytes, _ := json.Marshal(metasRaw)
		// Try [][]map first
		var metas2D [][]map[string]interface{}
		if err := json.Unmarshal(metasBytes, &metas2D); err == nil && len(metas2D) > 0 {
			metadatas = metas2D[0]
		} else {
			// Try []map
			json.Unmarshal(metasBytes, &metadatas)
		}
	}

	// Score chunks based on keyword matches (in text and metadata key_terms)
	type scoredChunk struct {
		result ChunkResult
		score  int
	}
	var scoredChunks []scoredChunk

	if len(documents) > 0 && len(ids) == len(documents) && len(metadatas) == len(documents) {
		for i := 0; i < len(documents); i++ {
			text := strings.ToLower(documents[i])
			metadata := metadatas[i]

			// Calculate keyword match score
			score := 0
			matchedKeywords := 0

			// Check text content
			for _, keyword := range keywords {
				lowerKeyword := strings.ToLower(keyword)
				if strings.Contains(text, lowerKeyword) {
					score += 10 // Higher weight for text matches
					matchedKeywords++
				}
			}

			// Check metadata key_terms (if available)
			if keyTermsStr, ok := metadata["key_terms"].(string); ok {
				keyTermsLower := strings.ToLower(keyTermsStr)
				for _, keyword := range keywords {
					lowerKeyword := strings.ToLower(keyword)
					if strings.Contains(keyTermsLower, lowerKeyword) {
						score += 5 // Lower weight for metadata matches
						if !strings.Contains(text, lowerKeyword) {
							matchedKeywords++
						}
					}
				}
			}

			// Only include chunks that match at least one keyword
			if matchedKeywords > 0 {
				scoredChunks = append(scoredChunks, scoredChunk{
					result: ChunkResult{
						ID:       ids[i],
						Text:     documents[i],
						Metadata: metadata,
						Distance: float64(1000 - score), // Convert score to distance (lower is better)
					},
					score: score,
				})
			}
		}
	}

	// Sort by score (descending) and take topK
	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].score > scoredChunks[j].score
	})

	var results []ChunkResult
	for i := 0; i < len(scoredChunks) && i < topK; i++ {
		results = append(results, scoredChunks[i].result)
	}

	return results, nil
}

// HybridSearch performs both semantic and keyword search, then merges results
func (s *ChromaService) HybridSearch(queryEmbedding []float64, keywords []string, fileID string, topK int) ([]ChunkResult, error) {
	// Perform semantic search
	semanticResults, err := s.QuerySimilar(queryEmbedding, fileID, topK)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Perform keyword search
	keywordResults, err := s.KeywordSearch(keywords, fileID, topK/2)
	if err != nil {
		log.Printf("Warning: keyword search failed: %v", err)
		// Continue with just semantic results
		return semanticResults, nil
	}

	// Merge results using Reciprocal Rank Fusion (RRF)
	return s.ReciprocalRankFusion(semanticResults, keywordResults, topK)
}

// ReciprocalRankFusion merges two result sets using RRF algorithm
// score = 1 / (k + rank)
func (s *ChromaService) ReciprocalRankFusion(semantic, keyword []ChunkResult, topK int) ([]ChunkResult, error) {
	const k = 60.0 // Standard RRF constant
	scores := make(map[string]float64)
	chunkMap := make(map[string]ChunkResult)

	// Process semantic results
	for rank, result := range semantic {
		scores[result.ID] += 1.0 / (k + float64(rank+1))
		chunkMap[result.ID] = result
	}

	// Process keyword results
	for rank, result := range keyword {
		scores[result.ID] += 1.0 / (k + float64(rank+1))
		// If not in semantic, add to map
		if _, exists := chunkMap[result.ID]; !exists {
			chunkMap[result.ID] = result
		}
	}

	// Convert to slice
	var merged []ChunkResult
	for id, score := range scores {
		result := chunkMap[id]
		// Store RRF score in Distance field (inverted, so lower is better for sorting?)
		// Wait, our system expects Distance where lower is better.
		// RRF score: higher is better.
		// So we can store 1.0 - normalized_score or just 1/score.
		// Let's use 1/score as distance proxy.
		result.Distance = 1.0 / score
		merged = append(merged, result)
	}

	// Sort by Distance (ascending) -> effectively sorting by RRF score (descending)
	sortResultsByDistance(merged)

	// Limit
	if len(merged) > topK {
		merged = merged[:topK]
	}

	return merged, nil
}

// sortResultsByDistance sorts chunk results by distance (ascending)
func sortResultsByDistance(results []ChunkResult) {
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Distance > results[j+1].Distance {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}
