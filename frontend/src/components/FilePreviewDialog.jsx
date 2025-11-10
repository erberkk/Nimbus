import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  CircularProgress,
  IconButton,
  Chip,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import CloseIcon from '@mui/icons-material/Close';
import DownloadIcon from '@mui/icons-material/Download';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import ImageIcon from '@mui/icons-material/Image';
import AudioFileIcon from '@mui/icons-material/AudioFile';
import VideoFileIcon from '@mui/icons-material/VideoFile';
import DescriptionIcon from '@mui/icons-material/Description';
import TableChartIcon from '@mui/icons-material/TableChart';
import mammoth from 'mammoth';
import * as XLSX from 'xlsx';
import { fileApi } from '../services/api';
import { getFileType, isPreviewable as checkIsPreviewable, formatFileSize } from '../utils/fileUtils';

const MotionDialog = motion.create(Dialog);

const FilePreviewDialog = ({ open, onClose, file, onDownload }) => {
  const [previewUrl, setPreviewUrl] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Get file type using utility function
  const fileType = file ? getFileType(file.content_type, file.filename) : 'unknown';
  const fileIsPreviewable = file ? checkIsPreviewable(file.content_type, file.filename) : false;

  // Load preview URL when dialog opens
  useEffect(() => {
    if (!open || !file) {
      setPreviewUrl(null);
      setError(null);
      setLoading(false);
      return;
    }

    if (fileIsPreviewable) {
      if (fileType !== 'word-doc') {
        loadPreviewUrl(fileType);
      } else {
        // For word-doc, just set loading to false so the message can be shown
        setLoading(false);
      }
    }

    return () => {
      // Cleanup: revoke object URL if created
      if (previewUrl && previewUrl.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, file]);

  const loadPreviewUrl = async (currentFileType) => {
    if (!file) return;

    try {
      setLoading(true);
      setError(null);

      // Use file.id if available (for shared files), otherwise use filename
      const response = await fileApi.getPreviewPresignedURL(file.id, file.filename);
      const presignedUrl = response.presigned_url;

      // For Word/Excel, fetch and convert
      if (currentFileType === 'word-docx' || currentFileType === 'excel') {
        await loadAndConvertFile(presignedUrl, currentFileType);
      } else {
        // For other types, just set the URL
        setPreviewUrl(presignedUrl);
      }
    } catch (err) {
      console.error('Preview URL yükleme hatası:', err);
      setError('Dosya önizlemesi yüklenemedi');
      window.toast?.error('Dosya önizlemesi yüklenemedi');
    } finally {
      setLoading(false);
    }
  };

  const loadAndConvertFile = async (url, currentFileType) => {
    try {
      // Fetch file as array buffer
      const response = await fetch(url);
      if (!response.ok) throw new Error('Dosya yüklenemedi');
      const arrayBuffer = await response.arrayBuffer();

      if (currentFileType === 'word-docx') {
        // Convert Word to HTML
        const result = await mammoth.convertToHtml({ arrayBuffer });
        setPreviewUrl(result.value); // Store HTML content
        if (result.messages.length > 0) {
          console.warn('Word conversion warnings:', result.messages);
        }
      } else if (currentFileType === 'excel') {
        // Convert Excel to JSON and render as table
        const workbook = XLSX.read(arrayBuffer, { type: 'array' });
        const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
        const html = XLSX.utils.sheet_to_html(firstSheet);
        setPreviewUrl(html);
      }
    } catch (err) {
      console.error('Dosya dönüştürme hatası:', err);
      setError('Dosya önizleme için dönüştürülemedi');
      throw err;
    }
  };

  const handleDownload = () => {
    if (onDownload && file) {
      onDownload(file);
    }
    onClose();
  };

  const renderPreview = () => {
    if (loading) {
      return (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '400px',
            gap: 2,
          }}
        >
          <CircularProgress size={48} />
          <Typography variant="body2" color="text.secondary">
            {fileType === 'word-docx' || fileType === 'excel' 
              ? 'Dosya dönüştürülüyor...' 
              : 'Dosya yükleniyor...'}
          </Typography>
        </Box>
      );
    }

    if (error) {
      return (
        <Box
          sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '400px',
            gap: 2,
            p: 3,
          }}
        >
          <Typography variant="h6" color="error">
            Hata
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {error}
          </Typography>
        </Box>
      );
    }

    // For word-doc, show message without previewUrl
    if (!previewUrl && fileType !== 'word-doc') {
      return null;
    }

    switch (fileType) {
      case 'pdf':
        return (
          <Box
            sx={{
              width: '100%',
              height: '80vh',
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 1,
              overflow: 'hidden',
            }}
          >
            <iframe
              src={previewUrl}
              style={{
                width: '100%',
                height: '100%',
                border: 'none',
              }}
              title={file.filename}
            />
          </Box>
        );

      case 'image':
        return (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              p: 2,
            }}
          >
            <img
              src={previewUrl}
              alt={file.filename}
              style={{
                maxWidth: '100%',
                maxHeight: '80vh',
                objectFit: 'contain',
                borderRadius: 8,
              }}
              onError={() => {
                setError('Resim yüklenemedi');
              }}
            />
          </Box>
        );

      case 'audio':
        return (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              p: 3,
              gap: 2,
            }}
          >
            <AudioFileIcon sx={{ fontSize: 64, color: 'primary.main' }} />
            <audio
              controls
              src={previewUrl}
              style={{
                width: '100%',
                maxWidth: '600px',
              }}
            >
              Tarayıcınız audio elementi desteklemiyor.
            </audio>
            <Typography variant="body2" color="text.secondary">
              {file.filename}
            </Typography>
          </Box>
        );

      case 'video':
        return (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              p: 2,
            }}
          >
            <video
              controls
              src={previewUrl}
              style={{
                maxWidth: '100%',
                maxHeight: '80vh',
                borderRadius: 8,
              }}
              onError={() => {
                setError('Video yüklenemedi');
              }}
            >
              Tarayıcınız video elementi desteklemiyor.
            </video>
          </Box>
        );

      case 'word-docx':
        return (
          <Box
            sx={{
              width: '100%',
              maxHeight: '80vh',
              overflow: 'auto',
              p: 3,
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 1,
              bgcolor: 'background.paper',
            }}
          >
            <Box
              component="div"
              dangerouslySetInnerHTML={{ __html: previewUrl }}
              sx={{
                '& p': {
                  margin: '0.5em 0',
                },
                '& ul, & ol': {
                  margin: '0.5em 0',
                  paddingLeft: '1.5em',
                },
                '& table': {
                  borderCollapse: 'collapse',
                  width: '100%',
                  margin: '1em 0',
                },
                '& td, & th': {
                  border: '1px solid',
                  borderColor: 'divider',
                  padding: '0.5em',
                },
              }}
            />
          </Box>
        );

      case 'excel':
        return (
          <Box
            sx={{
              width: '100%',
              maxHeight: '80vh',
              overflow: 'auto',
              p: 2,
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 1,
            }}
          >
            <Box
              component="div"
              dangerouslySetInnerHTML={{ __html: previewUrl }}
              sx={{
                '& table': {
                  borderCollapse: 'collapse',
                  width: '100%',
                },
                '& td, & th': {
                  border: '1px solid',
                  borderColor: 'divider',
                  padding: 1,
                  textAlign: 'left',
                  fontSize: '0.875rem',
                },
                '& th': {
                  bgcolor: 'grey.100',
                  fontWeight: 600,
                  position: 'sticky',
                  top: 0,
                  zIndex: 1,
                },
                '& tr:nth-of-type(even)': {
                  bgcolor: 'grey.50',
                },
              }}
            />
          </Box>
        );

      case 'word-doc':
        return (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              gap: 2,
              p: 3,
            }}
          >
            <DescriptionIcon sx={{ fontSize: 64, color: 'text.secondary' }} />
            <Typography variant="h6" color="text.secondary">
              Eski format Word dosyası (.doc) için önizleme desteklenmiyor
            </Typography>
            <Typography variant="body2" color="text.secondary" textAlign="center">
              Dosyayı görüntülemek için indirin veya .docx formatına dönüştürün
            </Typography>
            <Button variant="contained" onClick={handleDownload} startIcon={<DownloadIcon />}>
              İndir
            </Button>
          </Box>
        );

      default:
        return (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              gap: 2,
            }}
          >
            <Typography variant="h6" color="text.secondary">
              Bu dosya türü için önizleme desteklenmiyor
            </Typography>
            <Button variant="contained" onClick={handleDownload} startIcon={<DownloadIcon />}>
              İndir
            </Button>
          </Box>
        );
    }
  };

  const getFileIcon = () => {
    switch (fileType) {
      case 'pdf':
        return <PictureAsPdfIcon sx={{ fontSize: 24 }} />;
      case 'image':
        return <ImageIcon sx={{ fontSize: 24 }} />;
      case 'audio':
        return <AudioFileIcon sx={{ fontSize: 24 }} />;
      case 'video':
        return <VideoFileIcon sx={{ fontSize: 24 }} />;
      case 'word-docx':
      case 'word-doc':
        return <DescriptionIcon sx={{ fontSize: 24 }} />;
      case 'excel':
        return <TableChartIcon sx={{ fontSize: 24 }} />;
      default:
        return null;
    }
  };

  if (!file) return null;

  return (
    <MotionDialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      PaperProps={{
        sx: {
          borderRadius: 2,
          maxHeight: '90vh',
        },
      }}
      TransitionComponent={motion.div}
      transition={{
        type: 'spring',
        damping: 25,
        stiffness: 200,
      }}
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, flex: 1, minWidth: 0 }}>
            {getFileIcon()}
            <Box sx={{ flex: 1, minWidth: 0 }}>
              <Typography variant="h6" noWrap title={file.filename}>
                {file.filename}
              </Typography>
              <Box sx={{ display: 'flex', gap: 1, mt: 0.5 }}>
                <Chip
                  label={file.content_type?.split('/')[1] || 'dosya'}
                  size="small"
                  variant="outlined"
                />
                <Chip label={formatFileSize(file.size)} size="small" variant="outlined" />
              </Box>
            </Box>
          </Box>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent dividers sx={{ p: 0, position: 'relative' }}>
        <AnimatePresence mode="wait">
          <motion.div
            key={previewUrl || 'loading'}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            style={{ width: '100%' }}
          >
            {renderPreview()}
          </motion.div>
        </AnimatePresence>
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={onClose} variant="outlined">
          Kapat
        </Button>
        {onDownload && (
          <Button
            onClick={handleDownload}
            variant="contained"
            startIcon={<DownloadIcon />}
            sx={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #5568d3 0%, #6a4190 100%)',
              },
            }}
          >
            İndir
          </Button>
        )}
      </DialogActions>
    </MotionDialog>
  );
};

export default FilePreviewDialog;

