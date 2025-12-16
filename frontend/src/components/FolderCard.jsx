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
import { useTranslation } from 'react-i18next';
import FolderIcon from '@mui/icons-material/Folder';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import FileItemMenu from './FileItemMenu';

const MotionCard = motion.create(Card);
const FolderCard = ({ folder, onOpen, onDelete, onEdit, onShare, onMove, onToggleStar, onRestore, onMenuOpen }) => {
  const { t } = useTranslation();
  const [anchorEl, setAnchorEl] = useState(null);

  const handleMenuOpen = event => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleContextMenu = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (onMenuOpen) {
      onMenuOpen(e, { ...folder, type: 'folder' });
    }
  };

  const getFolderColor = () => {
    if (folder.color) return folder.color;
    const colors = ['#667eea', '#764ba2', '#f093fb', '#4facfe', '#43e97b'];
    return colors[Math.floor(Math.random() * colors.length)];
  };

  return (
    <MotionCard
      data-context-menu-handled
      initial={false}
      whileHover={{ y: -4, boxShadow: '0px 12px 40px rgba(102, 126, 234, 0.15)' }}
      whileTap={{ scale: 0.98 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
      onClick={() => onOpen && onOpen(folder)}
      onContextMenu={handleContextMenu}
      sx={{
        cursor: 'pointer',
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
              background: `linear-gradient(135deg, ${getFolderColor()} 20%, ${getFolderColor()}99 100%)`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mb: 2,
              backdropFilter: 'blur(10px)',
              boxShadow: `0 4px 20px ${getFolderColor()}40`,
            }}
          >
            <FolderIcon sx={{ fontSize: 36, color: 'white' }} />
          </Box>

          <IconButton size="small" onClick={handleMenuOpen} sx={{ mt: -1, mr: -1 }}>
            <MoreVertIcon />
          </IconButton>
        </Box>

        <Typography variant="h6" fontWeight={600} noWrap sx={{ mb: 0.5 }}>
          {folder.name}
        </Typography>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Typography variant="body2" color="text.secondary">
            {folder.item_count
              ? t('folder.items', { count: folder.item_count })
              : t('folder.items_zero')}
          </Typography>
          {folder.isShared && (
            <>
              <Typography variant="body2" color="text.secondary">
                •
              </Typography>
              <Chip
                label={folder.access_type === 'read' ? t('access.read') : t('access.write')}
                size="small"
                color={folder.access_type === 'read' ? 'info' : 'warning'}
                sx={{ fontSize: '0.7rem' }}
              />
              {folder.owner && (
                <>
                  <Typography variant="body2" color="text.secondary">
                    •
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {folder.owner.name || folder.owner.email}
                  </Typography>
                </>
              )}
            </>
          )}
        </Box>
      </CardContent>

      <FileItemMenu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        item={{ ...folder, isTrash: !!folder.deleted_at }}
        itemType="folder"
        onEdit={onEdit}
        onShare={onShare}
        onMove={(item) => onMove(item, 'folder')}
        onToggleStar={onToggleStar}
        onRestore={onRestore}
        onDelete={onDelete}
      />
    </MotionCard>
  );
};

export default FolderCard;
