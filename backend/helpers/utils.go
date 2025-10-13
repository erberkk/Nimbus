package helpers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"nimbus-backend/database"
	"nimbus-backend/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GeneratePublicLink - Rastgele 16 karakterlik public link oluştur
func GeneratePublicLink() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetAncestors - Belirtilen resource'un tüm üst klasörlerini döndürür
func GetAncestors(resourceID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx := context.Background()

	// Önce bu resource'un parent'ını bul
	var resource interface{}
	err := database.FolderCollection.FindOne(ctx, bson.M{"_id": resourceID}).Decode(&resource)
	if err != nil {
		// Folder bulunamadı, file olarak ara
		err = database.FileCollection.FindOne(ctx, bson.M{"_id": resourceID}).Decode(&resource)
		if err != nil {
			return nil, err
		}
	}

	// ParentID varsa ancestors'ı döndür, yoksa boş array
	ancestors := []primitive.ObjectID{}
	if folder, ok := resource.(models.Folder); ok {
		ancestors = folder.Ancestors
	} else if file, ok := resource.(models.File); ok {
		ancestors = file.Ancestors
	}

	return ancestors, nil
}

// UpdateAncestorsRecursive - Belirtilen klasörün tüm alt öğelerinin ancestors'ını günceller
func UpdateAncestorsRecursive(parentID primitive.ObjectID, ancestors []primitive.ObjectID) error {
	ctx := context.Background()

	// Güncellenmiş ancestors array'i oluştur (parent'ı da dahil et)
	newAncestors := append(ancestors, parentID)

	// Tüm alt klasörleri bul ve güncelle
	folderCursor, err := database.FolderCollection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return err
	}
	defer folderCursor.Close(ctx)

	for folderCursor.Next(ctx) {
		var folder models.Folder
		if err := folderCursor.Decode(&folder); err != nil {
			continue
		}

		// Bu klasörün ancestors'ını güncelle
		_, err = database.FolderCollection.UpdateOne(
			ctx,
			bson.M{"_id": folder.ID},
			bson.M{"$set": bson.M{
				"ancestors":  newAncestors,
				"updated_at": time.Now(),
			}},
		)
		if err != nil {
			continue
		}

		// Bu klasörün alt öğelerini de recursive güncelle
		UpdateAncestorsRecursive(folder.ID, newAncestors)
	}

	// Tüm alt dosyaları bul ve güncelle
	fileCursor, err := database.FileCollection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return err
	}
	defer fileCursor.Close(ctx)

	for fileCursor.Next(ctx) {
		var file models.File
		if err := fileCursor.Decode(&file); err != nil {
			continue
		}

		// Bu dosyanın ancestors'ını güncelle
		database.FileCollection.UpdateOne(
			ctx,
			bson.M{"_id": file.ID},
			bson.M{"$set": bson.M{
				"ancestors":  newAncestors,
				"updated_at": time.Now(),
			}},
		)
	}

	return nil
}

// MergeAccessLevels - Birden fazla erişim seviyesini birleştirir (en yüksek seviyeyi alır)
func MergeAccessLevels(accessEntries []models.AccessEntry) string {
	if len(accessEntries) == 0 {
		return "none"
	}

	priority := map[string]int{
		"none":  0,
		"read":  1,
		"write": 2,
	}

	maxLevel := "none"
	for _, entry := range accessEntries {
		if priority[entry.AccessType] > priority[maxLevel] {
			maxLevel = entry.AccessType
		}
	}

	return maxLevel
}

// GetAllChildrenRecursive - Belirtilen klasörün tüm alt öğelerini recursive olarak bulur
func GetAllChildrenRecursive(parentID primitive.ObjectID) ([]models.Folder, []models.File, error) {
	ctx := context.Background()
	var folders []models.Folder
	var files []models.File

	// Önce alt klasörleri bul
	folderCursor, err := database.FolderCollection.Find(ctx, bson.M{"ancestors": parentID})
	if err != nil {
		return nil, nil, err
	}
	defer folderCursor.Close(ctx)

	for folderCursor.Next(ctx) {
		var folder models.Folder
		if err := folderCursor.Decode(&folder); err != nil {
			continue
		}
		folders = append(folders, folder)

		// Bu klasörün alt öğelerini de recursive al
		childFolders, childFiles, err := GetAllChildrenRecursive(folder.ID)
		if err != nil {
			continue
		}
		folders = append(folders, childFolders...)
		files = append(files, childFiles...)
	}

	// Alt dosyaları bul
	fileCursor, err := database.FileCollection.Find(ctx, bson.M{"ancestors": parentID})
	if err != nil {
		return nil, nil, err
	}
	defer fileCursor.Close(ctx)

	for fileCursor.Next(ctx) {
		var file models.File
		if err := fileCursor.Decode(&file); err != nil {
			continue
		}
		files = append(files, file)
	}

	return folders, files, nil
}

// HasHierarchicalAccess - Belirtilen kullanıcı için resource'a erişimi kontrol eder
func HasHierarchicalAccess(resourceID primitive.ObjectID, userID string) (string, error) {
	ctx := context.Background()

	// Resource'un ancestors'ını al
	ancestors, err := GetAncestors(resourceID)
	if err != nil {
		return "none", err
	}

	// Tüm ancestors'larda bu kullanıcının erişimini ara
	var accessEntries []models.AccessEntry

	// Folder ancestors'larında ara
	if len(ancestors) > 0 {
		folderCursor, err := database.FolderCollection.Find(ctx, bson.M{
			"_id":                 bson.M{"$in": ancestors},
			"access_list.user_id": userID,
		})
		if err == nil {
			defer folderCursor.Close(ctx)
			for folderCursor.Next(ctx) {
				var folder models.Folder
				if err := folderCursor.Decode(&folder); err != nil {
					continue
				}
				for _, access := range folder.AccessList {
					if access.UserID == userID {
						accessEntries = append(accessEntries, access)
					}
				}
			}
		}
	}

	// File ancestors'larında ara (dosyanın klasör zincirinde)
	if len(ancestors) > 0 {
		fileCursor, err := database.FileCollection.Find(ctx, bson.M{
			"ancestors":           bson.M{"$in": ancestors},
			"access_list.user_id": userID,
		})
		if err == nil {
			defer fileCursor.Close(ctx)
			for fileCursor.Next(ctx) {
				var file models.File
				if err := fileCursor.Decode(&file); err != nil {
					continue
				}
				for _, access := range file.AccessList {
					if access.UserID == userID {
						accessEntries = append(accessEntries, access)
					}
				}
			}
		}
	}

	// En yüksek erişim seviyesini döndür
	return MergeAccessLevels(accessEntries), nil
}

// AddUserToResourceAccess - Belirtilen resource'a kullanıcı erişimi ekler
func AddUserToResourceAccess(resourceID primitive.ObjectID, userID, accessType string, grantedBy string) error {
	ctx := context.Background()

	// Önce bu resource'un türünü belirle
	var collection interface{}
	var update bson.M

	// Folder mı kontrol et
	err := database.FolderCollection.FindOne(ctx, bson.M{"_id": resourceID}).Decode(&collection)
	if err == nil {
		// Folder ise
		update = bson.M{
			"$push": bson.M{"access_list": models.AccessEntry{
				UserID:     userID,
				AccessType: accessType,
				GrantedAt:  time.Now(),
				GrantedBy:  grantedBy,
			}},
			"$set": bson.M{"updated_at": time.Now()},
		}
		database.FolderCollection.UpdateOne(ctx, bson.M{"_id": resourceID}, update)
	} else {
		// File ise
		update = bson.M{
			"$push": bson.M{"access_list": models.AccessEntry{
				UserID:     userID,
				AccessType: accessType,
				GrantedAt:  time.Now(),
				GrantedBy:  grantedBy,
			}},
			"$set": bson.M{"updated_at": time.Now()},
		}
		database.FileCollection.UpdateOne(ctx, bson.M{"_id": resourceID}, update)
	}

	return err
}

// RemoveUserFromResourceAccess - Belirtilen resource'dan kullanıcı erişimini kaldırır
func RemoveUserFromResourceAccess(resourceID primitive.ObjectID, userID string) error {
	ctx := context.Background()

	// Önce bu resource'un türünü belirle
	var collection interface{}

	// Folder mı kontrol et
	err := database.FolderCollection.FindOne(ctx, bson.M{"_id": resourceID}).Decode(&collection)
	if err == nil {
		// Folder ise
		database.FolderCollection.UpdateOne(
			ctx,
			bson.M{"_id": resourceID},
			bson.M{
				"$pull": bson.M{"access_list": bson.M{"user_id": userID}},
				"$set":  bson.M{"updated_at": time.Now()},
			},
		)
	} else {
		// File ise
		database.FileCollection.UpdateOne(
			ctx,
			bson.M{"_id": resourceID},
			bson.M{
				"$pull": bson.M{"access_list": bson.M{"user_id": userID}},
				"$set":  bson.M{"updated_at": time.Now()},
			},
		)
	}

	return err
}

// PropagateAccessToChildren - Ana klasör paylaşıldığında tüm alt öğeleri günceller
func PropagateAccessToChildren(parentID primitive.ObjectID, userID, accessType string, grantedBy string) error {
	ctx := context.Background()

	// Tüm alt öğeleri bul
	childrenFolders, childrenFiles, err := GetAllChildrenRecursive(parentID)
	if err != nil {
		return err
	}

	// Bulk operations için array oluştur
	var operations []mongo.WriteModel

	// Tüm alt klasörleri güncelle
	for _, folder := range childrenFolders {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": folder.ID}).
			SetUpdate(bson.M{
				"$push": bson.M{"access_list": models.AccessEntry{
					UserID:        userID,
					AccessType:    accessType,
					GrantedAt:     time.Now(),
					GrantedBy:     grantedBy,
					InheritedFrom: &parentID,
				}},
				"$set": bson.M{"updated_at": time.Now()},
			}),
		)
	}

	// Tüm alt dosyaları güncelle
	for _, file := range childrenFiles {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": file.ID}).
			SetUpdate(bson.M{
				"$push": bson.M{"access_list": models.AccessEntry{
					UserID:        userID,
					AccessType:    accessType,
					GrantedAt:     time.Now(),
					GrantedBy:     grantedBy,
					InheritedFrom: &parentID,
				}},
				"$set": bson.M{"updated_at": time.Now()},
			}),
		)
	}

	// Bulk update yap
	if len(operations) > 0 {
		_, err = database.FolderCollection.BulkWrite(ctx, operations[:len(childrenFolders)])
		if err != nil {
			return err
		}

		if len(childrenFiles) > 0 {
			_, err = database.FileCollection.BulkWrite(ctx, operations[len(childrenFolders):])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveAccessFromChildren - Erişim kaldırıldığında tüm alt öğelerden de kaldırır
func RemoveAccessFromChildren(parentID primitive.ObjectID, userID string) error {
	ctx := context.Background()

	// Tüm alt öğeleri bul
	childrenFolders, childrenFiles, err := GetAllChildrenRecursive(parentID)
	if err != nil {
		return err
	}

	// Bulk operations için array oluştur
	var operations []mongo.WriteModel

	// Tüm alt klasörlerden kullanıcıyı çıkar
	for _, folder := range childrenFolders {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": folder.ID}).
			SetUpdate(bson.M{
				"$pull": bson.M{"access_list": bson.M{"user_id": userID}},
				"$set":  bson.M{"updated_at": time.Now()},
			}),
		)
	}

	// Tüm alt dosyalardan kullanıcıyı çıkar
	for _, file := range childrenFiles {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": file.ID}).
			SetUpdate(bson.M{
				"$pull": bson.M{"access_list": bson.M{"user_id": userID}},
				"$set":  bson.M{"updated_at": time.Now()},
			}),
		)
	}

	// Bulk update yap
	if len(operations) > 0 {
		_, err = database.FolderCollection.BulkWrite(ctx, operations[:len(childrenFolders)])
		if err != nil {
			return err
		}

		if len(childrenFiles) > 0 {
			_, err = database.FileCollection.BulkWrite(ctx, operations[len(childrenFolders):])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UpdateAccessInHierarchy - Hiyerarşik erişim güncellemesi (mevcut erişimi değiştir)
func UpdateAccessInHierarchy(resourceID primitive.ObjectID, userID, newAccessType string, grantedBy string) error {
	ctx := context.Background()

	// Tüm alt öğeleri bul
	childrenFolders, childrenFiles, err := GetAllChildrenRecursive(resourceID)
	if err != nil {
		return err
	}

	// Bulk operations için array oluştur
	var operations []mongo.WriteModel

	// Tüm alt klasörlerdeki erişimi güncelle
	for _, folder := range childrenFolders {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"_id":                 folder.ID,
				"access_list.user_id": userID,
			}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"access_list.$.access_type": newAccessType,
					"access_list.$.granted_at":  time.Now(),
					"access_list.$.granted_by":  grantedBy,
					"updated_at":                time.Now(),
				},
			}),
		)
	}

	// Tüm alt dosyalardaki erişimi güncelle
	for _, file := range childrenFiles {
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(bson.M{
				"_id":                 file.ID,
				"access_list.user_id": userID,
			}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"access_list.$.access_type": newAccessType,
					"access_list.$.granted_at":  time.Now(),
					"access_list.$.granted_by":  grantedBy,
					"updated_at":                time.Now(),
				},
			}),
		)
	}

	// Bulk update yap
	if len(operations) > 0 {
		_, err = database.FolderCollection.BulkWrite(ctx, operations[:len(childrenFolders)])
		if err != nil {
			return err
		}

		if len(childrenFiles) > 0 {
			_, err = database.FileCollection.BulkWrite(ctx, operations[len(childrenFolders):])
			if err != nil {
				return err
			}
		}
	}

	return nil
}
