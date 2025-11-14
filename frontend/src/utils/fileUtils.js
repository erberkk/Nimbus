/**
 * File utility functions
 * Centralized file type detection, preview checking, and formatting
 */

// Content type mappings for Office documents
const OFFICE_DOCUMENT_MAPPINGS = {
  word: {
    patterns: ['wordprocessingml.document', 'application/msword'],
    fileType: 'word-docx',
    extension: 'docx',
  },
  excel: {
    patterns: ['spreadsheetml.sheet', 'application/vnd.ms-excel'],
    fileType: 'excel',
    extension: 'xlsx',
  },
  powerpoint: {
    patterns: ['presentationml.presentation', 'application/vnd.ms-powerpoint'],
    fileType: 'powerpoint',
    extension: 'pptx',
  },
};

// Monaco Editor language mappings
const MONACO_LANGUAGE_MAP = {
  // Python
  py: 'python',
  pyw: 'python',
  pyi: 'python',
  // JavaScript/TypeScript
  js: 'javascript',
  jsx: 'javascript',
  mjs: 'javascript',
  cjs: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  // C#
  cs: 'csharp',
  csx: 'csharp',
  // Java
  java: 'java',
  // Kotlin
  kt: 'kotlin',
  kts: 'kotlin',
  ktm: 'kotlin',
  // JSON
  json: 'json',
  jsonc: 'json',
  // Markdown
  md: 'markdown',
  markdown: 'markdown',
  mdown: 'markdown',
  mkdn: 'markdown',
  // Text
  txt: 'plaintext',
  text: 'plaintext',
  // XML
  xml: 'xml',
  xsd: 'xml',
  xsl: 'xml',
  xslt: 'xml',
  // HTML
  html: 'html',
  htm: 'html',
  xhtml: 'html',
  // CSS
  css: 'css',
  scss: 'scss',
  sass: 'sass',
  less: 'less',
  // Shell
  sh: 'shell',
  bash: 'shell',
  zsh: 'shell',
  fish: 'shell',
  // YAML
  yaml: 'yaml',
  yml: 'yaml',
  // Go
  go: 'go',
  // Rust
  rs: 'rust',
  // PHP
  php: 'php',
  phtml: 'php',
  // Ruby
  rb: 'ruby',
  rake: 'ruby',
  // Perl
  pl: 'perl',
  pm: 'perl',
  // Scala
  scala: 'scala',
  sc: 'scala',
  // C/C++
  c: 'c',
  cpp: 'cpp',
  cc: 'cpp',
  cxx: 'cpp',
  h: 'c',
  hpp: 'cpp',
  hxx: 'cpp',
  // SQL
  sql: 'sql',
  // Vue
  vue: 'vue',
  // Svelte
  svelte: 'svelte',
  // Swift
  swift: 'swift',
  // Dart
  dart: 'dart',
  // Lua
  lua: 'lua',
  // R
  r: 'r',
  rdata: 'r',
  rds: 'r',
  // Objective-C
  m: 'objective-c',
  mm: 'objective-cpp',
  // PowerShell
  ps1: 'powershell',
  psm1: 'powershell',
  psd1: 'powershell',
  // Dockerfile
  dockerfile: 'dockerfile',
  // Makefile
  makefile: 'makefile',
  mk: 'makefile',
};

// Code file extensions for detection
const CODE_FILE_EXTENSIONS = Object.keys(MONACO_LANGUAGE_MAP);

/**
 * Check if content type matches office document pattern
 * @param {string} contentTypeLower - Lowercase content type
 * @param {string} pattern - Pattern to match
 * @returns {boolean}
 */
const matchesOfficePattern = (contentTypeLower, pattern) => {
  return contentTypeLower.includes(pattern) || contentTypeLower === pattern;
};

/**
 * Get Monaco Editor language from filename
 * @param {string} filename - Name of the file
 * @returns {string} Monaco language ID
 */
export const getMonacoLanguage = (filename = '') => {
  if (!filename) return 'plaintext';

  const ext = filename.split('.').pop()?.toLowerCase();
  return MONACO_LANGUAGE_MAP[ext] || 'plaintext';
};

/**
 * Check if a file is a code file
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file (optional)
 * @returns {boolean} True if file is a code file
 */
export const isCodeFile = (contentType, filename = '') => {
  if (filename) {
    const ext = filename.split('.').pop()?.toLowerCase();
    if (CODE_FILE_EXTENSIONS.includes(ext)) {
      return true;
    }
  }

  if (!contentType) return false;

  const contentTypeLower = contentType.toLowerCase();
  const codeContentTypes = [
    'text/plain',
    'text/x-',
    'text/javascript',
    'application/javascript',
    'text/typescript',
    'application/typescript',
    'application/json',
    'text/json',
    'text/markdown',
    'text/xml',
    'application/xml',
    'text/css',
    'text/html',
    'application/x-sh',
    'text/x-sh',
    'application/x-yaml',
    'text/yaml',
  ];

  return codeContentTypes.some(type => contentTypeLower.includes(type));
};

/**
 * Get file type from content type and filename
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file
 * @returns {string} File type: 'pdf', 'image', 'audio', 'video', 'word-docx', 'word-doc', 'excel', 'powerpoint', 'code', 'unknown'
 */
export const getFileType = (contentType, filename = '') => {
  if (!contentType) {
    // If no contentType, check if it's a code file by extension
    if (filename && isCodeFile(null, filename)) {
      return 'code';
    }
    return 'unknown';
  }

  const contentTypeLower = contentType.toLowerCase();
  const filenameLower = filename.toLowerCase();

  if (isCodeFile(contentType, filename)) {
    return 'code';
  }

  if (contentTypeLower.includes('pdf')) return 'pdf';
  if (contentTypeLower.startsWith('image/')) return 'image';
  if (contentTypeLower.startsWith('audio/')) return 'audio';
  if (contentTypeLower.startsWith('video/')) return 'video';

  // Word documents
  if (
    matchesOfficePattern(contentTypeLower, 'wordprocessingml.document') ||
    (contentTypeLower === 'application/msword' && filenameLower.endsWith('.docx'))
  ) {
    return 'word-docx';
  }
  if (contentTypeLower === 'application/msword' && filenameLower.endsWith('.doc')) {
    return 'word-doc'; // Legacy format - no preview
  }

  // Excel files
  if (
    matchesOfficePattern(contentTypeLower, 'spreadsheetml.sheet') ||
    contentTypeLower === 'application/vnd.ms-excel'
  ) {
    return 'excel';
  }

  // PowerPoint files
  if (
    matchesOfficePattern(contentTypeLower, 'presentationml.presentation') ||
    contentTypeLower === 'application/vnd.ms-powerpoint'
  ) {
    return 'powerpoint';
  }

  return 'unknown';
};

/**
 * Check if a file is previewable
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file (optional)
 * @returns {boolean} True if file can be previewed
 */
export const isPreviewable = (contentType, filename = '') => {
  const fileType = getFileType(contentType, filename);
  const previewableTypes = [
    'pdf',
    'image',
    'audio',
    'video',
    'word-docx',
    'word-doc',
    'excel',
    'powerpoint',
    'code',
  ];
  return previewableTypes.includes(fileType);
};

/**
 * Check if a file can be used with Nimbus AI (PDF and Word documents)
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file (optional)
 * @returns {boolean} True if file can be used with Nimbus AI
 */
export const isAskableFile = (contentType, filename = '') => {
  if (!contentType) return false;

  const contentTypeLower = contentType.toLowerCase();
  const filenameLower = filename.toLowerCase();

  return (
    contentTypeLower.includes('pdf') ||
    contentTypeLower.includes('document') ||
    contentTypeLower.includes('wordprocessingml') ||
    filenameLower.endsWith('.doc') ||
    filenameLower.endsWith('.docx') ||
    filenameLower.endsWith('.pdf')
  );
};

/**
 * Check if a file can be edited with OnlyOffice (Word, Excel, PowerPoint) or Monaco Editor (code files)
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file (optional)
 * @returns {boolean} True if file can be edited
 */
export const isEditable = (contentType, filename = '') => {
  if (isCodeFile(contentType, filename)) {
    return true;
  }

  // Check if it's an Office document (editable with OnlyOffice)
  if (!contentType) return false;

  const contentTypeLower = contentType.toLowerCase();
  const filenameLower = filename.toLowerCase();

  // Word documents (.docx)
  if (
    contentTypeLower.includes('wordprocessingml.document') ||
    (contentTypeLower === 'application/msword' && filenameLower.endsWith('.docx'))
  ) {
    return true;
  }

  // Excel files (.xlsx)
  if (
    contentTypeLower.includes('spreadsheetml.sheet') ||
    contentTypeLower === 'application/vnd.ms-excel'
  ) {
    return true;
  }

  // PowerPoint files (.pptx)
  if (
    contentTypeLower.includes('presentationml.presentation') ||
    contentTypeLower === 'application/vnd.ms-powerpoint'
  ) {
    return true;
  }

  return false;
};

/**
 * Format file size in bytes to human-readable format
 * @param {number} bytes - File size in bytes
 * @returns {string} Formatted file size (e.g., "1.5 MB")
 */
export const formatFileSize = bytes => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

/**
 * Format date to relative or localized format
 * @param {string|Date} dateString - Date string or Date object
 * @param {Object} options - Formatting options
 * @returns {string} Formatted date string
 */
export const formatDate = (dateString, options = {}) => {
  if (!dateString) return '';

  const date = new Date(dateString);
  if (isNaN(date.getTime())) return '';

  // If relative format is requested (default)
  if (options.relative !== false) {
    const now = new Date();
    const diff = now - date;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return 'Bugün';
    if (days === 1) return 'Dün';
    if (days < 7) return `${days} gün önce`;
  }

  // Return localized date string
  return date.toLocaleDateString('tr-TR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    ...options.dateOptions,
  });
};

/**
 * Format content type to short file extension format
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file (optional, used as fallback)
 * @returns {string} Short file extension (e.g., "docx", "pdf", "jpg")
 */
export const formatContentType = (contentType, filename = '') => {
  if (!contentType) {
    // Try to get extension from filename as fallback
    if (filename) {
      const ext = filename.split('.').pop()?.toLowerCase();
      return ext || 'dosya';
    }
    return 'dosya';
  }

  const contentTypeLower = contentType.toLowerCase();
  const filenameLower = filename.toLowerCase();

  if (
    matchesOfficePattern(contentTypeLower, 'presentationml.presentation') ||
    contentTypeLower === 'application/vnd.ms-powerpoint'
  ) {
    return OFFICE_DOCUMENT_MAPPINGS.powerpoint.extension;
  }

  if (
    matchesOfficePattern(contentTypeLower, 'spreadsheetml.sheet') ||
    contentTypeLower === 'application/vnd.ms-excel'
  ) {
    return OFFICE_DOCUMENT_MAPPINGS.excel.extension;
  }

  if (
    matchesOfficePattern(contentTypeLower, 'wordprocessingml.document') ||
    (contentTypeLower === 'application/msword' && filenameLower.endsWith('.docx'))
  ) {
    return OFFICE_DOCUMENT_MAPPINGS.word.extension;
  }

  if (contentTypeLower === 'application/msword' && filenameLower.endsWith('.doc')) {
    return 'doc';
  }

  if (contentTypeLower.includes('pdf')) {
    return 'pdf';
  }

  if (contentTypeLower.startsWith('image/')) {
    const imageTypes = {
      'image/jpeg': 'jpg',
      'image/jpg': 'jpg',
      'image/png': 'png',
      'image/gif': 'gif',
      'image/webp': 'webp',
      'image/svg+xml': 'svg',
      'image/bmp': 'bmp',
      'image/tiff': 'tiff',
    };
    return imageTypes[contentTypeLower] || contentTypeLower.split('/')[1] || 'img';
  }

  if (contentTypeLower.startsWith('audio/')) {
    const audioTypes = {
      'audio/mpeg': 'mp3',
      'audio/mp3': 'mp3',
      'audio/wav': 'wav',
      'audio/flac': 'flac',
      'audio/aac': 'aac',
      'audio/ogg': 'ogg',
      'audio/mp4': 'm4a',
      'audio/x-m4a': 'm4a',
    };
    return audioTypes[contentTypeLower] || contentTypeLower.split('/')[1] || 'audio';
  }

  if (contentTypeLower.startsWith('video/')) {
    const videoTypes = {
      'video/mp4': 'mp4',
      'video/avi': 'avi',
      'video/quicktime': 'mov',
      'video/x-msvideo': 'avi',
      'video/x-ms-wmv': 'wmv',
      'video/webm': 'webm',
      'video/x-matroska': 'mkv',
    };
    return videoTypes[contentTypeLower] || contentTypeLower.split('/')[1] || 'video';
  }

  if (contentTypeLower.includes('zip')) {
    return 'zip';
  }
  if (contentTypeLower.includes('rar')) {
    return 'rar';
  }
  if (contentTypeLower.includes('7z') || contentTypeLower.includes('x-7z')) {
    return '7z';
  }
  if (contentTypeLower.includes('gzip')) {
    return 'gz';
  }

  if (contentTypeLower.includes('text/')) {
    if (contentTypeLower.includes('csv')) {
      return 'csv';
    }
    if (contentTypeLower.includes('plain')) {
      return 'txt';
    }
    return contentTypeLower.split('/')[1] || 'txt';
  }

  if (contentTypeLower.includes('rtf')) {
    return 'rtf';
  }

  const parts = contentTypeLower.split('/');
  if (parts.length > 1) {
    const subtype = parts[1].split(';')[0].trim();
    if (subtype.includes('.')) {
      if (filename) {
        const ext = filename.split('.').pop()?.toLowerCase();
        if (ext) return ext;
      }
      return subtype.split('.').pop() || subtype;
    }
    return subtype;
  }

  if (filename) {
    const ext = filename.split('.').pop()?.toLowerCase();
    return ext || 'dosya';
  }

  return 'dosya';
};
