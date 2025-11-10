import {
  Card,
  CardContent,
  Typography,
  IconButton,
  Box,
  Menu,
  MenuItem,
  Chip,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useState } from 'react';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import ImageIcon from '@mui/icons-material/Image';
import DescriptionIcon from '@mui/icons-material/Description';
import ArchiveIcon from '@mui/icons-material/Archive';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMove';
import ShareIcon from '@mui/icons-material/Share';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import EditIcon from '@mui/icons-material/Edit';
import InfoIcon from '@mui/icons-material/Info';
import { isPreviewable, isAskableFile, isEditable, formatFileSize, formatDate, formatContentType } from '../utils/fileUtils';

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


const FileCard = ({ file, onDownload, onDelete, onMove, onShare, onAskNimbus, onPreview, onEdit, onInfo }) => {
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

  const handleCardClick = (e) => {
    // Don't trigger preview if clicking on menu button
    if (e.target.closest('button') || e.target.closest('[role="button"]')) {
      return;
    }
    if (fileIsPreviewable && onPreview) {
      onPreview(file);
    }
  };

  const handleDownload = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onDownload) {
      onDownload(file);
    }
  };

  const handleDelete = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onDelete) {
      onDelete(file);
    }
  };

  const handleMove = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onMove) {
      onMove(file);
    }
  };

  const handleShare = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onShare) {
      onShare(file);
    }
  };

  const handleAskNimbus = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onAskNimbus) {
      onAskNimbus(file);
    }
  };

  const handleEdit = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onEdit) {
      onEdit(file);
    }
  };

  const handleInfo = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onInfo) {
      onInfo(file);
    }
  };

  // Word ve PDF dosyaları için Nimbus'a Sor seçeneğini göster
  const fileIsAskable = isAskableFile(file?.content_type, file?.filename);
  
  // Word, Excel, PowerPoint dosyaları için Düzenle seçeneğini göster
  const fileIsEditable = isEditable(file?.content_type, file?.filename);

  return (
    <MotionCard
      whileHover={{ y: -4, boxShadow: '0px 8px 24px rgba(0,0,0,0.12)' }}
      whileTap={{ scale: 0.98 }}
      onClick={handleCardClick}
      sx={{
        cursor: fileIsPreviewable ? 'pointer' : 'default',
        position: 'relative',
        transition: 'all 0.3s ease',
        height: '100%',
      }}
    >
      <CardContent sx={{ p: 2.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box
            sx={{
              width: 60,
              height: 60,
              borderRadius: 2,
              backgroundColor: 'grey.100',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mb: 2,
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

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Chip
            label={formatContentType(file.content_type, file.filename)}
            size="small"
            variant="outlined"
            sx={{ fontSize: '0.7rem' }}
          />
          {file.isShared && (
            <>
              <Chip
                label={file.access_type === 'read' ? 'Görüntüleme' : 'Düzenleme'}
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

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        onClick={e => e.stopPropagation()}
      >
        <MenuItem onClick={handleInfo}>
          <InfoIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Bilgi
        </MenuItem>
        <MenuItem onClick={handleDownload}>
          <DownloadIcon sx={{ mr: 1.5, fontSize: 20 }} />
          İndir
        </MenuItem>
        {fileIsEditable && (
          <MenuItem onClick={handleEdit} sx={{ color: '#667eea' }}>
            <EditIcon sx={{ mr: 1.5, fontSize: 20 }} />
            Düzenle
          </MenuItem>
        )}
        {fileIsAskable && (
          <MenuItem onClick={handleAskNimbus} sx={{ color: '#667eea' }}>
            <SmartToyIcon sx={{ mr: 1.5, fontSize: 20 }} />
            Nimbus'a Sor
          </MenuItem>
        )}
        <MenuItem onClick={handleShare}>
          <ShareIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Paylaş
        </MenuItem>
        <MenuItem onClick={handleMove}>
          <DriveFileMoveIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Taşı
        </MenuItem>
        <MenuItem onClick={handleDelete} sx={{ color: 'error.main' }}>
          <DeleteIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Sil
        </MenuItem>
      </Menu>
    </MotionCard>
  );
};

export default FileCard;
