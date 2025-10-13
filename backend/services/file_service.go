package services

import (
	"context"
	"fmt"
	"nimbus-backend/database"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FileService struct{}

var FileServiceInstance = &FileService{}

// CreateFileRecord - Dosya metadata'sını MongoDB'ye kaydet
func (fs *FileService) CreateFileRecord(userID, filename string, size int64, contentType, minioPath string, folderID *string) (*models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Public link oluştur
	publicLink, err := helpers.GeneratePublicLink()
	if err != nil {
		return nil, fmt.Errorf("public link oluşturulamadı: %v", err)
	}

	file := &models.File{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		FolderID:    folderID,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		MinioPath:   minioPath,
		PublicLink:  publicLink,
		AccessList:  []models.AccessEntry{}, // Initialize empty access list
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = database.FileCollection.InsertOne(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("dosya kaydı oluşturulamadı: %v", err)
	}

	return file, nil
}

// GetUserFiles - Kullanıcının dosyalarını listele
func (fs *FileService) GetUserFiles(userID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := database.FileCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("dosyalar listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("dosyalar decode edilemedi: %v", err)
	}

	return files, nil
}

// GetFileByID - ID'ye göre dosya getir
func (fs *FileService) GetFileByID(fileID string) (*models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	var file models.File
	err = database.FileCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&file)
	if err != nil {
		return nil, fmt.Errorf("dosya bulunamadı: %v", err)
	}

	return &file, nil
}

// GetFileByMinioPath - MinIO path'e göre dosya getir
func (fs *FileService) GetFileByMinioPath(minioPath string) (*models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var file models.File
	err := database.FileCollection.FindOne(ctx, bson.M{"minio_path": minioPath}).Decode(&file)
	if err != nil {
		return nil, fmt.Errorf("dosya bulunamadı: %v", err)
	}

	return &file, nil
}

// DeleteFileRecord - Dosya kaydını sil
func (fs *FileService) DeleteFileRecord(fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	result, err := database.FileCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("dosya silinemedi: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}

// UpdateFileRecord - Dosya kaydını güncelle
func (fs *FileService) UpdateFileRecord(fileID string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	updates["updated_at"] = time.Now()

	result, err := database.FileCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("dosya güncellenemedi: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}
