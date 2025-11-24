package retrieval

import (
	"fmt"
	"log"
	"time"
)

// PerformanceMetrics tracks performance data for retrieval operations
type PerformanceMetrics struct {
	TotalQueries      int64
	TotalLatencyMs    int64
	MinLatencyMs      int64
	MaxLatencyMs      int64
	AvgLatencyMs      float64
	CacheHits         int64
	CacheMisses       int64
	AdaptiveTopKs     []int
	FixedTopKs        []int
	RetrievalTimes    []int64
	EmbeddingTimes    []int64
	GenerationTimes   []int64
}

// BenchmarkResult represents the result of a benchmark run
type BenchmarkResult struct {
	Name             string
	TotalDuration    time.Duration
	AvgLatency       time.Duration
	MinLatency       time.Duration
	MaxLatency       time.Duration
	ThroughputQPS    float64
	CacheHitRate     float64
	Timestamp        time.Time
	AdditionalMetrics map[string]interface{}
}

// PerformanceBenchmark provides utilities for benchmarking retrieval performance
type PerformanceBenchmark struct {
	metrics PerformanceMetrics
	started time.Time
}

// NewPerformanceBenchmark creates a new benchmark instance
func NewPerformanceBenchmark() *PerformanceBenchmark {
	return &PerformanceBenchmark{
		started: time.Now(),
		metrics: PerformanceMetrics{
			MinLatencyMs:    1000000, // Set to high value initially
			AdaptiveTopKs:   []int{},
			FixedTopKs:      []int{},
			RetrievalTimes:  []int64{},
			EmbeddingTimes:  []int64{},
			GenerationTimes: []int64{},
		},
	}
}

// RecordQuery records a query execution with timing information
func (pb *PerformanceBenchmark) RecordQuery(latencyMs int64, cacheHit bool) {
	pb.metrics.TotalQueries++
	pb.metrics.TotalLatencyMs += latencyMs
	
	if latencyMs < pb.metrics.MinLatencyMs {
		pb.metrics.MinLatencyMs = latencyMs
	}
	
	if latencyMs > pb.metrics.MaxLatencyMs {
		pb.metrics.MaxLatencyMs = latencyMs
	}
	
	if cacheHit {
		pb.metrics.CacheHits++
	} else {
		pb.metrics.CacheMisses++
	}
	
	pb.metrics.AvgLatencyMs = float64(pb.metrics.TotalLatencyMs) / float64(pb.metrics.TotalQueries)
}

// RecordRetrievalBreakdown records detailed timing breakdown
func (pb *PerformanceBenchmark) RecordRetrievalBreakdown(embeddingMs, retrievalMs, generationMs int64) {
	pb.metrics.EmbeddingTimes = append(pb.metrics.EmbeddingTimes, embeddingMs)
	pb.metrics.RetrievalTimes = append(pb.metrics.RetrievalTimes, retrievalMs)
	pb.metrics.GenerationTimes = append(pb.metrics.GenerationTimes, generationMs)
}

// RecordAdaptiveTopK records an adaptive top-k value
func (pb *PerformanceBenchmark) RecordAdaptiveTopK(topK int) {
	pb.metrics.AdaptiveTopKs = append(pb.metrics.AdaptiveTopKs, topK)
}

// RecordFixedTopK records a fixed top-k value
func (pb *PerformanceBenchmark) RecordFixedTopK(topK int) {
	pb.metrics.FixedTopKs = append(pb.metrics.FixedTopKs, topK)
}

// GetMetrics returns current performance metrics
func (pb *PerformanceBenchmark) GetMetrics() PerformanceMetrics {
	return pb.metrics
}

// GetResult generates a benchmark result summary
func (pb *PerformanceBenchmark) GetResult() BenchmarkResult {
	totalDuration := time.Since(pb.started)
	
	var throughput float64
	if totalDuration.Seconds() > 0 {
		throughput = float64(pb.metrics.TotalQueries) / totalDuration.Seconds()
	}
	
	var cacheHitRate float64
	totalCacheAccess := pb.metrics.CacheHits + pb.metrics.CacheMisses
	if totalCacheAccess > 0 {
		cacheHitRate = float64(pb.metrics.CacheHits) / float64(totalCacheAccess) * 100
	}
	
	additionalMetrics := map[string]interface{}{
		"total_queries":    pb.metrics.TotalQueries,
		"cache_hits":       pb.metrics.CacheHits,
		"cache_misses":     pb.metrics.CacheMisses,
		"avg_embedding_ms": pb.getAvgInt64(pb.metrics.EmbeddingTimes),
		"avg_retrieval_ms": pb.getAvgInt64(pb.metrics.RetrievalTimes),
		"avg_generation_ms": pb.getAvgInt64(pb.metrics.GenerationTimes),
	}
	
	if len(pb.metrics.AdaptiveTopKs) > 0 {
		additionalMetrics["avg_adaptive_topk"] = pb.getAvgInt(pb.metrics.AdaptiveTopKs)
		additionalMetrics["adaptive_topk_variance"] = pb.getVarianceInt(pb.metrics.AdaptiveTopKs)
	}
	
	return BenchmarkResult{
		Name:              "Retrieval Performance",
		TotalDuration:     totalDuration,
		AvgLatency:        time.Duration(int64(pb.metrics.AvgLatencyMs)) * time.Millisecond,
		MinLatency:        time.Duration(pb.metrics.MinLatencyMs) * time.Millisecond,
		MaxLatency:        time.Duration(pb.metrics.MaxLatencyMs) * time.Millisecond,
		ThroughputQPS:     throughput,
		CacheHitRate:      cacheHitRate,
		Timestamp:         time.Now(),
		AdditionalMetrics: additionalMetrics,
	}
}

// PrintResult prints a formatted benchmark result
func (pb *PerformanceBenchmark) PrintResult() {
	result := pb.GetResult()
	
	log.Println("===============================================")
	log.Printf("Benchmark: %s", result.Name)
	log.Println("===============================================")
	log.Printf("Total Duration:    %v", result.TotalDuration)
	log.Printf("Total Queries:     %d", result.AdditionalMetrics["total_queries"])
	log.Printf("Throughput:        %.2f QPS", result.ThroughputQPS)
	log.Printf("Average Latency:   %v", result.AvgLatency)
	log.Printf("Min Latency:       %v", result.MinLatency)
	log.Printf("Max Latency:       %v", result.MaxLatency)
	log.Printf("Cache Hit Rate:    %.2f%%", result.CacheHitRate)
	log.Printf("Cache Hits:        %d", result.AdditionalMetrics["cache_hits"])
	log.Printf("Cache Misses:      %d", result.AdditionalMetrics["cache_misses"])
	
	if avgEmbedding, ok := result.AdditionalMetrics["avg_embedding_ms"].(float64); ok {
		log.Printf("Avg Embedding:     %.2f ms", avgEmbedding)
	}
	if avgRetrieval, ok := result.AdditionalMetrics["avg_retrieval_ms"].(float64); ok {
		log.Printf("Avg Retrieval:     %.2f ms", avgRetrieval)
	}
	if avgGeneration, ok := result.AdditionalMetrics["avg_generation_ms"].(float64); ok {
		log.Printf("Avg Generation:    %.2f ms", avgGeneration)
	}
	
	if avgTopK, ok := result.AdditionalMetrics["avg_adaptive_topk"].(float64); ok {
		log.Printf("Avg Adaptive Top-K: %.2f", avgTopK)
	}
	if variance, ok := result.AdditionalMetrics["adaptive_topk_variance"].(float64); ok {
		log.Printf("Top-K Variance:    %.2f", variance)
	}
	
	log.Println("===============================================")
}

// CompareStrategies compares fixed vs adaptive retrieval strategies
func CompareStrategies(fixedMetrics, adaptiveMetrics PerformanceMetrics) string {
	comparison := "\n==================== Strategy Comparison ====================\n"
	
	fixedAvg := fixedMetrics.AvgLatencyMs
	adaptiveAvg := adaptiveMetrics.AvgLatencyMs
	improvement := ((fixedAvg - adaptiveAvg) / fixedAvg) * 100
	
	comparison += fmt.Sprintf("Fixed Strategy Avg Latency:    %.2f ms\n", fixedAvg)
	comparison += fmt.Sprintf("Adaptive Strategy Avg Latency: %.2f ms\n", adaptiveAvg)
	comparison += fmt.Sprintf("Performance Improvement:       %.2f%%\n", improvement)
	
	fixedCacheHitRate := float64(fixedMetrics.CacheHits) / float64(fixedMetrics.CacheHits+fixedMetrics.CacheMisses) * 100
	adaptiveCacheHitRate := float64(adaptiveMetrics.CacheHits) / float64(adaptiveMetrics.CacheHits+adaptiveMetrics.CacheMisses) * 100
	
	comparison += fmt.Sprintf("\nFixed Cache Hit Rate:          %.2f%%\n", fixedCacheHitRate)
	comparison += fmt.Sprintf("Adaptive Cache Hit Rate:       %.2f%%\n", adaptiveCacheHitRate)
	comparison += "=============================================================\n"
	
	return comparison
}

// getAvgInt64 calculates average of int64 slice
func (pb *PerformanceBenchmark) getAvgInt64(values []int64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	
	return float64(sum) / float64(len(values))
}

// getAvgInt calculates average of int slice
func (pb *PerformanceBenchmark) getAvgInt(values []int) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	sum := 0
	for _, v := range values {
		sum += v
	}
	
	return float64(sum) / float64(len(values))
}

// getVarianceInt calculates variance of int slice
func (pb *PerformanceBenchmark) getVarianceInt(values []int) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	avg := pb.getAvgInt(values)
	
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := float64(v) - avg
		sumSquaredDiff += diff * diff
	}
	
	return sumSquaredDiff / float64(len(values))
}

// RetrievalTimer is a helper for timing individual operations
type RetrievalTimer struct {
	startTime time.Time
	stages    map[string]time.Duration
}

// NewRetrievalTimer creates a new retrieval timer
func NewRetrievalTimer() *RetrievalTimer {
	return &RetrievalTimer{
		startTime: time.Now(),
		stages:    make(map[string]time.Duration),
	}
}

// Mark marks a stage completion
func (rt *RetrievalTimer) Mark(stageName string) {
	rt.stages[stageName] = time.Since(rt.startTime)
}

// GetStage returns duration for a specific stage
func (rt *RetrievalTimer) GetStage(stageName string) time.Duration {
	return rt.stages[stageName]
}

// GetTotal returns total elapsed time
func (rt *RetrievalTimer) GetTotal() time.Duration {
	return time.Since(rt.startTime)
}

// GetBreakdown returns a formatted breakdown of stage timings
func (rt *RetrievalTimer) GetBreakdown() string {
	breakdown := "\n--- Retrieval Breakdown ---\n"
	
	stageOrder := []string{"embedding", "retrieval", "generation"}
	for _, stage := range stageOrder {
		if duration, exists := rt.stages[stage]; exists {
			breakdown += fmt.Sprintf("  %s: %v\n", stage, duration)
		}
	}
	
	breakdown += fmt.Sprintf("  total: %v\n", rt.GetTotal())
	breakdown += "---------------------------\n"
	
	return breakdown
}

// LogPerformance logs performance metrics to console
func LogPerformance(operation string, duration time.Duration, additionalInfo map[string]interface{}) {
	log.Printf("[PERF] %s completed in %v", operation, duration)
	
	if len(additionalInfo) > 0 {
		for key, value := range additionalInfo {
			log.Printf("  - %s: %v", key, value)
		}
	}
}

