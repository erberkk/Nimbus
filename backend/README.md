# Nimbus Backend

Nimbus projesinin Go Fiber backend servisi.

## Kurulum

### Gereksinimler
- Go 1.19+
- MongoDB

### Bağımlılıkları Yükleme
```bash
go mod tidy
```

### Environment Variables

`.env` dosyası oluşturun ve aşağıdaki değerleri ayarlayın:

```env
# =============================================================================
# SERVER CONFIGURATION
# =============================================================================
# Sunucu portu
PORT=8080

# =============================================================================
# DATABASE CONFIGURATION
# =============================================================================
# MongoDB bağlantı URI'si
MONGO_URI=mongodb://localhost:27017

# Kullanılacak veritabanı adı
MONGO_DB=nimbus

# =============================================================================
# JWT CONFIGURATION
# =============================================================================
# JWT token imzası için kullanılacak secret key (güçlü bir key kullanın)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# =============================================================================
# GOOGLE OAUTH CONFIGURATION (ZORUNLU)
# =============================================================================
# Google Cloud Console'dan aldığınız Client ID
GOOGLE_CLIENT_ID=your-google-oauth-client-id

# Google Cloud Console'dan aldığınız Client Secret
GOOGLE_CLIENT_SECRET=your-google-oauth-client-secret

# OAuth callback URL'i (Google Cloud Console'da aynı olmalı)
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# =============================================================================
# MINIO OBJECT STORAGE CONFIGURATION
# =============================================================================
# MinIO server endpoint (localhost:9000 for local MinIO)
MINIO_ENDPOINT=localhost:9000

# MinIO access credentials
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# MinIO SSL kullanımı (development için false)
MINIO_USE_SSL=false

# =============================================================================
# FRONTEND CONFIGURATION
# =============================================================================
# Frontend uygulamanızın URL'i
FRONTEND_URL=http://localhost:5173
```

#### ⚠️ Önemli Notlar:

1. **Google OAuth için**:
   - Google Cloud Console'da OAuth 2.0 Client ID oluşturun
   - `GOOGLE_CLIENT_ID` ve `GOOGLE_CLIENT_SECRET` değerlerini gerçek değerlerle değiştirin
   - `GOOGLE_REDIRECT_URL` Google Cloud Console'da tanımladığınız redirect URI ile aynı olmalı

2. **JWT Secret için**:
   - Production ortamında güçlü ve rastgele bir değer kullanın (en az 32 karakter)
   - Aynı secret'i tüm sunucularınızda kullanın

3. **Production ayarları için**:
   ```env
   JWT_SECRET=super-secure-production-jwt-secret-key-minimum-32-chars
   FRONTEND_URL=https://yourdomain.com
   ```

#### 🚀 Hızlı Başlangıç (Development)

`.env` dosyası oluşturmak için:

**Adım 1:** Backend klasörüne gidin
```bash
cd backend
```

**Adım 2:** .env dosyasını oluşturun (aşağıdaki içeriği kopyalayın)
```bash
cat > .env << 'EOF'
# Nimbus Backend Environment Variables
PORT=8080
MONGO_URI=mongodb://localhost:27017
MONGO_DB=nimbus
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
GOOGLE_CLIENT_ID=your-actual-google-client-id
GOOGLE_CLIENT_SECRET=your-actual-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
FRONTEND_URL=http://localhost:5173
EOF
```

**Adım 3:** Google OAuth bilgilerini ekleyin
- Google Cloud Console'da OAuth 2.0 Client ID oluşturun
- `GOOGLE_CLIENT_ID` ve `GOOGLE_CLIENT_SECRET` alanlarını gerçek değerlerle değiştirin

**Adım 4:** Backend'i başlatın
```bash
go run main.go
```

### Çalıştırma
```bash
go run main.go
```

## API Endpoints

### Auth
- `GET /api/v1/auth/google` - Google OAuth login
- `GET /api/v1/auth/google/callback` - Google OAuth callback
- `POST /api/v1/auth/logout` - Logout

### User (Protected)
- `GET /api/v1/user/profile` - Kullanıcı profili

### Files (Protected)

**Desteklenen Dosya Türleri:**
- **Resimler:** PNG, JPEG, JPG, GIF, WebP, BMP, TIFF, SVG
- **Belgeler:** PDF, Word (.doc, .docx), Excel (.xls, .xlsx), PowerPoint (.ppt, .pptx), TXT, CSV, RTF
- **Arşivler:** ZIP, RAR, 7Z, GZIP
- **Maksimum Boyut:** 100MB

**API Endpoints:**
- `GET /api/v1/files/upload-url?filename=dosya.pdf&content_type=application/pdf` - Upload için presigned URL al
- `GET /api/v1/files/download-url?filename=dosya.pdf` - Download için presigned URL al
- `GET /api/v1/files/list` - Kullanıcının dosyalarını listele
- `GET /api/v1/files/:filename` - Dosya bilgilerini al

## MinIO Kurulumu (İsteğe Bağlı)

MinIO object storage kullanmak için:

### 1. MinIO'yu İndirin ve Başlatın

**Docker ile (Önerilen):**
```bash
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  quay.io/minio/minio server /data --console-address ":9001"
```

**Erişim:** http://localhost:9001 (Web UI)

### 2. Bucket Oluşturun

MinIO Web UI'den veya mc komutuyla `user-files` bucket'ını oluşturun:
```bash
mc alias set myminio http://localhost:9000 minioadmin minioadmin
mc mb myminio/user-files
```

### 3. Environment Variables'ı Güncelleyin

`.env` dosyasına MinIO konfigürasyonunu ekleyin:
```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
```

## Google OAuth Kurulumu

### 1. Google Cloud Console Kurulumu

1. [Google Cloud Console](https://console.cloud.google.com/) gidin
2. Yeni proje oluşturun veya mevcut projeyi seçin
3. APIs & Services > Credentials bölümüne gidin
4. **"CREATE CREDENTIALS"** > **"OAuth 2.0 Client IDs"** seçin
5. Application type olarak **"Web application"** seçin

### 2. Authorized JavaScript Origins (Frontend için)

**"Authorized JavaScript origins"** bölümüne şunu ekleyin:
```
http://localhost:5173
```

*Bu, frontend'inizin çalıştığı origin'dir.*

### 3. Authorized Redirect URIs (Backend için)

**"Authorized redirect URIs"** bölümüne şunu ekleyin:
```
http://localhost:8080/auth/google/callback
```

*Bu, Google'dan gelen OAuth callback'inin gideceği URL'dir.*

### 4. Production İçin Ekleyin

Production ortamında şu URL'leri de ekleyin:
```
https://yourdomain.com
https://yourdomain.com/auth/google/callback
```

### 5. Client ID ve Secret'i Alın

- Oluşturulan Client ID ve Client Secret'i kopyalayın
- Backend `.env` dosyasına ekleyin:
```env
GOOGLE_CLIENT_ID=your-actual-client-id
GOOGLE_CLIENT_SECRET=your-actual-client-secret
```
