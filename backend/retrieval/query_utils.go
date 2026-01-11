package retrieval

import (
	"regexp"
	"strings"
)

// KeyTermExtractor extracts important terms from a query
type KeyTermExtractor struct {
	stopWords map[string]bool
}

// NewKeyTermExtractor creates a new key term extractor
func NewKeyTermExtractor() *KeyTermExtractor {
	stopWords := map[string]bool{
		"the": true, "is": true, "are": true, "was": true, "were": true,
		"what": true, "whats": true, "what's": true, "how": true, "why": true, "when": true, "where": true,
		"who": true, "which": true, "a": true, "an": true, "and": true,
		"or": true, "but": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "of": true, "with": true, "by": true,
		"from": true, "up": true, "about": true, "into": true, "through": true,
		"during": true, "before": true, "after": true, "above": true, "below": true,
		"between": true, "under": true, "again": true, "further": true, "then": true,
		"once": true, "here": true, "there": true, "all": true, "both": true,
		"each": true, "few": true, "more": true, "most": true, "other": true,
		"some": true, "such": true, "no": true, "nor": true, "not": true,
		"only": true, "own": true, "same": true, "so": true, "than": true,
		"too": true, "very": true, "can": true, "will": true, "just": true,
		"difference": true, "differences": true, "compare": true, "comparison": true,
		"versus": true, "vs": true, "among": true,
		"you": true, "your": true, "yours": true, "this": true, "that": true, "these": true, "those": true,
		"me": true, "my": true, "mine": true, "we": true, "our": true, "ours": true,
		"file": true, "document": true, "text": true, "content": true,

		"nedir": true, "ne": true, "nasıl": true, "neden": true, "niçin": true,
		"nerede": true, "kim": true, "hangi": true, "bir": true, "ve": true,
		"veya": true, "ile": true, "için": true, "üzerinde": true, "altında": true,
		"arasında": true, "içinde": true, "dışında": true, "önce": true, "sonra": true,
		"bu": true, "şu": true, "o": true, "bunlar": true, "şunlar": true,
		"onlar": true, "ben": true, "sen": true, "biz": true, "siz": true,
		"fark": true, "farkı": true, "farklar": true, "karşılaştır": true,
		"karşılaştırma": true, "arasındaki": true,
	}

	return &KeyTermExtractor{
		stopWords: stopWords,
	}
}

// Extract extracts key terms from a query
func (e *KeyTermExtractor) Extract(query string) []string {
	normalized := strings.ToLower(query)
	normalized = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(normalized, " ")

	words := strings.Fields(normalized)

	var keyTerms []string
	seen := make(map[string]bool)

	for _, word := range words {
		word = strings.TrimSpace(word)

		isNumber := regexp.MustCompile(`^\d+$`).MatchString(word)

		if (len(word) < 3 && !isNumber) || e.stopWords[word] || seen[word] {
			continue
		}

		keyTerms = append(keyTerms, word)
		seen[word] = true
	}

	return keyTerms
}

// ExtractNamedTerms extracts specific named terms from comparison queries
// Example: "smishing vishing phishing farkı" → ["smishing", "vishing", "phishing"]
// Example: "wifi 5 6 7 comparison" → ["wifi", "5", "6", "7"]
func (e *KeyTermExtractor) ExtractNamedTerms(query string) []string {
	keyTerms := e.Extract(query)

	var namedTerms []string

	normalized := strings.ToLower(query)

	for _, term := range keyTerms {
		isNumber := regexp.MustCompile(`^\d+$`).MatchString(term)
		if len(term) >= 3 || isNumber {
			namedTerms = append(namedTerms, term)
		}
	}

	comparisonPattern := regexp.MustCompile(`([a-zA-Z0-9]+)\s+(and|ve|,)\s+([a-zA-Z0-9]+)`)
	matches := comparisonPattern.FindAllStringSubmatch(normalized, -1)

	for _, match := range matches {
		if len(match) > 3 {
			term1 := strings.TrimSpace(match[1])
			term2 := strings.TrimSpace(match[3])

			if len(term1) >= 3 && !e.stopWords[term1] {
				namedTerms = append(namedTerms, term1)
			}
			if len(term2) >= 3 && !e.stopWords[term2] {
				namedTerms = append(namedTerms, term2)
			}
		}
	}

	seen := make(map[string]bool)
	var uniqueTerms []string
	for _, term := range namedTerms {
		if !seen[term] {
			uniqueTerms = append(uniqueTerms, term)
			seen[term] = true
		}
	}

	return uniqueTerms
}

// ExpandQueryForComparison generates multiple query variations for comparison
func ExpandQueryForComparison(originalQuery string, terms []string) []string {
	var expandedQueries []string

	expandedQueries = append(expandedQueries, originalQuery)

	for _, term := range terms {
		expandedQueries = append(expandedQueries, "what is "+term)
		expandedQueries = append(expandedQueries, term+" definition")
		expandedQueries = append(expandedQueries, term)
	}

	if len(terms) >= 2 {
		for i := 0; i < len(terms)-1; i++ {
			for j := i + 1; j < len(terms); j++ {
				expandedQueries = append(expandedQueries,
					"difference between "+terms[i]+" and "+terms[j])
				expandedQueries = append(expandedQueries,
					terms[i]+" vs "+terms[j])
			}
		}
	}

	return expandedQueries
}

// ChunkDeduplicator removes duplicate chunks based on ID
type ChunkDeduplicator struct{}

// NewChunkDeduplicator creates a new chunk deduplicator
func NewChunkDeduplicator() *ChunkDeduplicator {
	return &ChunkDeduplicator{}
}

// Deduplicate removes duplicate chunks, keeping the one with best score (lowest distance)
func (d *ChunkDeduplicator) Deduplicate(chunks []ChunkResult) []ChunkResult {
	seen := make(map[string]*ChunkResult)

	for i := range chunks {
		chunk := &chunks[i]
		existing, exists := seen[chunk.ID]

		if !exists {
			seen[chunk.ID] = chunk
		} else {
			if chunk.Distance < existing.Distance {
				seen[chunk.ID] = chunk
			}
		}
	}

	var uniqueChunks []ChunkResult
	for _, chunk := range seen {
		uniqueChunks = append(uniqueChunks, *chunk)
	}

	return uniqueChunks
}

// ChunkResult represents a search result chunk
type ChunkResult struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
	Distance float64
	Score    float64
}

// RankAndMerge merges multiple chunk result sets and re-ranks them
func (d *ChunkDeduplicator) RankAndMerge(resultSets [][]ChunkResult, maxResults int) []ChunkResult {
	var allChunks []ChunkResult
	for _, resultSet := range resultSets {
		allChunks = append(allChunks, resultSet...)
	}

	uniqueChunks := d.Deduplicate(allChunks)

	sortChunksByDistance(uniqueChunks)

	if len(uniqueChunks) > maxResults {
		uniqueChunks = uniqueChunks[:maxResults]
	}

	return uniqueChunks
}

// sortChunksByDistance sorts chunks by distance (ascending)
func sortChunksByDistance(chunks []ChunkResult) {
	n := len(chunks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if chunks[j].Distance > chunks[j+1].Distance {
				chunks[j], chunks[j+1] = chunks[j+1], chunks[j]
			}
		}
	}
}
