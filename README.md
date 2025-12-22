# Nimbus - AI Destekli Bulut Depolama ve Doküman Analiz Sistemi

Nimbus, modern web teknolojileri ve yapay zeka kullanılarak geliştirilmiş, Dropbox benzeri bir bulut depolama platformudur. RAG (Retrieval-Augmented Generation) teknolojisi ile dokümanlarınızı yükleyip, içeriklerine doğal dil ile sorular sorabilirsiniz.

## 🚀 Özellikler

### Temel Özellikler
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

### 🤖 AI ve RAG Özellikleri
- **RAG (Retrieval-Augmented Generation)**: Dokümanlarınıza doğal dil ile sorular sorun
- **Semantic Search**: Embedding model ile anlamsal benzerlik araması
- **Hybrid Search**: Semantic + Keyword arama kombinasyonu
- **Intent Classification**: Sorgu niyetini otomatik algılama (karşılaştırma, tanım, özet vb.)
- **Adaptive Retrieval**: Sorgu tipine göre dinamik top-k seçimi
- **Document Processing**: PDF ve DOCX dosyalarını otomatik işleme ve chunk'lara ayırma
- **Table Extraction**: Dokümanlardaki karşılaştırma tablolarını otomatik tespit
- **Conversation History**: Dosya bazlı sohbet geçmişi yönetimi

## 🛠️ Teknoloji Stack'i

### Backend
- **Go**: Yüksek performanslı backend dili
- **Fiber**: Express.js benzeri hızlı web framework
- **MongoDB**: NoSQL veritabanı
- **JWT**: Token tabanlı kimlik doğrulama
- **Ollama**: LLM ve embedding model servisi
- **ChromaDB**: Vektör veritabanı (HNSW algoritması ile ANN araması)
- **MinIO**: S3 uyumlu object storage

### Frontend
- **React**: Modern kullanıcı arayüzü kütüphanesi
- **Vite**: Hızlı geliştirme ve build aracı
- **Tailwind CSS**: Utility-first CSS framework

### AI/ML Teknolojileri
- **Embedding Models**: Vektör temsilleri için (örn: all-minilm:l6-v2)
- **LLM Models**: Metin üretimi için (örn: llama3:8b)
- **HNSW (Hierarchical Navigable Small World)**: Yaklaşık en yakın komşu (ANN) araması
- **Cosine Similarity**: Vektör benzerlik hesaplama
- **Reciprocal Rank Fusion (RRF)**: Hybrid search sonuç birleştirme

## 📁 Proje Yapısı

```
nimbus/
├── backend/              # Go Fiber backend
│   ├── config/          # Konfigürasyon yönetimi
│   ├── database/        # Veritabanı bağlantıları
│   ├── handlers/        # HTTP request handlers
│   │   ├── ai.go        # RAG ve AI sorgu handler'ları
│   │   └── files.go     # Dosya yönetim handler'ları
│   ├── middleware/      # JWT ve diğer middleware'ler
│   ├── models/          # Veri modelleri
│   ├── routes/          # API route tanımları
│   ├── services/        # İş mantığı servisleri
│   │   ├── chroma_service.go      # ChromaDB vektör arama servisi
│   │   ├── ollama_service.go      # LLM ve embedding servisi
│   │   └── document_processor.go # Doküman işleme pipeline'ı
│   ├── retrieval/       # RAG retrieval bileşenleri
│   │   ├── file_router.go         # In-memory dosya bazlı arama
│   │   ├── intent_classifier.go   # Sorgu niyet analizi
│   │   ├── query_utils.go         # Sorgu yardımcı fonksiyonları
│   │   └── adaptive.go            # Adaptif top-k retrieval
│   ├── chunks/          # Metin chunking ve işleme
│   │   ├── semantic_splitter.go   # Semantic chunking
│   │   ├── table_processor.go     # Tablo tespit ve işleme
│   │   └── normalizer.go          # Metin normalizasyonu
│   └── cache/           # Cache mekanizmaları
│       └── query_cache.go          # Semantic query cache
├── frontend/            # React Vite frontend
│   ├── src/
│   │   ├── components/  # React bileşenleri
│   │   ├── hooks/         # Custom React hooks
│   │   ├── services/      # API servisleri
│   │   └── pages/         # Sayfa bileşenleri
└── README.md
```

## 🏃‍♂️ Kurulum ve Çalıştırma

### Gereksinimler
- Go 1.19+
- Node.js 20+
- MongoDB
- MinIO (dosya depolama için)
- Ollama (AI model servisi için)
- ChromaDB (vektör veritabanı için)

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

# MinIO Object Storage
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false

# Ollama AI Service
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_EMBED_MODEL=all-minilm:l6-v2
OLLAMA_LLM_MODEL=llama3:8b

# ChromaDB Vector Database
CHROMA_BASE_URL=http://localhost:6006
CHROMA_TENANT=default_tenant
CHROMA_DATABASE=default_database
CHROMA_COLLECTION=nimbus_documents

# RAG Optimization Flags
ENABLE_QUERY_CACHE=true
ENABLE_CHUNK_CACHE=true
ENABLE_ADAPTIVE=true
ENABLE_FILE_ROUTING=true
ENABLE_DEDUPLICATION=false

# Cache Settings
QUERY_CACHE_TTL=60
CHUNK_CACHE_SIZE=1000

# RAG Thresholds
HIGH_SIMIL_THRESHOLD=0.85
MED_SIMIL_THRESHOLD=0.70
MIN_SIMIL_THRESHOLD=0.50
CONTEXT_WINDOW_SIZE=4096
MAX_RAG_CHUNKS=10
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
   - PDF ve DOCX dosyaları otomatik olarak işlenir ve RAG için hazırlanır

2. **Dosya Listesi:**
   - "Dosyalarım" bölümünden yüklediğiniz dosyaları görün
   - Her dosyanın boyutu, türü ve işleme durumu gösterilir
   - İndirme butonu ile dosyaları indirebilirsiniz

3. **AI ile Doküman Sorgulama:**
   - İşlenmiş PDF/DOCX dosyalarınıza doğal dil ile sorular sorun
   - Örnek sorular:
     - "Bu dokümanda X nedir?"
     - "X ve Y'yi karşılaştır"
     - "Dokümanın özetini ver"
     - "X'in özellikleri nelerdir?"
   - Sistem sorgu niyetini otomatik algılar ve uygun arama stratejisini kullanır

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

## 🧠 RAG Sistemi Mimarisi

### Vektör Arama ve ANN (Approximate Nearest Neighbor)

Nimbus, büyük vektör koleksiyonları üzerinde yüksek performanslı arama için **HNSW (Hierarchical Navigable Small World)** algoritması kullanmaktadır. ChromaDB varsayılan olarak HNSW ile ANN araması gerçekleştirir.

#### Arama Stratejileri

1. **Semantic Search (Anlamsal Arama)**
   - Kullanıcı sorguları embedding model ile vektör temsillerine dönüştürülür
   - ChromaDB üzerinde cosine similarity ile en benzer chunk'lar bulunur
   - HNSW algoritması ile yaklaşık en yakın komşu (ANN) araması yapılır

2. **Keyword Search (Kelime Eşleşmesi)**
   - Sorgudan çıkarılan anahtar kelimeler chunk metinlerinde aranır
   - Metadata'daki `key_terms` alanında da arama yapılır
   - Özellikle karşılaştırma ve tanım sorguları için kullanılır

3. **Hybrid Search (Karma Arama)**
   - Semantic ve keyword arama sonuçları birleştirilir
   - Reciprocal Rank Fusion (RRF) algoritması ile sonuçlar merge edilir
   - Karşılaştırma ve tanım sorguları için otomatik aktif olur

#### Performans Optimizasyonları

- **File Router**: Sık kullanılan dosyalar için in-memory index
- **Query Cache**: Benzer sorgular için semantic cache
- **Chunk Cache**: Popüler chunk'lar için embedding cache
- **Adaptive Top-K**: Sorgu tipine göre dinamik chunk sayısı seçimi

### Doküman İşleme Pipeline

1. **Metin Çıkarma**: PDF/DOCX dosyalarından metin çıkarılır
2. **Normalizasyon**: Metin temizlenir ve normalize edilir
3. **Tablo Tespiti**: Karşılaştırma tabloları otomatik tespit edilir
4. **Semantic Chunking**: Metin anlamsal olarak chunk'lara ayrılır
5. **Embedding Üretimi**: Her chunk için embedding vektörü oluşturulur
6. **Vektör Depolama**: Embedding'ler ChromaDB'ye kaydedilir

## 🚀 Kurulum: AI Servisleri

### Ollama Kurulumu

```bash
# Ollama'yı indirin ve başlatın
# macOS/Linux:
curl -fsSL https://ollama.ai/install.sh | sh

# Windows: https://ollama.ai/download adresinden indirin

# Gerekli modelleri yükleyin
ollama pull all-minilm:l6-v2  # Embedding model
ollama pull llama3:8b          # LLM model
```

### ChromaDB Kurulumu

```bash
# Docker ile ChromaDB başlatın
docker run -d \
  -p 6006:8000 \
  --name chromadb \
  chromadb/chroma:latest
```

ChromaDB `http://localhost:6006` adresinde çalışacak.
