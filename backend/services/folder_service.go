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

type FolderService struct{}

var FolderServiceInstance = &FolderService{}

// CreateFolder - Yeni klasör oluştur
func (fs *FolderService) CreateFolder(userID, name, color, folderID string) (*models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aynı isimde klasör var mı kontrol et (parent folder'a göre)
	var existingFolder models.Folder
	filter := bson.M{
		"user_id": userID,
		"name":    name,
	}

	// Eğer parent folder varsa, o klasör içinde aynı isimde klasör var mı kontrol et
	if folderID != "" {
		filter["folder_id"] = folderID
	} else {
		// Root seviyede kontrol
		filter["folder_id"] = bson.M{"$exists": false}
	}

	err := database.FolderCollection.FindOne(ctx, filter).Decode(&existingFolder)

	if err == nil {
		return nil, fmt.Errorf("bu isimde bir klasör zaten mevcut")
	}

	// Public link oluştur
	publicLink, err := helpers.GeneratePublicLink()
	if err != nil {
		return nil, fmt.Errorf("public link oluşturulamadı: %v", err)
	}

	folder := &models.Folder{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		Name:       name,
		Color:      color,
		PublicLink: publicLink,
		AccessList: []models.AccessEntry{}, // Initialize empty access list
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Eğer parent folder varsa, folder_id'yi set et ve ancestors'ı hesapla
	if folderID != "" {
		folder.FolderID = &folderID

		// Parent klasörün ancestors'ını al ve kendi ID'mizi ekle
		parentFolder, err := fs.GetFolderByID(folderID)
		if err == nil {
			// Parent'ın ancestors'ına kendi ID'mizi ekle
			folder.Ancestors = append(parentFolder.Ancestors, parentFolder.ID)
			folder.ParentID = &parentFolder.ID
		}
	} else {
		// Root klasör - ancestors boş
		folder.Ancestors = []primitive.ObjectID{}
	}

	_, err = database.FolderCollection.InsertOne(ctx, folder)
	if err != nil {
		return nil, fmt.Errorf("klasör oluşturulamadı: %v", err)
	}

	return folder, nil
}

// GetUserFolders - Kullanıcının tüm klasörlerini listele
func (fs *FolderService) GetUserFolders(userID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := database.FolderCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("klasörler listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("klasörler decode edilemedi: %v", err)
	}

	return folders, nil
}

// GetFolderByID - ID'ye göre klasör getir
func (fs *FolderService) GetFolderByID(folderID string) (*models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return nil, fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	var folder models.Folder
	err = database.FolderCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		return nil, fmt.Errorf("klasör bulunamadı: %v", err)
	}

	return &folder, nil
}

// GetFolderFiles - Klasördeki dosyaları getir
func (fs *FolderService) GetFolderFiles(folderID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"folder_id": folderID}
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

// GetRootFiles - Root'taki dosyaları getir (folder_id = null)
func (fs *FolderService) GetRootFiles(userID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":   userID,
		"folder_id": nil,
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

// GetFolderItemCount - Klasördeki toplam item sayısı (recursive - tüm alt öğeler)
func (fs *FolderService) GetFolderItemCount(folderID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return fs.getFolderItemCountRecursive(ctx, folderID)
}

// getFolderItemCountRecursive - Recursive olarak klasördeki tüm item sayısını hesapla
func (fs *FolderService) getFolderItemCountRecursive(ctx context.Context, folderID string) (int64, error) {
	var totalCount int64 = 0

	// Direkt dosya sayısını al
	fileCount, err := database.FileCollection.CountDocuments(ctx, bson.M{"folder_id": folderID})
	if err != nil {
		return 0, fmt.Errorf("dosya sayısı alınamadı: %v", err)
	}
	totalCount += fileCount

	// Alt klasörleri al
	cursor, err := database.FolderCollection.Find(ctx, bson.M{"folder_id": folderID})
	if err != nil {
		return 0, fmt.Errorf("alt klasörler alınamadı: %v", err)
	}
	defer cursor.Close(ctx)

	var subFolders []models.Folder
	if err := cursor.All(ctx, &subFolders); err != nil {
		return 0, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
	}

	// Her alt klasör için recursive olarak say
	for _, subFolder := range subFolders {
		subCount, err := fs.getFolderItemCountRecursive(ctx, subFolder.ID.Hex())
		if err != nil {
			return 0, err
		}
		totalCount += subCount + 1 // +1 alt klasörün kendisi için
	}

	return totalCount, nil
}

// GetUserStorageUsage - Kullanıcının toplam depolama kullanımını hesapla
func (fs *FolderService) GetUserStorageUsage(userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var totalSize int64 = 0

	// Kullanıcının sahibi olduğu dosyaları al
	ownedFiles, err := database.FileCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, fmt.Errorf("owned files alınamadı: %v", err)
	}
	defer ownedFiles.Close(ctx)

	for ownedFiles.Next(ctx) {
		var file models.File
		if err := ownedFiles.Decode(&file); err != nil {
			continue
		}
		totalSize += file.Size
	}

	// Kullanıcının access list'te bulunduğu dosyaları al
	accessFiles, err := database.FileCollection.Find(ctx, bson.M{
		"access_list.user_id": userID,
	})
	if err != nil {
		return 0, fmt.Errorf("access files alınamadı: %v", err)
	}
	defer accessFiles.Close(ctx)

	for accessFiles.Next(ctx) {
		var file models.File
		if err := accessFiles.Decode(&file); err != nil {
			continue
		}

		// Kullanıcının bu dosyaya erişimi var mı kontrol et
		for _, access := range file.AccessList {
			if access.UserID == userID {
				totalSize += file.Size
				break
			}
		}
	}

	return totalSize, nil
}

// UpdateFolder - Klasör güncelle
func (fs *FolderService) UpdateFolder(folderID string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	updates["updated_at"] = time.Now()

	result, err := database.FolderCollection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("klasör güncellenemedi: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("klasör bulunamadı")
	}

	return nil
}

// DeleteFolder - Klasör sil (boşsa)
func (fs *FolderService) DeleteFolder(folderID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Klasörde dosya var mı kontrol et
	count, err := fs.GetFolderItemCount(folderID)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("klasör boş değil, önce içindeki dosyaları silin")
	}

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	result, err := database.FolderCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("klasör silinemedi: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("klasör bulunamadı")
	}

	return nil
}

// MoveFileToFolder - Dosyayı başka klasöre taşı
func (fs *FolderService) MoveFileToFolder(fileID, newFolderID *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fileObjectID, err := primitive.ObjectIDFromHex(*fileID)
	if err != nil {
		return fmt.Errorf("geçersiz dosya ID'si: %v", err)
	}

	updates := bson.M{
		"folder_id":  newFolderID,
		"updated_at": time.Now(),
	}

	result, err := database.FileCollection.UpdateOne(
		ctx,
		bson.M{"_id": fileObjectID},
		bson.M{"$set": updates},
	)
	if err != nil {
		return fmt.Errorf("dosya taşınamadı: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("dosya bulunamadı")
	}

	return nil
}

// GetSubFolders - Belirtilen klasörün alt klasörlerini getir
func (fs *FolderService) GetSubFolders(parentFolderID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var folders []models.Folder

	// Eğer parentFolderID boşsa root klasörleri getir
	if parentFolderID == "" {
		cursor, err := database.FolderCollection.Find(ctx, bson.M{
			"folder_id": bson.M{"$exists": false},
		})
		if err != nil {
			return nil, fmt.Errorf("alt klasörler sorgulanamadı: %v", err)
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &folders); err != nil {
			return nil, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
		}
	} else {
		// Belirtilen klasörün alt klasörlerini getir
		cursor, err := database.FolderCollection.Find(ctx, bson.M{
			"folder_id": parentFolderID,
		})
		if err != nil {
			return nil, fmt.Errorf("alt klasörler sorgulanamadı: %v", err)
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &folders); err != nil {
			return nil, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
		}
	}

	return folders, nil
}
