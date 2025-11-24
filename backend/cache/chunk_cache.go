package cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

// LRUChunkCache implements ChunkCache using LRU eviction policy
type LRUChunkCache struct {
	maxSize      int
	cache        map[string]*list.Element
	lruList      *list.List
	accessCount  map[string]int64
	stats        CacheStats
	mutex        sync.RWMutex
}

// cacheItem represents an item in the LRU cache
type cacheItem struct {
	key       string
	embedding []float64
	lastAccess time.Time
}

// NewLRUChunkCache creates a new LRU chunk cache with the specified maximum size
func NewLRUChunkCache(maxSize int) *LRUChunkCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default size
	}
	
	return &LRUChunkCache{
		maxSize:     maxSize,
		cache:       make(map[string]*list.Element),
		lruList:     list.New(),
		accessCount: make(map[string]int64),
		stats: CacheStats{
			MaxSize: maxSize,
		},
	}
}

// Get retrieves a cached chunk embedding
func (c *LRUChunkCache) Get(chunkID string) ([]float64, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	element, exists := c.cache[chunkID]
	if !exists {
		c.stats.Misses++
		return nil, false
	}
	
	// Move to front (most recently used)
	c.lruList.MoveToFront(element)
	
	item := element.Value.(*cacheItem)
	item.lastAccess = time.Now()
	
	c.stats.Hits++
	
	return item.embedding, true
}

// Set stores a chunk embedding
func (c *LRUChunkCache) Set(chunkID string, embedding []float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if already exists
	if element, exists := c.cache[chunkID]; exists {
		// Update existing entry
		c.lruList.MoveToFront(element)
		item := element.Value.(*cacheItem)
		item.embedding = embedding
		item.lastAccess = time.Now()
		return
	}
	
	// Add new entry
	item := &cacheItem{
		key:       chunkID,
		embedding: embedding,
		lastAccess: time.Now(),
	}
	
	element := c.lruList.PushFront(item)
	c.cache[chunkID] = element
	c.stats.Size++
	
	// Evict if over capacity
	if c.stats.Size > c.maxSize {
		c.evictOldest()
	}
}

// RecordAccess tracks chunk access for popularity tracking
func (c *LRUChunkCache) RecordAccess(chunkID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.accessCount[chunkID]++
	
	// If chunk exists in cache, move to front
	if element, exists := c.cache[chunkID]; exists {
		c.lruList.MoveToFront(element)
		item := element.Value.(*cacheItem)
		item.lastAccess = time.Now()
	}
}

// GetPopularChunks returns the most frequently accessed chunks
func (c *LRUChunkCache) GetPopularChunks(limit int) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// Create slice of chunks with access counts
	type chunkAccess struct {
		chunkID string
		count   int64
	}
	
	chunks := make([]chunkAccess, 0, len(c.accessCount))
	for chunkID, count := range c.accessCount {
		chunks = append(chunks, chunkAccess{chunkID, count})
	}
	
	// Sort by access count (simple bubble sort for small lists)
	for i := 0; i < len(chunks); i++ {
		for j := i + 1; j < len(chunks); j++ {
			if chunks[j].count > chunks[i].count {
				chunks[i], chunks[j] = chunks[j], chunks[i]
			}
		}
	}
	
	// Return top N
	if limit > len(chunks) {
		limit = len(chunks)
	}
	
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = chunks[i].chunkID
	}
	
	return result
}

// Delete removes a chunk from cache
func (c *LRUChunkCache) Delete(chunkID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if element, exists := c.cache[chunkID]; exists {
		c.lruList.Remove(element)
		delete(c.cache, chunkID)
		delete(c.accessCount, chunkID)
		c.stats.Size--
	}
}

// Clear removes all cached chunks
func (c *LRUChunkCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.accessCount = make(map[string]int64)
	c.stats.Size = 0
	now := time.Now()
	c.stats.LastCleared = &now
	
	log.Println("Chunk cache cleared")
}

// Stats returns cache statistics
func (c *LRUChunkCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	stats := c.stats
	stats.HitRate = stats.ComputeHitRate()
	return stats
}

// evictOldest removes the least recently used item from cache
// Must be called with mutex locked
func (c *LRUChunkCache) evictOldest() {
	element := c.lruList.Back()
	if element == nil {
		return
	}
	
	item := element.Value.(*cacheItem)
	c.lruList.Remove(element)
	delete(c.cache, item.key)
	// Keep access count for popularity tracking
	c.stats.Size--
	c.stats.Evictions++
	
	log.Printf("Chunk cache: evicted chunk %s (LRU policy)", item.key)
}

// WarmCache preloads frequently accessed chunks into cache
func (c *LRUChunkCache) WarmCache(chunks map[string][]float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	warmed := 0
	for chunkID, embedding := range chunks {
		// Only warm cache if not already full or if this is a popular chunk
		if c.stats.Size < c.maxSize || c.accessCount[chunkID] > 0 {
			if _, exists := c.cache[chunkID]; !exists {
				item := &cacheItem{
					key:       chunkID,
					embedding: embedding,
					lastAccess: time.Now(),
				}
				element := c.lruList.PushFront(item)
				c.cache[chunkID] = element
				c.stats.Size++
				warmed++
				
				// Evict if necessary
				if c.stats.Size > c.maxSize {
					c.evictOldest()
				}
			}
		}
	}
	
	if warmed > 0 {
		log.Printf("Chunk cache: warmed %d chunks", warmed)
	}
}

// GetAccessCount returns the access count for a specific chunk
func (c *LRUChunkCache) GetAccessCount(chunkID string) int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.accessCount[chunkID]
}

// GetCacheKeys returns all chunk IDs currently in cache
func (c *LRUChunkCache) GetCacheKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	keys := make([]string, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	
	return keys
}

