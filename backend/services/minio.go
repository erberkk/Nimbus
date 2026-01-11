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
	"code": {
		"text/plain",
		"text/x-python",
		"text/x-java",
		"text/x-csharp",
		"text/x-c++",
		"text/x-c",
		"text/javascript",
		"application/javascript",
		"text/typescript",
		"application/typescript",
		"application/json",
		"text/json",
		"text/markdown",
		"text/x-markdown",
		"text/xml",
		"application/xml",
		"text/css",
		"text/html",
		"application/x-sh",
		"text/x-sh",
		"text/x-kotlin",
		"text/x-scala",
		"text/x-go",
		"text/x-rust",
		"text/x-php",
		"text/x-ruby",
		"text/x-perl",
		"application/x-yaml",
		"text/yaml",
		"text/x-yaml",
		"text/x-sql",
		"application/x-sql",
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

// Tehlikeli dosya uzantıları (sadece executable'lar)
var BlockedExtensions = []string{
	".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs",
	".msi", ".dll", ".so", ".dylib", ".deb", ".rpm", ".apk",
	".jar",
}

func InitMinIO(cfg *config.Config) error {
	endpoint := cfg.MinIOEndpoint
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	accessKey := cfg.MinIOAccessKey
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := cfg.MinIOSecretKey
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	useSSL := cfg.MinIOUseSSL
	if !useSSL {
		useSSL = false
	}

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
	if size > MaxFileSize {
		return fmt.Errorf("dosya boyutu çok büyük: maksimum %d MB", MaxFileSize/(1024*1024))
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, blocked := range BlockedExtensions {
		if ext == blocked {
			return fmt.Errorf("bu dosya türü desteklenmiyor: %s", ext)
		}
	}

	effectiveContentType := contentType
	if contentType == "application/octet-stream" || contentType == "" {
		effectiveContentType = m.GuessMimeTypeFromExtension(ext)
	}

	// MIME type kontrolü
	allowed := false
	for _, types := range AllowedMimeTypes {
		for _, allowedType := range types {
			if effectiveContentType == allowedType || strings.HasPrefix(effectiveContentType, allowedType) {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}

	if !allowed {
		codeExtensions := []string{".py", ".js", ".jsx", ".ts", ".tsx", ".cs", ".java", ".kt", ".kts",
			".json", ".md", ".txt", ".xml", ".html", ".css", ".sh", ".bash", ".yaml", ".yml",
			".go", ".rs", ".php", ".rb", ".pl", ".scala", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp",
			".sql", ".vue", ".svelte", ".swift", ".dart", ".lua", ".r", ".m", ".mm", ".ps1"}
		extLower := strings.ToLower(ext)
		for _, codeExt := range codeExtensions {
			if extLower == codeExt {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return fmt.Errorf("desteklenmeyen dosya türü: %s (uzantı: %s)", effectiveContentType, ext)
	}

	return nil
}

func (m *MinIOService) GuessMimeTypeFromExtension(ext string) string {
	mimeTypes := map[string]string{
		// Documents
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".rtf":  "application/rtf",
		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".svg":  "image/svg+xml",
		// Code files
		".py":     "text/x-python",
		".js":     "application/javascript",
		".jsx":    "application/javascript",
		".ts":     "application/typescript",
		".tsx":    "application/typescript",
		".json":   "application/json",
		".xml":    "application/xml",
		".html":   "text/html",
		".css":    "text/css",
		".md":     "text/markdown",
		".yaml":   "application/x-yaml",
		".yml":    "application/x-yaml",
		".sh":     "application/x-sh",
		".bash":   "application/x-sh",
		".go":     "text/x-go",
		".rs":     "text/x-rust",
		".java":   "text/x-java",
		".kt":     "text/x-kotlin",
		".kts":    "text/x-kotlin",
		".cs":     "text/x-csharp",
		".php":    "text/x-php",
		".rb":     "text/x-ruby",
		".pl":     "text/x-perl",
		".scala":  "text/x-scala",
		".c":      "text/x-c",
		".cpp":    "text/x-c++",
		".cc":     "text/x-c++",
		".cxx":    "text/x-c++",
		".h":      "text/x-c",
		".hpp":    "text/x-c++",
		".sql":    "application/x-sql",
		".vue":    "text/plain",
		".svelte": "text/plain",
		".swift":  "text/plain",
		".dart":   "text/plain",
		".lua":    "text/plain",
		".r":      "text/plain",
		".m":      "text/plain",
		".mm":     "text/plain",
		".ps1":    "text/plain",
		// Archives
		".zip": "application/zip",
		".rar": "application/x-rar-compressed",
		".7z":  "application/x-7z-compressed",
		".gz":  "application/gzip",
		// Audio
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		// Video
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".webm": "video/webm",
		".mkv":  "video/x-matroska",
	}

	extLower := strings.ToLower(ext)
	if mimeType, ok := mimeTypes[extLower]; ok {
		return mimeType
	}

	return "text/plain"
}

// Kullanıcı dosyası için path oluştur
func (m *MinIOService) GetUserFilePath(userID, filename string) string {
	safeFilename := strings.ReplaceAll(filename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "..", "_")

	return fmt.Sprintf("user-%s/%s", userID, safeFilename)
}

// Upload için presigned URL oluştur
func (m *MinIOService) GenerateUploadPresignedURL(userID, filename string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	objectName := m.GetUserFilePath(userID, filename)

	presignedURL, err := m.Client.PresignedPutObject(ctx, "user-files", objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("presigned URL oluşturma hatası: %v", err)
	}

	return presignedURL.String(), nil
}

// Download için presigned URL oluştur
func (m *MinIOService) GenerateDownloadPresignedURL(userID, filename string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	objectName := m.GetUserFilePath(userID, filename)

	_, err := m.Client.StatObject(ctx, "user-files", objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("dosya bulunamadı: %v", err)
	}

	presignedURL, err := m.Client.PresignedGetObject(ctx, "user-files", objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned URL oluşturma hatası: %v", err)
	}

	return presignedURL.String(), nil
}

// Download için presigned URL oluştur (External endpoint ile - OnlyOffice gibi external servislere için)
func (m *MinIOService) GenerateDownloadPresignedURLExternal(userID, filename string, expiry time.Duration, externalEndpoint string) (string, error) {
	ctx := context.Background()

	objectName := m.GetUserFilePath(userID, filename)

	_, err := m.Client.StatObject(ctx, "user-files", objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("dosya bulunamadı: %v", err)
	}

	presignedURL, err := m.Client.PresignedGetObject(ctx, "user-files", objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned URL oluşturma hatası: %v", err)
	}

	urlStr := presignedURL.String()

	if externalEndpoint != "" && m.Config.MinIOEndpoint != "" {
		internalEndpoint := m.Config.MinIOEndpoint
		if m.Config.MinIOUseSSL {
			urlStr = strings.Replace(urlStr, "https://"+internalEndpoint, "http://"+externalEndpoint, 1)
		} else {
			urlStr = strings.Replace(urlStr, "http://"+internalEndpoint, "http://"+externalEndpoint, 1)
		}
	}

	return urlStr, nil
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
