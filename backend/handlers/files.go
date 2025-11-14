package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nimbus-backend/config"
	"nimbus-backend/helpers"
	"nimbus-backend/middleware"
	"nimbus-backend/models"
	"nimbus-backend/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/minio/minio-go/v7"
)

// Upload için presigned URL al
func GetUploadPresignedURL(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.Claims)

		filename := c.Query("filename")
		if filename == "" {
			return middleware.BadRequestResponse(c, "filename parametresi gerekli")
		}

		contentType := c.Query("content_type")
		if contentType == "" {
			return middleware.BadRequestResponse(c, "content_type parametresi gerekli")
		}

		// Güvenlik kontrolü
		if err := services.MinioService.ValidateFile(filename, contentType, 0); err != nil {
			return middleware.BadRequestResponse(c, err.Error())
		}

		// 1 saatlik presigned URL oluştur
		presignedURL, err := services.MinioService.GenerateUploadPresignedURL(
			user.UserID,
			filename,
			time.Hour,
		)
		if err != nil {
			log.Printf("Presigned URL oluşturma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Presigned URL oluşturulamadı")
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

		// Auto-trigger processing for PDF/DOCX files
		contentTypeLower := strings.ToLower(file.ContentType)
		isAskable := strings.Contains(contentTypeLower, "pdf") ||
			strings.Contains(contentTypeLower, "wordprocessingml") ||
			strings.Contains(contentTypeLower, "msword")

		if isAskable && services.DocumentProcessorInstance != nil {
			log.Printf("Auto-triggering document processing for file %s (%s)", file.ID.Hex(), file.Filename)
			services.DocumentProcessorInstance.ProcessDocumentAsync(file.ID.Hex(), file.MinioPath, file.ContentType)
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "Dosya başarıyla kaydedildi",
			"file": models.FileResponse{
				ID:               file.ID.Hex(),
				Filename:         file.Filename,
				Size:             file.Size,
				ContentType:      file.ContentType,
				PublicLink:       file.PublicLink,
				AccessList:       file.AccessList,
				ProcessingStatus: file.ProcessingStatus,
				ProcessingError:  file.ProcessingError,
				ProcessedAt:      file.ProcessedAt,
				ChunkCount:       file.ChunkCount,
				CreatedAt:        file.CreatedAt,
				UpdatedAt:        file.UpdatedAt,
			},
		})
	}
}

// ProcessDocument - Trigger document processing for PDF/DOCX files
func ProcessDocument(cfg *config.Config) fiber.Handler {
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

		// Get file record
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Check if user has access to the file
		if file.UserID != userID {
			// Check if file is shared with read or write access
			hasAccess := false
			for _, access := range file.AccessList {
				if access.UserID == userID {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				return c.Status(403).JSON(fiber.Map{
					"error": "Bu dosyaya erişim yetkiniz yok",
				})
			}
		}

		// Check if file is PDF or DOCX
		contentTypeLower := strings.ToLower(file.ContentType)
		isAskable := strings.Contains(contentTypeLower, "pdf") ||
			strings.Contains(contentTypeLower, "wordprocessingml") ||
			strings.Contains(contentTypeLower, "msword")

		if !isAskable {
			return c.Status(400).JSON(fiber.Map{
				"error": "Bu dosya türü işlenemiyor. Sadece PDF ve DOCX dosyaları desteklenmektedir.",
			})
		}

		// Check if already processing or completed
		if file.ProcessingStatus == "processing" {
			return c.Status(409).JSON(fiber.Map{
				"error":  "Dosya zaten işleniyor",
				"status": file.ProcessingStatus,
			})
		}

		if file.ProcessingStatus == "completed" {
			return c.JSON(fiber.Map{
				"message":     "Dosya zaten işlenmiş",
				"status":      file.ProcessingStatus,
				"chunk_count": file.ChunkCount,
			})
		}

		// Start async processing
		services.DocumentProcessorInstance.ProcessDocumentAsync(fileID, file.MinioPath, file.ContentType)

		return c.Status(202).JSON(fiber.Map{
			"message": "Dosya işleme başlatıldı",
			"status":  "pending",
		})
	}
}

// Download için presigned URL al
func GetDownloadPresignedURL(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.Claims)

		filename := c.Query("filename")
		if filename == "" {
			return middleware.BadRequestResponse(c, "filename parametresi gerekli")
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

// Preview için presigned URL al (inline görüntüleme için)
func GetPreviewPresignedURL(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return middleware.BadRequestResponse(c, err.Error())
		}

		fileID := c.Query("file_id")
		if fileID == "" {
			// Fallback: use filename if file_id not provided (for backward compatibility)
			filename := c.Query("filename")
			if filename == "" {
				return middleware.BadRequestResponse(c, "file_id veya filename parametresi gerekli")
			}

			// For own files, use current user's ID
			presignedURL, err := services.MinioService.GenerateDownloadPresignedURL(
				userID,
				filename,
				time.Hour,
			)
			if err != nil {
				log.Printf("Preview presigned URL oluşturma hatası: %v", err)
				return middleware.InternalServerErrorResponse(c, "Preview URL oluşturulamadı")
			}

			return c.JSON(fiber.Map{
				"presigned_url": presignedURL,
				"filename":      filename,
				"expires_in":    3600,
			})
		}

		// Get file from MongoDB
		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			log.Printf("Dosya bulunamadı: %v", err)
			return c.Status(404).JSON(fiber.Map{
				"error": "Dosya bulunamadı",
			})
		}

		// Check access using helper function
		hasAccess, err := helpers.CanUserAccess(userID, "file", fileID, helpers.AccessLevelRead)
		if err != nil || !hasAccess {
			return c.Status(403).JSON(fiber.Map{
				"error": "Bu dosyaya erişim yetkiniz yok",
			})
		}

		// Use file owner's UserID to generate presigned URL (file is stored in owner's folder)
		presignedURL, err := services.MinioService.GenerateDownloadPresignedURL(
			file.UserID,
			file.Filename,
			time.Hour,
		)
		if err != nil {
			log.Printf("Preview presigned URL oluşturma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Preview URL oluşturulamadı")
		}

		return c.JSON(fiber.Map{
			"presigned_url": presignedURL,
			"filename":      file.Filename,
			"expires_in":    3600,
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
				ID:               file.ID.Hex(),
				Filename:         file.Filename,
				Size:             file.Size,
				ContentType:      file.ContentType,
				PublicLink:       file.PublicLink,
				AccessList:       file.AccessList,
				ParentID:         file.ParentID,
				Ancestors:        file.Ancestors,
				ProcessingStatus: file.ProcessingStatus,
				ProcessingError:  file.ProcessingError,
				ProcessedAt:      file.ProcessedAt,
				ChunkCount:       file.ChunkCount,
				CreatedAt:        file.CreatedAt,
				UpdatedAt:        file.UpdatedAt,
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

		// Chroma'dan sil (if document was processed)
		if file.ProcessingStatus == "completed" && services.DocumentProcessorInstance != nil {
			chromaService := services.NewChromaService(cfg)
			if err := chromaService.DeleteDocumentChunks(fileID); err != nil {
				log.Printf("Chroma'dan chunks silme hatası: %v (kayıt silindi)", err)
			} else {
				log.Printf("Chroma'dan %s dosyası için chunks silindi", fileID)
			}
		}

		// Delete conversation history
		if err := services.ConversationServiceInstance.DeleteConversationsByFileID(fileID); err != nil {
			log.Printf("Sohbet geçmişi silme hatası: %v (kayıt silindi)", err)
		} else {
			log.Printf("%s dosyası için sohbet geçmişi silindi", fileID)
		}

		return c.JSON(fiber.Map{
			"message": "Dosya başarıyla silindi",
		})
	}
}

// OnlyOfficeFileMapping represents mapping from content type to OnlyOffice types
type OnlyOfficeFileMapping struct {
	FileType     string
	DocumentType string
}

// getOnlyOfficeMapping - Content type'dan OnlyOffice mapping'e çevir
func getOnlyOfficeMapping(contentType string) OnlyOfficeFileMapping {
	contentTypeLower := strings.ToLower(contentType)

	if strings.Contains(contentTypeLower, "wordprocessingml.document") ||
		contentTypeLower == "application/msword" {
		return OnlyOfficeFileMapping{
			FileType:     "docx",
			DocumentType: "word",
		}
	}

	if strings.Contains(contentTypeLower, "spreadsheetml.sheet") ||
		contentTypeLower == "application/vnd.ms-excel" {
		return OnlyOfficeFileMapping{
			FileType:     "xlsx",
			DocumentType: "cell",
		}
	}

	if strings.Contains(contentTypeLower, "presentationml.presentation") ||
		contentTypeLower == "application/vnd.ms-powerpoint" {
		return OnlyOfficeFileMapping{
			FileType:     "pptx",
			DocumentType: "slide",
		}
	}

	if strings.Contains(contentTypeLower, "pdf") {
		return OnlyOfficeFileMapping{
			FileType:     "pdf",
			DocumentType: "pdf",
		}
	}

	return OnlyOfficeFileMapping{
		FileType:     "docx",
		DocumentType: "word",
	}
}

func getOnlyOfficeFileType(contentType string) string {
	return getOnlyOfficeMapping(contentType).FileType
}

func getOnlyOfficeDocumentType(contentType string) string {
	return getOnlyOfficeMapping(contentType).DocumentType
}

func generateDocumentKey(fileID string) string {
	return fmt.Sprintf("%s_%d", fileID, time.Now().Unix())
}

// signOnlyOfficeConfig - OnlyOffice config'i JWT ile imzala
func signOnlyOfficeConfig(payload map[string]interface{}, secret string) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("payload marshal hatası: %v", err)
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signatureInput := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signatureInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := signatureInput + "." + signature

	return token, nil
}

func GetOnlyOfficeConfig(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return middleware.BadRequestResponse(c, err.Error())
		}

		fileID := c.Query("file_id")
		if fileID == "" {
			return middleware.BadRequestResponse(c, "file_id parametresi gerekli")
		}

		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			log.Printf("Dosya bulunamadı: %v", err)
			return middleware.NotFoundResponse(c, "Dosya bulunamadı")
		}

		hasWriteAccess, err := helpers.CheckFileAccessWithOwnerFallback(userID, fileID, file.UserID, helpers.AccessLevelWrite)
		if err != nil {
			log.Printf("Access check hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Erişim kontrolü yapılamadı")
		}
		if !hasWriteAccess {
			return middleware.ForbiddenResponse(c, "Bu dosyayı düzenleme yetkiniz yok")
		}

		user, err := services.UserServiceInstance.GetUserByID(userID)
		userName := userID
		if err == nil && user != nil {
			userName = user.Name
		} else {
			log.Printf("Kullanıcı bulunamadı: %v", err)
		}

		docToken, err := generateOnlyOfficeDocumentToken(fileID, userID, cfg.JWTSecret, time.Hour)
		if err != nil {
			log.Printf("Document token oluşturma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Doküman token'ı oluşturulamadı")
		}

		docURL := fmt.Sprintf("%s/api/v1/files/onlyoffice-document?file_id=%s&token=%s",
			cfg.BackendExternalURL, fileID, docToken)

		docKey := generateDocumentKey(fileID)

		mode := c.Query("mode", "edit")
		if mode != "edit" && mode != "view" {
			mode = "edit"
		}

		var callbackURL string
		if mode == "edit" {
			callbackURL = fmt.Sprintf("%s/api/v1/files/onlyoffice-callback", cfg.BackendExternalURL)
		}

		editorConfig := map[string]interface{}{
			"mode": mode,
			"user": map[string]interface{}{
				"id":   userID,
				"name": userName,
			},
		}

		if mode == "edit" && callbackURL != "" {
			editorConfig["callbackUrl"] = callbackURL
		}

		documentConfig := map[string]interface{}{
			"fileType": getOnlyOfficeFileType(file.ContentType),
			"key":      docKey,
			"title":    file.Filename,
			"url":      docURL,
		}

		config := map[string]interface{}{
			"document":     documentConfig,
			"documentType": getOnlyOfficeDocumentType(file.ContentType),
			"editorConfig": editorConfig,
			"type":         "desktop",
		}

		if cfg.OnlyOfficeJWTSecret != "" {
			// JWT payload should only contain document info
			jwtPayload := map[string]interface{}{
				"document":     documentConfig,
				"documentType": getOnlyOfficeDocumentType(file.ContentType),
				"editorConfig": editorConfig,
			}

			token, err := signOnlyOfficeConfig(jwtPayload, cfg.OnlyOfficeJWTSecret)
			if err != nil {
				log.Printf("JWT token oluşturma hatası: %v", err)
			} else {
				config["token"] = token
			}
		}

		return c.JSON(config)
	}
}

func OnlyOfficeCallback(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Key        string `json:"key"`
			Status     int    `json:"status"`
			URL        string `json:"url,omitempty"`
			ChangesURL string `json:"changesurl,omitempty"`
			History    struct {
				ServerVersion string `json:"serverVersion"`
				Changes       []struct {
					Created string `json:"created"`
					User    struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"user"`
				} `json:"changes"`
			} `json:"history,omitempty"`
			Users   []string `json:"users,omitempty"`
			Actions []struct {
				Type   int    `json:"type"`
				UserID string `json:"userid"`
			} `json:"actions,omitempty"`
			Forcesavetype int `json:"forcesavetype,omitempty"`
		}

		if err := c.BodyParser(&req); err != nil {
			log.Printf("Callback parse hatası: %v", err)
			return c.Status(400).JSON(fiber.Map{"error": 0})
		}

		if req.Status == 2 || req.Status == 6 {
			if req.URL == "" {
				log.Printf("Callback URL boş: key=%s, status=%d", req.Key, req.Status)
				return c.JSON(fiber.Map{"error": 0})
			}

			keyParts := strings.Split(req.Key, "_")
			if len(keyParts) == 0 {
				log.Printf("Geçersiz key formatı: %s", req.Key)
				return c.JSON(fiber.Map{"error": 0})
			}
			fileID := keyParts[0]

			file, err := services.FileServiceInstance.GetFileByID(fileID)
			if err != nil {
				log.Printf("Dosya bulunamadı: %v", err)
				return c.JSON(fiber.Map{"error": 0})
			}

			resp, err := http.Get(req.URL)
			if err != nil {
				log.Printf("OnlyOffice'den dosya indirme hatası: %v", err)
				return c.JSON(fiber.Map{"error": 0})
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("OnlyOffice'den dosya indirme hatası: status=%d", resp.StatusCode)
				return c.JSON(fiber.Map{"error": 0})
			}

			fileContent, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Dosya okuma hatası: %v", err)
				return c.JSON(fiber.Map{"error": 0})
			}

			objectName := file.MinioPath
			ctx := context.Background()

			_, err = services.MinioService.Client.PutObject(
				ctx,
				"user-files",
				objectName,
				bytes.NewReader(fileContent),
				int64(len(fileContent)),
				minio.PutObjectOptions{
					ContentType: file.ContentType,
				},
			)
			if err != nil {
				log.Printf("MinIO'ya yükleme hatası: %v", err)
				return c.JSON(fiber.Map{"error": 0})
			}

			// Update file metadata (size, updated_at)
			updates := map[string]interface{}{
				"size": int64(len(fileContent)),
			}
			if err := services.FileServiceInstance.UpdateFileRecord(fileID, updates); err != nil {
				log.Printf("Dosya metadata güncelleme hatası: %v", err)
			}

			log.Printf("Dosya başarıyla kaydedildi: fileID=%s, size=%d", fileID, len(fileContent))
		}

		return c.JSON(fiber.Map{"error": 0})
	}
}

func generateOnlyOfficeDocumentToken(fileID, userID, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"file_id": fileID,
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "onlyoffice_document",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateOnlyOfficeDocumentToken(tokenString, secret string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("beklenmeyen signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "onlyoffice_document" {
			return "", "", fmt.Errorf("geçersiz token tipi")
		}

		fileID, ok1 := claims["file_id"].(string)
		userID, ok2 := claims["user_id"].(string)

		if !ok1 || !ok2 {
			return "", "", fmt.Errorf("token'da file_id veya user_id bulunamadı")
		}

		return fileID, userID, nil
	}

	return "", "", fmt.Errorf("geçersiz token")
}

func GetOnlyOfficeDocument(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileID := c.Query("file_id")
		tokenStr := c.Query("token")

		if fileID == "" || tokenStr == "" {
			return middleware.BadRequestResponse(c, "file_id ve token parametreleri gerekli")
		}

		validatedFileID, userID, err := validateOnlyOfficeDocumentToken(tokenStr, cfg.JWTSecret)
		if err != nil {
			log.Printf("Token validation hatası: %v", err)
			return middleware.UnauthorizedResponse(c, "Geçersiz veya süresi dolmuş token")
		}

		if validatedFileID != fileID {
			return middleware.ForbiddenResponse(c, "Token ve file_id eşleşmiyor")
		}

		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			log.Printf("Dosya bulunamadı: %v", err)
			return middleware.NotFoundResponse(c, "Dosya bulunamadı")
		}

		hasReadAccess, err := helpers.CheckFileAccessWithOwnerFallback(userID, fileID, file.UserID, helpers.AccessLevelRead)
		if err != nil {
			log.Printf("Access check hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Erişim kontrolü yapılamadı")
		}
		if !hasReadAccess {
			return middleware.ForbiddenResponse(c, "Bu dosyaya erişim yetkiniz yok")
		}

		objectName := file.MinioPath
		if objectName == "" {
			objectName = services.MinioService.GetUserFilePath(file.UserID, file.Filename)
		}
		ctx := context.Background()

		objInfo, err := services.MinioService.Client.StatObject(ctx, "user-files", objectName, minio.StatObjectOptions{})
		if err != nil {
			log.Printf("Dosya MinIO'da bulunamadı: %v (objectName: %s)", err, objectName)
			return middleware.NotFoundResponse(c, "Dosya bulunamadı")
		}

		object, err := services.MinioService.Client.GetObject(ctx, "user-files", objectName, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("MinIO'dan dosya okuma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Dosya okunamadı")
		}
		defer object.Close()

		c.Set("Content-Type", file.ContentType)
		c.Set("Content-Length", fmt.Sprintf("%d", objInfo.Size))
		c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Filename))
		c.Set("Cache-Control", "private, max-age=3600")

		_, err = io.Copy(c.Response().BodyWriter(), object)
		if err != nil {
			log.Printf("Dosya stream hatası: %v", err)
			return err
		}

		return nil
	}
}

// GetFileContent - Kod dosyası içeriğini text olarak döndür
func GetFileContent(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return middleware.BadRequestResponse(c, err.Error())
		}

		fileID := c.Query("file_id")
		if fileID == "" {
			return middleware.BadRequestResponse(c, "file_id parametresi gerekli")
		}

		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			log.Printf("Dosya bulunamadı: %v", err)
			return middleware.NotFoundResponse(c, "Dosya bulunamadı")
		}

		hasReadAccess, err := helpers.CheckFileAccessWithOwnerFallback(userID, fileID, file.UserID, helpers.AccessLevelRead)
		if err != nil {
			log.Printf("Access check hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Erişim kontrolü yapılamadı")
		}
		if !hasReadAccess {
			return middleware.ForbiddenResponse(c, "Bu dosyaya erişim yetkiniz yok")
		}

		objectName := file.MinioPath
		if objectName == "" {
			objectName = services.MinioService.GetUserFilePath(file.UserID, file.Filename)
		}
		ctx := context.Background()

		object, err := services.MinioService.Client.GetObject(ctx, "user-files", objectName, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("MinIO'dan dosya okuma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Dosya okunamadı")
		}
		defer object.Close()

		fileContent, err := io.ReadAll(object)
		if err != nil {
			log.Printf("Dosya içeriği okuma hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Dosya içeriği okunamadı")
		}

		c.Set("Content-Type", "text/plain; charset=utf-8")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		return c.Send(fileContent)
	}
}

// UpdateFileContent - Kod dosyası içeriğini güncelle
func UpdateFileContent(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := helpers.GetCurrentUserID(c)
		if err != nil {
			return middleware.BadRequestResponse(c, err.Error())
		}

		fileID := c.Params("id")
		if fileID == "" {
			return middleware.BadRequestResponse(c, "file_id parametresi gerekli")
		}

		file, err := services.FileServiceInstance.GetFileByID(fileID)
		if err != nil {
			log.Printf("Dosya bulunamadı: %v", err)
			return middleware.NotFoundResponse(c, "Dosya bulunamadı")
		}

		hasWriteAccess, err := helpers.CheckFileAccessWithOwnerFallback(userID, fileID, file.UserID, helpers.AccessLevelWrite)
		if err != nil {
			log.Printf("Access check hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Erişim kontrolü yapılamadı")
		}
		if !hasWriteAccess {
			return middleware.ForbiddenResponse(c, "Bu dosyayı düzenleme yetkiniz yok")
		}

		var req struct {
			Content string `json:"content"`
		}
		if err := c.BodyParser(&req); err != nil {
			return middleware.BadRequestResponse(c, "Geçersiz istek verisi")
		}

		objectName := file.MinioPath
		if objectName == "" {
			objectName = services.MinioService.GetUserFilePath(file.UserID, file.Filename)
		}
		ctx := context.Background()

		fileContent := []byte(req.Content)
		_, err = services.MinioService.Client.PutObject(
			ctx,
			"user-files",
			objectName,
			bytes.NewReader(fileContent),
			int64(len(fileContent)),
			minio.PutObjectOptions{
				ContentType: file.ContentType,
			},
		)
		if err != nil {
			log.Printf("MinIO'ya yükleme hatası: %v", err)
			return middleware.InternalServerErrorResponse(c, "Dosya güncellenemedi")
		}

		updates := map[string]interface{}{
			"size": int64(len(fileContent)),
		}
		if err := services.FileServiceInstance.UpdateFileRecord(fileID, updates); err != nil {
			log.Printf("Dosya metadata güncelleme hatası: %v", err)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Dosya başarıyla güncellendi",
			"size":    len(fileContent),
		})
	}
}
