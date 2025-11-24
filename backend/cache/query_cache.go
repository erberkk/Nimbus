package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

// InMemoryQueryCache implements QueryCache using sync.Map with expiration tracking
type InMemoryQueryCache struct {
	cache      sync.Map
	stats      CacheStats
	statsMutex sync.RWMutex
	ttl        time.Duration
	
	// Background cleanup
	stopCleanup chan struct{}
	cleanupOnce sync.Once
}

// cacheEntry wraps CachedQuery with expiration time
type cacheEntry struct {
	value      *CachedQuery
	expiresAt  time.Time
}

// NewInMemoryQueryCache creates a new in-memory query cache with the specified TTL
func NewInMemoryQueryCache(ttl time.Duration) *InMemoryQueryCache {
	cache := &InMemoryQueryCache{
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
		stats: CacheStats{
			MaxSize: -1, // Unlimited for query cache
		},
	}
	
	// Start background cleanup goroutine
	go cache.cleanupExpired()
	
	return cache
}

// NormalizeQuery normalizes a query string for consistent caching
// Converts to lowercase, trims whitespace, and removes extra spaces
func NormalizeQuery(query string) string {
	// Convert to lowercase
	normalized := strings.ToLower(query)
	
	// Trim whitespace
	normalized = strings.TrimSpace(normalized)
	
	// Replace multiple spaces with single space
	normalized = strings.Join(strings.Fields(normalized), " ")
	
	return normalized
}

// GenerateQueryKey generates a SHA256 hash key for a normalized query
func GenerateQueryKey(query string) string {
	normalized := NormalizeQuery(query)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached query by key
func (c *InMemoryQueryCache) Get(key string) (*CachedQuery, bool) {
	value, exists := c.cache.Load(key)
	if !exists {
		c.recordMiss()
		return nil, false
	}
	
	entry := value.(*cacheEntry)
	
	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.cache.Delete(key)
		c.recordMiss()
		return nil, false
	}
	
	c.recordHit()
	return entry.value, true
}

// Set stores a query result with TTL
func (c *InMemoryQueryCache) Set(key string, value *CachedQuery, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}
	
	entry := &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	
	c.cache.Store(key, entry)
	c.incrementSize()
	
	return nil
}

// Delete removes a cached query
func (c *InMemoryQueryCache) Delete(key string) error {
	_, loaded := c.cache.LoadAndDelete(key)
	if loaded {
		c.decrementSize()
	}
	return nil
}

// Clear removes all cached queries
func (c *InMemoryQueryCache) Clear() error {
	c.cache = sync.Map{}
	
	c.statsMutex.Lock()
	c.stats.Size = 0
	now := time.Now()
	c.stats.LastCleared = &now
	c.statsMutex.Unlock()
	
	log.Println("Query cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *InMemoryQueryCache) Stats() CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	
	stats := c.stats
	stats.HitRate = stats.ComputeHitRate()
	return stats
}

// Close stops the background cleanup goroutine
func (c *InMemoryQueryCache) Close() {
	c.cleanupOnce.Do(func() {
		close(c.stopCleanup)
	})
}

// cleanupExpired periodically removes expired entries
func (c *InMemoryQueryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.performCleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// performCleanup removes all expired entries
func (c *InMemoryQueryCache) performCleanup() {
	now := time.Now()
	expiredKeys := []string{}
	
	c.cache.Range(func(key, value interface{}) bool {
		entry := value.(*cacheEntry)
		if now.After(entry.expiresAt) {
			expiredKeys = append(expiredKeys, key.(string))
		}
		return true
	})
	
	for _, key := range expiredKeys {
		c.cache.Delete(key)
		c.decrementSize()
	}
	
	if len(expiredKeys) > 0 {
		log.Printf("Query cache cleanup: removed %d expired entries", len(expiredKeys))
	}
}

// recordHit increments cache hit counter
func (c *InMemoryQueryCache) recordHit() {
	c.statsMutex.Lock()
	c.stats.Hits++
	c.statsMutex.Unlock()
}

// recordMiss increments cache miss counter
func (c *InMemoryQueryCache) recordMiss() {
	c.statsMutex.Lock()
	c.stats.Misses++
	c.statsMutex.Unlock()
}

// incrementSize increments cache size counter
func (c *InMemoryQueryCache) incrementSize() {
	c.statsMutex.Lock()
	c.stats.Size++
	c.statsMutex.Unlock()
}

// decrementSize decrements cache size counter
func (c *InMemoryQueryCache) decrementSize() {
	c.statsMutex.Lock()
	if c.stats.Size > 0 {
		c.stats.Size--
	}
	c.statsMutex.Unlock()
}

// CosineSimilarity computes the cosine similarity between two embedding vectors
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same length: %d vs %d", len(a), len(b))
	}
	
	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
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
		return 0, fmt.Errorf("vector has zero magnitude")
	}
	
	return dotProduct / (normA * normB), nil
}

// FindSimilarQuery checks if a similar query exists in cache
// Returns the cached query if similarity > threshold (default 0.95)
func (c *InMemoryQueryCache) FindSimilarQuery(embedding []float64, threshold float64) (*CachedQuery, string, bool) {
	if threshold == 0 {
		threshold = 0.95 // Default high threshold for semantic similarity
	}
	
	var bestMatch *CachedQuery
	var bestKey string
	var bestSimilarity float64
	
	now := time.Now()
	
	c.cache.Range(func(key, value interface{}) bool {
		entry := value.(*cacheEntry)
		
		// Skip expired entries
		if now.After(entry.expiresAt) {
			return true
		}
		
		// Compute similarity
		similarity, err := CosineSimilarity(embedding, entry.value.Embedding)
		if err != nil {
			return true
		}
		
		// Track best match
		if similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = entry.value
			bestKey = key.(string)
		}
		
		return true
	})
	
	if bestSimilarity >= threshold {
		c.recordHit()
		return bestMatch, bestKey, true
	}
	
	return nil, "", false
}

