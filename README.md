# Nimbus - Dropbox Benzeri Bulut Depolama Sistemi

Nimbus, modern web teknolojileri kullanılarak geliştirilmiş, Dropbox benzeri bir bulut depolama ve dosya paylaşım platformudur.

## 🚀 Özellikler

- **Google OAuth 2.0 Kimlik Doğrulama**: Güvenli ve kolay giriş sistemi
- **JWT Token Yönetimi**: Stateless authentication
- **MongoDB Veritabanı**: Kullanıcı ve dosya bilgileri için NoSQL veritabanı
- **MinIO Object Storage**: Dosya depolama için S3 uyumlu object storage
- **Güvenlik Taramalı Dosya Yönetimi**: Tehlikeli dosya türleri ve boyut kontrolleri
- **Çoklu Dil Desteği (i18n)**: Türkçe ve İngilizce dil seçenekleri
- **Toast Bildirim Sistemi**: Modern kullanıcı geri bildirimi
- **Modern UI Tasarımı**: Material-UI ve Framer Motion animasyonları
- **Responsive Tasarım**: Mobil ve desktop uyumlu arayüz
- **RESTful API**: Temiz ve ölçeklenebilir API tasarımı

## 🛠️ Teknoloji Stack'i

### Backend
- **Go**: Yüksek performanslı backend dili
- **Fiber**: Express.js benzeri hızlı web framework
- **MongoDB**: NoSQL veritabanı
- **JWT**: Token tabanlı kimlik doğrulama

### Frontend
- **React**: Modern kullanıcı arayüzü kütüphanesi
- **Vite**: Hızlı geliştirme ve build aracı
- **Tailwind CSS**: Utility-first CSS framework

## 📁 Proje Yapısı

```
nimbus/
├── backend/           # Go Fiber backend
│   ├── config/       # Konfigürasyon yönetimi
│   ├── database/     # Veritabanı bağlantıları
│   ├── handlers/     # HTTP request handlers
│   ├── middleware/   # JWT ve diğer middleware'ler
│   ├── models/       # Veri modelleri
│   └── routes/       # API route tanımları
├── frontend/         # React Vite frontend
│   ├── src/
│   │   ├── components/  # React bileşenleri
│   │   ├── hooks/       # Custom React hooks
│   │   ├── services/    # API servisleri
│   │   └── pages/       # Sayfa bileşenleri
└── README.md
```

## 🏃‍♂️ Kurulum ve Çalıştırma

### Gereksinimler
- Go 1.19+
- Node.js 20+
- MongoDB
- MinIO (isteğe bağlı, dosya depolama için)

### 1. Backend Kurulumu

```bash
cd backend

# Bağımlılıkları yükle
go mod tidy

# Environment variables (.env dosyası oluşturun)
cp .env.example .env
# .env dosyasını düzenleyin

# Sunucuyu başlat
go run main.go
```

Backend `http://localhost:8080` adresinde çalışacak.

### 2. Frontend Kurulumu

```bash
cd frontend

# Bağımlılıkları yükle
npm install

# Environment variables (.env dosyası oluşturun)
echo "VITE_API_URL=http://localhost:8080/api/v1" > .env

# Geliştirme sunucusunu başlat
npm run dev
```

Frontend `http://localhost:5173` adresinde çalışacak.

### 3. Google OAuth Kurulumu

1. [Google Cloud Console](https://console.cloud.google.com/) gidin
2. Yeni proje oluşturun veya mevcut projeyi seçin
3. APIs & Services > Credentials bölümüne gidin
4. OAuth 2.0 Client IDs oluşturun
5. Application type: Web application seçin
6. Authorized redirect URIs'e şunu ekleyin:
   ```
   http://localhost:8080/auth/google/callback
   ```
7. Client ID ve Client Secret'i backend `.env` dosyasına ekleyin:
   ```env
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

## 🔧 Environment Variables

### Backend (.env)
```env
# Server Configuration
PORT=8080

# Database Configuration
MONGO_URI=mongodb://localhost:27017
MONGO_DB=nimbus

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# Frontend Configuration
FRONTEND_URL=http://localhost:5173
```

### Frontend (.env)
```env
VITE_API_URL=http://localhost:8080/api/v1
```

## 🌟 Kullanım

1. Backend ve frontend sunucularını başlatın
2. Tarayıcıda `http://localhost:5173` adresine gidin
3. "Google ile Giriş Yap" butonuna tıklayın
4. Google hesabınızla giriş yapın
5. Dashboard'a yönlendirileceksiniz

### Dosya Yönetimi Kullanımı

1. **Dosya Yükleme:**
   - Dashboard'daki "Dosya Yükleme" alanını kullanın
   - Sürükle-bırak veya tıklayarak dosya seçin
   - Desteklenen türler: Resimler, PDF, Word, Excel, PPT, TXT, ZIP

2. **Dosya Listesi:**
   - "Dosyalarım" bölümünden yüklediğiniz dosyaları görün
   - Her dosyanın boyutu ve türü gösterilir
   - İndirme butonu ile dosyaları indirebilirsiniz

## 🔒 Güvenlik

- JWT token'ları httpOnly cookie'lerde saklanır
- OAuth state parametresi CSRF saldırılarını önler
- CORS konfigürasyonu sadece güvenilir origin'lere izin verir
- Input validation ve sanitization uygulanır

### Dosya Güvenliği

- **Dosya Türü Kontrolü**: Sadece güvenli dosya türleri kabul edilir (resimler, PDF, Word, Excel, PPT, TXT, ZIP)
- **Boyut Limiti**: Maksimum 100MB dosya boyutu
- **Uzantı Engelleme**: Tehlikeli dosya uzantıları (.exe, .bat, .js, .php vs.) otomatik engellenir
- **MIME Type Doğrulama**: İçerik türü doğrulaması
- **User-Based Storage**: Her kullanıcının kendi klasörü (`user-{user_id}/`)
- **Presigned URL'ler**: Güvenli ve zaman sınırlı erişim
- CORS konfigürasyonu sadece güvenilir origin'lere izin verir
- Input validation ve sanitization uygulanır

## 🚧 Geliştirme Notları

- Proje adım adım geliştirilmekte olup, dosya yükleme, paylaşım ve diğer özellikler sonraki adımlarda eklenecektir
- Kod kalitesi için ESLint ve Go fmt kullanılmaktadır
- Responsive tasarım Tailwind CSS ile sağlanmaktadır

## 📝 Lisans

Bu proje eğitim amaçlı geliştirilmekte olup, ticari kullanım için uygun değildir.
