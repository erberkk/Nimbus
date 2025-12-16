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
	"go.mongodb.org/mongo-driver/mongo"
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
		Ancestors:   []primitive.ObjectID{}, // Başlangıçta boş
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Eğer folderID varsa, parent klasörün ancestors'ını al ve kendi ID'mizi ekle
	if folderID != nil && *folderID != "" {
		// Önce bu klasörün bilgilerini al
		parentFolder, err := FolderServiceInstance.GetFolderByID(*folderID)
		if err == nil {
			// Parent'ın ancestors'ına kendi ID'mizi ekle
			file.Ancestors = append(parentFolder.Ancestors, parentFolder.ID)
			file.ParentID = &parentFolder.ID
		}
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

	filter := bson.M{
		"user_id":    userID,
		"deleted_at": nil,
	}
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

// GetRecentFiles - Son güncellenen dosyaları getir (silinmemiş)
func (fs *FileService) GetRecentFiles(userID string, limit int64) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"deleted_at": nil,
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := database.FileCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("son dosyalar listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("dosyalar decode edilemedi: %v", err)
	}

	return files, nil
}

// GetStarredFiles - Yıldızlı dosyaları getir (silinmemiş)
func (fs *FileService) GetStarredFiles(userID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_starred": true,
		"deleted_at": nil,
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})

	cursor, err := database.FileCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("yıldızlı dosyalar listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("dosyalar decode edilemedi: %v", err)
	}

	return files, nil
}

// GetTrashFiles - Çöp kutusundaki dosyaları getir
func (fs *FileService) GetTrashFiles(userID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation pipeline:
	// 1. Match deleted files for user
	// 2. Lookup parent folder
	// 3. Filter out files whose parent folder is also deleted
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "user_id", Value: userID},
			{Key: "deleted_at", Value: bson.D{{Key: "$ne", Value: nil}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "folder_id_obj", Value: bson.D{{Key: "$toObjectId", Value: "$folder_id"}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "folders"},
			{Key: "localField", Value: "folder_id_obj"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "parent_folder"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$parent_folder"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "parent_folder", Value: nil}},
				bson.D{{Key: "parent_folder.deleted_at", Value: nil}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "deleted_at", Value: -1}}}},
	}

	cursor, err := database.FileCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("çöp kutusu listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("dosyalar decode edilemedi: %v", err)
	}

	return files, nil
}

// ToggleFileStar - Dosya yıldız durumunu değiştir
func (fs *FileService) ToggleFileStar(fileID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return false, fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	// Dosyayı bul ve is_starred değerini al
	var file models.File
	err = database.FileCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&file)
	if err != nil {
		return false, fmt.Errorf("dosya bulunamadı: %v", err)
	}

	newStatus := !file.IsStarred

	update := bson.M{
		"$set": bson.M{
			"is_starred": newStatus,
			"updated_at": time.Now(),
		},
	}

	_, err = database.FileCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return false, fmt.Errorf("yıldız durumu güncellenemedi: %v", err)
	}

	return newStatus, nil
}

// SoftDeleteFile - Dosyayı çöp kutusuna taşı (Soft Delete)
func (fs *FileService) SoftDeleteFile(fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	result, err := database.FileCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("dosya silinemedi: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}

// RestoreFile - Dosyayı çöp kutusundan geri yükle
func (fs *FileService) RestoreFile(fileID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	// Dosyayı getir
	var file models.File
	err = database.FileCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&file)
	if err != nil {
		return fmt.Errorf("dosya bulunamadı: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": nil,
			"updated_at": time.Now(),
		},
	}

	// Eğer dosya bir klasöre bağlıysa, o klasörün parent hiyerarşisini kontrol et
	if file.FolderID != nil && *file.FolderID != "" {
		parentFolder, err := FolderServiceInstance.GetFolderByID(*file.FolderID)
		if err != nil || parentFolder == nil {
			// Parent klasör bulunamadı - dosya root'a taşın
			update["$set"].(bson.M)["folder_id"] = nil
			update["$set"].(bson.M)["parent_id"] = nil
			update["$set"].(bson.M)["ancestors"] = []primitive.ObjectID{}
		} else if parentFolder.DeletedAt != nil {
			// Parent klasör silinmiş - parent hiyerarşisini restore et
			if err := FolderServiceInstance.restoreParentHierarchy(ctx, parentFolder); err != nil {
				// Hata durumunda bile dosyayı restore et (parent'ı root'a taşı)
				update["$set"].(bson.M)["folder_id"] = nil
				update["$set"].(bson.M)["parent_id"] = nil
				update["$set"].(bson.M)["ancestors"] = []primitive.ObjectID{}
			}
			// Aksi takdirde dosya original parent'ında kalır (parent hiyerarşisi restore edildi)
		}
	}

	result, err := database.FileCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("dosya geri yüklenemedi: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}

// MoveFile - Dosyayı klasöre taşı
func (fs *FileService) MoveFile(fileID string, targetFolderID *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	// Yeni Ancestors ve ParentID hazırla
	var newAncestors []primitive.ObjectID
	var newParentID *primitive.ObjectID

	if targetFolderID != nil && *targetFolderID != "" {
		targetFolder, err := FolderServiceInstance.GetFolderByID(*targetFolderID)
		if err != nil {
			return fmt.Errorf("hedef klasör bulunamadı: %v", err)
		}
		newAncestors = append(targetFolder.Ancestors, targetFolder.ID)
		newParentID = &targetFolder.ID
	} else {
		newAncestors = []primitive.ObjectID{}
		newParentID = nil
	}

	update := bson.M{
		"$set": bson.M{
			"folder_id":  targetFolderID,
			"parent_id":  newParentID,
			"ancestors":  newAncestors,
			"updated_at": time.Now(),
		},
	}

	result, err := database.FileCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("dosya taşınamadı: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}

