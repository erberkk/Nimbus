import { useState, useRef, useEffect } from 'react';
import {
  Box,
  Button,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  LinearProgress,
  Alert,
  Stack,
  IconButton,
  Chip,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import CloseIcon from '@mui/icons-material/Close';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import { fileApi, folderApi } from '../services/api';
import { formatFileSize } from '../utils/fileUtils';

const MotionBox = motion.create(Box);

const FileUpload = ({ open, onClose, onUploadSuccess, currentFolderId, mode = 'both' }) => {
  const { t } = useTranslation();
  const [dragActive, setDragActive] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [uploadedFile, setUploadedFile] = useState(null);
  const [isFolderUpload, setIsFolderUpload] = useState(mode === 'folder');
  const fileInputRef = useRef(null);

  useEffect(() => {
    if (mode === 'file') {
      setIsFolderUpload(false);
    } else if (mode === 'folder') {
      setIsFolderUpload(true);
    }
  }, [mode]);

  useEffect(() => {
    if (open) {
      setError('');
      setSuccess('');
      setUploadedFile(null);
      setProgress(0);
      setUploading(false);
    }
  }, [open]);

  const handleDrag = e => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = e => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      const files = Array.from(e.dataTransfer.files);
      
      // Klasör entry'si kontrolü (size: 0, type: '' -> klasör)
      const hasEmptyFolder = files.some(file => file.size === 0 && file.type === '');
      if (hasEmptyFolder) {
        setError('Tarayıcılar drag & drop ile klasör yüklemeyi desteklemiyor. Lütfen "Klasör Seç" butonunu kullanın.');
        window.toast?.error('Klasör yüklemek için "Klasör Seç" butonunu kullanın');
        return;
      }
      
      const hasFolderStructure = files.some(file => file.webkitRelativePath && file.webkitRelativePath.includes('/'));
      
      if (isFolderUpload || files.length > 1 || hasFolderStructure) {
        handleMultipleFiles(files);
      } else {
        handleFile(files[0]);
      }
    }
  };

  const handleFile = async file => {
    // Reset states
    setError('');
    setSuccess('');
    setUploadedFile(null);
    setProgress(0);

    if (!file) {
      setError(t('select_file'));
      return;
    }

    if (file.size > 100 * 1024 * 1024) {
      setError(t('file_too_large'));
      return;
    }

    try {
      setUploading(true);

      const presignedResponse = await fileApi.getUploadPresignedURL(file.name, file.type);
      const { presigned_url, minio_path } = presignedResponse;

      await uploadToMinIO(presigned_url, file);

      await fileApi.createFile({
        filename: file.name,
        size: file.size,
        content_type: file.type || 'application/octet-stream',
        minio_path: minio_path,
        folder_id: currentFolderId || null,
      });

      // Step 4: Show success
      setSuccess(t('upload_success'));
      setUploadedFile({
        name: file.name,
        size: file.size,
        type: file.type,
      });

      if (onUploadSuccess) {
        onUploadSuccess();
      }
    } catch (err) {
      console.error('Upload error:', err);
      const errorMessage = err.response?.data?.error || t('upload_error');
      setError(errorMessage);
      window.toast?.error(errorMessage);
    } finally {
      setUploading(false);
      setProgress(0);
    }
  };

  // MIME type tahmin fonksiyonu
  const guessMimeType = (filename) => {
    const ext = filename.split('.').pop()?.toLowerCase();
    const mimeTypes = {
      // Documents
      pdf: 'application/pdf',
      doc: 'application/msword',
      docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      xls: 'application/vnd.ms-excel',
      xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      ppt: 'application/vnd.ms-powerpoint',
      pptx: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
      txt: 'text/plain',
      csv: 'text/csv',
      rtf: 'application/rtf',
      // Images
      jpg: 'image/jpeg',
      jpeg: 'image/jpeg',
      png: 'image/png',
      gif: 'image/gif',
      webp: 'image/webp',
      bmp: 'image/bmp',
      tiff: 'image/tiff',
      svg: 'image/svg+xml',
      // Code files
      py: 'text/x-python',
      js: 'application/javascript',
      jsx: 'application/javascript',
      ts: 'application/typescript',
      tsx: 'application/typescript',
      json: 'application/json',
      xml: 'application/xml',
      html: 'text/html',
      css: 'text/css',
      md: 'text/markdown',
      yaml: 'application/x-yaml',
      yml: 'application/x-yaml',
      sh: 'application/x-sh',
      bash: 'application/x-sh',
      go: 'text/x-go',
      rs: 'text/x-rust',
      java: 'text/x-java',
      kt: 'text/x-kotlin',
      kts: 'text/x-kotlin',
      cs: 'text/x-csharp',
      php: 'text/x-php',
      rb: 'text/x-ruby',
      pl: 'text/x-perl',
      scala: 'text/x-scala',
      c: 'text/x-c',
      cpp: 'text/x-c++',
      cc: 'text/x-c++',
      cxx: 'text/x-c++',
      h: 'text/x-c',
      hpp: 'text/x-c++',
      sql: 'application/x-sql',
      vue: 'text/plain',
      svelte: 'text/plain',
      swift: 'text/plain',
      dart: 'text/plain',
      lua: 'text/plain',
      r: 'text/plain',
      m: 'text/plain',
      mm: 'text/plain',
      ps1: 'text/plain',
      // Archives
      zip: 'application/zip',
      rar: 'application/x-rar-compressed',
      '7z': 'application/x-7z-compressed',
      gz: 'application/gzip',
      // Audio
      mp3: 'audio/mpeg',
      wav: 'audio/wav',
      flac: 'audio/flac',
      aac: 'audio/aac',
      ogg: 'audio/ogg',
      m4a: 'audio/mp4',
      // Video
      mp4: 'video/mp4',
      avi: 'video/x-msvideo',
      mov: 'video/quicktime',
      wmv: 'video/x-ms-wmv',
      webm: 'video/webm',
      mkv: 'video/x-matroska',
    };
    return mimeTypes[ext] || 'application/octet-stream';
  };

  const handleMultipleFiles = async (files) => {
    setError('');
    setSuccess('');
    setUploadedFile(null);
    setProgress(0);
    setUploading(true);

    try {
      const folderStructure = {};
      const filesByPath = {};

      for (const file of files) {
        if (!file.size || file.size === 0) {
          continue;
        }

        const relativePath = file.webkitRelativePath || file.name;
        const pathParts = relativePath.split('/');
        
        const fileName = pathParts[pathParts.length - 1];
        
        if (!fileName || fileName.trim() === '') {
          continue;
        }
        
        const folderPath = pathParts.slice(0, -1).join('/');
        
        if (folderPath) {
          if (!folderStructure[folderPath]) {
            folderStructure[folderPath] = [];
          }
          folderStructure[folderPath].push(file);
          filesByPath[relativePath] = { file, fileName, folderPath };
        } else {
          filesByPath[relativePath] = { file, fileName, folderPath: '' };
        }
      }

      const totalFiles = Object.keys(filesByPath).length;

      if (totalFiles === 0) {
        setError('Seçilen klasörde yüklenebilir dosya bulunamadı');
        window.toast?.error('Seçilen klasörde yüklenebilir dosya bulunamadı');
        setUploading(false);
        return;
      }

      const folderIdCache = {};
      const sortedFolderPaths = Object.keys(folderStructure).sort((a, b) => {
        return a.split('/').length - b.split('/').length;
      });

      for (const folderPath of sortedFolderPaths) {
        const pathParts = folderPath.split('/');
        let currentParentId = currentFolderId || null;

        for (let i = 0; i < pathParts.length; i++) {
          const partialPath = pathParts.slice(0, i + 1).join('/');
          
          if (folderIdCache[partialPath]) {
            currentParentId = folderIdCache[partialPath];
            continue;
          }

          const folderName = pathParts[i];
          try {
            const folderResponse = await folderApi.createFolder({
              name: folderName,
              folder_id: currentParentId,
              color: '#3b82f6',
            });
            
            const folderId = folderResponse.folder?.id || folderResponse.folder?.ID;
            folderIdCache[partialPath] = folderId;
            currentParentId = folderId;
          } catch (err) {
            console.error(`Klasör oluşturma hatası (${folderName}):`, err);
            if (err.response?.data?.error?.includes('zaten mevcut')) {
              continue;
            }
            throw err;
          }
        }
      }

      let uploadedCount = 0;

      for (const [relativePath, { file, fileName, folderPath }] of Object.entries(filesByPath)) {
        const targetFolderId = folderPath ? folderIdCache[folderPath] : (currentFolderId || null);
        
        const contentType = file.type || guessMimeType(fileName);

        try {
          const presignedResponse = await fileApi.getUploadPresignedURL(fileName, contentType);
          const { presigned_url, minio_path } = presignedResponse;

          await uploadToMinIO(presigned_url, file);

          await fileApi.createFile({
            filename: fileName,
            size: file.size,
            content_type: contentType,
            minio_path: minio_path,
            folder_id: targetFolderId,
          });

          uploadedCount++;
          setProgress((uploadedCount / totalFiles) * 100);
        } catch (err) {
          console.error(`Dosya yükleme hatası (${fileName}):`, err);
          throw err;
        }
      }

      setSuccess(`${uploadedCount} dosya başarıyla yüklendi`);
      if (onUploadSuccess) {
        onUploadSuccess();
      }
    } catch (err) {
      console.error('Klasör yükleme hatası:', err);
      const errorMessage = err.response?.data?.error || 'Klasör yükleme başarısız';
      setError(errorMessage);
      window.toast?.error(errorMessage);
    } finally {
      setUploading(false);
      setProgress(0);
    }
  };

  const uploadToMinIO = async (presignedURL, file) => {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      xhr.upload.addEventListener('progress', e => {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100;
          setProgress(percentComplete);
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status === 200) {
          resolve();
        } else {
          reject(new Error(`Upload failed: ${xhr.status}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      xhr.open('PUT', presignedURL);
      xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
      xhr.send(file);
    });
  };

  // Tek dosya için tam yükleme işlemi (MinIO + MongoDB) - DEPRECATED, handleMultipleFiles kullan
  const uploadSingleFileComplete = async file => {
    try {
      // Step 1: Get presigned URL
      const presignedResponse = await fileApi.getUploadPresignedURL(file.name, file.type || 'application/octet-stream');
      const { presigned_url, minio_path } = presignedResponse;

      // Step 2: Upload file to MinIO using presigned URL
      await uploadToMinIO(presigned_url, file);

      // Step 3: Save file metadata to MongoDB
      await fileApi.createFile({
        filename: file.name,
        size: file.size,
        content_type: file.type || 'application/octet-stream',
        minio_path: minio_path,
        folder_id: currentFolderId || null,
      });
    } catch (error) {
      console.error('File upload error:', error);
      throw error;
    }
  };

  const handleFileSelect = e => {
    if (e.target.files && e.target.files.length > 0) {
      if (isFolderUpload) {
        const files = Array.from(e.target.files);
        handleMultipleFiles(files);
      } else {
        handleFile(e.target.files[0]);
      }
    }
  };

  const resetUpload = () => {
    setError('');
    setSuccess('');
    setUploadedFile(null);
    setProgress(0);
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(102, 126, 234, 0.2)',
          boxShadow: '0 12px 48px 0 rgba(31, 38, 135, 0.25)',
          borderRadius: 3,
        },
      }}
    >
      <DialogTitle
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: 'white',
          fontWeight: 600,
          py: 2,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <CloudUploadIcon />
        {isFolderUpload ? t('upload.title_folder') : t('upload.title_file')}
      </DialogTitle>

      <DialogContent sx={{ pt: 3 }}>
        {/* Upload Mode Toggle - Only show if mode is 'both' */}
        {mode === 'both' && (
          <Box
            sx={{
              mb: 3,
              display: 'flex',
              gap: 1,
              p: 1,
              borderRadius: 2,
              background: 'rgba(102, 126, 234, 0.05)',
              border: '1px solid rgba(102, 126, 234, 0.1)',
            }}
          >
            <Button
              variant={!isFolderUpload ? 'contained' : 'outlined'}
              size="small"
              onClick={() => setIsFolderUpload(false)}
              sx={{
                flex: 1,
                ...(!isFolderUpload && {
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  color: 'white',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
                  },
                }),
                transition: 'all 0.3s ease',
              }}
            >
              {t('upload.mode_single')}
            </Button>
            <Button
              variant={isFolderUpload ? 'contained' : 'outlined'}
              size="small"
              onClick={() => setIsFolderUpload(true)}
              sx={{
                flex: 1,
                ...(isFolderUpload && {
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  color: 'white',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
                  },
                }),
                transition: 'all 0.3s ease',
              }}
            >
              {t('upload.mode_folder')}
            </Button>
          </Box>
        )}

        <input
          ref={fileInputRef}
          type="file"
          onChange={handleFileSelect}
          style={{ display: 'none' }}
          accept={isFolderUpload ? undefined : "image/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.csv,.zip,.rar,.7z,.mp3,.wav,.flac,.aac,.ogg,.m4a,.mp4,.avi,.mov,.wmv,.webm,.mkv,.py,.js,.jsx,.ts,.tsx,.cs,.java,.kt,.kts,.json,.md,.xml,.html,.css,.sh,.bash,.yaml,.yml,.go,.rs,.php,.rb,.pl,.scala,.c,.cpp,.cc,.cxx,.h,.hpp,.sql,.vue,.svelte,.swift,.dart,.lua,.r,.m,.mm,.ps1"}
          multiple={isFolderUpload}
          webkitdirectory={isFolderUpload ? '' : undefined}
          directory={isFolderUpload ? '' : undefined}
          mozdirectory={isFolderUpload ? '' : undefined}
        />

        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <Alert
                severity="error"
                sx={{ mb: 2 }}
                onClose={resetUpload}
                action={
                  <IconButton size="small" onClick={resetUpload}>
                    <CloseIcon fontSize="small" />
                  </IconButton>
                }
              >
                {error}
              </Alert>
            </motion.div>
          )}

          {success && (
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
            >
              <Alert
                severity="success"
                sx={{ mb: 2 }}
                onClose={resetUpload}
                action={
                  <IconButton size="small" onClick={resetUpload}>
                    <CloseIcon fontSize="small" />
                  </IconButton>
                }
              >
                {success}
              </Alert>
            </motion.div>
          )}
        </AnimatePresence>

        <MotionBox
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          sx={{
            border: '2px dashed',
            borderColor: dragActive ? '#667eea' : 'rgba(102, 126, 234, 0.3)',
            background: dragActive
              ? 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)'
              : 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)',
            cursor: uploading ? 'not-allowed' : 'pointer',
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            opacity: uploading ? 0.7 : 1,
            borderRadius: 3,
            p: 4,
            textAlign: 'center',
            boxShadow: dragActive ? '0 8px 24px rgba(102, 126, 234, 0.15)' : 'none',
          }}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
          onClick={() => !uploading && fileInputRef.current?.click()}
        >
          <AnimatePresence mode="wait">
            {uploading ? (
              <motion.div
                key="uploading"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              >
                <Stack spacing={2} alignItems="center">
                  <CloudUploadIcon sx={{ fontSize: 48, color: 'primary.main' }} />
                  <Typography variant="h6" color="primary">
                    {t('uploading')}
                  </Typography>
                  <Box sx={{ width: '100%', maxWidth: 300 }}>
                    <LinearProgress variant="determinate" value={progress} />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                      {t('upload_progress', { progress: Math.round(progress) })}
                    </Typography>
                  </Box>
                </Stack>
              </motion.div>
            ) : uploadedFile ? (
              <motion.div
                key="success"
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.8 }}
              >
                <Stack spacing={2} alignItems="center">
                  <CheckCircleIcon sx={{ fontSize: 48, color: 'success.main' }} />
                  <Typography variant="h6" color="success.main">
                    {t('upload_success')}
                  </Typography>
                  <Box sx={{ textAlign: 'center' }}>
                    <Typography variant="body1" fontWeight={600}>
                      {uploadedFile.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {formatFileSize(uploadedFile.size)}
                    </Typography>
                    <Chip label={uploadedFile.type || t('unknown')} size="small" sx={{ mt: 1 }} />
                  </Box>
                </Stack>
              </motion.div>
            ) : (
              <motion.div
                key="upload"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              >
                <Stack spacing={2} alignItems="center">
                  <CloudUploadIcon sx={{ fontSize: 48, color: 'primary.main' }} />
                  <Typography variant="h6">{t('upload_files')}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {t('drag_drop')}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {t('supported_types')}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {t('max_size')}
                  </Typography>
                </Stack>
              </motion.div>
            )}
          </AnimatePresence>
        </MotionBox>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 3, gap: 1.5 }}>
        <Button
          onClick={onClose}
          variant="outlined"
          sx={{
            borderColor: 'rgba(102, 126, 234, 0.3)',
            color: 'text.secondary',
            '&:hover': {
              borderColor: 'rgba(102, 126, 234, 0.5)',
              backgroundColor: 'rgba(102, 126, 234, 0.05)',
            },
            transition: 'all 0.3s ease',
          }}
        >
          İptal
        </Button>
        {uploadedFile && (
          <Button
            onClick={() => {
              onUploadSuccess();
              resetUpload();
              onClose();
            }}
            variant="contained"
            sx={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              color: 'white',
              fontWeight: 600,
              px: 3,
              '&:hover': {
                background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
                transform: 'translateY(-2px)',
                boxShadow: '0 8px 24px rgba(102, 126, 234, 0.4)',
              },
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            Tamam
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default FileUpload;
