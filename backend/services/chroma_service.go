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

type ChromaService struct {
	baseURL        string
	tenant         string
	database       string
	collectionName string
	collectionID   string // UUID of the collection
	httpClient     *http.Client
}

type ChunkData struct {
	ID        string                 `json:"id"`
	Embedding []float64              `json:"embedding"`
	Text      string                 `json:"text"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ChunkResult struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"document"`
	Metadata map[string]interface{} `json:"metadata"`
	Distance float64                `json:"distance"`
}

type AddDocumentsRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float64              `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
}

type QueryRequest struct {
	QueryEmbeddings [][]float64 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Where           interface{} `json:"where,omitempty"`
}

type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Documents [][]string                 `json:"documents"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Distances [][]float64                `json:"distances"`
}

type CollectionResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewChromaService(cfg *config.Config) *ChromaService {
	return &ChromaService{
		baseURL:        cfg.ChromaBaseURL,
		tenant:         cfg.ChromaTenant,
		database:       cfg.ChromaDatabase,
		collectionName: cfg.ChromaCollection,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// EnsureCollection gets or creates a collection and stores its UUID
func (s *ChromaService) EnsureCollection() error {
	if s.collectionID != "" {
		return nil // Already initialized
	}

	// Try to get existing collection
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s",
		s.baseURL, s.tenant, s.database, s.collectionName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create get collection request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	// If collection exists, get its ID
	if resp.StatusCode == http.StatusOK {
		var collResp CollectionResponse
		if err := json.NewDecoder(resp.Body).Decode(&collResp); err != nil {
			return fmt.Errorf("failed to decode collection response: %w", err)
		}
		s.collectionID = collResp.ID
		return nil
	}

	// If collection doesn't exist (404), create it
	if resp.StatusCode == http.StatusNotFound {
		createURL := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections",
			s.baseURL, s.tenant, s.database)

		createBody := map[string]interface{}{
			"name": s.collectionName,
		}

		jsonData, err := json.Marshal(createBody)
		if err != nil {
			return fmt.Errorf("failed to marshal create collection request: %w", err)
		}

		createReq, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create collection request: %w", err)
		}
		createReq.Header.Set("Content-Type", "application/json")

		createResp, err := s.httpClient.Do(createReq)
		if err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		defer createResp.Body.Close()

		if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(createResp.Body)
			return fmt.Errorf("failed to create collection, status %d: %s", createResp.StatusCode, string(bodyBytes))
		}

		var collResp CollectionResponse
		if err := json.NewDecoder(createResp.Body).Decode(&collResp); err != nil {
			return fmt.Errorf("failed to decode create collection response: %w", err)
		}
		s.collectionID = collResp.ID
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status when getting collection: %d, body: %s", resp.StatusCode, string(bodyBytes))
}

func (s *ChromaService) AddDocuments(chunks []ChunkData) error {
	if len(chunks) == 0 {
		return nil
	}

	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	ids := make([]string, len(chunks))
	embeddings := make([][]float64, len(chunks))
	documents := make([]string, len(chunks))
	metadatas := make([]map[string]interface{}, len(chunks))

	for i, chunk := range chunks {
		ids[i] = chunk.ID
		embeddings[i] = chunk.Embedding
		documents[i] = chunk.Text
		metadatas[i] = chunk.Metadata
	}

	reqBody := AddDocumentsRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal add documents request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/upsert",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create add documents request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add documents to Chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Chroma add documents API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (s *ChromaService) QuerySimilar(queryEmbedding []float64, fileID string, topK int) ([]ChunkResult, error) {
	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	where := map[string]interface{}{
		"file_id": fileID,
	}

	reqBody := QueryRequest{
		QueryEmbeddings: [][]float64{queryEmbedding},
		NResults:        topK,
		Where:           where,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Chroma query API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	var results []ChunkResult
	if len(queryResp.Documents) > 0 && len(queryResp.Documents[0]) > 0 {
		for i := 0; i < len(queryResp.Documents[0]); i++ {
			result := ChunkResult{
				ID:       queryResp.IDs[0][i],
				Text:     queryResp.Documents[0][i],
				Metadata: queryResp.Metadatas[0][i],
				Distance: queryResp.Distances[0][i],
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (s *ChromaService) DeleteDocumentChunks(fileID string) error {
	// Ensure collection exists and we have its UUID
	if err := s.EnsureCollection(); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	where := map[string]interface{}{
		"file_id": fileID,
	}

	deleteReq := map[string]interface{}{
		"where": where,
	}

	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/delete",
		s.baseURL, s.tenant, s.database, s.collectionID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete chunks from Chroma: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Chroma delete API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
