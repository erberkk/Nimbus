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
import { Slide } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DownloadIcon from '@mui/icons-material/Download';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import ImageIcon from '@mui/icons-material/Image';
import AudioFileIcon from '@mui/icons-material/AudioFile';
import VideoFileIcon from '@mui/icons-material/VideoFile';
import DescriptionIcon from '@mui/icons-material/Description';
import TableChartIcon from '@mui/icons-material/TableChart';
import { fileApi } from '../services/api';
import { getFileType, isPreviewable as checkIsPreviewable, formatFileSize, isEditable, formatContentType, isCodeFile } from '../utils/fileUtils';
import OnlyOfficeEditor from './OnlyOfficeEditor';
import CodeEditor from './CodeEditor';

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const FilePreviewDialog = ({ open, onClose, file, onDownload, onSave }) => {
  const [previewUrl, setPreviewUrl] = useState(null);
  const [codeContent, setCodeContent] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  const fileType = file ? getFileType(file.content_type, file.filename) : 'unknown';
  const fileIsPreviewable = file ? checkIsPreviewable(file.content_type, file.filename) : false;
  const fileIsEditable = file ? isEditable(file.content_type, file.filename) : false;
  const fileIsCodeFile = file ? isCodeFile(file.content_type, file.filename) : false;

  useEffect(() => {
    if (!open || !file) {
      setPreviewUrl(null);
      setCodeContent(null);
      setError(null);
      setLoading(false);
      return;
    }

    // If it's a code file, load content
    if (fileIsCodeFile) {
      loadCodeContent();
      return;
    }

    // If it's an Office document editable with OnlyOffice
    if (fileIsEditable && !fileIsCodeFile) {
      setLoading(false);
      return;
    }

    // For other previewable files (not code, not Office)
    if (fileIsPreviewable && fileType !== 'word-doc' && !fileIsCodeFile) {
      loadPreviewUrl();
    } else if (fileType === 'word-doc') {
      setLoading(false);
    }

    return () => {
      if (previewUrl && previewUrl.startsWith('blob:')) {
        URL.revokeObjectURL(previewUrl);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, file, fileIsEditable, fileIsPreviewable, fileIsCodeFile]);

  const loadPreviewUrl = async () => {
    if (!file) return;

    try {
      setLoading(true);
      setError(null);

      const response = await fileApi.getPreviewPresignedURL(file.id, file.filename);
      const presignedUrl = response.presigned_url;

      setPreviewUrl(presignedUrl);
    } catch (err) {
      setError('Dosya önizlemesi yüklenemedi');
      window.toast?.error('Dosya önizlemesi yüklenemedi');
    } finally {
      setLoading(false);
    }
  };

  const loadCodeContent = async (forceRefresh = false) => {
    if (!file) return;

    try {
      setLoading(true);
      setError(null);

      const content = await fileApi.getFileContent(file.id, forceRefresh);
      setCodeContent(content);
    } catch (err) {
      setError('Dosya içeriği yüklenemedi');
      window.toast?.error('Dosya içeriği yüklenemedi');
    } finally {
      setLoading(false);
    }
  };

  const handleCodeSave = async (content) => {
    if (!file) return;

    try {
      setSaving(true);
      await fileApi.updateFileContent(file.id, content);
      setCodeContent(content);
      window.toast?.success('Dosya başarıyla kaydedildi');
      if (onSave) {
        onSave();
      }
    } catch (err) {
      window.toast?.error('Dosya kaydedilemedi');
    } finally {
      setSaving(false);
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
            {fileType === 'code'
              ? 'Kod dosyası yükleniyor...'
              : fileType === 'word-docx' || fileType === 'excel' 
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
    // For code files, codeContent will be loaded separately
    if (!previewUrl && fileType !== 'word-doc' && fileType !== 'code') {
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

      case 'code':
        return (
          <Box sx={{ width: '100%', height: '100%' }}>
            <CodeEditor
              file={file}
              content={codeContent}
              readOnly={!fileIsEditable || (file?.isShared && file?.access_type === 'read')}
              onChange={(content) => {
                // Content changed, will be saved on Ctrl+S
              }}
              onSave={fileIsEditable && (!file?.isShared || file?.access_type !== 'read') ? handleCodeSave : null}
            />
          </Box>
        );

      case 'word-docx':
      case 'excel':
      case 'powerpoint':
        // These file types now use OnlyOffice preview, so this shouldn't be reached
        // But keep as fallback
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
            <Typography variant="body2" color="text.secondary">
              Bu dosya türü için OnlyOffice önizleme kullanılmalı
            </Typography>
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

  // If it's an Office document (editable with OnlyOffice), use OnlyOffice
  if (fileIsEditable && !fileIsCodeFile) {
    return (
      <OnlyOfficeEditor
        open={open}
        onClose={onClose}
        file={file}
        mode="view"
        onSave={null}
      />
    );
  }

  // If it's a code file, it will be handled in renderPreview

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="lg"
      fullWidth
      TransitionComponent={Transition}
      PaperProps={{
        sx: {
          borderRadius: 2,
          maxHeight: '90vh',
        },
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
                  label={formatContentType(file.content_type, file.filename)}
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

      <DialogContent 
        dividers 
        sx={{ 
          p: 0, 
          position: 'relative',
          minHeight: fileIsCodeFile ? '70vh' : 'auto',
          height: fileIsCodeFile ? '70vh' : 'auto',
        }}
      >
        <AnimatePresence mode="wait">
          <motion.div
            key={previewUrl || 'loading'}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            style={{ width: '100%', height: '100%' }}
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
    </Dialog>
  );
};

export default FilePreviewDialog;

