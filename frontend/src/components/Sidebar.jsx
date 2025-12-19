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
  IconButton,
  Tooltip,
  Collapse,
  CircularProgress,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useState, useEffect } from 'react';
import { folderApi, fileApi } from '../services/api';
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
import SmartToyIcon from '@mui/icons-material/SmartToy';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import DeleteIcon from '@mui/icons-material/Delete';
import { formatRelativeTime } from '../utils/fileTypeUtils';

const MotionBox = motion.create(Box);

const Sidebar = ({ onCreateFolder, onFileUpload, onMenuChange, selectedMenu, onConversationClick }) => {
  const { t } = useTranslation();
  const [selected, setSelected] = useState(selectedMenu || 'home');
  const [newMenuAnchor, setNewMenuAnchor] = useState(null);
  const [conversationsOpen, setConversationsOpen] = useState(true);
  const [conversations, setConversations] = useState([]);
  const [conversationsLoading, setConversationsLoading] = useState(false);

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
    loadConversations();
  }, []);

  const loadConversations = async () => {
    try {
      setConversationsLoading(true);
      const response = await fileApi.getUserConversations();
      setConversations(response.conversations || []);
    } catch (error) {
      console.error('Failed to load conversations:', error);
      setConversations([]);
    } finally {
      setConversationsLoading(false);
    }
  };

  const handleConversationClick = async conversation => {
    if (onConversationClick) {
      // Create file object from conversation data
      const file = {
        id: conversation.file_id,
        filename: conversation.file.filename,
        content_type: conversation.file.content_type,
        size: conversation.file.size,
        processing_status: 'completed', // Assume completed if conversation exists
      };
      onConversationClick(file);
      // Refresh conversations after opening
      await loadConversations();
    }
  };

  const handleDeleteConversation = async (e, conversation) => {
    e.stopPropagation();
    if (window.confirm(t('ai.delete_conversation_confirm') || 'Bu sohbet geçmişini silmek istediğinize emin misiniz?')) {
      try {
        await fileApi.clearConversationHistory(conversation.file_id);
        window.toast?.success(t('ai.conversation_deleted') || 'Sohbet geçmişi silindi');
        await loadConversations();
      } catch (error) {
        console.error('Failed to delete conversation:', error);
        window.toast?.error(t('ai.delete_error') || 'Sohbet geçmişi silinemedi');
      }
    }
  };

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
      <Box sx={{ px: 2, py: 1.5, pt: 2.5, flexShrink: 0, display: 'flex', justifyContent: 'center' }}>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleNewMenuOpen}
          sx={{
            width: 'fit-content',
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
        {menuItems.map(item => (
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

      {/* Conversation History Section */}
      <Box sx={{ px: 2, py: 1, flexShrink: 0 }}>
        <Divider sx={{ mb: 1, borderColor: 'rgba(255,255,255,0.24)' }} />
        <ListItemButton
          onClick={() => setConversationsOpen(!conversationsOpen)}
          sx={{
            borderRadius: 1.5,
            py: 0.75,
            px: 1,
            mb: 0.5,
            '&:hover': {
              bgcolor: 'rgba(255, 255, 255, 0.10)',
            },
          }}
        >
          <ListItemIcon sx={{ minWidth: 32, color: 'white' }}>
            <SmartToyIcon sx={{ fontSize: 20 }} />
          </ListItemIcon>
          <ListItemText
            primary={t('sidebar.conversations') || 'Sohbetler'}
            primaryTypographyProps={{
              fontSize: '0.85rem',
              fontWeight: 500,
              color: 'white',
            }}
          />
          {conversationsOpen ? (
            <ExpandLessIcon sx={{ color: 'white', fontSize: 20 }} />
          ) : (
            <ExpandMoreIcon sx={{ color: 'white', fontSize: 20 }} />
          )}
        </ListItemButton>

        <Collapse in={conversationsOpen} timeout="auto" unmountOnExit>
          <Box
            sx={{
              maxHeight: 200,
              overflowY: 'auto',
              overflowX: 'hidden',
              '&::-webkit-scrollbar': {
                width: '4px',
              },
              '&::-webkit-scrollbar-track': {
                background: 'rgba(255,255,255,0.1)',
                borderRadius: '2px',
              },
              '&::-webkit-scrollbar-thumb': {
                background: 'rgba(255,255,255,0.3)',
                borderRadius: '2px',
                '&:hover': {
                  background: 'rgba(255,255,255,0.5)',
                },
              },
            }}
          >
            {conversationsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                <CircularProgress size={20} sx={{ color: 'white' }} />
              </Box>
            ) : conversations.length === 0 ? (
              <Typography
                variant="caption"
                sx={{
                  color: 'rgba(255,255,255,0.7)',
                  fontSize: '0.75rem',
                  px: 1,
                  py: 1,
                  display: 'block',
                }}
              >
                {t('sidebar.no_conversations') || 'Henüz sohbet yok'}
              </Typography>
            ) : (
              conversations.map(conv => (
                <ListItem
                  key={conv.id}
                  disablePadding
                  sx={{ mb: 0.5 }}
                >
                  <ListItemButton
                    onClick={() => handleConversationClick(conv)}
                    sx={{
                      borderRadius: 1,
                      py: 0.75,
                      px: 1,
                      '&:hover': {
                        bgcolor: 'rgba(255, 255, 255, 0.10)',
                      },
                    }}
                  >
                    <ListItemIcon sx={{ minWidth: 28, color: 'white' }}>
                      <SmartToyIcon sx={{ fontSize: 16 }} />
                    </ListItemIcon>
                    <ListItemText
                      primary={
                        <Typography
                          variant="caption"
                          sx={{
                            color: 'white',
                            fontSize: '0.75rem',
                            fontWeight: 500,
                            display: 'block',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }}
                        >
                          {conv.file.filename}
                        </Typography>
                      }
                      secondary={
                        <Typography
                          variant="caption"
                          sx={{
                            color: 'rgba(255,255,255,0.7)',
                            fontSize: '0.7rem',
                            display: 'block',
                          }}
                        >
                          {conv.messages.length > 0
                            ? formatRelativeTime(conv.updated_at)
                            : t('sidebar.no_messages') || 'Mesaj yok'}
                        </Typography>
                      }
                    />
                    <IconButton
                      size="small"
                      onClick={e => handleDeleteConversation(e, conv)}
                      sx={{
                        color: 'rgba(255,255,255,0.7)',
                        '&:hover': {
                          color: 'white',
                          bgcolor: 'rgba(255,255,255,0.1)',
                        },
                      }}
                    >
                      <DeleteIcon sx={{ fontSize: 16 }} />
                    </IconButton>
                  </ListItemButton>
                </ListItem>
              ))
            )}
          </Box>
        </Collapse>
      </Box>

      {/* Storage Section */}
      <Box sx={{ px: 2, py: 1.5, mt: 'auto', mb: 1, flexShrink: 0 }}>
        <Divider sx={{ mb: 1.5, borderColor: 'rgba(255,255,255,0.24)' }} />
        <Box sx={{ 
          background: 'rgba(255, 255, 255, 0.7)',
          backdropFilter: 'blur(10px)',
          border: '1px solid rgba(255, 255, 255, 0.5)',
          borderRadius: 2, 
          p: 1.25,
          transition: 'all 0.3s ease',
          '&:hover': {
            background: 'rgba(255, 255, 255, 0.85)',
            border: '1px solid rgba(255, 255, 255, 0.6)',
            boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.1)',
          }
        }}>
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
              bgcolor: 'rgba(0,0,0,0.1)',
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
