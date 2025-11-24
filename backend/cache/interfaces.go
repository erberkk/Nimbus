package cache

import "time"

// CachedQuery represents a cached query result with embedding and retrieved chunks
type CachedQuery struct {
	Embedding     []float64              `json:"embedding"`
	ChunkIDs      []string               `json:"chunk_ids"`
	RetrievedText []string               `json:"retrieved_text"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// QueryCache defines the interface for semantic query caching
// All implementations are in-memory for optimal performance.
type QueryCache interface {
	// Get retrieves a cached query by key
	Get(key string) (*CachedQuery, bool)

	// Set stores a query result with TTL
	Set(key string, value *CachedQuery, ttl time.Duration) error

	// Delete removes a cached query
	Delete(key string) error

	// Clear removes all cached queries
	Clear() error

	// Stats returns cache statistics
	Stats() CacheStats
}

// ChunkCache defines the interface for chunk popularity caching
// Stores frequently accessed chunk embeddings in memory
type ChunkCache interface {
	// Get retrieves a cached chunk embedding
	Get(chunkID string) ([]float64, bool)

	// Set stores a chunk embedding
	Set(chunkID string, embedding []float64)

	// RecordAccess tracks chunk access for popularity tracking
	RecordAccess(chunkID string)

	// GetPopularChunks returns the most frequently accessed chunks
	GetPopularChunks(limit int) []string

	// Delete removes a chunk from cache
	Delete(chunkID string)

	// Clear removes all cached chunks
	Clear()

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats provides metrics about cache performance
type CacheStats struct {
	Hits        int64      `json:"hits"`
	Misses      int64      `json:"misses"`
	Size        int        `json:"size"`
	MaxSize     int        `json:"max_size"`
	HitRate     float64    `json:"hit_rate"`
	Evictions   int64      `json:"evictions,omitempty"`
	LastCleared *time.Time `json:"last_cleared,omitempty"`
}

// ComputeHitRate calculates the cache hit rate
func (s *CacheStats) ComputeHitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}
