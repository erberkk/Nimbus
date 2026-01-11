package retrieval

import (
	"fmt"
	"math"

	"nimbus-backend/config"
)

// AdaptiveRetriever handles dynamic top-k computation based on similarity distribution
type AdaptiveRetriever struct {
	MinTopK            int
	MaxTopK            int
	HighSimilThreshold float64
	MedSimilThreshold  float64
	MinSimilThreshold  float64
}

// NewAdaptiveRetriever creates a new adaptive retriever with settings from config
func NewAdaptiveRetriever(cfg *config.Config) *AdaptiveRetriever {
	return &AdaptiveRetriever{
		MinTopK:            cfg.MaxRAGChunks / 4,
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

	mean := computeMean(similarities)
	stddev := computeStdDev(similarities, mean)

	var topK int

	if mean >= a.HighSimilThreshold {
		topK = a.MinTopK + int(float64(a.MaxTopK-a.MinTopK)*0.3)
	} else if mean >= a.MedSimilThreshold {
		topK = a.MinTopK + int(float64(a.MaxTopK-a.MinTopK)*0.6)
	} else {
		topK = a.MaxTopK
	}

	if stddev > 0.2 {
		topK = int(float64(topK) * 1.2)
	}

	if topK < a.MinTopK {
		topK = a.MinTopK
	}
	if topK > a.MaxTopK {
		topK = a.MaxTopK
	}

	if topK > len(similarities) {
		topK = len(similarities)
	}

	return topK
}

// FilterByThreshold filters chunks based on minimum similarity threshold
func (a *AdaptiveRetriever) FilterByThreshold(chunks []SimilarityResult) []SimilarityResult {
	var filtered []SimilarityResult

	for _, chunk := range chunks {
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

	threshold := mean - 1.5*stddev

	if threshold < a.MinSimilThreshold {
		threshold = a.MinSimilThreshold
	}

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

	similarities := make([]float64, len(allResults))
	for i, result := range allResults {
		similarities[i] = 1.0 - result.Distance
	}

	topK := a.ComputeAdaptiveTopK(similarities)

	filtered := a.FilterByThreshold(allResults)

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
	StrategyFixed     RetrievalStrategy = "fixed"
	StrategyAdaptive  RetrievalStrategy = "adaptive"
	StrategyThreshold RetrievalStrategy = "threshold"
)

// RetrievalConfig defines configuration for retrieval
type RetrievalConfig struct {
	Strategy            RetrievalStrategy
	FixedTopK           int
	MinTopK             int
	MaxTopK             int
	SimilarityThreshold float64
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
