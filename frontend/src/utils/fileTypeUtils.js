/**
 * Get file extension from filename
 */
export const getFileExtension = (filename) => {
    if (!filename) return '';
    const parts = filename.split('.');
    return parts.length > 1 ? parts.pop().toUpperCase() : '';
};

/**
 * Get color for file type
 */
export const getFileTypeColor = (contentType, filename) => {
    const type = (contentType || '').toLowerCase();
    const ext = getFileExtension(filename).toLowerCase();

    // PDFs
    if (type.includes('pdf') || ext === 'pdf') {
        return { primary: '#E53935', light: '#FFEBEE', dark: '#C62828' };
    }

    // Word documents
    if (type.includes('word') || type.includes('document') || ['doc', 'docx'].includes(ext)) {
        return { primary: '#1976D2', light: '#E3F2FD', dark: '#1565C0' };
    }

    // Excel
    if (type.includes('sheet') || type.includes('excel') || ['xls', 'xlsx'].includes(ext)) {
        return { primary: '#388E3C', light: '#E8F5E9', dark: '#2E7D32' };
    }

    // PowerPoint
    if (type.includes('presentation') || type.includes('powerpoint') || ['ppt', 'pptx'].includes(ext)) {
        return { primary: '#F57C00', light: '#FFF3E0', dark: '#E65100' };
    }

    // Images
    if (type.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'].includes(ext)) {
        return { primary: '#8E24AA', light: '#F3E5F5', dark: '#6A1B9A' };
    }

    // Videos
    if (type.startsWith('video/') || ['mp4', 'avi', 'mov', 'mkv'].includes(ext)) {
        return { primary: '#D32F2F', light: '#FFEBEE', dark: '#C62828' };
    }

    // Text files
    if (type.startsWith('text/') || ['txt', 'md', 'json', 'xml'].includes(ext)) {
        return { primary: '#616161', light: '#FAFAFA', dark: '#424242' };
    }

    // Code files
    if (['js', 'jsx', 'ts', 'tsx', 'py', 'java', 'go', 'cpp', 'c', 'css', 'html'].includes(ext)) {
        return { primary: '#0288D1', light: '#E1F5FE', dark: '#01579B' };
    }

    // Archives
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
        return { primary: '#F57F17', light: '#FFFDE7', dark: '#F57F17' };
    }

    // Default
    return { primary: '#757575', light: '#EEEEEE', dark: '#616161' };
};

/**
 * Get icon name for file type (MUI icons)
 */
export const getFileTypeIcon = (contentType, filename) => {
    const type = (contentType || '').toLowerCase();
    const ext = getFileExtension(filename).toLowerCase();

    if (type.includes('pdf') || ext === 'pdf') return 'PictureAsPdf';
    if (type.includes('word') || type.includes('document') || ['doc', 'docx'].includes(ext)) return 'Description';
    if (type.includes('sheet') || type.includes('excel') || ['xls', 'xlsx'].includes(ext)) return 'GridOn';
    if (type.includes('presentation') || type.includes('powerpoint') || ['ppt', 'pptx'].includes(ext)) return 'Slideshow';
    if (type.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'].includes(ext)) return 'Image';
    if (type.startsWith('video/') || ['mp4', 'avi', 'mov', 'mkv'].includes(ext)) return 'VideoFile';
    if (type.startsWith('text/') || ['txt', 'md', 'json', 'xml'].includes(ext)) return 'TextSnippet';
    if (['js', 'jsx', 'ts', 'tsx', 'py', 'java', 'go', 'cpp', 'c', 'css', 'html'].includes(ext)) return 'Code';
    if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return 'FolderZip';

    return 'InsertDriveFile';
};

/**
 * Format relative time
 */
export const formatRelativeTime = (date) => {
    if (!date) return '';

    const now = new Date();
    const then = new Date(date);
    const diffInSeconds = Math.floor((now - then) / 1000);

    if (diffInSeconds < 60) return 'Az önce';
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} dakika önce`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} saat önce`;
    if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)} gün önce`;
    if (diffInSeconds < 31536000) return `${Math.floor(diffInSeconds / 2592000)} ay önce`;

    return `${Math.floor(diffInSeconds / 31536000)} yıl önce`;
};
