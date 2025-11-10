/**
 * File utility functions
 * Centralized file type detection, preview checking, and formatting
 */

/**
 * Get file type from content type and filename
 * @param {string} contentType - MIME type of the file
 * @param {string} filename - Name of the file
 * @returns {string} File type: 'pdf', 'image', 'audio', 'video', 'word-docx', 'word-doc', 'excel', 'unknown'
 */
export const getFileType = (contentType, filename = '') => {
  if (!contentType) return 'unknown';
  
  const contentTypeLower = contentType.toLowerCase();
  const filenameLower = filename.toLowerCase();
  
  if (contentTypeLower.includes('pdf')) return 'pdf';
  if (contentTypeLower.startsWith('image/')) return 'image';
  if (contentTypeLower.startsWith('audio/')) return 'audio';
  if (contentTypeLower.startsWith('video/')) return 'video';
  
  // Word documents
  if (contentTypeLower.includes('wordprocessingml.document') || 
      (contentTypeLower === 'application/msword' && filenameLower.endsWith('.docx'))) {
    return 'word-docx';
  }
  if (contentTypeLower === 'application/msword' && filenameLower.endsWith('.doc')) {
    return 'word-doc'; // Legacy format - no preview
  }
  
  // Excel files
  if (contentTypeLower.includes('spreadsheetml.sheet') || 
      contentTypeLower === 'application/vnd.ms-excel') {
    return 'excel';
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
  if (!contentType) return false;
  
  const fileType = getFileType(contentType, filename);
  const previewableTypes = ['pdf', 'image', 'audio', 'video', 'word-docx', 'word-doc', 'excel'];
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
 * Format file size in bytes to human-readable format
 * @param {number} bytes - File size in bytes
 * @returns {string} Formatted file size (e.g., "1.5 MB")
 */
export const formatFileSize = (bytes) => {
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

