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

	"nimbus-backend/chunks"
	"nimbus-backend/config"
	"nimbus-backend/retrieval"
)

type DocumentProcessor struct {
	ollamaService  *OllamaService
	chromaService  *ChromaService
	minioService   *MinIOService
	fileCollection *mongo.Collection
	config         *config.Config
	deduplicator   *chunks.FileDeduplicator
	textNormalizer *chunks.TextNormalizer
	textSplitter   *chunks.SemanticTextSplitter
}

var DocumentProcessorInstance *DocumentProcessor

func NewDocumentProcessor(cfg *config.Config, minioService *MinIOService, fileCollection *mongo.Collection) *DocumentProcessor {
	return &DocumentProcessor{
		ollamaService:  NewOllamaService(cfg),
		chromaService:  NewChromaService(cfg),
		minioService:   minioService,
		fileCollection: fileCollection,
		config:         cfg,
		deduplicator:   chunks.NewFileDeduplicator(fileCollection),
		textNormalizer: chunks.NewTextNormalizer(chunks.DefaultNormalizerConfig()),
		textSplitter:   chunks.NewSemanticTextSplitter(chunks.DefaultChunkerConfig()),
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
		if err := p.processDocument(fileID, minioPath, contentType, nil); err != nil {
			log.Printf("Error processing document %s: %v", fileID, err)
			// Update file status to failed
			p.updateFileStatus(fileID, "failed", err.Error(), 0)
		}
	}()
}

// ProcessDocumentWithDeduplication checks for duplicate file and processes if unique
func (p *DocumentProcessor) ProcessDocumentWithDeduplication(fileID, minioPath, contentType string, fileBytes []byte) {
	go func() {
		// Step 1: Check deduplication if enabled
		if p.config.EnableDeduplication && fileBytes != nil {
			fileHash := chunks.ComputeFileHash(fileBytes)

			// Store hash for this file
			if err := p.deduplicator.StoreFileHash(fileID, fileHash); err != nil {
				log.Printf("Warning: failed to store file hash: %v", err)
			}

			// Check for duplicates
			duplicate, err := p.deduplicator.CheckDuplicate(fileHash)
			if err != nil {
				log.Printf("Warning: deduplication check failed: %v", err)
				// Continue with normal processing
			} else if duplicate != nil {
				// Found duplicate - reuse embeddings
				log.Printf("Deduplication hit: file %s is duplicate of %s", fileID, duplicate.ExistingFileID)
				if err := p.deduplicator.LinkToExistingEmbeddings(fileID, duplicate.ExistingFileID, duplicate.ChunkCount); err != nil {
					log.Printf("Error linking to existing embeddings: %v", err)
					// Fall back to normal processing
				} else {
					log.Printf("Successfully reused embeddings for file %s", fileID)
					return // Done - no need to process
				}
			}
		}

		// Step 2: Process normally if not duplicate or deduplication disabled
		if err := p.processDocument(fileID, minioPath, contentType, fileBytes); err != nil {
			log.Printf("Error processing document %s: %v", fileID, err)
			p.updateFileStatus(fileID, "failed", err.Error(), 0)
		}
	}()
}

// processDocument is the main processing pipeline
func (p *DocumentProcessor) processDocument(fileID, minioPath, contentType string, fileBytes []byte) error {
	startTime := time.Now()
	log.Printf("Starting document processing for file %s", fileID)

	// Update status to processing
	if err := p.updateFileStatus(fileID, "processing", "", 0); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Step 1: Extract text from document
	var text string
	var err error
	if fileBytes != nil {
		// Use provided bytes (for deduplication flow)
		text, err = p.extractTextFromBytes(fileBytes, contentType)
	} else {
		// Download from MinIO
		text, err = p.extractText(minioPath, contentType)
	}
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	if len(strings.TrimSpace(text)) == 0 {
		return fmt.Errorf("extracted text is empty")
	}

	log.Printf("Extracted %d characters from document %s", len(text), fileID)

	// Step 2: Normalize Text
	log.Printf("Normalizing text for %s...", fileID)
	normalizer := chunks.NewTextNormalizer(chunks.DefaultNormalizerConfig())
	normalizedText := normalizer.Normalize(text)

	// Step 3: Process Tables
	log.Printf("Processing tables for %s...", fileID)
	tableProcessor := chunks.NewTableProcessor()
	segments := tableProcessor.Process(normalizedText)

	// Step 4: Split into Chunks
	log.Printf("Splitting text into chunks for %s...", fileID)
	splitter := chunks.NewSemanticTextSplitter(chunks.DefaultChunkerConfig())
	semanticChunks := splitter.SplitSegments(segments)

	log.Printf("Generated %d chunks for %s", len(semanticChunks), fileID)

	if len(semanticChunks) == 0 {
		return fmt.Errorf("no chunks created from text")
	}

	// Step 5: Generate embeddings and store in Chroma
	// Extract key terms for cross-referencing (optional optimization)
	termExtractor := retrieval.NewKeyTermExtractor()

	var chromaChunks []ChunkData
	for _, chunk := range semanticChunks {
		// Normalize for embedding (lighter normalization)
		embeddingText := chunks.NormalizeForEmbedding(chunk.Text)

		// Generate embedding for this chunk
		embedding, err := p.ollamaService.GenerateEmbedding(embeddingText)
		if err != nil {
			log.Printf("Warning: failed to generate embedding for chunk %d: %v", chunk.Index, err)
			continue // Skip this chunk but continue with others
		}

		// Extract key terms from chunk for cross-referencing
		keyTerms := termExtractor.Extract(chunk.Text)

		// Extract additional metadata from table structures
		tableMetadata := extractTableMetadata(chunk.Text)
		if len(tableMetadata) > 0 {
			// Add table-specific terms to key terms
			keyTerms = append(keyTerms, tableMetadata...)
			// Remove duplicates
			keyTerms = removeDuplicates(keyTerms)
		}

		// Detect if chunk contains technical terms or definitions
		chunkType := detectChunkType(chunk.Text, keyTerms)

		// Merge chunk metadata with our standard metadata
		metadata := map[string]interface{}{
			"file_id":     fileID,
			"chunk_index": chunk.Index,
			"timestamp":   time.Now().Unix(),
			"start_char":  chunk.StartChar,
			"end_char":    chunk.EndChar,
		}

		// Add cross-referencing metadata if we found key terms
		if len(keyTerms) > 0 {
			// Chroma only accepts scalar values in metadata (string, number, bool)
			// Convert array to comma-separated string
			metadata["key_terms"] = strings.Join(keyTerms, ",")
			metadata["term_count"] = len(keyTerms)
		}

		// Add chunk type for better retrieval
		metadata["chunk_type"] = chunkType

		// Add chunk metadata
		for k, v := range chunk.Metadata {
			metadata[k] = v
		}

		chromaChunk := ChunkData{
			ID:        fmt.Sprintf("%s_%d", fileID, chunk.Index),
			Embedding: embedding,
			Text:      chunk.Text,
			Metadata:  metadata,
		}
		chromaChunks = append(chromaChunks, chromaChunk)

		log.Printf("Generated embedding for chunk %d/%d (type: %s, terms: %d)",
			chunk.Index+1, len(semanticChunks), chunkType, len(keyTerms))
	}

	if len(chromaChunks) == 0 {
		return fmt.Errorf("failed to generate any embeddings")
	}

	// Step 5: Add to Chroma
	if err := p.chromaService.AddDocuments(chromaChunks); err != nil {
		return fmt.Errorf("failed to add documents to Chroma: %w", err)
	}

	log.Printf("Successfully added %d chunks to Chroma for document %s", len(chromaChunks), fileID)

	// Step 6: Update file status to completed
	if err := p.updateFileStatus(fileID, "completed", "", len(chromaChunks)); err != nil {
		return fmt.Errorf("failed to update status to completed: %w", err)
	}

	processingTime := time.Since(startTime)
	log.Printf("Document processing completed for file %s in %v", fileID, processingTime)
	return nil
}

// extractText extracts text from PDF or DOCX files
func (p *DocumentProcessor) extractText(minioPath string, contentType string) (string, error) {
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

	return p.extractTextFromBytes(fileBytes, contentType)
}

// extractTextFromBytes extracts text from file bytes
func (p *DocumentProcessor) extractTextFromBytes(fileBytes []byte, contentType string) (string, error) {
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

// detectChunkType identifies the type of chunk based on its content and key terms
func detectChunkType(text string, keyTerms []string) string {
	lowerText := strings.ToLower(text)

	// Check for comparison/table patterns FIRST (more specific than definitions)
	if strings.Contains(lowerText, "comparison") ||
		strings.Contains(lowerText, "difference") ||
		strings.Contains(lowerText, "versus") ||
		strings.Contains(lowerText, "vs") {
		return "comparison"
	}

	// Check for list patterns
	listPatterns := []string{
		`^\s*[\d•\-\*]`,                // Starts with bullet or number
		`\n\s*[\d•\-\*]`,               // Contains bullets
		`(first|second|third|finally)`, // Enumeration
	}
	for _, pattern := range listPatterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return "list"
		}
	}

	// Check for table-like structure (multiple columns/rows)
	lines := strings.Split(text, "\n")
	if len(lines) > 3 {
		tabCount := 0
		for _, line := range lines {
			if strings.Count(line, "\t") > 1 || strings.Count(line, "  ") > 3 {
				tabCount++
			}
		}
		if tabCount > 2 {
			return "table"
		}
	}

	// Check for high density of technical terms
	if len(keyTerms) > 5 && len(strings.Fields(text)) < 200 {
		return "technical"
	}

	// Default to narrative
	return "narrative"
}

// extractTableMetadata extracts key terms from restructured table content
// Specifically handles tables like "Comparison of WiFi 5-6-7"
func extractTableMetadata(text string) []string {
	var metadata []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract from table titles (e.g., "Comparison of WiFi 5-6-7")
		if strings.Contains(strings.ToLower(line), "comparison") {
			// Extract technology names with numbers (e.g., "WiFi 5", "WiFi 6")
			techPattern := regexp.MustCompile(`(?i)(wifi|wi-fi|802\.11[a-z]*)\s*(\d+|[a-z]{1,2})`)
			matches := techPattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					// Add both "wifi" and "wifi 5" variants
					tech := strings.ToLower(match[1])
					version := match[2]
					metadata = append(metadata, tech, tech+" "+version)
				}
			}

			// Add "comparison" as a key term
			metadata = append(metadata, "comparison")
		}

		// Extract from bullet points (e.g., "• Wi-Fi 5: 2013")
		if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") {
			// Extract technology names before colons
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				tech := strings.TrimSpace(strings.TrimLeft(parts[0], "•-"))
				tech = strings.ToLower(tech)
				// Clean and add
				tech = regexp.MustCompile(`[^\w\s-]`).ReplaceAllString(tech, "")
				if len(tech) > 2 {
					metadata = append(metadata, tech)
				}
			}
		}

		// Extract from row headers (e.g., "Release Year:", "Frequency Bands:")
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "•") {
			header := strings.ToLower(strings.TrimSuffix(line, ":"))
			words := strings.Fields(header)
			// Add individual significant words
			for _, word := range words {
				if len(word) >= 4 { // Skip short words
					metadata = append(metadata, word)
				}
			}
		}
	}

	return metadata
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
