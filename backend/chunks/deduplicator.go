package chunks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FileDeduplicator handles file hash-based deduplication
type FileDeduplicator struct {
	fileCollection *mongo.Collection
}

// DuplicateFileInfo contains information about a duplicate file
type DuplicateFileInfo struct {
	ExistingFileID string
	Hash           string
	OriginalName   string
	UploadedAt     time.Time
	ChunkCount     int
}

// NewFileDeduplicator creates a new file deduplicator
func NewFileDeduplicator(fileCollection *mongo.Collection) *FileDeduplicator {
	return &FileDeduplicator{
		fileCollection: fileCollection,
	}
}

// ComputeFileHash computes SHA256 hash of file bytes
func ComputeFileHash(fileBytes []byte) string {
	hash := sha256.Sum256(fileBytes)
	return hex.EncodeToString(hash[:])
}

// CheckDuplicate checks if a file with the same hash already exists
// Returns duplicate info if found, nil if unique
func (d *FileDeduplicator) CheckDuplicate(fileHash string) (*DuplicateFileInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Query for existing file with same hash
	var result struct {
		ID               primitive.ObjectID `bson:"_id"`
		Filename         string             `bson:"filename"`
		FileHash         string             `bson:"file_hash"`
		CreatedAt        time.Time          `bson:"created_at"`
		ChunkCount       int                `bson:"chunk_count"`
		ProcessingStatus string             `bson:"processing_status"`
	}
	
	filter := bson.M{
		"file_hash":         fileHash,
		"processing_status": "completed", // Only reuse successfully processed files
	}
	
	err := d.fileCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No duplicate found (this is normal)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check for duplicate: %w", err)
	}
	
	// Found a duplicate
	info := &DuplicateFileInfo{
		ExistingFileID: result.ID.Hex(),
		Hash:           result.FileHash,
		OriginalName:   result.Filename,
		UploadedAt:     result.CreatedAt,
		ChunkCount:     result.ChunkCount,
	}
	
	log.Printf("Deduplication: Found existing file %s with hash %s", info.ExistingFileID, fileHash[:16])
	
	return info, nil
}

// StoreFileHash stores the hash for a new file
func (d *FileDeduplicator) StoreFileHash(fileID, fileHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}
	
	update := bson.M{
		"$set": bson.M{
			"file_hash": fileHash,
		},
	}
	
	_, err = d.fileCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to store file hash: %w", err)
	}
	
	return nil
}

// LinkToExistingEmbeddings creates a reference to existing file's embeddings
// This allows a new file record to reuse embeddings without re-processing
func (d *FileDeduplicator) LinkToExistingEmbeddings(newFileID, sourceFileID string, chunkCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := primitive.ObjectIDFromHex(newFileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}
	
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"source_file_id":    sourceFileID, // Reference to original file
			"processing_status": "completed",   // Mark as processed
			"chunk_count":       chunkCount,
			"processed_at":      &now,
			"deduplication_hit": true, // Flag for tracking
		},
	}
	
	_, err = d.fileCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to link to existing embeddings: %w", err)
	}
	
	log.Printf("Deduplication: Linked file %s to existing embeddings from %s (%d chunks)", 
		newFileID, sourceFileID, chunkCount)
	
	return nil
}

// GetSourceFileID retrieves the source file ID if this file is a duplicate
func (d *FileDeduplicator) GetSourceFileID(fileID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	objID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return "", fmt.Errorf("invalid file ID: %w", err)
	}
	
	var result struct {
		SourceFileID string `bson:"source_file_id"`
	}
	
	err = d.fileCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to get source file ID: %w", err)
	}
	
	return result.SourceFileID, nil
}

// GetDeduplicationStats returns statistics about deduplication
func (d *FileDeduplicator) GetDeduplicationStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Count total files
	totalFiles, err := d.fileCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total files: %w", err)
	}
	
	// Count deduplicated files
	deduplicatedFiles, err := d.fileCollection.CountDocuments(ctx, bson.M{"deduplication_hit": true})
	if err != nil {
		return nil, fmt.Errorf("failed to count deduplicated files: %w", err)
	}
	
	// Calculate savings rate
	var savingsRate float64
	if totalFiles > 0 {
		savingsRate = float64(deduplicatedFiles) / float64(totalFiles) * 100
	}
	
	stats := map[string]interface{}{
		"total_files":        totalFiles,
		"deduplicated_files": deduplicatedFiles,
		"unique_files":       totalFiles - deduplicatedFiles,
		"savings_rate":       fmt.Sprintf("%.2f%%", savingsRate),
	}
	
	return stats, nil
}

