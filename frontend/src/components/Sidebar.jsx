import {
  Box,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Divider,
  Typography,
  LinearProgress,
  Button,
  Menu,
  MenuItem,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useState, useEffect } from 'react';
import { folderApi } from '../services/api';
import HomeIcon from '@mui/icons-material/Home';
import FolderSharedIcon from '@mui/icons-material/FolderShared';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import StarBorderIcon from '@mui/icons-material/StarBorder';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import CloudQueueIcon from '@mui/icons-material/CloudQueue';
import AddIcon from '@mui/icons-material/Add';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import FolderIcon from '@mui/icons-material/Folder';

const MotionBox = motion.create(Box);

const Sidebar = ({ onCreateFolder, onFileUpload, onMenuChange, selectedMenu }) => {
  const { t } = useTranslation();
  const [selected, setSelected] = useState(selectedMenu || 'home');
  const [newMenuAnchor, setNewMenuAnchor] = useState(null);

  const menuItems = [
    { id: 'home', icon: <HomeIcon />, label: t('sidebar.home') || 'Home', primary: true },
    {
      id: 'shared',
      icon: <FolderSharedIcon />,
      label: t('sidebar.shared') || 'Shared with me',
    },
    { id: 'recent', icon: <AccessTimeIcon />, label: t('sidebar.recent') || 'Recent' },
    { id: 'starred', icon: <StarBorderIcon />, label: t('sidebar.starred') || 'Starred' },
    { id: 'trash', icon: <DeleteOutlineIcon />, label: t('sidebar.trash') || 'Trash' },
  ];

  const [storageInfo, setStorageInfo] = useState({
    usage: '0 MB',
    usageGB: 0,
    percent: 0,
  });

  useEffect(() => {
    loadStorageInfo();
  }, []);

  useEffect(() => {
    setSelected(selectedMenu || 'home');
  }, [selectedMenu]);

  const loadStorageInfo = async () => {
    try {
      const response = await folderApi.getStorageUsage();
      const totalGB = 15; // Toplam 15 GB
      const usedGB = response.usage_gb || 0;
      const percent = (usedGB / totalGB) * 100;

      setStorageInfo({
        usage: response.usage || '0 MB',
        usageGB: usedGB,
        percent: Math.min(percent, 100), // Maksimum %100
      });
    } catch (error) {
      console.error('Depolama bilgisi yüklenemedi:', error);
      // Hata durumunda varsayılan değerleri kullan
      setStorageInfo({
        usage: '0 MB',
        usageGB: 0,
        percent: 0,
      });
    }
  };

  const handleNewMenuOpen = event => {
    setNewMenuAnchor(event.currentTarget);
  };

  const handleNewMenuClose = () => {
    setNewMenuAnchor(null);
  };

  const handleCreateFolder = () => {
    handleNewMenuClose();
    onCreateFolder();
  };

  const handleFileUpload = () => {
    handleNewMenuClose();
    onFileUpload();
  };

  return (
    <Box
      sx={{
        width: 220,
        height: 'calc(100vh - 64px)',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        color: 'white',
        display: 'flex',
        flexDirection: 'column',
        position: 'fixed',
        top: 64,
        left: 0,
        overflowY: 'auto',
        overflowX: 'hidden',
      }}
    >
      {/* Yeni Button */}
      <Box sx={{ px: 2, py: 1.5, flexShrink: 0 }}>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleNewMenuOpen}
          sx={{
            width: 'auto',
            px: 2,
            height: 36,
            borderRadius: 1.5,
            backgroundColor: 'white',
            color: '#667eea',
            textTransform: 'none',
            fontSize: '0.85rem',
            fontWeight: 600,
            boxShadow: '0 6px 16px rgba(0,0,0,0.15)',
            '& .MuiSvgIcon-root': { color: '#667eea' },
            '&:hover': {
              backgroundColor: 'rgba(255,255,255,0.95)',
              boxShadow: '0 8px 20px rgba(0,0,0,0.2)',
            },
          }}
        >
          Yeni
        </Button>

        {/* New Menu */}
        <Menu
          anchorEl={newMenuAnchor}
          open={Boolean(newMenuAnchor)}
          onClose={handleNewMenuClose}
          PaperProps={{
            sx: {
              mt: 1,
              minWidth: 200,
              borderRadius: 2,
              boxShadow: '0 8px 32px rgba(0,0,0,0.12)',
            },
          }}
        >
          <MenuItem onClick={handleCreateFolder} sx={{ py: 1.5, px: 2 }}>
            <ListItemIcon sx={{ minWidth: 36 }}>
              <CreateNewFolderIcon sx={{ fontSize: 20, color: '#4285f4' }} />
            </ListItemIcon>
            <ListItemText
              primary={t('folder.new_folder_menu')}
              primaryTypographyProps={{ fontSize: '0.9rem' }}
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
              Alt+C, F
            </Typography>
          </MenuItem>

          <Divider sx={{ my: 0.5 }} />

          <MenuItem onClick={handleFileUpload} sx={{ py: 1.5, px: 2 }}>
            <ListItemIcon sx={{ minWidth: 36 }}>
              <CloudUploadIcon sx={{ fontSize: 20, color: '#34a853' }} />
            </ListItemIcon>
            <ListItemText
              primary={t('folder.upload_menu')}
              primaryTypographyProps={{ fontSize: '0.9rem' }}
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
              Alt+C, U
            </Typography>
          </MenuItem>

          <MenuItem onClick={handleFileUpload} sx={{ py: 1.5, px: 2 }}>
            <ListItemIcon sx={{ minWidth: 36 }}>
              <FolderIcon sx={{ fontSize: 20, color: '#ea4335' }} />
            </ListItemIcon>
            <ListItemText
              primary={t('folder.upload')}
              primaryTypographyProps={{ fontSize: '0.9rem' }}
            />
            <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
              Alt+C, I
            </Typography>
          </MenuItem>
        </Menu>
      </Box>

      <Divider sx={{ mx: 2, flexShrink: 0 }} />

      <List sx={{ px: 2, py: 1.5, flex: 1, overflowY: 'auto', minHeight: 0 }}>
        {menuItems.map((item, index) => (
          <ListItem key={item.id} disablePadding sx={{ mb: 0.5 }}>
            <MotionBox
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              sx={{ width: '100%' }}
            >
              <ListItemButton
                selected={selected === item.id}
                onClick={() => {
                  setSelected(item.id);
                  if (onMenuChange) {
                    onMenuChange(item.id);
                  }
                }}
                sx={{
                  borderRadius: 1.5,
                  py: 1,
                  '&.Mui-selected': {
                    bgcolor: 'rgba(255, 255, 255, 0.16)',
                    color: 'white',
                    '&:hover': {
                      bgcolor: 'rgba(255, 255, 255, 0.22)',
                    },
                    '& .MuiListItemIcon-root': {
                      color: 'white',
                    },
                  },
                  '&:hover': {
                    bgcolor: 'rgba(255, 255, 255, 0.10)',
                  },
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 32,
                    color: 'white',
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                  primary={item.label}
                  primaryTypographyProps={{
                    fontSize: '0.85rem',
                    fontWeight: selected === item.id ? 600 : 400,
                    color: 'white',
                  }}
                />
              </ListItemButton>
            </MotionBox>
          </ListItem>
        ))}
      </List>

      {/* Storage Section */}
      <Box sx={{ px: 2, py: 1.5, mt: 'auto', mb: 1, flexShrink: 0 }}>
        <Divider sx={{ mb: 1.5, borderColor: 'rgba(255,255,255,0.24)' }} />
        <Box sx={{ bgcolor: 'white', borderRadius: 2, p: 1.25 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.75 }}>
            <CloudQueueIcon sx={{ fontSize: 18, color: 'text.secondary', mr: 1 }} />
            <Typography variant="body2" color="text.secondary" fontWeight={600} fontSize="0.8rem">
              {t('sidebar.storage') || 'Storage'}
            </Typography>
          </Box>
          <LinearProgress
            variant="determinate"
            value={storageInfo.percent}
            sx={{
              height: 6,
              borderRadius: 4,
              bgcolor: 'grey.200',
              '& .MuiLinearProgress-bar': {
                borderRadius: 4,
                backgroundColor: '#FBBF24',
              },
            }}
          />
          <Typography
            variant="caption"
            color="text.secondary"
            sx={{ mt: 0.5, display: 'block', fontSize: '0.7rem' }}
          >
            15 GB {t('sidebar.of') || 'of'} {storageInfo.usage} {t('sidebar.used') || 'used'}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default Sidebar;
