import { Card, CardContent, Typography, IconButton, Box, Menu, MenuItem } from '@mui/material';
import { motion } from 'framer-motion';
import { useState } from 'react';
import FolderIcon from '@mui/icons-material/Folder';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import ShareIcon from '@mui/icons-material/Share';

const MotionCard = motion.create(Card);

const FolderCard = ({ folder, onOpen, onDelete, onEdit, onShare }) => {
  const [anchorEl, setAnchorEl] = useState(null);

  const handleMenuOpen = event => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleDelete = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onDelete) {
      onDelete(folder);
    }
  };

  const handleEdit = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onEdit) {
      onEdit(folder);
    }
  };

  const handleShare = e => {
    e.stopPropagation();
    handleMenuClose();
    if (onShare) {
      onShare(folder);
    }
  };

  const getFolderColor = () => {
    if (folder.color) return folder.color;
    const colors = ['#667eea', '#764ba2', '#f093fb', '#4facfe', '#43e97b'];
    return colors[Math.floor(Math.random() * colors.length)];
  };

  return (
    <MotionCard
      whileHover={{ y: -4, boxShadow: '0px 8px 24px rgba(0,0,0,0.12)' }}
      whileTap={{ scale: 0.98 }}
      onClick={() => onOpen && onOpen(folder)}
      sx={{
        cursor: 'pointer',
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
              background: `linear-gradient(135deg, ${getFolderColor()} 0%, ${getFolderColor()}dd 100%)`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mb: 2,
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
            {folder.item_count || 0} öğe
          </Typography>
          {folder.isShared && (
            <>
              <Typography variant="body2" color="text.secondary">
                •
              </Typography>
              <Chip
                label={folder.access_type === 'read' ? 'Görüntüleme' : 'Düzenleme'}
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

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
        onClick={e => e.stopPropagation()}
      >
        <MenuItem onClick={handleEdit}>
          <EditIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Yeniden Adlandır
        </MenuItem>
        <MenuItem onClick={handleShare}>
          <ShareIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Paylaş
        </MenuItem>
        <MenuItem onClick={handleDelete} sx={{ color: 'error.main' }}>
          <DeleteIcon sx={{ mr: 1.5, fontSize: 20 }} />
          Sil
        </MenuItem>
      </Menu>
    </MotionCard>
  );
};

export default FolderCard;
