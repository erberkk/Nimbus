package retrieval

import (
	"log"
	"math"
	"sync"
	"time"
)

// FileRouter maintains in-memory indexes for fast file-level retrieval
type FileRouter struct {
	indexes    map[string]*FileIndex
	mutex      sync.RWMutex
	maxAge     time.Duration // TTL for unused indexes
	lastAccess map[string]time.Time
}

// FileIndex holds chunk embeddings for a single file
type FileIndex struct {
	FileID     string
	Chunks     []ChunkEmbedding
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ChunkCount int
}

// ChunkEmbedding represents a chunk with its embedding vector
type ChunkEmbedding struct {
	ChunkID   string
	Embedding []float64
	Text      string
	Metadata  map[string]interface{}
}

// NewFileRouter creates a new file router with TTL-based eviction
func NewFileRouter(maxAge time.Duration) *FileRouter {
	if maxAge == 0 {
		maxAge = 24 * time.Hour // Default 24 hours
	}
	
	router := &FileRouter{
		indexes:    make(map[string]*FileIndex),
		lastAccess: make(map[string]time.Time),
		maxAge:     maxAge,
	}
	
	// Start background cleanup goroutine
	go router.cleanupExpired()
	
	return router
}

// AddFileIndex adds or updates a file index
func (fr *FileRouter) AddFileIndex(fileID string, chunks []ChunkEmbedding) {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	now := time.Now()
	index := &FileIndex{
		FileID:     fileID,
		Chunks:     chunks,
		CreatedAt:  now,
		UpdatedAt:  now,
		ChunkCount: len(chunks),
	}
	
	fr.indexes[fileID] = index
	fr.lastAccess[fileID] = now
	
	log.Printf("File router: added index for file %s with %d chunks", fileID, len(chunks))
}

// GetFileIndex retrieves a file index
func (fr *FileRouter) GetFileIndex(fileID string) (*FileIndex, bool) {
	fr.mutex.RLock()
	defer fr.mutex.RUnlock()
	
	index, exists := fr.indexes[fileID]
	if exists {
		// Update access time (note: this is read-locked, so we'll update in a separate goroutine)
		go fr.updateAccessTime(fileID)
	}
	
	return index, exists
}

// RemoveFileIndex removes a file index
func (fr *FileRouter) RemoveFileIndex(fileID string) {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	delete(fr.indexes, fileID)
	delete(fr.lastAccess, fileID)
	
	log.Printf("File router: removed index for file %s", fileID)
}

// SearchInFile performs optimized similarity search within a specific file's index
// Uses parallel goroutines for fast dot product computation
func (fr *FileRouter) SearchInFile(fileID string, queryEmbedding []float64, topK int) []SimilarityResult {
	index, exists := fr.GetFileIndex(fileID)
	if !exists {
		return []SimilarityResult{}
	}
	
	// Compute similarities in parallel
	similarities := fr.computeSimilaritiesParallel(queryEmbedding, index.Chunks)
	
	// Sort by similarity (descending)
	sortedResults := fr.sortBySimilarity(similarities)
	
	// Return top-k
	if len(sortedResults) > topK {
		return sortedResults[:topK]
	}
	
	return sortedResults
}

// computeSimilaritiesParallel computes cosine similarity using goroutines
func (fr *FileRouter) computeSimilaritiesParallel(queryEmbedding []float64, chunks []ChunkEmbedding) []SimilarityResult {
	numChunks := len(chunks)
	results := make([]SimilarityResult, numChunks)
	
	// Determine optimal number of workers (use CPU cores, capped at 8)
	numWorkers := 4
	if numChunks < numWorkers {
		numWorkers = numChunks
	}
	
	// Channel for work distribution
	type workItem struct {
		index int
		chunk ChunkEmbedding
	}
	
	workChan := make(chan workItem, numChunks)
	var wg sync.WaitGroup
	
	// Worker function
	worker := func() {
		defer wg.Done()
		for item := range workChan {
			similarity := cosineSimilarity(queryEmbedding, item.chunk.Embedding)
			distance := 1.0 - similarity // Convert to distance
			
			results[item.index] = SimilarityResult{
				ChunkID:  item.chunk.ChunkID,
				Distance: distance,
				Text:     item.chunk.Text,
				Metadata: item.chunk.Metadata,
			}
		}
	}
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}
	
	// Distribute work
	for i, chunk := range chunks {
		workChan <- workItem{index: i, chunk: chunk}
	}
	close(workChan)
	
	// Wait for completion
	wg.Wait()
	
	return results
}

// sortBySimilarity sorts results by distance (ascending = most similar first)
func (fr *FileRouter) sortBySimilarity(results []SimilarityResult) []SimilarityResult {
	// Simple bubble sort for small lists (efficient enough for typical chunk counts)
	// For very large lists, could use sort.Slice, but this avoids the overhead
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Distance > results[j+1].Distance {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
	return results
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	
	var dotProduct, normA, normB float64
	
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)
	
	if normA == 0 || normB == 0 {
		return 0.0
	}
	
	return dotProduct / (normA * normB)
}

// updateAccessTime updates the last access time for a file (thread-safe)
func (fr *FileRouter) updateAccessTime(fileID string) {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	fr.lastAccess[fileID] = time.Now()
}

// cleanupExpired periodically removes expired indexes based on TTL
func (fr *FileRouter) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		fr.performCleanup()
	}
}

// performCleanup removes indexes that haven't been accessed within maxAge
func (fr *FileRouter) performCleanup() {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	now := time.Now()
	expiredFiles := []string{}
	
	for fileID, lastAccess := range fr.lastAccess {
		if now.Sub(lastAccess) > fr.maxAge {
			expiredFiles = append(expiredFiles, fileID)
		}
	}
	
	for _, fileID := range expiredFiles {
		delete(fr.indexes, fileID)
		delete(fr.lastAccess, fileID)
	}
	
	if len(expiredFiles) > 0 {
		log.Printf("File router cleanup: removed %d expired indexes", len(expiredFiles))
	}
}

// GetStats returns statistics about the file router
func (fr *FileRouter) GetStats() map[string]interface{} {
	fr.mutex.RLock()
	defer fr.mutex.RUnlock()
	
	totalChunks := 0
	for _, index := range fr.indexes {
		totalChunks += index.ChunkCount
	}
	
	return map[string]interface{}{
		"indexed_files": len(fr.indexes),
		"total_chunks":  totalChunks,
		"max_age_hours": fr.maxAge.Hours(),
	}
}

// WarmCache preloads indexes for frequently accessed files
func (fr *FileRouter) WarmCache(fileIDs []string, chunkProvider func(string) []ChunkEmbedding) {
	for _, fileID := range fileIDs {
		if _, exists := fr.GetFileIndex(fileID); !exists {
			chunks := chunkProvider(fileID)
			if len(chunks) > 0 {
				fr.AddFileIndex(fileID, chunks)
				log.Printf("File router: warmed cache for file %s", fileID)
			}
		}
	}
}

// Clear removes all indexes
func (fr *FileRouter) Clear() {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	fr.indexes = make(map[string]*FileIndex)
	fr.lastAccess = make(map[string]time.Time)
	
	log.Println("File router: cleared all indexes")
}

// SyncWithChroma syncs the file index with Chroma (called after document update)
// This ensures the in-memory index stays consistent with the vector DB
func (fr *FileRouter) SyncWithChroma(fileID string, chunks []ChunkEmbedding) {
	fr.mutex.Lock()
	defer fr.mutex.Unlock()
	
	if index, exists := fr.indexes[fileID]; exists {
		// Update existing index
		index.Chunks = chunks
		index.UpdatedAt = time.Now()
		index.ChunkCount = len(chunks)
		log.Printf("File router: synced index for file %s (updated)", fileID)
	} else {
		// Create new index
		now := time.Now()
		fr.indexes[fileID] = &FileIndex{
			FileID:     fileID,
			Chunks:     chunks,
			CreatedAt:  now,
			UpdatedAt:  now,
			ChunkCount: len(chunks),
		}
		fr.lastAccess[fileID] = now
		log.Printf("File router: synced index for file %s (created)", fileID)
	}
}

// GetIndexedFileIDs returns a list of all indexed file IDs
func (fr *FileRouter) GetIndexedFileIDs() []string {
	fr.mutex.RLock()
	defer fr.mutex.RUnlock()
	
	fileIDs := make([]string, 0, len(fr.indexes))
	for fileID := range fr.indexes {
		fileIDs = append(fileIDs, fileID)
	}
	
	return fileIDs
}

// HasIndex checks if a file has an index
func (fr *FileRouter) HasIndex(fileID string) bool {
	fr.mutex.RLock()
	defer fr.mutex.RUnlock()
	
	_, exists := fr.indexes[fileID]
	return exists
}

