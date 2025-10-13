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
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: 0,
				FolderID:  folder.FolderID,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
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

			folderList = append(folderList, models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: int(count),
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

		// Klasördeki alt klasörleri getir
		subFolders, err := services.FolderServiceInstance.GetSubFolders(folderID)
		if err != nil {
			log.Printf("Alt klasörleri alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Alt klasörler listelenemedi",
			})
		}

		// Klasördeki dosyaları getir
		files, err := services.FolderServiceInstance.GetFolderFiles(folderID)
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

			folderList = append(folderList, models.FolderResponse{
				ID:        subFolder.ID.Hex(),
				Name:      subFolder.Name,
				Color:     subFolder.Color,
				ItemCount: int(count),
				FolderID:  subFolder.FolderID,
				CreatedAt: subFolder.CreatedAt,
				UpdatedAt: subFolder.UpdatedAt,
			})
		}

		// Dosya response'ları formatla
		fileList := make([]models.FileResponse, 0, len(files))
		for _, file := range files {
			fileList = append(fileList, models.FileResponse{
				ID:          file.ID.Hex(),
				Filename:    file.Filename,
				Size:        file.Size,
				ContentType: file.ContentType,
				CreatedAt:   file.CreatedAt,
				UpdatedAt:   file.UpdatedAt,
			})
		}

		return c.JSON(fiber.Map{
			"folder": models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: len(subFolders) + len(files),
				FolderID:  folder.FolderID,
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
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
			folderList = append(folderList, models.FolderResponse{
				ID:        folder.ID.Hex(),
				Name:      folder.Name,
				Color:     folder.Color,
				ItemCount: int(count),
				CreatedAt: folder.CreatedAt,
				UpdatedAt: folder.UpdatedAt,
			})
		}

		// Dosya response'ları
		fileList := make([]models.FileResponse, 0, len(files))
		for _, file := range files {
			fileList = append(fileList, models.FileResponse{
				ID:          file.ID.Hex(),
				Filename:    file.Filename,
				Size:        file.Size,
				ContentType: file.ContentType,
				CreatedAt:   file.CreatedAt,
				UpdatedAt:   file.UpdatedAt,
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

// DeleteFolder - Klasör sil
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

		// Klasörü sil (boşsa)
		if err := services.FolderServiceInstance.DeleteFolder(folderID); err != nil {
			log.Printf("Klasör silme hatası: %v", err)
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Klasör başarıyla silindi",
		})
	}
}

// MoveFile - Dosyayı başka klasöre taşı
func MoveFile(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		fileID := c.Params("id")
		if fileID == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "file ID parametresi gerekli",
			})
		}

		var req struct {
			FolderID *string `json:"folder_id"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek verisi",
			})
		}

		// Dosya owner kontrolü
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		if file.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyayı taşıma yetkiniz yok",
			})
		}

		// Hedef klasör kontrolü (varsa)
		if req.FolderID != nil && *req.FolderID != "" {
			folder, err := services.FolderServiceInstance.GetFolderByID(*req.FolderID)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{
					"error": "Hedef klasör bulunamadı",
				})
			}

			if folder.UserID != userID {
				return c.Status(403).JSON(fiber.Map{
					"error": "Hedef klasöre erişim yetkiniz yok",
				})
			}
		}

		// Dosyayı taşı
		if err := services.FolderServiceInstance.MoveFileToFolder(&fileID, req.FolderID); err != nil {
			log.Printf("Dosya taşıma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Dosya başarıyla taşındı",
		})
	}
}
