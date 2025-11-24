package retrieval

import (
	"fmt"
	"math"

	"nimbus-backend/config"
)

// AdaptiveRetriever handles dynamic top-k computation based on similarity distribution
type AdaptiveRetriever struct {
	MinTopK       int     // Minimum chunks to return (default 3)
	MaxTopK       int     // Maximum chunks to return (default 12)
	HighSimilThreshold float64 // Threshold for high similarity (default 0.8)
	MedSimilThreshold  float64 // Threshold for medium similarity (default 0.5)
	MinSimilThreshold  float64 // Minimum similarity to include (default 0.3)
}

// NewAdaptiveRetriever creates a new adaptive retriever with settings from config
func NewAdaptiveRetriever(cfg *config.Config) *AdaptiveRetriever {
	return &AdaptiveRetriever{
		MinTopK:            cfg.MaxRAGChunks / 4, // Heuristic: min is 1/4 of max
		MaxTopK:            cfg.MaxRAGChunks,
		HighSimilThreshold: cfg.HighSimilThreshold,
		MedSimilThreshold:  cfg.MedSimilThreshold,
		MinSimilThreshold:  cfg.MinSimilThreshold,
	}
}

// ComputeAdaptiveTopK calculates the optimal number of chunks to retrieve
// based on the similarity distribution
func (a *AdaptiveRetriever) ComputeAdaptiveTopK(similarities []float64) int {
	if len(similarities) == 0 {
		return a.MinTopK
	}
	
	// Calculate statistics
	mean := computeMean(similarities)
	stddev := computeStdDev(similarities, mean)
	
	var topK int
	
	// Adaptive strategy based on mean similarity
	if mean >= a.HighSimilThreshold {
		// High confidence - fewer chunks needed
		topK = a.MinTopK + int(float64(a.MaxTopK-a.MinTopK)*0.3)
	} else if mean >= a.MedSimilThreshold {
		// Medium confidence - moderate number of chunks
		topK = a.MinTopK + int(float64(a.MaxTopK-a.MinTopK)*0.6)
	} else {
		// Low confidence - more chunks to get better coverage
		topK = a.MaxTopK
	}
	
	// Adjust based on standard deviation
	// High stddev means mixed relevance - get more chunks to be safe
	if stddev > 0.2 {
		topK = int(float64(topK) * 1.2)
	}
	
	// Ensure within bounds
	if topK < a.MinTopK {
		topK = a.MinTopK
	}
	if topK > a.MaxTopK {
		topK = a.MaxTopK
	}
	
	// Don't exceed available similarities
	if topK > len(similarities) {
		topK = len(similarities)
	}
	
	return topK
}

// FilterByThreshold filters chunks based on minimum similarity threshold
func (a *AdaptiveRetriever) FilterByThreshold(chunks []SimilarityResult) []SimilarityResult {
	var filtered []SimilarityResult
	
	for _, chunk := range chunks {
		// Convert distance to similarity (assuming cosine distance: similarity = 1 - distance)
		similarity := 1.0 - chunk.Distance
		
		if similarity >= a.MinSimilThreshold {
			filtered = append(filtered, chunk)
		}
	}
	
	return filtered
}

// SimilarityResult represents a chunk with its similarity score
type SimilarityResult struct {
	ChunkID  string
	Distance float64
	Text     string
	Metadata map[string]interface{}
}

// ComputeDynamicThreshold computes a dynamic threshold based on similarity distribution
// This is more sophisticated than a fixed threshold
func (a *AdaptiveRetriever) ComputeDynamicThreshold(similarities []float64) float64 {
	if len(similarities) == 0 {
		return a.MinSimilThreshold
	}
	
	mean := computeMean(similarities)
	stddev := computeStdDev(similarities, mean)
	
	// Threshold = mean - 1.5 * stddev
	// This captures most relevant chunks while filtering outliers
	threshold := mean - 1.5*stddev
	
	// Ensure it's not too low
	if threshold < a.MinSimilThreshold {
		threshold = a.MinSimilThreshold
	}
	
	// Ensure it's not too high (to avoid filtering everything)
	if threshold > 0.7 {
		threshold = 0.7
	}
	
	return threshold
}

// GetAdaptiveResults retrieves and filters results using adaptive strategy
func (a *AdaptiveRetriever) GetAdaptiveResults(allResults []SimilarityResult) []SimilarityResult {
	if len(allResults) == 0 {
		return []SimilarityResult{}
	}
	
	// Extract similarities (convert distances)
	similarities := make([]float64, len(allResults))
	for i, result := range allResults {
		similarities[i] = 1.0 - result.Distance
	}
	
	// Compute adaptive top-k
	topK := a.ComputeAdaptiveTopK(similarities)
	
	// Filter by threshold
	filtered := a.FilterByThreshold(allResults)
	
	// Take top-k from filtered results
	if len(filtered) > topK {
		return filtered[:topK]
	}
	
	return filtered
}

// ExplainDecision provides a human-readable explanation of why a certain top-k was chosen
func (a *AdaptiveRetriever) ExplainDecision(similarities []float64, topK int) string {
	if len(similarities) == 0 {
		return "No similarities provided"
	}
	
	mean := computeMean(similarities)
	stddev := computeStdDev(similarities, mean)
	
	explanation := fmt.Sprintf(
		"Selected top-k=%d based on mean similarity=%.3f, stddev=%.3f. ",
		topK, mean, stddev,
	)
	
	if mean >= a.HighSimilThreshold {
		explanation += "High confidence match - fewer chunks needed."
	} else if mean >= a.MedSimilThreshold {
		explanation += "Medium confidence - moderate number of chunks."
	} else {
		explanation += "Low confidence - retrieving more chunks for better coverage."
	}
	
	if stddev > 0.2 {
		explanation += " High variance detected, increased chunk count."
	}
	
	return explanation
}

// computeMean calculates the arithmetic mean of a slice of floats
func computeMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	
	return sum / float64(len(values))
}

// computeStdDev calculates the standard deviation
func computeStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	
	variance := sumSquaredDiff / float64(len(values))
	return math.Sqrt(variance)
}

// RetrievalStrategy defines different retrieval strategies
type RetrievalStrategy string

const (
	StrategyFixed    RetrievalStrategy = "fixed"    // Fixed top-k
	StrategyAdaptive RetrievalStrategy = "adaptive" // Adaptive based on similarity
	StrategyThreshold RetrievalStrategy = "threshold" // All above threshold
)

// RetrievalConfig defines configuration for retrieval
type RetrievalConfig struct {
	Strategy          RetrievalStrategy
	FixedTopK         int     // Used when strategy is "fixed"
	MinTopK           int     // Used when strategy is "adaptive"
	MaxTopK           int     // Used when strategy is "adaptive"
	SimilarityThreshold float64 // Used when strategy is "threshold"
}

// DefaultRetrievalConfig returns default configuration
func DefaultRetrievalConfig() RetrievalConfig {
	return RetrievalConfig{
		Strategy:            StrategyAdaptive,
		FixedTopK:           5,
		MinTopK:             3,
		MaxTopK:             12,
		SimilarityThreshold: 0.3,
	}
}

