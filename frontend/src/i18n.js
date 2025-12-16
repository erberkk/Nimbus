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
      move: 'Taşı',
      info: 'Bilgi',
      share: 'Paylaş',
      'access.read': 'Görüntüleme',
      'access.write': 'Düzenleme',
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

      'starred.add': 'Yıldızla',
      'starred.remove': 'Yıldızlamayı Kaldır',
      'restore': 'Geri Yükle',

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

      // AI & Chat
      'ai.title': 'Nimbus AI',
      'ai.subtitle': 'Dosya Asistanı',
      'ai.ask_nimbus': "Nimbus'a Sor",
      'ai.ask_nimbus_processing': "Nimbus'a Sor (İşleniyor)",
      'ai.ask_nimbus_error': "Nimbus'a Sor (Hata)",
      'ai.processing': 'İşleniyor...',
      'ai.ready': 'AI Hazır',
      'ai.failed': 'İşleme Hatası',
      'ai.pending': 'Beklemede',
      'ai.welcome': 'Merhaba! "{{filename}}" dosyası hakkında ne öğrenmek istiyorsun?',
      'ai.processing_message':
        '"{{filename}}" dosyası şu anda işleniyor. Lütfen işlem tamamlanana kadar bekleyin.',
      'ai.failed_message':
        '"{{filename}}" dosyası işlenirken bir hata oluştu. Lütfen daha sonra tekrar deneyin.',
      'ai.pending_message': '"{{filename}}" henüz işlenmedi. Dosya işlenirken lütfen bekleyin.',
      'ai.placeholder': 'Dosya hakkında soru sor...',
      'ai.typing': 'Nimbus yazıyor',
      'ai.not_processed': 'Dosya henüz işlenmedi. Lütfen bekleyin.',
      'ai.query_error': 'Üzgünüm, bir hata oluştu. Lütfen tekrar deneyin.',
      'ai.unsaved_changes': 'Kaydedilmemiş değişiklikler',
      'ai.save_hint': 'Ctrl+S ile kaydet',

      // File Preview
      'file.loading_code': 'Kod dosyası yükleniyor...',
      'file.converting': 'Dosya dönüştürülüyor...',
      'file.old_word_format': 'Eski format Word dosyası (.doc) için önizleme desteklenmiyor',
      'file.old_word_hint': 'Dosyayı görüntülemek için indirin veya .docx formatına dönüştürün',

      // File Info Panel
      'file_info.title': 'Dosya Bilgileri',
      'file_info.size': 'Dosya Boyutu',
      'file_info.type': 'Dosya Türü',
      'file_info.share_link': 'PAYLAŞIM LİNKİ',
      'file.preview_error': 'Dosya önizlemesi yüklenemedi',
      'file.content_error': 'Dosya içeriği yüklenemedi',
      'file.save_success': 'Dosya başarıyla kaydedildi',
      'file.save_error': 'Dosya kaydedilemedi',
      'file.image_error': 'Resim yüklenemedi',
      'file.video_error': 'Video yüklenemedi',
      'file.video_not_supported': 'Tarayıcınız video elementi desteklemiyor',
      'file.onlyoffice_required': 'Bu dosya türü için OnlyOffice önizleme kullanılmalı',
      'file.preview_not_supported': 'Bu dosya türü için önizleme desteklenmiyor',

      // File Upload
      'upload.title_file': 'Dosya Yükle',
      'upload.title_folder': 'Klasör Yükle',
      'upload.mode_single': 'Tek Dosya',
      'upload.mode_folder': 'Klasör',

      // Folders
      'folder.title': 'Klasörler',
      'folder.create': 'Yeni Klasör Oluştur',
      'folder.name': 'Klasör Adı',
      'folder.color': 'Klasör Rengi',
      'folder.new': 'Yeni Klasör',
      'folder.items': '{{count}} öğe',
      'folder.items_zero': '0 öğe',
      'folder.upload': 'Klasör yükleme',
      'folder.upload_file': 'Dosya Yükle',
      'folder.upload_folder': 'Klasör Yükle',
      'folder.new_folder': 'Yeni Klasör',
      'folder.empty': 'Bu klasör boş',
      'folder.no_items': 'Henüz klasör veya dosya yok',
      'folder.empty_hint': 'Sağ tıklayarak yeni klasör oluşturun veya dosya yükleyin',
      'folder.drag_drop': 'Dosyaları buraya bırakın',
      'folder.drag_drop_hint': 'Sürükleyip bıraktığınız dosyalar yüklenecek',
      'folder.files_title': 'Dosyalar',
      'folder.new_folder_menu': 'Yeni klasör',
      'folder.upload_menu': 'Dosya yükleme',
      'folder.delete_confirm': '"{{name}}" klasörünü silmek istediğinizden emin misiniz?',
      'folder.rename': 'Yeniden Adlandır',
      'folder.uploading': '{{count}} dosya yükleniyor...',
      'folder.upload_success': '{{count}} dosya başarıyla yüklendi!',
      'folder.upload_error': 'Dosya yükleme hatası: {{error}}',
      'folder.upload_success_single': '{{count}} dosya başarıyla yüklendi',
      'folder.upload_error_single': 'Yükleme hatası: {{error}}',

      // OnlyOffice Editor
      'onlyoffice.api_error': 'OnlyOffice Document Server API yüklenemedi',
      'onlyoffice.connection_error': 'OnlyOffice Document Server bağlantı hatası',
      'onlyoffice.connection_failed': "OnlyOffice Document Server'a bağlanılamadı",
      'onlyoffice.config_error': 'Dosya düzenleme konfigürasyonu yüklenemedi',
      'onlyoffice.start_error': 'Dosya düzenleme başlatılamadı',
      'onlyoffice.editor_error': 'OnlyOffice editor yüklenemedi',
      'onlyoffice.edit_error': 'Dosya düzenlenirken hata oluştu',
      'onlyoffice.init_error': 'Dosya düzenleyici başlatılamadı',
      'onlyoffice.edit_mode': 'Düzenle',
      'onlyoffice.preview_mode': 'Önizle',
      'onlyoffice.loading': 'OnlyOffice editor yükleniyor...',
      'onlyoffice.invalid_config': 'Geçersiz OnlyOffice config: document URL bulunamadı',

      // Share Dialog
      'share.title': 'Paylaş',
      'share.viewer': 'Görüntüleyen',
      'share.editor': 'Düzenleyen',
      'share.no_permission':
        'Sadece görüntüleme yetkiniz var. Bu kaynağı paylaşamaz veya erişimleri düzenleyemezsiniz.',
      'share.public_hint':
        'Bu bağlantıyı herkesle paylaşabilirsiniz. Bağlantıya tıklayan giriş yapmış kullanıcılar dosyayı görüntüleyebilir.',
      'share.public_link': 'Public Bağlantı',
      'share.load_error': 'Paylaşımlar yüklenemedi',
      'share.success': '{{email}} ile paylaşıldı',
      'share.failed': 'Paylaşım başarısız',
      'share.remove_success': 'Paylaşım kaldırıldı',
      'share.remove_failed': 'Paylaşım kaldırılamadı',
      'share.link_copied': 'Bağlantı kopyalandı',
      'share.link_copy_failed': 'Bağlantı kopyalanamadı',

      // User Search
      'user_search.placeholder': 'Email ile kullanıcı ara...',

      // Create Folder
      'folder.name_placeholder': 'Belgeler, Resimler, Projeler...',

      // File Explorer Header
      'header.shared': 'Paylaşılanlarım',
      'file.delete_confirm': 'Bu dosyayı silmek istediğinizden emin misiniz?',
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
      move: 'Move',
      info: 'Info',
      share: 'Share',
      'access.read': 'View',
      'access.write': 'Edit',
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

      'starred.add': 'Add Star',
      'starred.remove': 'Remove Star',
      'restore': 'Restore',

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

      // AI & Chat
      'ai.title': 'Nimbus AI',
      'ai.subtitle': 'File Assistant',
      'ai.ask_nimbus': 'Ask Nimbus',
      'ai.ask_nimbus_processing': 'Ask Nimbus (Processing)',
      'ai.ask_nimbus_error': 'Ask Nimbus (Error)',
      'ai.processing': 'Processing...',
      'ai.ready': 'AI Ready',
      'ai.failed': 'Processing Error',
      'ai.pending': 'Pending',
      'ai.welcome': 'Hello! What would you like to know about "{{filename}}"?',
      'ai.processing_message':
        'The file "{{filename}}" is currently being processed. Please wait until processing is complete.',
      'ai.failed_message':
        'An error occurred while processing the file "{{filename}}". Please try again later.',
      'ai.pending_message':
        'The file "{{filename}}" has not been processed yet. Please wait while the file is being processed.',
      'ai.placeholder': 'Ask a question about the file...',
      'ai.typing': 'Nimbus is typing',
      'ai.not_processed': 'File has not been processed yet. Please wait.',
      'ai.query_error': 'Sorry, an error occurred. Please try again.',
      'ai.unsaved_changes': 'Unsaved changes',
      'ai.save_hint': 'Press Ctrl+S to save',

      // File Preview
      'file.loading_code': 'Loading code file...',
      'file.converting': 'Converting file...',
      'file.old_word_format': 'Preview is not supported for old format Word files (.doc)',
      'file.old_word_hint': 'Download the file or convert it to .docx format to view it',

      // File Info Panel
      'file_info.title': 'File Information',
      'file_info.size': 'File Size',
      'file_info.type': 'File Type',
      'file_info.share_link': 'SHARE LINK',
      'file.preview_error': 'Failed to load file preview',
      'file.content_error': 'Failed to load file content',
      'file.save_success': 'File saved successfully',
      'file.save_error': 'Failed to save file',
      'file.image_error': 'Failed to load image',
      'file.video_error': 'Failed to load video',
      'file.video_not_supported': 'Your browser does not support video element',
      'file.onlyoffice_required': 'OnlyOffice preview should be used for this file type',
      'file.preview_not_supported': 'Preview is not supported for this file type',

      // File Upload
      'upload.title_file': 'Upload File',
      'upload.title_folder': 'Upload Folder',
      'upload.mode_single': 'Single File',
      'upload.mode_folder': 'Folder',

      // Folders
      'folder.title': 'Folders',
      'folder.create': 'Create New Folder',
      'folder.name': 'Folder Name',
      'folder.color': 'Folder Color',
      'folder.new': 'New Folder',
      'folder.items': '{{count}} items',
      'folder.items_zero': '0 items',
      'folder.upload': 'Folder upload',
      'folder.upload_file': 'Upload File',
      'folder.upload_folder': 'Upload Folder',
      'folder.new_folder': 'New Folder',
      'folder.empty': 'This folder is empty',
      'folder.no_items': 'No folders or files yet',
      'folder.empty_hint': 'Right-click to create a new folder or upload files',
      'folder.drag_drop': 'Drop files here',
      'folder.drag_drop_hint': 'Files you drag and drop will be uploaded',
      'folder.files_title': 'Files',
      'folder.new_folder_menu': 'New folder',
      'folder.upload_menu': 'File upload',
      'folder.delete_confirm': 'Are you sure you want to delete the folder "{{name}}"?',
      'folder.rename': 'Rename',
      'folder.uploading': 'Uploading {{count}} files...',
      'folder.upload_success': '{{count}} files uploaded successfully!',
      'folder.upload_error': 'File upload error: {{error}}',
      'folder.upload_success_single': '{{count}} file uploaded successfully',
      'folder.upload_error_single': 'Upload error: {{error}}',

      // OnlyOffice Editor
      'onlyoffice.api_error': 'Failed to load OnlyOffice Document Server API',
      'onlyoffice.connection_error': 'OnlyOffice Document Server connection error',
      'onlyoffice.connection_failed': 'Failed to connect to OnlyOffice Document Server',
      'onlyoffice.config_error': 'Failed to load file editing configuration',
      'onlyoffice.start_error': 'Failed to start file editing',
      'onlyoffice.editor_error': 'Failed to load OnlyOffice editor',
      'onlyoffice.edit_error': 'Error occurred while editing file',
      'onlyoffice.init_error': 'Failed to initialize file editor',
      'onlyoffice.edit_mode': 'Edit',
      'onlyoffice.preview_mode': 'Preview',
      'onlyoffice.loading': 'Loading OnlyOffice editor...',
      'onlyoffice.invalid_config': 'Invalid OnlyOffice config: document URL not found',

      // Share Dialog
      'share.title': 'Share',
      'share.viewer': 'Viewer',
      'share.editor': 'Editor',
      'share.no_permission':
        'You only have view permission. You cannot share this resource or edit access.',
      'share.public_hint':
        'You can share this link with anyone. Users who click the link and are logged in can view the file.',
      'share.public_link': 'Public Link',
      'share.load_error': 'Failed to load shares',
      'share.success': 'Shared with {{email}}',
      'share.failed': 'Share failed',
      'share.remove_success': 'Share removed',
      'share.remove_failed': 'Failed to remove share',
      'share.link_copied': 'Link copied',
      'share.link_copy_failed': 'Failed to copy link',

      // User Search
      'user_search.placeholder': 'Search user by email...',

      // Create Folder
      'folder.name_placeholder': 'Documents, Images, Projects...',

      // File Explorer Header
      'header.shared': 'My Shared',
      'file.delete_confirm': 'Are you sure you want to delete this file?',
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
