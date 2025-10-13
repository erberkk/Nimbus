# Nimbus Frontend

Nimbus projesinin React Vite frontend uygulaması.

## Kurulum

### Gereksinimler

- Node.js 20+
- npm veya yarn

### Bağımlılıkları Yükleme

```bash
npm install
```

### Environment Variables

`.env` dosyası oluşturun:

```env
# =============================================================================
# API CONFIGURATION
# =============================================================================
# Backend API'nin URL'i
VITE_API_URL=http://localhost:8080/api/v1
```

#### ⚠️ Önemli Notlar:

1. **Development için**:
   - `VITE_API_URL` backend'in çalıştığı adrese işaret etmeli
   - Varsayılan: `http://localhost:8080/api/v1`

2. **Production için**:
   ```env
   VITE_API_URL=https://your-api-domain.com/api/v1
   ```

### Geliştirme Sunucusu

```bash
npm run dev
```

### Prodüksiyon Build

```bash
npm run build
```

## Kullanım

1. Backend sunucusunun çalıştığından emin olun (`http://localhost:8080`)
2. Frontend geliştirme sunucusunu başlatın (`http://localhost:5173`)
3. Google OAuth için Google Cloud Console'da gerekli ayarları yapın

### Dil Değiştirme

- Navbar'daki 🌐 ikonuna tıklayarak dil seçin
- Türkçe ve İngilizce arasında geçiş yapın
- Tüm arayüz otomatik olarak seçilen dile geçer

### Toast Bildirimleri

- Başarı, hata, uyarı ve bilgi mesajları otomatik gösterilir
- Sağ üst köşede belirir ve otomatik kapanır
- Manuel olarak da kapatılabilir

## Özellikler

- Google OAuth 2.0 ile kimlik doğrulama
- JWT token yönetimi
- Modern UI tasarımı (Material-UI / MUI)
- Framer Motion animasyonları
- React Router ile sayfa yönetimi
- Çoklu dil desteği (Türkçe/İngilizce)
- Toast bildirim sistemi
- Dosya yükleme ve yönetimi (MinIO entegrasyonu)
- Güvenlik taramalı dosya yükleme
- Responsive tasarım
- Modern React hooks kullanımı

## Kullanılan Paketler

- **@mui/material** - Material-UI component kütüphanesi
- **@mui/icons-material** - Material Icons
- **framer-motion** - Animasyon kütüphanesi
- **react-router-dom** - Routing
- **react-i18next** - Çoklu dil desteği
- **@emotion/react & @emotion/styled** - MUI için CSS-in-JS

## Dosya Yönetimi Özellikleri

### Güvenlik

- ✅ **Dosya türü kontrolü** - Sadece güvenli dosya türleri kabul edilir
- ✅ **Boyut limiti** - Maksimum 100MB dosya boyutu
- ✅ **Uzantı engelleme** - Tehlikeli dosya uzantıları (.exe, .bat, .js vs.) engellenir
- ✅ **MIME type validation** - İçerik türü doğrulaması

### Kullanıcı Deneyimi

- ✅ **Sürükle-bırak** yükleme
- ✅ **İlerleme çubuğu** - Gerçek zamanlı yükleme takibi
- ✅ **Dosya önizleme** - İkonlar ve tür etiketleri
- ✅ **Çoklu dil desteği** - Türkçe/İngilizce geçiş
- ✅ **Toast bildirimleri** - Modern kullanıcı geri bildirimi
- ✅ **Hata yönetimi** - Kullanıcı dostu hata mesajları
- ✅ **Başarı bildirimleri** - Yükleme tamamlanma onayları

### Teknik Özellikler

- ✅ **Presigned URL'ler** - Güvenli upload/download
- ✅ **User-based storage** - Her kullanıcının kendi klasörü
- ✅ **Responsive tasarım** - Mobil uyumlu arayüz
