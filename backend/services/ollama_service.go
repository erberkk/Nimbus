package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nimbus-backend/config"
)

type OllamaService struct {
	baseURL    string
	embedModel string
	llmModel   string
	httpClient *http.Client
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
	return &OllamaService{
		baseURL:    cfg.OllamaBaseURL,
		embedModel: cfg.OllamaEmbedModel,
		llmModel:   cfg.OllamaLLMModel,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 3 minutes for long-running operations
		},
	}
}

// GenerateEmbedding generates an embedding vector for the given text using Ollama
func (s *OllamaService) GenerateEmbedding(text string) ([]float64, error) {
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
	// Build secure, context-bound prompt with markdown formatting
	prompt := "You are a document analysis assistant. Your ONLY job is to answer questions based strictly on the provided document context below.\n\n"

	prompt += "CRITICAL RULES:\n"
	prompt += "1. ONLY answer questions that can be answered from the given context\n"
	prompt += "2. If the question is unrelated or cannot be answered from the context, respond EXACTLY:\n"
	prompt += "   \"Üzgünüm, bu soru dosya içeriğiyle ilgili değil. Lütfen belgede yer alan bilgilerle alakalı bir soru sorun.\"\n"
	prompt += "3. NEVER generate, suggest, execute, or help with:\n"
	prompt += "   - Database commands (SQL: DROP, DELETE, INSERT, UPDATE, etc.)\n"
	prompt += "   - System/shell commands (rm, del, format, shutdown, etc.)\n"
	prompt += "   - Code execution or malicious scripts\n"
	prompt += "   - Accessing credentials, passwords, or sensitive system data\n"
	prompt += "   - Any harmful, dangerous, or unethical content\n"
	prompt += "4. Your purpose is ONLY to: explain, summarize, analyze, and answer questions about the document content\n\n"

	prompt += "FORMATTING RULES (for valid answers only):\n"
	prompt += "- Use Markdown format\n"
	prompt += "- Use ## or ### for headers\n"
	prompt += "- Use - or numbers for lists\n"
	prompt += "- Use **bold** for emphasis\n"
	prompt += "- Use `backticks` for code or technical terms\n"
	prompt += "- Use blank lines between paragraphs\n"
	prompt += "- Present information in a clear, organized, and professional manner\n\n"

	prompt += "DOCUMENT CONTEXT:\n"
	prompt += "=================\n"
	for i, chunk := range contextChunks {
		prompt += fmt.Sprintf("--- Context Chunk %d ---\n%s\n\n", i+1, chunk)
	}

	prompt += "=================\n\n"
	prompt += fmt.Sprintf("USER QUESTION: %s\n\n", question)
	prompt += "YOUR ANSWER (in Markdown format, based ONLY on the context above):\n"

	return s.GenerateResponse(prompt)
}
