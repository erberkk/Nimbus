package handlers

import (
	"log"
	"nimbus-backend/config"
	"nimbus-backend/helpers"
	"nimbus-backend/models"
	"nimbus-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Upload için presigned URL al
func GetUploadPresignedURL(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.Claims)

		filename := c.Query("filename")
		if filename == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "filename parametresi gerekli",
			})
		}

		contentType := c.Query("content_type")
		if contentType == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "content_type parametresi gerekli",
			})
		}

		// Güvenlik kontrolü
		if err := services.MinioService.ValidateFile(filename, contentType, 0); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// 1 saatlik presigned URL oluştur
		presignedURL, err := services.MinioService.GenerateUploadPresignedURL(
			user.UserID,
			filename,
			time.Hour,
		)
		if err != nil {
			log.Printf("Presigned URL oluşturma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Presigned URL oluşturulamadı",
			})
		}

		// MinIO path oluştur
		minioPath := services.MinioService.GetUserFilePath(user.UserID, filename)

		return c.JSON(fiber.Map{
			"presigned_url": presignedURL,
			"filename":      filename,
			"minio_path":    minioPath,
			"expires_in":    3600,
		})
	}
}

// CreateFile - Dosya metadata'sını MongoDB'ye kaydet
func CreateFile(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var req models.CreateFileRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Geçersiz istek verisi",
			})
		}

		// Dosya kaydı oluştur
		file, err := services.FileServiceInstance.CreateFileRecord(
			userID,
			req.Filename,
			req.Size,
			req.ContentType,
			req.MinioPath,
			req.FolderID,
		)
		if err != nil {
			log.Printf("Dosya kaydı oluşturma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosya kaydı oluşturulamadı",
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Dosya başarıyla kaydedildi",
			"file": models.FileResponse{
				ID:          file.ID.Hex(),
				Filename:    file.Filename,
				Size:        file.Size,
				ContentType: file.ContentType,
				PublicLink:  file.PublicLink,
				AccessList:  file.AccessList,
				CreatedAt:   file.CreatedAt,
				UpdatedAt:   file.UpdatedAt,
			},
		})
	}
}

// Download için presigned URL al
func GetDownloadPresignedURL(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.Claims)

		filename := c.Query("filename")
		if filename == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "filename parametresi gerekli",
			})
		}

		// 1 saatlik presigned URL oluştur
		presignedURL, err := services.MinioService.GenerateDownloadPresignedURL(
			user.UserID,
			filename,
			time.Hour,
		)
		if err != nil {
			log.Printf("Download presigned URL oluşturma hatası: %v", err)
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı veya presigned URL oluşturulamadı",
			})
		}

		return c.JSON(fiber.Map{
			"presigned_url": presignedURL,
			"filename":      filename,
			"expires_in":    3600, // saniye cinsinden
		})
	}
}

// Kullanıcının dosyalarını listele (MongoDB'den)
func ListUserFiles(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		files, err := services.FileServiceInstance.GetUserFiles(userID)
		if err != nil {
			log.Printf("Dosya listesi alma hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosya listesi alınamadı",
			})
		}

		// Dosya bilgilerini formatla
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
			"files": fileList,
			"count": len(fileList),
		})
	}
}

// DeleteFile - Dosyayı MongoDB ve MinIO'dan sil
func DeleteFile(cfg *config.Config) fiber.Handler {
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

		// Dosya kaydını getir
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Dosyanın sahibi mi kontrol et
		if file.UserID != userID {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyayı silme yetkiniz yok",
			})
		}

		// MongoDB'den sil
		if err := services.FileServiceInstance.DeleteFileRecord(fileID); err != nil {
			log.Printf("Dosya kaydı silme hatası: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Dosya kaydı silinemedi",
			})
		}

		// MinIO'dan sil (opsiyonel - hata olsa bile kayıt silindi)
		if err := services.MinioService.DeleteFile(file.MinioPath); err != nil {
			log.Printf("MinIO'dan dosya silme hatası: %v (kayıt silindi)", err)
		}

		return c.JSON(fiber.Map{
			"message": "Dosya başarıyla silindi",
		})
	}
}
