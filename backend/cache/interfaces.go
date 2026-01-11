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
	Get(key string) (*CachedQuery, bool)
	Set(key string, value *CachedQuery, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	Stats() CacheStats
}

// ChunkCache defines the interface for chunk popularity caching
// Stores frequently accessed chunk embeddings in memory
type ChunkCache interface {
	Get(chunkID string) ([]float64, bool)
	Set(chunkID string, embedding []float64)
	RecordAccess(chunkID string)
	GetPopularChunks(limit int) []string
	Delete(chunkID string)
	Clear()
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
