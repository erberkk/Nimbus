package handlers

import (
	"fmt"
	"log"
	"nimbus-backend/config"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"nimbus-backend/services"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// CreateFolder - Yeni klasör oluştur
func CreateFolder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var req models.CreateFolderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek verisi",
			})
		}

		if req.Name == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Klasör adı gerekli",
			})
		}

		// Klasör oluştur
		folder, err := services.FolderServiceInstance.CreateFolder(userID, req.Name, req.Color, req.FolderID)
		if err != nil {
			log.Printf("Klasör oluşturma hatası: %v", err)
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Klasör başarıyla oluşturuldu",
			"folder": models.FolderResponse{
				ID:         folder.ID.Hex(),
				Name:       folder.Name,
				Color:      folder.Color,
				PublicLink: folder.PublicLink,
				ItemCount:  0,
				Size:       0,
				AccessList: folder.AccessList,
				FolderID:   folder.FolderID,
				CreatedAt:  folder.CreatedAt,
				UpdatedAt:  folder.UpdatedAt,
				Owner:      services.UserServiceInstance.GetUserResponse(folder.UserID),
			},
		})
	}
}

// GetUserFolders - Kullanıcının klasörlerini listele
func GetUserFolders(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folders, err := services.FolderServiceInstance.GetUserFolders(userID)
		if err != nil {
			log.Printf("Klasör listesi alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Klasörler listelenemedi",
			})
		}

		// Klasör response'ları formatla
		folderList := make([]models.FolderResponse, 0, len(folders))
		for _, folder := range folders {
			// Her klasördeki dosya sayısını hesapla
			count, _ := services.FolderServiceInstance.GetFolderItemCount(folder.ID.Hex())

			size, _ := services.FolderServiceInstance.GetFolderSize(folder.ID.Hex())
			folderList = append(folderList, models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: int(count),
				Size:      size,
				Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
				IsStarred: folder.IsStarred,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
			})
		}

		return c.JSON(fiber.Map{
			"folders": folderList,
			"count":   len(folderList),
		})
	}
}

// GetFolderContents - Klasörün içeriğini getir (alt klasörler ve dosyalar)
func GetFolderContents(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		// Klasör bilgisini getir ve erişim kontrolü yap
		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		// Owner veya access list kontrolü
		canAccess, err := helpers.CanUserAccess(userID, "folder", folderID, helpers.AccessLevelRead)
		if err != nil || !canAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu klasöre erişim yetkiniz yok",
			})
		}

		// Starred only kontrolü (query parametresi)
		starredOnly := c.Query("starred_only") == "true"

		// Klasördeki alt klasörleri ve dosyaları getir
		var subFolders []models.Folder
		var files []models.File

		// Eğer klasör silinmişse (DeletedAt != nil), silinmiş içeriği getir.
		// Değilse normal içeriği getir.
		if folder.DeletedAt != nil {
			subFolders, err = services.FolderServiceInstance.GetTrashedSubFolders(folderID)
		} else {
			if starredOnly {
				subFolders, err = services.FolderServiceInstance.GetStarredSubFolders(folderID, userID)
			} else {
				subFolders, err = services.FolderServiceInstance.GetSubFolders(folderID)
			}
		}

		if err != nil {
			log.Printf("Alt klasörleri alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Alt klasörler listelenemedi",
			})
		}

		if folder.DeletedAt != nil {
			files, err = services.FolderServiceInstance.GetTrashedFolderFiles(folderID)
		} else {
			if starredOnly {
				files, err = services.FileServiceInstance.GetStarredFolderFiles(folderID, userID)
			} else {
				files, err = services.FolderServiceInstance.GetFolderFiles(folderID)
			}
		}

		if err != nil {
			log.Printf("Klasör dosyaları alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosyalar listelenemedi",
			})
		}

		// Klasör response'ları formatla
		folderList := make([]models.FolderResponse, 0, len(subFolders))
		for _, subFolder := range subFolders {
			// Her alt klasörün toplam item sayısını hesapla
			count, err := services.FolderServiceInstance.GetFolderItemCount(subFolder.ID.Hex())
			if err != nil {
				log.Printf("Klasör item count hesaplama hatası: %v", err)
				count = 0
			}

			size, _ := services.FolderServiceInstance.GetFolderSize(subFolder.ID.Hex())
			folderList = append(folderList, models.FolderResponse{
				ID:        subFolder.ID.Hex(),
				Name:      subFolder.Name,
				Color:     subFolder.Color,
				ItemCount: int(count),
				Size:      size,
				IsStarred: subFolder.IsStarred,
				FolderID:  subFolder.FolderID,
				DeletedAt: subFolder.DeletedAt,
				CreatedAt: subFolder.CreatedAt,
				UpdatedAt: subFolder.UpdatedAt,
				Owner:     services.UserServiceInstance.GetUserResponse(subFolder.UserID),
			})
		}

		// Dosya response'ları formatla
		fileList := make([]models.FileResponse, 0, len(files))
		for _, file := range files {
			fileList = append(fileList, models.FileResponse{
				ID:               file.ID.Hex(),
				UserID:           file.UserID,
				Filename:         file.Filename,
				Size:             file.Size,
				ContentType:      file.ContentType,
				MinioPath:        file.MinioPath,
				PublicLink:       file.PublicLink,
				AccessList:       file.AccessList,
				ParentID:         file.ParentID,
				Ancestors:        file.Ancestors,
				IsStarred:        file.IsStarred,
				ProcessingStatus: file.ProcessingStatus,
				ProcessingError:  file.ProcessingError,
				ProcessedAt:      file.ProcessedAt,
				ChunkCount:       file.ChunkCount,
				DeletedAt:        file.DeletedAt,
				CreatedAt:        file.CreatedAt,
				UpdatedAt:        file.UpdatedAt,
			})
		}

		size, _ := services.FolderServiceInstance.GetFolderSize(folderID)
		return c.JSON(fiber.Map{
			"folder": models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: len(subFolders) + len(files),
				Size:      size,
				FolderID:  folder.FolderID,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
				Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
			},
			"folders": folderList,
			"files":   fileList,
			"count":   len(folderList) + len(fileList),
		})
	}
}

// GetStorageUsage - Kullanıcının depolama kullanımını getir
func GetStorageUsage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Kullanıcının toplam depolama kullanımını hesapla
		totalSize, err := services.FolderServiceInstance.GetUserStorageUsage(userID)
		if err != nil {
			log.Printf("Depolama kullanımı hesaplama hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Depolama kullanımı hesaplanamadı",
			})
		}

		// Byte'dan GB'a çevir ve format'la
		const GB = 1024 * 1024 * 1024
		totalGB := float64(totalSize) / GB

		var usage string
		if totalGB >= 1.0 {
			usage = fmt.Sprintf("%.1f GB", totalGB)
		} else {
			totalMB := float64(totalSize) / (1024 * 1024)
			usage = fmt.Sprintf("%.0f MB", totalMB)
		}

		return c.JSON(fiber.Map{
			"total_size": totalSize,
			"usage":      usage,
			"usage_gb":   totalGB,
		})
	}
}

// GetRootContents - Root klasörün içeriğini getir (klasörler + root dosyalar)
func GetRootContents(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Root klasörleri getir (folder_id olmayan klasörler)
		folders, err := services.FolderServiceInstance.GetSubFolders("")
		if err != nil {
			log.Printf("Root klasör listesi alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Klasörler listelenemedi",
			})
		}

		// Sadece kullanıcının klasörlerini filtrele
		userFolders := make([]models.Folder, 0)
		for _, folder := range folders {
			if folder.UserID == userID {
				userFolders = append(userFolders, folder)
			}
		}

		// Root dosyaları getir (folder_id = null)
		files, err := services.FolderServiceInstance.GetRootFiles(userID)
		if err != nil {
			log.Printf("Root dosyaları alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosyalar listelenemedi",
			})
		}

		// Klasör response'ları
		folderList := make([]models.FolderResponse, 0, len(userFolders))
		for _, folder := range userFolders {
			count, _ := services.FolderServiceInstance.GetFolderItemCount(folder.ID.Hex())

			size, _ := services.FolderServiceInstance.GetFolderSize(folder.ID.Hex())
			folderList = append(folderList, models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: int(count),
				Size:      size,
				IsStarred: folder.IsStarred,
				DeletedAt: folder.DeletedAt,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
				Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
			})
		}

		// Dosya response'ları
		fileList := make([]models.FileResponse, 0, len(files))
		for _, file := range files {
			fileList = append(fileList, models.FileResponse{
				ID:               file.ID.Hex(),
				UserID:           file.UserID,
				Filename:         file.Filename,
				Size:             file.Size,
				ContentType:      file.ContentType,
				MinioPath:        file.MinioPath,
				PublicLink:       file.PublicLink,
				AccessList:       file.AccessList,
				ParentID:         file.ParentID,
				Ancestors:        file.Ancestors,
				IsStarred:        file.IsStarred,
				ProcessingStatus: file.ProcessingStatus,
				ProcessingError:  file.ProcessingError,
				ProcessedAt:      file.ProcessedAt,
				ChunkCount:       file.ChunkCount,
				DeletedAt:        file.DeletedAt,
				CreatedAt:        file.CreatedAt,
				UpdatedAt:        file.UpdatedAt,
				Owner:            services.UserServiceInstance.GetUserResponse(file.UserID),
			})
		}

		return c.JSON(fiber.Map{
			"folders": folderList,
			"files":   fileList,
		})
	}
}

// UpdateFolder - Klasör güncelle
func UpdateFolder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		var req models.UpdateFolderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek verisi",
			})
		}

		// Klasör owner kontrolü
		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		if folder.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu klasörü güncelleme yetkiniz yok",
			})
		}

		// Güncellemeleri hazırla
		updates := bson.M{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Color != "" {
			updates["color"] = req.Color
		}

		if len(updates) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Güncellenecek alan belirtilmedi",
			})
		}

		// Klasörü güncelle
		if err := services.FolderServiceInstance.UpdateFolder(folderID, updates); err != nil {
			log.Printf("Klasör güncelleme hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Klasör başarıyla güncellendi",
		})
	}
}

// DeleteFolder - Klasör sil (Soft veya Hard)
func DeleteFolder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		isPermanent := c.Query("permanent") == "true"

		// Klasör owner kontrolü
		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		if folder.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu klasörü silme yetkiniz yok",
			})
		}

		if isPermanent {
			// Klasörü sil (boşsa)
			if err := services.FolderServiceInstance.DeleteFolder(folderID); err != nil {
				log.Printf("Klasör silme hatası: %v", err)
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.JSON(fiber.Map{
				"message": "Klasör kalıcı olarak silindi",
			})
		} else {
			// Soft Delete
			if err := services.FolderServiceInstance.SoftDeleteFolder(folderID); err != nil {
				log.Printf("Klasör silme hatası: %v", err)
				return c.Status(400).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.JSON(fiber.Map{
				"message": "Klasör çöp kutusuna taşındı",
			})
		}
	}
}

// GetStarredFolders - Yıldızlı klasörleri getir
func GetStarredFolders(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folders, err := services.FolderServiceInstance.GetStarredFolders(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Klasörler alınamadı",
			})
		}

		return c.JSON(fiber.Map{
			"folders": formatFolderResponse(folders),
			"count":   len(folders),
		})
	}
}

// GetTrashFolders - Çöp kutusundaki klasörleri getir
func GetTrashFolders(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folders, err := services.FolderServiceInstance.GetTrashFolders(userID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Klasörler alınamadı",
			})
		}

		return c.JSON(fiber.Map{
			"folders": formatFolderResponse(folders),
			"count":   len(folders),
		})
	}
}

// ToggleFolderStar - Klasör yıldız durumunu değiştir
func ToggleFolderStar(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		if folder.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Yetkiniz yok",
			})
		}

		newStatus, err := services.FolderServiceInstance.ToggleFolderStar(folderID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "İşlem başarısız",
			})
		}

		return c.JSON(fiber.Map{
			"is_starred": newStatus,
		})
	}
}

// RestoreFolder - Klasörü geri yükle
func RestoreFolder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")
		if folderID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "folder ID parametresi gerekli",
			})
		}

		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		if folder.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Yetkiniz yok",
			})
		}

		if err := services.FolderServiceInstance.RestoreFolder(folderID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Geri yükleme başarısız",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Klasör geri yüklendi",
		})
	}
}

// MoveFolder - Klasör taşı
func MoveFolder(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		folderID := c.Params("id")

		var req struct {
			FolderID *string `json:"folder_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek",
			})
		}

		folder, err := services.FolderServiceInstance.GetFolderByID(folderID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Klasör bulunamadı",
			})
		}

		if folder.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Yetkiniz yok",
			})
		}

		if err := services.FolderServiceInstance.MoveFolder(folderID, req.FolderID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Klasör taşındı",
		})
	}
}

func formatFolderResponse(folders []models.Folder) []models.FolderResponse {
	folderList := make([]models.FolderResponse, 0, len(folders))
	for _, folder := range folders {
		count, _ := services.FolderServiceInstance.GetFolderItemCount(folder.ID.Hex())
		size, _ := services.FolderServiceInstance.GetFolderSize(folder.ID.Hex())

		folderList = append(folderList, models.FolderResponse{
			ID:        folder.ID.Hex(),
			Name:      folder.Name,
			Color:     folder.Color,
			ItemCount: int(count),
			Size:      size,
			IsStarred: folder.IsStarred,
			FolderID:  folder.FolderID,
			DeletedAt: folder.DeletedAt,
			CreatedAt: folder.CreatedAt,
			UpdatedAt: folder.UpdatedAt,
			Owner:     services.UserServiceInstance.GetUserResponse(folder.UserID),
		})
	}
	return folderList
}
