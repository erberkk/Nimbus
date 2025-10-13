import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Çeviri kaynakları
const resources = {
  tr: {
    translation: {
      // Genel
      welcome: 'Hoş Geldiniz',
      login: 'Giriş Yap',
      logout: 'Çıkış Yap',
      dashboard: 'Dashboard',
      loading: 'Yükleniyor...',
      error: 'Hata',

      // Sidebar
      'sidebar.home': 'Ana Sayfa',
      'sidebar.shared': 'Benimle Paylaşılan',
      'sidebar.recent': 'Son Kullanılanlar',
      'sidebar.starred': 'Yıldızlananlar',
      'sidebar.trash': 'Çöp Kutusu',
      'sidebar.storage': 'Depolama',
      'sidebar.of': 'üzerinden',
      'sidebar.used': 'kullanıldı',
      success: 'Başarılı',
      warning: 'Uyarı',
      info: 'Bilgi',

      // Auth
      google_login: 'Google ile Giriş Yap',
      login_welcome: 'Devam etmek için Google hesabınızla giriş yapın',
      login_terms: 'Giriş yaparak',
      terms_of_use: 'Kullanım Koşulları',
      privacy_policy: 'Gizlilik Politikası',
      accept_terms: 'kabul etmiş olursunuz.',

      // Navbar
      nimbus: 'Nimbus',
      cloud_storage: 'Bulut Depolama',
      cloud_storage_desc: 'Dosyalarınızı güvenle bulutta saklayın',
      secure: 'Güvenli',
      secure_desc: 'End-to-end şifreleme ile korumalı',
      fast_access: 'Hızlı Erişim',
      fast_access_desc: 'Dosyalarınıza her yerden anında erişin',

      // Home Page
      welcome_user: 'Hoş Geldiniz',
      dashboard_desc: 'Dosyalarınızı yönetin, paylaşın ve her yerden erişin',
      pro_member: 'Pro Üye',

      // Stats
      total_files: 'Toplam Dosya',
      folders: 'Klasörler',
      shared: 'Paylaşılanlar',

      // Actions
      quick_actions: 'Hızlı İşlemler',
      upload_file: 'Dosya Yükle',
      upload_desc: 'Bilgisayarınızdan dosya yükleyin',
      create_folder: 'Klasör Oluştur',
      create_folder_desc: 'Yeni klasör oluşturun',
      share: 'Paylaş',
      share_desc: 'Dosyalarınızı paylaşın',

      // Files
      my_files: 'Dosyalarım',
      refresh: 'Yenile',
      files_loading: 'Dosyalar yükleniyor...',
      files_error: 'Dosya listesi yüklenemedi',
      no_files: 'Henüz dosya yüklemediniz',
      no_files_desc: 'İlk dosyanızı yüklemek için yukarıdaki "Dosya Yükle" alanını kullanın',

      // File Upload
      upload_files: 'Dosya Yükleyin',
      drag_drop: 'Sürükleyip bırakın veya tıklayarak seçin',
      supported_types: 'Desteklenen türler: PNG, JPEG, PDF, Word, Excel, PPT, TXT, ZIP, RAR, 7Z',
      max_size: 'Maksimum: 100MB',
      uploading: 'Dosya Yükleniyor...',
      upload_progress: '% tamamlandı',
      upload_success: 'Dosya başarıyla yüklendi!',
      upload_error: 'Dosya yüklenirken hata oluştu',
      file_too_large: "Dosya boyutu 100MB'dan büyük olamaz",
      select_file: 'Lütfen bir dosya seçin',

      // File Types
      image: 'Resim',
      document: 'Belge',
      archive: 'Arşiv',
      unknown: 'Bilinmeyen',

      // Common Actions
      download: 'İndir',
      delete: 'Sil',
      edit: 'Düzenle',
      save: 'Kaydet',
      cancel: 'İptal',
      confirm: 'Onayla',
      close: 'Kapat',
      confirm_delete: 'Bu dosyayı silmek istediğinizden emin misiniz?',
      delete_success: 'Dosya başarıyla silindi',
      delete_error: 'Dosya silinirken hata oluştu',

      // Errors
      network_error: 'Ağ hatası',
      server_error: 'Sunucu hatası',
      unauthorized: 'Yetkisiz erişim',
      not_found: 'Bulunamadı',
      forbidden: 'Yasak',
      language_changed: 'Dil {{lng}} olarak değiştirildi',

      // Landing Page
      'landing.nav_features': 'Özellikler',
      'landing.nav_about': 'Hakkında',
      'landing.hero_title1': 'Dosyalarınız,',
      'landing.hero_title2': 'Her Yerde',
      'landing.hero_subtitle':
        'Dosyalarınızı bulutta güvenle saklayın, senkronize edin ve paylaşın. Verilerinize dünyanın her yerinden, herhangi bir cihazdan erişin.',
      'landing.cta_start': 'Ücretsiz Başlayın',
      'landing.cta_demo': 'Demo İzle',
      'landing.features_title': 'Güçlü Özellikler',
      'landing.features_subtitle': 'Modern bulut depolama için ihtiyacınız olan her şey',
      'landing.feature1_title': 'Kolay Yükleme',
      'landing.feature1_desc':
        'Dosyaları veya klasörleri sürükleyip bırakın. Sezgisel arayüzümüzle aynı anda birden fazla dosya yükleyin.',
      'landing.feature2_title': 'Akıllı Paylaşım',
      'landing.feature2_desc':
        'Dosya ve klasörleri paylaşmak için güvenli bağlantılar oluşturun. İzinleri ve erişim seviyelerini kontrol edin.',
      'landing.feature3_title': 'Güvenli Depolama',
      'landing.feature3_desc':
        'Uçtan uca şifreleme dosyalarınızı güvende tutar. Verileriniz kurumsal düzeyde güvenlikle korunur.',
      'landing.stat1': 'Depolanan Dosya',
      'landing.stat2': 'Aktif Kullanıcı',
      'landing.stat3': 'Çalışma Süresi',
      'landing.stat4': 'Destek',
      'landing.cta_title': 'Başlamaya Hazır mısınız?',
      'landing.cta_subtitle': "Dosyalarını Nimbus'a emanet eden binlerce kullanıcıya katılın",
      'landing.cta_button': 'Ücretsiz Denemenizi Başlatın',
      'landing.cta_dashboard': "Dashboard'a Git",
      'landing.footer_desc': 'Modern ekipler ve bireyler için güvenilir bulut depolama çözümünüz.',
      'landing.footer_product': 'Ürün',
      'landing.footer_company': 'Şirket',
      'landing.footer_support': 'Destek',
      'landing.footer_rights': 'Tüm hakları saklıdır.',
    },
  },
  en: {
    translation: {
      // General
      welcome: 'Welcome',
      login: 'Login',
      logout: 'Logout',
      dashboard: 'Dashboard',
      loading: 'Loading...',
      error: 'Error',

      // Sidebar
      'sidebar.home': 'Home',
      'sidebar.shared': 'Shared with me',
      'sidebar.recent': 'Recent',
      'sidebar.starred': 'Starred',
      'sidebar.trash': 'Trash',
      'sidebar.storage': 'Storage',
      'sidebar.of': 'of',
      'sidebar.used': 'used',
      success: 'Success',
      warning: 'Warning',
      info: 'Info',

      // Auth
      google_login: 'Sign in with Google',
      login_welcome: 'Sign in with your Google account to continue',
      login_terms: 'By signing in, you agree to our',
      terms_of_use: 'Terms of Use',
      privacy_policy: 'Privacy Policy',
      accept_terms: 'and accept them.',

      // Navbar
      nimbus: 'Nimbus',
      cloud_storage: 'Cloud Storage',
      cloud_storage_desc: 'Store your files securely in the cloud',
      secure: 'Secure',
      secure_desc: 'Protected with end-to-end encryption',
      fast_access: 'Fast Access',
      fast_access_desc: 'Access your files instantly from anywhere',

      // Home Page
      welcome_user: 'Welcome',
      dashboard_desc: 'Manage, share and access your files from anywhere',
      pro_member: 'Pro Member',

      // Stats
      total_files: 'Total Files',
      folders: 'Folders',
      shared: 'Shared',

      // Actions
      quick_actions: 'Quick Actions',
      upload_file: 'Upload File',
      upload_desc: 'Upload files from your computer',
      create_folder: 'Create Folder',
      create_folder_desc: 'Create a new folder',
      share: 'Share',
      share_desc: 'Share your files',

      // Files
      my_files: 'My Files',
      refresh: 'Refresh',
      files_loading: 'Loading files...',
      files_error: 'Failed to load file list',
      no_files: "You haven't uploaded any files yet",
      no_files_desc: 'Use the "Upload File" area above to upload your first file',

      // File Upload
      upload_files: 'Upload Files',
      drag_drop: 'Drag and drop or click to select',
      supported_types: 'Supported types: PNG, JPEG, PDF, Word, Excel, PPT, TXT, ZIP, RAR, 7Z',
      max_size: 'Maximum: 100MB',
      uploading: 'Uploading File...',
      upload_progress: '% completed',
      upload_success: 'File uploaded successfully!',
      upload_error: 'Error uploading file',
      file_too_large: 'File size cannot exceed 100MB',
      select_file: 'Please select a file',

      // File Types
      image: 'Image',
      document: 'Document',
      archive: 'Archive',
      unknown: 'Unknown',

      // Common Actions
      download: 'Download',
      delete: 'Delete',
      edit: 'Edit',
      save: 'Save',
      cancel: 'Cancel',
      confirm: 'Confirm',
      close: 'Close',
      confirm_delete: 'Are you sure you want to delete this file?',
      delete_success: 'File deleted successfully',
      delete_error: 'Error deleting file',

      // Errors
      network_error: 'Network error',
      server_error: 'Server error',
      unauthorized: 'Unauthorized access',
      not_found: 'Not found',
      forbidden: 'Forbidden',
      language_changed: 'Language changed to {{lng}}',

      // Landing Page
      'landing.nav_features': 'Features',
      'landing.nav_about': 'About',
      'landing.hero_title1': 'Your Files,',
      'landing.hero_title2': 'Anywhere',
      'landing.hero_subtitle':
        'Store, sync, and share your files securely in the cloud. Access your data from any device, anywhere in the world.',
      'landing.cta_start': 'Get Started Free',
      'landing.cta_demo': 'Watch Demo',
      'landing.features_title': 'Powerful Features',
      'landing.features_subtitle': 'Everything you need for modern cloud storage',
      'landing.feature1_title': 'Easy Upload',
      'landing.feature1_desc':
        'Drag and drop files or folders. Upload multiple files at once with our intuitive interface.',
      'landing.feature2_title': 'Smart Sharing',
      'landing.feature2_desc':
        'Generate secure links to share files and folders. Control permissions and access levels.',
      'landing.feature3_title': 'Secure Storage',
      'landing.feature3_desc':
        'End-to-end encryption keeps your files safe. Your data is protected with enterprise-grade security.',
      'landing.stat1': 'Files Stored',
      'landing.stat2': 'Active Users',
      'landing.stat3': 'Uptime',
      'landing.stat4': 'Support',
      'landing.cta_title': 'Ready to Get Started?',
      'landing.cta_subtitle': 'Join thousands of users who trust Nimbus with their files',
      'landing.cta_button': 'Start Your Free Trial',
      'landing.cta_dashboard': 'Go to Dashboard',
      'landing.footer_desc':
        'Your trusted cloud storage solution for modern teams and individuals.',
      'landing.footer_product': 'Product',
      'landing.footer_company': 'Company',
      'landing.footer_support': 'Support',
      'landing.footer_rights': 'All rights reserved.',
    },
  },
};

i18n.use(initReactI18next).init({
  resources,
  lng: 'tr', // default language
  fallbackLng: 'tr',
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
