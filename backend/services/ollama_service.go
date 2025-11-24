package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"nimbus-backend/cache"
	"nimbus-backend/config"
)

type OllamaService struct {
	baseURL    string
	embedModel string
	llmModel   string
	httpClient *http.Client
	queryCache cache.QueryCache
	config     *config.Config
}

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllamaService(cfg *config.Config) *OllamaService {
	var queryCache cache.QueryCache

	// Initialize query cache if enabled
	if cfg.EnableQueryCache {
		ttl := time.Duration(cfg.QueryCacheTTL) * time.Minute
		queryCache = cache.NewInMemoryQueryCache(ttl)
		log.Printf("Query cache enabled with TTL: %v", ttl)
	}

	return &OllamaService{
		baseURL:    cfg.OllamaBaseURL,
		embedModel: cfg.OllamaEmbedModel,
		llmModel:   cfg.OllamaLLMModel,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 3 minutes for long-running operations
		},
		queryCache: queryCache,
		config:     cfg,
	}
}

// GenerateEmbedding generates an embedding vector for the given text using Ollama
// Includes caching support if query cache is enabled
func (s *OllamaService) GenerateEmbedding(text string) ([]float64, error) {
	// Check cache if enabled
	if s.queryCache != nil && s.config.EnableQueryCache {
		queryKey := cache.GenerateQueryKey(text)

		// Try exact cache hit
		if cachedQuery, found := s.queryCache.Get(queryKey); found {
			log.Printf("Cache hit for embedding (exact match)")
			return cachedQuery.Embedding, nil
		}
	}

	// Generate embedding from Ollama
	embedding, err := s.generateEmbeddingFromAPI(text)
	if err != nil {
		return nil, err
	}

	// Cache the result if caching is enabled
	if s.queryCache != nil && s.config.EnableQueryCache {
		queryKey := cache.GenerateQueryKey(text)
		cachedQuery := &cache.CachedQuery{
			Embedding: embedding,
			Timestamp: time.Now(),
		}

		ttl := time.Duration(s.config.QueryCacheTTL) * time.Minute
		if err := s.queryCache.Set(queryKey, cachedQuery, ttl); err != nil {
			log.Printf("Warning: failed to cache embedding: %v", err)
		}
	}

	return embedding, nil
}

// generateEmbeddingFromAPI calls the Ollama API to generate an embedding
func (s *OllamaService) generateEmbeddingFromAPI(text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Model:  s.embedModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama embedding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama embedding API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var embResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embResp.Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding from Ollama")
	}

	return embResp.Embedding, nil
}

// GetCacheStats returns query cache statistics if cache is enabled
func (s *OllamaService) GetCacheStats() *cache.CacheStats {
	if s.queryCache != nil {
		stats := s.queryCache.Stats()
		return &stats
	}
	return nil
}

// GenerateResponse generates a text response using the LLM model
func (s *OllamaService) GenerateResponse(prompt string) (string, error) {
	reqBody := GenerateRequest{
		Model:  s.llmModel,
		Prompt: prompt,
		Stream: false, // Non-streaming for simplicity
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generate request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create generate request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama generate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama generate API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to decode generate response: %w", err)
	}

	return genResp.Response, nil
}

// GenerateRAGResponse generates a response with context chunks (for RAG)
func (s *OllamaService) GenerateRAGResponse(question string, contextChunks []string) (string, error) {
	var sb strings.Builder
	qLower := strings.ToLower(question)

	// Detect intent categories
	isComparison := strings.Contains(qLower, "compare") ||
		strings.Contains(qLower, "comparison") ||
		strings.Contains(qLower, "vs") ||
		strings.Contains(qLower, "versus")

	isDefinition := strings.HasPrefix(qLower, "what is ") ||
		strings.HasPrefix(qLower, "what's ") ||
		strings.HasPrefix(qLower, "whats ") ||
		strings.HasPrefix(qLower, "define ") ||
		strings.Contains(qLower, "definition of") ||
		strings.Contains(qLower, "meaning of")

	isSummary := strings.Contains(qLower, "summarize") ||
		strings.Contains(qLower, "summary") ||
		strings.Contains(qLower, "overview") ||
		strings.Contains(qLower, "brief") ||
		strings.Contains(qLower, "outline")

	// Extract definition term once (optimized)
	defTerm := ""
	if isDefinition {
		defTerm = strings.TrimPrefix(qLower, "what is ")
		defTerm = strings.TrimPrefix(defTerm, "what's ")
		defTerm = strings.TrimPrefix(defTerm, "whats ")
		defTerm = strings.TrimPrefix(defTerm, "define ")
		defTerm = strings.TrimPrefix(defTerm, "the ")
		defTerm = strings.TrimSpace(defTerm)
	}

	// === CORE SYSTEM RULES ===
	sb.WriteString("You are a document analysis assistant. You ONLY answer questions using the document context below.\n\n")

	sb.WriteString("CRITICAL RULES:\n")
	sb.WriteString("1. Only answer using the provided context.\n")
	sb.WriteString("2. If the answer is NOT in the context, reply EXACTLY:\n")
	sb.WriteString("   \"Sorry, this question is not related to the document content. Please ask a question that is related to the document.\"\n")
	sb.WriteString("3. NEVER generate or assist with:\n")
	sb.WriteString("   - database/system/shell commands\n")
	sb.WriteString("   - harmful or unethical actions\n")
	sb.WriteString("   - code execution or scripts\n")
	sb.WriteString("4. Your ONLY job: summarize, analyze, explain, compare **based on the context**.\n\n")

	// ==== INTENT-SPECIFIC RULES ====
	if isComparison {
		sb.WriteString("SPECIAL MODE: COMPARISON QUERY\n")
		sb.WriteString("- Extract ALL features and ALL item values from any comparison-like text.\n")
		sb.WriteString("- Comparison tables may appear as bullet lists or inline patterns.\n")
		sb.WriteString("- Do NOT omit any features.\n")
		sb.WriteString("- Present the final result in a clean Markdown table.\n\n")
	}

	if isDefinition {
		sb.WriteString("SPECIAL MODE: DEFINITION QUERY\n")
		sb.WriteString(fmt.Sprintf("- The term you must define is: **%s**\n", defTerm))
		sb.WriteString("- Look for chunks where this term appears prominently.\n")
		sb.WriteString("- Extract the complete meaning, purpose, components, and key characteristics.\n\n")
	}

	if isSummary {
		sb.WriteString("SPECIAL MODE: SUMMARY QUERY\n")
		sb.WriteString("- Produce a structured, topic-based summary.\n")
		sb.WriteString("- Cover all major themes, lists, and important details.\n")
		sb.WriteString("- Use headers and bullet points.\n\n")
	}

	// Formatting rules
	sb.WriteString("FORMATTING RULES:\n")
	sb.WriteString("- Use Markdown.\n")
	sb.WriteString("- Use headings, lists, bold text.\n")
	sb.WriteString("- Keep paragraphs clean and well-structured.\n")
	sb.WriteString("- Final answer MUST be clean and professional.\n\n")

	// ==== DOCUMENT CONTEXT ====
	sb.WriteString("DOCUMENT CONTEXT:\n")
	sb.WriteString("====================================================\n\n")

	// Calculate available tokens for context
	// Reserve 1000 tokens for system prompt and answer
	maxContextTokens := s.config.ContextWindowSize - 1000
	currentTokens := 0

	for i, chunk := range contextChunks {
		chunkTokens := s.CountTokens(chunk)
		
		// Check if adding this chunk exceeds the limit
		if currentTokens + chunkTokens > maxContextTokens {
			log.Printf("Context limit reached (%d/%d tokens). Truncating remaining %d chunks.", 
				currentTokens, maxContextTokens, len(contextChunks)-i)
			break
		}

		chLower := strings.ToLower(chunk)
		header := fmt.Sprintf("--- Context Chunk %d ---", i+1)

		// Highlight comparison chunks
		if isComparison && (strings.Contains(chLower, "vs") ||
			strings.Contains(chLower, "versus") ||
			strings.Contains(chLower, "comparison")) {
			header = fmt.Sprintf("--- Context Chunk %d (comparison data) ---", i+1)
		}

		// Highlight definition chunks
		if isDefinition && defTerm != "" {
			if strings.HasPrefix(chLower, defTerm) ||
				strings.Contains(chLower, defTerm+" ") ||
				strings.Contains(chLower, defTerm+"\n") {
				header = fmt.Sprintf("--- Context Chunk %d (definition of %s) ---", i+1, defTerm)
			}
		}

		sb.WriteString(fmt.Sprintf("%s\n%s\n\n", header, chunk))
		currentTokens += chunkTokens
	}

	sb.WriteString("====================================================\n\n")
	sb.WriteString(fmt.Sprintf("USER QUESTION: %s\n\n", question))
	
	// Chain of Thought Prompting (Internal only - do not output reasoning)
	sb.WriteString("INSTRUCTIONS:\n")
	sb.WriteString("1. Analyze the user's question and the provided context.\n")
	sb.WriteString("2. Synthesize the information into a cohesive, natural answer.\n")
	sb.WriteString("3. Do NOT mention 'Chunk 1', 'Chunk 2', etc. in your final answer.\n")
	sb.WriteString("4. Do NOT output your internal reasoning or analysis steps.\n")
	sb.WriteString("5. If the context has conflicting info, mention the conflict naturally.\n")
	sb.WriteString("6. Provide a direct, professional response in Markdown.\n\n")
	
	sb.WriteString("YOUR ANSWER (Markdown only):\n")

	return s.GenerateResponse(sb.String())
}

// CountTokens estimates the number of tokens in a string
// Uses a simple heuristic (1 token ~= 4 chars) as we don't have a tokenizer
func (s *OllamaService) CountTokens(text string) int {
	return len(text) / 4
}
