import {
  Card,
  CardContent,
  Typography,
  IconButton,
  Box,
  Chip,
  CircularProgress,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import ImageIcon from '@mui/icons-material/Image';
import DescriptionIcon from '@mui/icons-material/Description';
import ArchiveIcon from '@mui/icons-material/Archive';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import FileItemMenu from './FileItemMenu';
import {
  isPreviewable,
  isAskableFile,
  formatFileSize,
  formatDate,
  formatContentType,
} from '../utils/fileUtils';

const MotionCard = motion.create(Card);

const getFileIcon = contentType => {
  if (!contentType) return <InsertDriveFileIcon sx={{ fontSize: 36 }} />;

  if (contentType.startsWith('image/')) {
    return <ImageIcon sx={{ fontSize: 36, color: '#4facfe' }} />;
  } else if (contentType.includes('pdf') || contentType.includes('document')) {
    return <DescriptionIcon sx={{ fontSize: 36, color: '#f093fb' }} />;
  } else if (contentType.includes('zip') || contentType.includes('rar')) {
    return <ArchiveIcon sx={{ fontSize: 36, color: '#43e97b' }} />;
  }

  return <InsertDriveFileIcon sx={{ fontSize: 36, color: '#667eea' }} />;
};

const FileCard = ({
  file,
  onDownload,
  onDelete,
  onMove,
  onShare,
  onAskNimbus,
  onPreview,
  onEdit,
  onInfo,
  onToggleStar,
  onRestore,
  onMenuOpen,
}) => {
  const { t } = useTranslation();
  const [anchorEl, setAnchorEl] = useState(null);

  const handleMenuOpen = event => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  // Check if file is previewable using utility function
  const fileIsPreviewable = isPreviewable(file?.content_type, file?.filename);

  const handleCardClick = e => {
    // Don't trigger preview if clicking on menu button
    if (e.target.closest('button') || e.target.closest('[role="button"]')) {
      return;
    }
    if (fileIsPreviewable && onPreview) {
      onPreview(file);
    }
  };

  // Word ve PDF dosyaları için Nimbus'a Sor seçeneğini göster
  const fileIsAskable = isAskableFile(file?.content_type, file?.filename);

  const handleContextMenu = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (onMenuOpen) {
      onMenuOpen(e, { ...file, type: 'file' });
    }
  };

  return (
    <MotionCard
      data-context-menu-handled
      initial={false}
      whileHover={{ y: -4, boxShadow: '0px 12px 40px rgba(102, 126, 234, 0.15)' }}
      whileTap={{ scale: 0.98 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
      onClick={handleCardClick}
      onContextMenu={handleContextMenu}
      sx={{
        cursor: fileIsPreviewable ? 'pointer' : 'default',
        position: 'relative',
        transition: 'all 0.3s ease',
        height: '100%',
        background: 'rgba(255, 255, 255, 0.7)',
        backdropFilter: 'blur(10px)',
        border: '1px solid rgba(255, 255, 255, 0.5)',
        borderRadius: 3,
        '&:hover': {
          background: 'rgba(255, 255, 255, 0.85)',
          border: '1px solid rgba(102, 126, 234, 0.3)',
        },
      }}
    >
      <CardContent sx={{ p: 2.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box
            sx={{
              width: 60,
              height: 60,
              borderRadius: 2,
              background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mb: 2,
              backdropFilter: 'blur(10px)',
              border: '1px solid rgba(102, 126, 234, 0.2)',
            }}
          >
            {getFileIcon(file.content_type)}
          </Box>

          <IconButton size="small" onClick={handleMenuOpen} sx={{ mt: -1, mr: -1 }}>
            <MoreVertIcon />
          </IconButton>
        </Box>

        <Typography variant="body1" fontWeight={600} noWrap sx={{ mb: 0.5 }} title={file.filename}>
          {file.filename}
        </Typography>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Typography variant="body2" color="text.secondary">
            {formatFileSize(file.size)}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            •
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {formatDate(file.created_at)}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1, flexWrap: 'wrap' }}>
          <Chip
            label={formatContentType(file.content_type, file.filename)}
            size="small"
            variant="outlined"
            sx={{ fontSize: '0.7rem' }}
          />
          {fileIsAskable && file.processing_status && (
            <Chip
              label={
                file.processing_status === 'processing'
                  ? t('ai.processing')
                  : file.processing_status === 'completed'
                    ? t('ai.ready')
                    : file.processing_status === 'failed'
                      ? t('ai.failed')
                      : t('ai.pending')
              }
              size="small"
              color={
                file.processing_status === 'completed'
                  ? 'success'
                  : file.processing_status === 'failed'
                    ? 'error'
                    : 'warning'
              }
              icon={
                file.processing_status === 'processing' ? (
                  <CircularProgress size={12} sx={{ color: 'inherit' }} />
                ) : undefined
              }
              sx={{ fontSize: '0.7rem' }}
            />
          )}
          {file.isShared && (
            <>
              <Chip
                label={file.access_type === 'read' ? t('access.read') : t('access.write')}
                size="small"
                color={file.access_type === 'read' ? 'info' : 'warning'}
                sx={{ fontSize: '0.7rem' }}
              />
              {file.owner && (
                <Typography variant="caption" color="text.secondary">
                  {file.owner.name || file.owner.email}
                </Typography>
              )}
            </>
          )}
        </Box>
      </CardContent>

      <FileItemMenu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        item={{ ...file, isTrash: !!file.deleted_at }}
        itemType="file"
        onInfo={onInfo}
        onDownload={onDownload}
        onEdit={onEdit}
        onAskNimbus={onAskNimbus}
        onShare={onShare}
        onMove={(item) => onMove(item, 'file')}
        onToggleStar={onToggleStar}
        onRestore={onRestore}
        onDelete={onDelete}
      />
    </MotionCard>
  );
};

export default FileCard;
