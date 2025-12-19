package services

import (
	"context"
	"fmt"
	"log"
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

	filter := bson.M{
		"user_id":    userID,
		"deleted_at": nil,
	}
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

	filter := bson.M{
		"folder_id":  folderID,
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

// GetRootFiles - Root'taki dosyaları getir (folder_id = null)
func (fs *FolderService) GetRootFiles(userID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"folder_id":  nil,
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

// GetFolderItemCount - Klasördeki toplam item sayısı (recursive - tüm alt öğeler)
func (fs *FolderService) GetFolderItemCount(folderID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return fs.getFolderItemCountRecursive(ctx, folderID)
}

// getFolderItemCountRecursive - Recursive olarak klasördeki tüm item sayısını hesapla
func (fs *FolderService) getFolderItemCountRecursive(ctx context.Context, folderID string) (int64, error) {
	var totalCount int64 = 0

	// Direkt dosya sayısını al (sadece silinmemiş dosyalar)
	fileCount, err := database.FileCollection.CountDocuments(ctx, bson.M{
		"folder_id":  folderID,
		"deleted_at": nil,
	})
	if err != nil {
		return 0, fmt.Errorf("dosya sayısı alınamadı: %v", err)
	}
	totalCount += fileCount

	// Alt klasörleri al (sadece silinmemiş klasörler)
	cursor, err := database.FolderCollection.Find(ctx, bson.M{
		"folder_id":  folderID,
		"deleted_at": nil,
	})
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

// GetFolderSize - Klasördeki toplam dosya boyutunu hesapla (recursive - tüm alt öğeler)
func (fs *FolderService) GetFolderSize(folderID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return fs.getFolderSizeRecursive(ctx, folderID)
}

// getFolderSizeRecursive - Recursive olarak klasördeki tüm dosyaların toplam boyutunu hesapla
func (fs *FolderService) getFolderSizeRecursive(ctx context.Context, folderID string) (int64, error) {
	var totalSize int64 = 0

	// Direkt dosyaların boyutunu al (sadece silinmemiş dosyalar)
	cursor, err := database.FileCollection.Find(ctx, bson.M{
		"folder_id":  folderID,
		"deleted_at": nil,
	})
	if err != nil {
		return 0, fmt.Errorf("dosyalar alınamadı: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var file models.File
		if err := cursor.Decode(&file); err != nil {
			continue
		}
		totalSize += file.Size
	}

	// Alt klasörleri al (sadece silinmemiş klasörler)
	folderCursor, err := database.FolderCollection.Find(ctx, bson.M{
		"folder_id":  folderID,
		"deleted_at": nil,
	})
	if err != nil {
		return 0, fmt.Errorf("alt klasörler alınamadı: %v", err)
	}
	defer folderCursor.Close(ctx)

	var subFolders []models.Folder
	if err := folderCursor.All(ctx, &subFolders); err != nil {
		return 0, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
	}

	// Her alt klasör için recursive olarak boyut hesapla
	for _, subFolder := range subFolders {
		subSize, err := fs.getFolderSizeRecursive(ctx, subFolder.ID.Hex())
		if err != nil {
			return 0, err
		}
		totalSize += subSize
	}

	return totalSize, nil
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
			"folder_id":  bson.M{"$exists": false},
			"deleted_at": nil,
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
			"folder_id":  parentFolderID,
			"deleted_at": nil,
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

// GetStarredSubFolders - Belirtilen klasörün star'lanmış alt klasörlerini getir
func (fs *FolderService) GetStarredSubFolders(parentFolderID string, userID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"folder_id":  parentFolderID,
		"user_id":    userID,
		"is_starred": true,
		"deleted_at": nil,
	}

	cursor, err := database.FolderCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("star'lanmış alt klasörler sorgulanamadı: %v", err)
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
	}

	return folders, nil
}

// GetStarredFolders - Yıldızlı klasörleri getir (silinmemiş, sadece root seviyedeki)
// Recursive star yapıldığı için sadece root seviyedeki (folder_id null) star'lanmış klasörleri döndürür
// Alt klasörler normal klasör navigasyonu ile gösterilir
func (fs *FolderService) GetStarredFolders(userID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"is_starred": true,
		"deleted_at": nil,
		"$or": []bson.M{
			{"folder_id": nil},
			{"folder_id": bson.M{"$exists": false}},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})

	cursor, err := database.FolderCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("yıldızlı klasörler listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("klasörler decode edilemedi: %v", err)
	}

	return folders, nil
}

// GetTrashFolders - Çöp kutusundaki klasörleri getir
func (fs *FolderService) GetTrashFolders(userID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"deleted_at": bson.M{"$ne": nil},
	}
	opts := options.Find().SetSort(bson.D{{Key: "deleted_at", Value: -1}})

	cursor, err := database.FolderCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("çöp kutusu klasörleri listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("klasörler decode edilemedi: %v", err)
	}

	return folders, nil
}

// GetTrashedFolderFiles - Silinmiş bir klasörün içindeki (silinmiş) dosyaları getir
func (fs *FolderService) GetTrashedFolderFiles(folderID string) ([]models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"folder_id":  folderID,
		"deleted_at": bson.M{"$ne": nil},
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := database.FileCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("trashed dosyalar listelenemedi: %v", err)
	}
	defer cursor.Close(ctx)

	var files []models.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("dosyalar decode edilemedi: %v", err)
	}

	return files, nil
}

// GetTrashedSubFolders - Silinmiş bir klasörün içindeki (silinmiş) alt klasörleri getir
func (fs *FolderService) GetTrashedSubFolders(parentFolderID string) ([]models.Folder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"folder_id":  parentFolderID,
		"deleted_at": bson.M{"$ne": nil},
	}

	cursor, err := database.FolderCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("trashed alt klasörler sorgulanamadı: %v", err)
	}
	defer cursor.Close(ctx)

	var folders []models.Folder
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("alt klasörler decode edilemedi: %v", err)
	}

	return folders, nil
}

// ToggleFolderStar - Klasör yıldız durumunu değiştir (recursive - alt klasörler ve dosyalar da star'lanır)
func (fs *FolderService) ToggleFolderStar(folderID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return false, fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	var folder models.Folder
	err = database.FolderCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		return false, fmt.Errorf("klasör bulunamadı: %v", err)
	}

	newStatus := !folder.IsStarred

	// 1. Ana klasörü star'la/unstar'la
	update := bson.M{
		"$set": bson.M{
			"is_starred": newStatus,
			"updated_at": time.Now(),
		},
	}

	_, err = database.FolderCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return false, fmt.Errorf("klasör güncellenemedi: %v", err)
	}

	// 2. Recursive olarak alt klasörleri ve dosyaları star'la/unstar'la
	if err := fs.toggleStarRecursive(ctx, objectID, newStatus); err != nil {
		log.Printf("Recursive star işlemi hatası: %v", err)
		// Ana klasör zaten star'landı, hata döndürmeyelim
	}

	return newStatus, nil
}

// toggleStarRecursive - Alt klasörleri ve dosyaları recursive olarak star'la/unstar'la
func (fs *FolderService) toggleStarRecursive(ctx context.Context, folderID primitive.ObjectID, starStatus bool) error {
	// 1. Bu klasördeki dosyaları star'la/unstar'la
	_, err := database.FileCollection.UpdateMany(
		ctx,
		bson.M{"folder_id": folderID.Hex()},
		bson.M{
			"$set": bson.M{
				"is_starred": starStatus,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("dosyalar güncellenemedi: %v", err)
	}

	// 2. Bu klasörün doğrudan alt klasörlerini bul (folder_id = folderID)
	cursor, err := database.FolderCollection.Find(ctx, bson.M{"folder_id": folderID.Hex()})
	if err != nil {
		return fmt.Errorf("alt klasörler bulunamadı: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var subFolder models.Folder
		if err := cursor.Decode(&subFolder); err != nil {
			continue
		}

		// Alt klasörü star'la/unstar'la
		_, err := database.FolderCollection.UpdateOne(
			ctx,
			bson.M{"_id": subFolder.ID},
			bson.M{
				"$set": bson.M{
					"is_starred": starStatus,
					"updated_at": time.Now(),
				},
			},
		)
		if err != nil {
			log.Printf("Alt klasör güncellenemedi: %v", err)
			continue
		}

		// Recursive olarak alt klasörün çocuklarını da star'la/unstar'la
		if err := fs.toggleStarRecursive(ctx, subFolder.ID, starStatus); err != nil {
			log.Printf("Recursive star işlemi hatası (alt klasör): %v", err)
			// Devam et, diğer alt klasörleri de işle
		}
	}

	return nil
}

// SoftDeleteFolder - Klasörü ve içeriğini çöp kutusuna taşı
func (fs *FolderService) SoftDeleteFolder(folderID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Recursion might take time
	defer cancel()

	return fs.softDeleteFolderRecursive(ctx, folderID)
}

func (fs *FolderService) softDeleteFolderRecursive(ctx context.Context, folderID string) error {
	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	// 1. Klasörü sil
	_, err = database.FolderCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("klasör silinemedi: %v", err)
	}

	// 2. Bu klasördeki dosyaları sil
	_, err = database.FileCollection.UpdateMany(ctx, bson.M{"folder_id": folderID}, update)
	if err != nil {
		return fmt.Errorf("klasör içeriği silinemedi: %v", err)
	}

	// 3. Alt klasörleri bul ve recursive çağır
	cursor, err := database.FolderCollection.Find(ctx, bson.M{"folder_id": folderID})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var subFolder models.Folder
		if err := cursor.Decode(&subFolder); err != nil {
			continue
		}
		if err := fs.softDeleteFolderRecursive(ctx, subFolder.ID.Hex()); err != nil {
			return err
		}
	}

	return nil
}

// RestoreFolder - Klasörü ve içeriğini geri yükle
func (fs *FolderService) RestoreFolder(folderID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return fs.restoreFolderRecursive(ctx, folderID)
}

func (fs *FolderService) restoreFolderRecursive(ctx context.Context, folderID string) error {
	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	// Klasörü getir - parent kontrolü için
	var folder models.Folder
	err = database.FolderCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&folder)
	if err != nil {
		return fmt.Errorf("klasör bulunamadı: %v", err)
	}

	update := bson.M{
		"$set": bson.M{
			"deleted_at": nil,
			"updated_at": time.Now(),
		},
	}

	// Eğer klasör bir parent klasöre bağlıysa, o parent'ın aktif olup olmadığını kontrol et
	if folder.FolderID != nil && *folder.FolderID != "" {
		parentFolder, err := fs.GetFolderByID(*folder.FolderID)
		if err != nil || parentFolder == nil || parentFolder.DeletedAt != nil {
			// Parent klasör silinmiş - bu klasör root'a taşın
			update["$set"].(bson.M)["folder_id"] = nil
			update["$set"].(bson.M)["parent_id"] = nil
			update["$set"].(bson.M)["ancestors"] = []primitive.ObjectID{}
		}
	}

	// 1. Klasörü geri yükle
	_, err = database.FolderCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("klasör geri yüklenemedi: %v", err)
	}

	// 2. Bu klasördeki dosyaları geri yükle
	// Her dosya için parent klasörü kontrol et
	_, err = database.FileCollection.UpdateMany(ctx, bson.M{"folder_id": folderID}, bson.M{
		"$set": bson.M{
			"deleted_at": nil,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("klasör içeriği geri yüklenemedi: %v", err)
	}

	// 3. Alt klasörleri bul ve recursive çağır
	cursor, err := database.FolderCollection.Find(ctx, bson.M{"folder_id": folderID})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var subFolder models.Folder
		if err := cursor.Decode(&subFolder); err != nil {
			continue
		}
		if err := fs.restoreFolderRecursive(ctx, subFolder.ID.Hex()); err != nil {
			return err
		}
	}

	return nil
}

// restoreParentHierarchy - Parent hiyerarşisini (silinmiş parent'ları) restore et
// Bu fonksiyon bir dosya/klasör restore edilirken, parent chain'ini de restore eder
func (fs *FolderService) restoreParentHierarchy(ctx context.Context, folder *models.Folder) error {
	if folder == nil || folder.DeletedAt == nil {
		// Klasör zaten aktif veya nil
		return nil
	}

	// 1. Klasörü restore et (deleted_at: nil)
	update := bson.M{
		"$set": bson.M{
			"deleted_at": nil,
			"updated_at": time.Now(),
		},
	}

	// Eğer parent klasörü silinmişse, parent'ı root'a taşı
	if folder.FolderID != nil && *folder.FolderID != "" {
		parentFolder, err := fs.GetFolderByID(*folder.FolderID)
		if err != nil || parentFolder == nil {
			// Parent bulunamadı - root'a taşı
			update["$set"].(bson.M)["folder_id"] = nil
			update["$set"].(bson.M)["parent_id"] = nil
			update["$set"].(bson.M)["ancestors"] = []primitive.ObjectID{}
		} else if parentFolder.DeletedAt != nil {
			// Parent silinmiş - parent'ı da restore et (recursive)
			if err := fs.restoreParentHierarchy(ctx, parentFolder); err != nil {
				// Hata durumunda bu klasörü root'a taşı
				update["$set"].(bson.M)["folder_id"] = nil
				update["$set"].(bson.M)["parent_id"] = nil
				update["$set"].(bson.M)["ancestors"] = []primitive.ObjectID{}
			}
			// Aksi takdirde, parent'ın ancestors'ı güncelle
		}
	}

	_, err := database.FolderCollection.UpdateOne(ctx, bson.M{"_id": folder.ID}, update)
	return err
}

// MoveFolder - Klasörü başka bir klasöre taşı
func (fs *FolderService) MoveFolder(folderID string, targetFolderID *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return fmt.Errorf("geçersiz klasör ID'si: %v", err)
	}

	// 1. Klasörü bul
	folder, err := fs.GetFolderByID(folderID)
	if err != nil {
		return err
	}

	// 2. Circular dependency kontrolü
	if targetFolderID != nil && *targetFolderID != "" {
		if folderID == *targetFolderID {
			return fmt.Errorf("klasör kendi içine taşınamaz")
		}

		// Hedef klasör, taşınan klasörün alt klasörü mü?
		targetFolder, err := fs.GetFolderByID(*targetFolderID)
		if err != nil {
			return fmt.Errorf("hedef klasör bulunamadı")
		}

		for _, ancestorID := range targetFolder.Ancestors {
			if ancestorID == folder.ID {
				return fmt.Errorf("klasör kendi alt klasörüne taşınamaz")
			}
		}
	}

	// 3. Yeni Ancestors listesini hesapla
	var newAncestors []primitive.ObjectID
	var newParentID *primitive.ObjectID

	if targetFolderID != nil && *targetFolderID != "" {
		targetFolder, _ := fs.GetFolderByID(*targetFolderID)
		newAncestors = append(targetFolder.Ancestors, targetFolder.ID)
		newParentID = &targetFolder.ID
	} else {
		newAncestors = []primitive.ObjectID{}
		newParentID = nil
	}

	// 4. Güncelleme
	update := bson.M{
		"$set": bson.M{
			"folder_id":  targetFolderID,
			"parent_id":  newParentID,
			"ancestors":  newAncestors,
			"updated_at": time.Now(),
		},
	}

	_, err = database.FolderCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("klasör taşınamadı: %v", err)
	}

	// 5. Alt klasörlerin ve dosyaların ancestors'larını recursive güncelle
	return fs.updateDescendantsAncestors(ctx, folder.ID, newAncestors)
}

func (fs *FolderService) updateDescendantsAncestors(ctx context.Context, folderID primitive.ObjectID, parentAncestors []primitive.ObjectID) error {
	// Yeni ancestors listesi: Parent'ın ancestors'ı + Parent ID
	currentAncestors := append(parentAncestors, folderID)

	// 1. Bu klasörün altındaki dosyaları güncelle
	_, err := database.FileCollection.UpdateMany(
		ctx,
		bson.M{"folder_id": folderID.Hex()},
		bson.M{
			"$set": bson.M{
				"ancestors": currentAncestors,
				"parent_id": folderID,
			},
		},
	)
	if err != nil {
		return err
	}

	// 2. Bu klasörün altındaki klasörleri bul ve güncelle
	cursor, err := database.FolderCollection.Find(ctx, bson.M{"folder_id": folderID.Hex()})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var subFolder models.Folder
		if err := cursor.Decode(&subFolder); err != nil {
			continue
		}

		// Önce alt klasörün kaydını güncelle
		_, err := database.FolderCollection.UpdateOne(
			ctx,
			bson.M{"_id": subFolder.ID},
			bson.M{
				"$set": bson.M{
					"ancestors": currentAncestors,
					"parent_id": folderID,
				},
			},
		)
		if err != nil {
			return err
		}

		// Recursive olarak onun çocuklarını güncelle
		if err := fs.updateDescendantsAncestors(ctx, subFolder.ID, currentAncestors); err != nil {
			return err
		}
	}

	return nil
}
