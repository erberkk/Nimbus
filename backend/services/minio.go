package services

import (
	"context"
	"fmt"
	"log"
	"nimbus-backend/config"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	Client *minio.Client
	Config *config.Config
}

var MinioService *MinIOService

// Güvenlik için izin verilen dosya türleri
var AllowedMimeTypes = map[string][]string{
	"image": {
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp",
		"image/bmp", "image/tiff", "image/svg+xml",
	},
	"document": {
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain", "text/csv",
		"application/rtf",
	},
	"archive": {
		"application/zip", "application/x-rar-compressed",
		"application/x-7z-compressed", "application/gzip",
	},
	"audio": {
		"audio/mpeg",  // MP3
		"audio/wav",   // WAV
		"audio/flac",  // FLAC
		"audio/aac",   // AAC
		"audio/ogg",   // OGG
		"audio/mp4",   // M4A
		"audio/x-m4a", // M4A
	},
	"video": {
		"video/mp4",        // MP4
		"video/avi",        // AVI
		"video/quicktime",  // MOV
		"video/x-msvideo",  // AVI
		"video/x-ms-wmv",   // WMV
		"video/webm",       // WebM
		"video/x-matroska", // MKV
	},
}

// Güvenlik için maksimum dosya boyutu (100MB)
const MaxFileSize = 100 * 1024 * 1024

// Tehlikeli dosya uzantıları
var BlockedExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar",
	".msi", ".dll", ".so", ".dylib", ".deb", ".rpm", ".apk",
	".sh", ".ps1", ".py", ".pl", ".rb", ".php", ".asp", ".jsp",
}

func InitMinIO(cfg *config.Config) error {
	// MinIO endpoint'ini environment'dan al veya default kullan
	endpoint := cfg.MinIOEndpoint
	if endpoint == "" {
		endpoint = "localhost:9000" // Default MinIO endpoint
	}

	accessKey := cfg.MinIOAccessKey
	if accessKey == "" {
		accessKey = "minioadmin" // Default MinIO access key
	}

	secretKey := cfg.MinIOSecretKey
	if secretKey == "" {
		secretKey = "minioadmin" // Default MinIO secret key
	}

	useSSL := cfg.MinIOUseSSL
	if !useSSL {
		useSSL = false // Default HTTP for development
	}

	// MinIO client oluştur
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("MinIO client oluşturma hatası: %v", err)
	}

	MinioService = &MinIOService{
		Client: client,
		Config: cfg,
	}

	// Bucket oluştur/kontrol et
	if err := MinioService.CreateBucketIfNotExists("user-files"); err != nil {
		return fmt.Errorf("bucket oluşturma hatası: %v", err)
	}

	log.Println("✅ MinIO servisi başlatıldı!")
	return nil
}

// Bucket oluştur veya varlığını kontrol et
func (m *MinIOService) CreateBucketIfNotExists(bucketName string) error {
	ctx := context.Background()

	exists, err := m.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("✅ Bucket '%s' oluşturuldu", bucketName)
	} else {
		log.Printf("✅ Bucket '%s' zaten mevcut", bucketName)
	}

	return nil
}

// Güvenlik taraması
func (m *MinIOService) ValidateFile(filename string, contentType string, size int64) error {
	// Dosya boyutu kontrolü
	if size > MaxFileSize {
		return fmt.Errorf("dosya boyutu çok büyük: maksimum %d MB", MaxFileSize/(1024*1024))
	}

	// Tehlikeli uzantı kontrolü
	ext := strings.ToLower(filepath.Ext(filename))
	for _, blocked := range BlockedExtensions {
		if ext == blocked {
			return fmt.Errorf("bu dosya türü desteklenmiyor: %s", ext)
		}
	}

	// MIME type kontrolü
	allowed := false
	for _, types := range AllowedMimeTypes {
		for _, allowedType := range types {
			if contentType == allowedType {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}

	if !allowed {
		return fmt.Errorf("desteklenmeyen dosya türü: %s", contentType)
	}

	return nil
}

// Kullanıcı dosyası için path oluştur
func (m *MinIOService) GetUserFilePath(userID, filename string) string {
	// Güvenli dosya adı oluştur (sanitization)
	safeFilename := strings.ReplaceAll(filename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "..", "_")

	return fmt.Sprintf("user-%s/%s", userID, safeFilename)
}

// Upload için presigned URL oluştur
func (m *MinIOService) GenerateUploadPresignedURL(userID, filename string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	// Dosya path'i oluştur
	objectName := m.GetUserFilePath(userID, filename)

	// Presigned PUT URL oluştur
	presignedURL, err := m.Client.PresignedPutObject(ctx, "user-files", objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("presigned URL oluşturma hatası: %v", err)
	}

	return presignedURL.String(), nil
}

// Download için presigned URL oluştur
func (m *MinIOService) GenerateDownloadPresignedURL(userID, filename string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	// Dosya path'i oluştur
	objectName := m.GetUserFilePath(userID, filename)

	// Dosya varlığını kontrol et
	_, err := m.Client.StatObject(ctx, "user-files", objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("dosya bulunamadı: %v", err)
	}

	// Presigned GET URL oluştur
	presignedURL, err := m.Client.PresignedGetObject(ctx, "user-files", objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned URL oluşturma hatası: %v", err)
	}

	return presignedURL.String(), nil
}

// Dosya bilgilerini al
func (m *MinIOService) GetFileInfo(userID, filename string) (*minio.ObjectInfo, error) {
	ctx := context.Background()

	objectName := m.GetUserFilePath(userID, filename)

	info, err := m.Client.StatObject(ctx, "user-files", objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("dosya bilgisi alınamadı: %v", err)
	}

	return &info, nil
}

// Dosyaları listele
func (m *MinIOService) ListUserFiles(userID string) ([]minio.ObjectInfo, error) {
	ctx := context.Background()

	// User prefix ile listele
	prefix := fmt.Sprintf("user-%s/", userID)

	var files []minio.ObjectInfo
	for object := range m.Client.ListObjects(ctx, "user-files", minio.ListObjectsOptions{
		Prefix: prefix,
	}) {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object)
	}

	return files, nil
}

// DeleteFile - MinIO'dan dosya sil
func (m *MinIOService) DeleteFile(objectName string) error {
	ctx := context.Background()

	err := m.Client.RemoveObject(ctx, "user-files", objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("dosya silinemedi: %v", err)
	}

	return nil
}
