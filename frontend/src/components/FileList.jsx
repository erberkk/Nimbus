import { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Chip,
  Button,
  Stack,
  Alert,
  CircularProgress,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import ImageIcon from '@mui/icons-material/Image';
import DescriptionIcon from '@mui/icons-material/Description';
import ArchiveIcon from '@mui/icons-material/Archive';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import { fileApi } from '../services/api';
import { formatFileSize, formatDate } from '../utils/fileUtils';

const MotionCard = motion.create(Card);

const getFileIcon = contentType => {
  if (contentType.startsWith('image/')) {
    return <ImageIcon color="primary" />;
  }
  if (
    contentType.includes('pdf') ||
    contentType.includes('document') ||
    contentType.includes('text')
  ) {
    return <DescriptionIcon color="secondary" />;
  }
  if (contentType.includes('zip') || contentType.includes('rar') || contentType.includes('7z')) {
    return <ArchiveIcon color="warning" />;
  }
  return <InsertDriveFileIcon color="action" />;
};

const FileList = ({ refreshTrigger }) => {
  const { t } = useTranslation();
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [downloadingFile, setDownloadingFile] = useState(null);

  const loadFiles = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const response = await fileApi.listFiles();
      setFiles(response.files || []);
    } catch (err) {
      console.error('Dosya listesi yükleme hatası:', err);
      setError(t('files_error'));
      window.toast?.error(t('files_error'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadFiles();
  }, [refreshTrigger]);

  const handleDownload = async file => {
    try {
      setDownloadingFile(file.id);

      // Get presigned download URL
      const response = await fileApi.getDownloadPresignedURL(file.filename);
      const { presigned_url } = response;

      // Open in new tab
      window.open(presigned_url, '_blank');
    } catch (err) {
      console.error('İndirme hatası:', err);
      setError(t('network_error'));
      window.toast?.error(t('network_error'));
    } finally {
      setDownloadingFile(null);
    }
  };

  const handleDelete = async file => {
    if (!window.confirm(t('confirm_delete'))) {
      return;
    }

    try {
      await fileApi.deleteFile(file.id);
      window.toast?.success(t('delete_success'));
      loadFiles();
    } catch (err) {
      console.error('Silme hatası:', err);
      window.toast?.error(t('delete_error'));
    }
  };

  const handleRefresh = () => {
    loadFiles();
  };

  if (loading) {
    return (
      <MotionCard initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
        <CardContent sx={{ textAlign: 'center', py: 4 }}>
          <CircularProgress size={40} />
          <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
            {t('files_loading')}
          </Typography>
        </CardContent>
      </MotionCard>
    );
  }

  if (error) {
    return (
      <MotionCard initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
        <CardContent>
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
          <Button startIcon={<RefreshIcon />} onClick={handleRefresh} variant="outlined">
            {t('refresh')}
          </Button>
        </CardContent>
      </MotionCard>
    );
  }

  if (files.length === 0) {
    return (
      <MotionCard initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
        <CardContent sx={{ textAlign: 'center', py: 4 }}>
          <InsertDriveFileIcon sx={{ fontSize: 60, color: 'text.disabled', mb: 2 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            {t('no_files')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('no_files_desc')}
          </Typography>
        </CardContent>
      </MotionCard>
    );
  }

  return (
    <MotionCard initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}>
      <CardContent sx={{ p: 0 }}>
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
          <Stack direction="row" justifyContent="space-between" alignItems="center">
            <Typography variant="h6" fontWeight={600}>
              {t('my_files')} ({files.length})
            </Typography>
            <Button
              startIcon={<RefreshIcon />}
              onClick={handleRefresh}
              size="small"
              variant="outlined"
            >
              {t('refresh')}
            </Button>
          </Stack>
        </Box>

        <List sx={{ py: 0 }}>
          <AnimatePresence>
            {files.map((file, index) => (
              <motion.div
                key={file.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                transition={{ duration: 0.3, delay: index * 0.05 }}
              >
                <ListItem
                  sx={{
                    px: 2,
                    py: 1.5,
                    borderBottom: index < files.length - 1 ? 1 : 0,
                    borderColor: 'divider',
                    '&:hover': {
                      backgroundColor: 'action.hover',
                    },
                  }}
                >
                  <ListItemIcon>{getFileIcon(file.content_type)}</ListItemIcon>

                  <ListItemText
                    primary={
                      <Typography variant="body1" fontWeight={500}>
                        {file.filename}
                      </Typography>
                    }
                    secondary={
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Typography variant="body2" color="text.secondary">
                          {formatFileSize(file.size)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          •
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          {formatDate(file.created_at)}
                        </Typography>
                      </Stack>
                    }
                  />

                  <ListItemSecondaryAction>
                    <Stack direction="row" spacing={0.5}>
                      <Chip
                        label={file.content_type || t('unknown')}
                        size="small"
                        variant="outlined"
                        sx={{ maxWidth: 120 }}
                      />
                      <IconButton
                        size="small"
                        onClick={() => handleDownload(file)}
                        disabled={downloadingFile === file.id}
                      >
                        {downloadingFile === file.id ? (
                          <CircularProgress size={16} />
                        ) : (
                          <DownloadIcon />
                        )}
                      </IconButton>
                      <IconButton size="small" color="error" onClick={() => handleDelete(file)}>
                        <DeleteIcon />
                      </IconButton>
                    </Stack>
                  </ListItemSecondaryAction>
                </ListItem>
              </motion.div>
            ))}
          </AnimatePresence>
        </List>
      </CardContent>
    </MotionCard>
  );
};

export default FileList;
