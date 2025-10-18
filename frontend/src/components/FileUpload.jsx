import { useState, useRef } from 'react';
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
import { fileApi } from '../services/api';

const MotionBox = motion.create(Box);

const FileUpload = ({ open, onClose, onUploadSuccess, userId, currentFolderId }) => {
  const { t } = useTranslation();
  const [dragActive, setDragActive] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [uploadedFile, setUploadedFile] = useState(null);
  const fileInputRef = useRef(null);

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

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  const handleFile = async file => {
    // Reset states
    setError('');
    setSuccess('');
    setUploadedFile(null);
    setProgress(0);

    // Validate file
    if (!file) {
      setError(t('select_file'));
      return;
    }

    // Check file size (100MB limit)
    if (file.size > 100 * 1024 * 1024) {
      setError(t('file_too_large'));
      return;
    }

    try {
      setUploading(true);

      // Step 1: Get presigned URL
      const presignedResponse = await fileApi.getUploadPresignedURL(file.name, file.type);
      const { presigned_url, minio_path } = presignedResponse;

      // Step 2: Upload file to MinIO using presigned URL
      await uploadToMinIO(presigned_url, file);

      // Step 3: Save file metadata to MongoDB
      await fileApi.createFile({
        filename: file.name,
        size: file.size,
        content_type: file.type,
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

      // Call success callback
      if (onUploadSuccess) {
        onUploadSuccess(file.name);
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

  const handleFileSelect = e => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0]);
    }
  };

  const resetUpload = () => {
    setError('');
    setSuccess('');
    setUploadedFile(null);
    setProgress(0);
  };

  const formatFileSize = bytes => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <CloudUploadIcon sx={{ color: 'primary.main' }} />
          <Typography variant="h6">Dosya Yükle</Typography>
        </Box>
      </DialogTitle>

      <DialogContent>
        <input
          ref={fileInputRef}
          type="file"
          onChange={handleFileSelect}
          style={{ display: 'none' }}
          accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.csv,.zip,.rar,.7z"
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
            borderColor: dragActive ? 'primary.main' : 'grey.300',
            backgroundColor: dragActive ? 'primary.50' : 'background.paper',
            cursor: uploading ? 'not-allowed' : 'pointer',
            transition: 'all 0.3s ease',
            opacity: uploading ? 0.7 : 1,
            borderRadius: 2,
            p: 4,
            textAlign: 'center',
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

        <DialogActions sx={{ p: 2 }}>
          <Button onClick={onClose} variant="outlined">
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
