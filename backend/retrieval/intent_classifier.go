package retrieval

import (
	"regexp"
	"strings"
)

// QueryIntent represents the type of query
type QueryIntent string

const (
	IntentSummary         QueryIntent = "summary"
	IntentTableOfContents QueryIntent = "toc"
	IntentDefinition      QueryIntent = "definition"
	IntentSpecific        QueryIntent = "specific"
	IntentComparison      QueryIntent = "comparison"
	IntentList            QueryIntent = "list"
)

// IntentClassifier classifies user queries based on heuristics
type IntentClassifier struct {
	patterns map[QueryIntent][]string
}

// NewIntentClassifier creates a new intent classifier
func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{
		patterns: map[QueryIntent][]string{
			IntentSummary: {
				"^(summarize|summary|overview|brief|outline|abstract)",
				"(can you|could you|please) .* (summarize|summary)",
				"(summarize|summary) .* (this|the) .* (file|document|text|content)",
				"(what|tell me) .* (about|of) .* (document|file|text|content)",
				"(give me|provide) .* (summary|overview)",
				"main (point|idea|theme|topic)",
			},
			IntentTableOfContents: {
				"table of content",
				"^(structure|organization|layout|sections|chapters)",
				"what .* (cover|contain|include)",
				"list .* (section|chapter|topic|part)",
			},
			IntentDefinition: {
				"^(what is|what's|whats|define|definition of|meaning of|explain)",
				"(what does|what do) .* mean",
				"(tell me|explain) .* (definition|meaning)",
			},
			IntentComparison: {
				"(compar[ieaos]*|difference|versus|vs\\.?)",
				"(how .* differ|what .* difference)",
				"(similar|similarity) .* (between|and)",
				"\\d+.*\\d+",
				"(between|among).*(and|or)",
			},
			IntentList: {
				"^list",
				"what are .* (all|the)",
				"(enumerate|mention) .* ",
				"give me .* list",
			},
		},
	}
}

// Classify determines the intent of a user query
func (ic *IntentClassifier) Classify(query string) QueryIntent {
	normalized := strings.ToLower(strings.TrimSpace(query))

	for intent, patterns := range ic.patterns {
		for _, pattern := range patterns {
			matched, err := regexp.MatchString(pattern, normalized)
			if err != nil {
				continue
			}
			if matched {
				return intent
			}
		}
	}

	return IntentSpecific
}

// GetRetrievalStrategy returns the recommended retrieval strategy for an intent
func (ic *IntentClassifier) GetRetrievalStrategy(intent QueryIntent) RetrievalStrategy {
	switch intent {
	case IntentSummary, IntentTableOfContents:
		return StrategyAdaptive
	case IntentDefinition, IntentSpecific:
		return StrategyAdaptive
	case IntentComparison, IntentList:
		return StrategyAdaptive
	default:
		return StrategyAdaptive
	}
}

// GetRecommendedTopK returns the recommended top-k for an intent
func (ic *IntentClassifier) GetRecommendedTopK(intent QueryIntent) int {
	switch intent {
	case IntentSummary:
		return 10
	case IntentTableOfContents:
		return 8
	case IntentDefinition:
		return 3
	case IntentSpecific:
		return 5
	case IntentComparison:
		return 8
	case IntentList:
		return 7
	default:
		return 5
	}
}

// ShouldBypassVectorSearch checks if the intent requires special handling
// that bypasses normal vector search
func (ic *IntentClassifier) ShouldBypassVectorSearch(intent QueryIntent) bool {
	return false
}

// GetSearchHints returns search hints based on intent
func (ic *IntentClassifier) GetSearchHints(intent QueryIntent, query string) map[string]interface{} {
	hints := make(map[string]interface{})
	hints["intent"] = string(intent)
	hints["strategy"] = string(ic.GetRetrievalStrategy(intent))
	hints["recommended_top_k"] = ic.GetRecommendedTopK(intent)

	if intent == IntentDefinition {
		definitionPatterns := []string{
			`what is ([a-zA-Z0-9\s]+)`,
			`define ([a-zA-Z0-9\s]+)`,
			`definition of ([a-zA-Z0-9\s]+)`,
			`meaning of ([a-zA-Z0-9\s]+)`,
		}

		normalized := strings.ToLower(query)
		for _, pattern := range definitionPatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(normalized)
			if len(matches) > 1 {
				term := strings.TrimSpace(matches[1])
				hints["term"] = term
				hints["use_keyword_boost"] = true
				break
			}
		}
	}

	if intent == IntentComparison {
		hints["boost_comparative_chunks"] = true
	}

	if intent == IntentList {
		hints["boost_list_chunks"] = true
	}

	return hints
}

// ExplainIntent provides a human-readable explanation of the detected intent
func (ic *IntentClassifier) ExplainIntent(intent QueryIntent) string {
	switch intent {
	case IntentSummary:
		return "Detected summary intent - will retrieve broad coverage of document"
	case IntentTableOfContents:
		return "Detected table of contents intent - will focus on structure and organization"
	case IntentDefinition:
		return "Detected definition intent - will retrieve precise, concise explanations"
	case IntentSpecific:
		return "Detected specific question intent - will retrieve targeted relevant chunks"
	case IntentComparison:
		return "Detected comparison intent - will retrieve multiple perspectives"
	case IntentList:
		return "Detected list intent - will retrieve comprehensive enumeration"
	default:
		return "Unknown intent - using default retrieval strategy"
	}
}

// IntentMetadata provides metadata about query intent
type IntentMetadata struct {
	Intent          QueryIntent
	Confidence      float64 // 0.0 to 1.0
	Strategy        RetrievalStrategy
	RecommendedTopK int
	SearchHints     map[string]interface{}
	Explanation     string
}

// AnalyzeQuery performs full intent analysis on a query
func (ic *IntentClassifier) AnalyzeQuery(query string) IntentMetadata {
	intent := ic.Classify(query)

	return IntentMetadata{
		Intent:          intent,
		Confidence:      0.8,
		Strategy:        ic.GetRetrievalStrategy(intent),
		RecommendedTopK: ic.GetRecommendedTopK(intent),
		SearchHints:     ic.GetSearchHints(intent, query),
		Explanation:     ic.ExplainIntent(intent),
	}
}
