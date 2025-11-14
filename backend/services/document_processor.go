package services

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"nimbus-backend/config"
)

type DocumentProcessor struct {
	ollamaService  *OllamaService
	chromaService  *ChromaService
	minioService   *MinIOService
	fileCollection *mongo.Collection
}

type Chunk struct {
	Index int
	Text  string
}

var DocumentProcessorInstance *DocumentProcessor

func NewDocumentProcessor(cfg *config.Config, minioService *MinIOService, fileCollection *mongo.Collection) *DocumentProcessor {
	return &DocumentProcessor{
		ollamaService:  NewOllamaService(cfg),
		chromaService:  NewChromaService(cfg),
		minioService:   minioService,
		fileCollection: fileCollection,
	}
}

func InitDocumentProcessor(cfg *config.Config, fileCollection *mongo.Collection) error {
	if MinioService == nil {
		return fmt.Errorf("MinIO service must be initialized first")
	}
	DocumentProcessorInstance = NewDocumentProcessor(cfg, MinioService, fileCollection)
	log.Println("✅ Document processor initialized")
	return nil
}

// ProcessDocumentAsync starts document processing in a goroutine
func (p *DocumentProcessor) ProcessDocumentAsync(fileID, minioPath, contentType string) {
	go func() {
		if err := p.processDocument(fileID, minioPath, contentType); err != nil {
			log.Printf("Error processing document %s: %v", fileID, err)
			// Update file status to failed
			p.updateFileStatus(fileID, "failed", err.Error(), 0)
		}
	}()
}

// processDocument is the main processing pipeline
func (p *DocumentProcessor) processDocument(fileID, minioPath, contentType string) error {
	log.Printf("Starting document processing for file %s", fileID)

	// Update status to processing
	if err := p.updateFileStatus(fileID, "processing", "", 0); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Step 1: Extract text from document
	text, err := p.extractText(minioPath, contentType)
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	if len(strings.TrimSpace(text)) == 0 {
		return fmt.Errorf("extracted text is empty")
	}

	log.Printf("Extracted %d characters from document %s", len(text), fileID)

	// Step 2: Chunk the text
	chunks := p.chunkText(text)
	log.Printf("Created %d chunks for document %s", len(chunks), fileID)

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks created from text")
	}

	// Step 3: Generate embeddings and store in Chroma
	var chromaChunks []ChunkData
	for _, chunk := range chunks {
		// Generate embedding for this chunk
		embedding, err := p.ollamaService.GenerateEmbedding(chunk.Text)
		if err != nil {
			log.Printf("Warning: failed to generate embedding for chunk %d: %v", chunk.Index, err)
			continue // Skip this chunk but continue with others
		}

		chromaChunk := ChunkData{
			ID:        fmt.Sprintf("%s_%d", fileID, chunk.Index),
			Embedding: embedding,
			Text:      chunk.Text,
			Metadata: map[string]interface{}{
				"file_id":     fileID,
				"chunk_index": chunk.Index,
				"timestamp":   time.Now().Unix(),
			},
		}
		chromaChunks = append(chromaChunks, chromaChunk)

		log.Printf("Generated embedding for chunk %d/%d", chunk.Index+1, len(chunks))
	}

	if len(chromaChunks) == 0 {
		return fmt.Errorf("failed to generate any embeddings")
	}

	// Step 4: Add to Chroma
	if err := p.chromaService.AddDocuments(chromaChunks); err != nil {
		return fmt.Errorf("failed to add documents to Chroma: %w", err)
	}

	log.Printf("Successfully added %d chunks to Chroma for document %s", len(chromaChunks), fileID)

	// Step 5: Update file status to completed
	if err := p.updateFileStatus(fileID, "completed", "", len(chromaChunks)); err != nil {
		return fmt.Errorf("failed to update status to completed: %w", err)
	}

	log.Printf("Document processing completed for file %s", fileID)
	return nil
}

// extractText extracts text from PDF or DOCX files
func (p *DocumentProcessor) extractText(minioPath, contentType string) (string, error) {
	// Download file from MinIO
	ctx := context.Background()
	reader, err := p.minioService.Client.GetObject(ctx, "user-files", minioPath, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get file from MinIO: %w", err)
	}
	defer reader.Close()

	// Read entire file into memory (for PDF library)
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	contentTypeLower := strings.ToLower(contentType)

	// Extract based on content type
	if strings.Contains(contentTypeLower, "pdf") {
		return p.extractTextFromPDF(fileBytes)
	} else if strings.Contains(contentTypeLower, "wordprocessingml") || strings.Contains(contentTypeLower, "msword") {
		return p.extractTextFromDOCX(fileBytes)
	}

	return "", fmt.Errorf("unsupported content type: %s", contentType)
}

// extractTextFromPDF extracts text from PDF bytes
func (p *DocumentProcessor) extractTextFromPDF(fileBytes []byte) (string, error) {
	reader, err := pdf.NewReader(strings.NewReader(string(fileBytes)), int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var textBuilder strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("Warning: failed to extract text from page %d: %v", i, err)
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	return textBuilder.String(), nil
}

// extractTextFromDOCX extracts text from DOCX bytes
func (p *DocumentProcessor) extractTextFromDOCX(fileBytes []byte) (string, error) {
	reader := strings.NewReader(string(fileBytes))
	zipReader, err := zip.NewReader(reader, int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}

	var documentXML string
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			xmlBytes, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read document.xml: %w", err)
			}
			documentXML = string(xmlBytes)
			break
		}
	}

	if documentXML == "" {
		return "", fmt.Errorf("document.xml not found in DOCX")
	}

	text := p.extractTextFromXML(documentXML)
	return text, nil
}

// extractTextFromXML extracts text content from XML by removing tags
func (p *DocumentProcessor) extractTextFromXML(xml string) string {
	textRegex := regexp.MustCompile(`<w:t[^>]*>([^<]+)</w:t>`)
	matches := textRegex.FindAllStringSubmatch(xml, -1)

	var textBuilder strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			textBuilder.WriteString(match[1])
			textBuilder.WriteString(" ")
		}
	}

	paragraphRegex := regexp.MustCompile(`</w:p>`)
	text := paragraphRegex.ReplaceAllString(textBuilder.String(), "\n\n")

	return text
}

// chunkText splits text into chunks with overlap
func (p *DocumentProcessor) chunkText(text string) []Chunk {
	const (
		targetTokens  = 1000
		overlapTokens = 100
		charsPerToken = 4
	)

	targetChars := targetTokens * charsPerToken
	overlapChars := overlapTokens * charsPerToken

	sentences := p.splitIntoSentences(text)

	var chunks []Chunk
	var currentChunk strings.Builder
	var currentChunkSentences []string
	chunkIndex := 0

	for _, sentence := range sentences {
		potentialLength := currentChunk.Len() + len(sentence)

		if potentialLength > targetChars && currentChunk.Len() > 0 {
			chunks = append(chunks, Chunk{
				Index: chunkIndex,
				Text:  strings.TrimSpace(currentChunk.String()),
			})
			chunkIndex++

			currentChunk.Reset()
			currentChunkSentences = []string{}

			overlapSize := 0
			for i := len(currentChunkSentences) - 1; i >= 0 && overlapSize < overlapChars; i-- {
				sentence := currentChunkSentences[i]
				overlapSize += len(sentence)
				currentChunk.WriteString(sentence)
				currentChunk.WriteString(" ")
			}
		}

		currentChunk.WriteString(sentence)
		currentChunk.WriteString(" ")
		currentChunkSentences = append(currentChunkSentences, sentence)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			Index: chunkIndex,
			Text:  strings.TrimSpace(currentChunk.String()),
		})
	}

	return chunks
}

// splitIntoSentences splits text into sentences
func (p *DocumentProcessor) splitIntoSentences(text string) []string {
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRegex.Split(text, -1)

	var result []string
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 0 {
			result = append(result, trimmed)
		}
	}

	return result
}

// updateFileStatus updates the processing status of a file
func (p *DocumentProcessor) updateFileStatus(fileID, status, errorMsg string, chunkCount int) error {
	objID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"processing_status": status,
			"chunk_count":       chunkCount,
		},
	}

	if status == "completed" {
		now := time.Now()
		update["$set"].(bson.M)["processed_at"] = &now
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["processing_error"] = errorMsg
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = p.fileCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update file status: %w", err)
	}

	return nil
}
