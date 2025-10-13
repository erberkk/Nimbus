import { useState, useEffect, useCallback, forwardRef, useImperativeHandle } from 'react';
import {
  Box,
  Typography,
  Button,
  Breadcrumbs,
  Link,
  CircularProgress,
  Alert,
  IconButton,
  Menu,
  MenuItem,
  Toolbar,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Checkbox,
  Grid,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import HomeIcon from '@mui/icons-material/Home';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ViewListIcon from '@mui/icons-material/ViewList';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import FolderIcon from '@mui/icons-material/Folder';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMove';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import FolderCard from './FolderCard';
import FileCard from './FileCard';
import CreateFolderDialog from './CreateFolderDialog';
import FileUpload from './FileUpload';
import ShareDialog from './ShareDialog';
import { folderApi, fileApi, shareApi } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

const MotionBox = motion.create(Box);

const FileExplorerNew = forwardRef(({ selectedMenu = 'home' }, ref) => {
  useImperativeHandle(ref, () => ({
    handleCreateFolder
  }));
  const { t } = useTranslation();
  const { user } = useAuth();
  const [currentFolder, setCurrentFolder] = useState(null);
  const [folderPath, setFolderPath] = useState([]); // Breadcrumb path için
  const [folders, setFolders] = useState([]);
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [fileUploadOpen, setFileUploadOpen] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [viewMode, setViewMode] = useState('grid'); // 'grid' or 'list'
  const [menuAnchor, setMenuAnchor] = useState(null);
  const [selectedItem, setSelectedItem] = useState(null);
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const [shareResource, setShareResource] = useState(null);
  const [shareResourceType, setShareResourceType] = useState(null);

  const loadContents = useCallback(async () => {
    try {
      setLoading(true);
      setError('');

      if (selectedMenu === 'shared') {
        // Load shared files and folders
        const response = await shareApi.getSharedWithMe();
        const sharedFiles = response.filter(item => item.resource_type === 'file');
        const sharedFolders = response.filter(item => item.resource_type === 'folder');
        
        setFolders(sharedFolders.map(item => ({
          ...item.resource,
          access_type: item.access_type,
          owner: item.owner,
          isShared: true
        })));
        setFiles(sharedFiles.map(item => ({
          ...item.resource,
          access_type: item.access_type,
          owner: item.owner,
          isShared: true
        })));
      } else {
        // Load normal files and folders
        if (currentFolder) {
          const response = await folderApi.getFolderContents(currentFolder.id);
          setFolders(response.folders || []);
          setFiles(response.files || []);
        } else {
          const response = await folderApi.getRootContents();
          setFolders(response.folders || []);
          setFiles(response.files || []);
        }
      }
    } catch (err) {
      console.error('İçerik yükleme hatası:', err);
      setError(t('files_error'));
      window.toast?.error(t('files_error'));
    } finally {
      setLoading(false);
    }
  }, [currentFolder, selectedMenu, t]);

  useEffect(() => {
    loadContents();
  }, [loadContents, refreshTrigger]);

  const handleFolderOpen = folder => {
    setCurrentFolder(folder);
    setFolderPath(prev => [...prev, folder]); // Path'e yeni klasörü ekle
  };

  const handleBackToRoot = () => {
    setCurrentFolder(null);
    setFolderPath([]); // Path'i temizle
  };

  const handleBreadcrumbClick = (index) => {
    if (index === -1) {
      // Root'a git
      handleBackToRoot();
    } else {
      // Belirli bir seviyeye git
      const targetFolder = folderPath[index];
      setCurrentFolder(targetFolder);
      setFolderPath(prev => prev.slice(0, index + 1)); // O seviyeye kadar olan path'i al
    }
  };

  const handleCreateFolder = async folderData => {
    try {
      // Add current folder info to folder data
      const folderWithParent = {
        ...folderData,
        folder_id: currentFolder?.id || null
      };

      await folderApi.createFolder(folderWithParent);
      window.toast?.success('Klasör başarıyla oluşturuldu');
      setRefreshTrigger(prev => prev + 1);
    } catch (err) {
      console.error('Klasör oluşturma hatası:', err);
      window.toast?.error(err.response?.data?.error || 'Klasör oluşturulamadı');
    }
  };

  const handleDeleteFolder = async folder => {
    if (!window.confirm(`"${folder.name}" klasörünü silmek istediğinizden emin misiniz?`)) {
      return;
    }

    try {
      await folderApi.deleteFolder(folder.id);
      window.toast?.success('Klasör başarıyla silindi');
      setRefreshTrigger(prev => prev + 1);
    } catch (err) {
      console.error('Klasör silme hatası:', err);
      window.toast?.error(err.response?.data?.error || 'Klasör silinemedi');
    }
  };

  const handleDownloadFile = async file => {
    try {
      const response = await fileApi.getDownloadPresignedURL(file.filename);
      window.open(response.presigned_url, '_blank');
    } catch (err) {
      console.error('İndirme hatası:', err);
      window.toast?.error(t('network_error'));
    }
  };

  const handleDeleteFile = async file => {
    if (!window.confirm(t('confirm_delete'))) {
      return;
    }

    try {
      await fileApi.deleteFile(file.id);
      window.toast?.success(t('delete_success'));
      setRefreshTrigger(prev => prev + 1);
    } catch (err) {
      console.error('Silme hatası:', err);
      window.toast?.error(t('delete_error'));
    }
  };

  const handleUploadSuccess = () => {
    setRefreshTrigger(prev => prev + 1);
  };

  const handleMenuOpen = (event, item) => {
    setMenuAnchor(event.currentTarget);
    setSelectedItem(item);
  };

  const handleMenuClose = () => {
    setMenuAnchor(null);
    setSelectedItem(null);
  };

  const handleShare = (resource, type) => {
    setShareResource(resource);
    setShareResourceType(type);
    setShareDialogOpen(true);
  };

  const formatFileSize = bytes => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const formatDate = dateString => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) return 'Bugün';
    if (days === 1) return 'Dün';
    if (days < 7) return `${days} gün önce`;
    return date.toLocaleDateString('tr-TR');
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <CircularProgress size={60} />
      </Box>
    );
  }

  return (
    <Box sx={{ height: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      {/* Top Toolbar */}
      <Toolbar
        sx={{
          px: 0,
          py: 1.5,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: 1,
          borderColor: 'divider',
          mb: 1,
        }}
      >
        {/* Breadcrumb */}
        <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />}>
          {selectedMenu === 'shared' ? [
            <Typography key="shared" variant="body1" color="text.primary" sx={{ fontSize: '0.95rem', display: 'flex', alignItems: 'center', gap: 0.5 }}>
              <HomeIcon fontSize="small" />
              Paylaşılanlarım
            </Typography>
          ] : [
            <Link
              key="home"
              component="button"
              variant="body1"
              onClick={() => handleBreadcrumbClick(-1)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 0.5,
                color: currentFolder ? 'text.secondary' : 'text.primary',
                textDecoration: 'none',
                cursor: 'pointer',
                fontSize: '0.95rem',
                '&:hover': {
                  color: 'primary.main',
                },
              }}
            >
              <HomeIcon fontSize="small" />
              My Drive
            </Link>,
            // Path'deki her klasör için breadcrumb oluştur
            ...folderPath.map((folder, index) => (
              <Link
                key={`folder-${index}`}
                component="button"
                variant="body1"
                onClick={() => handleBreadcrumbClick(index)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5,
                  color: index === folderPath.length - 1 ? 'text.primary' : 'text.secondary',
                  textDecoration: 'none',
                  cursor: 'pointer',
                  fontSize: '0.95rem',
                  '&:hover': {
                    color: 'primary.main',
                  },
                }}
              >
                {folder.name}
              </Link>
            ))
          ]}
        </Breadcrumbs>

        {/* View Mode Toggle */}
        <Box sx={{ display: 'flex', gap: 1 }}>
          <IconButton
            size="small"
            onClick={() => setViewMode('list')}
            sx={{
              bgcolor: viewMode === 'list' ? 'action.selected' : 'transparent',
            }}
          >
            <ViewListIcon />
          </IconButton>
          <IconButton
            size="small"
            onClick={() => setViewMode('grid')}
            sx={{
              bgcolor: viewMode === 'grid' ? 'action.selected' : 'transparent',
            }}
          >
            <ViewModuleIcon />
          </IconButton>
          <IconButton size="small">
            <InfoOutlinedIcon />
          </IconButton>
        </Box>
      </Toolbar>


      {/* Error Alert */}
      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {/* Empty State - Always show when no content */}
      {!loading && folders.length === 0 && files.length === 0 && (
        <Box sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: 400,
          mt: 4
        }}>
          <MotionBox
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            sx={{
              textAlign: 'center',
              py: 4,
            }}
          >
          <CreateNewFolderIcon sx={{ fontSize: 80, color: 'text.disabled', mb: 1.5 }} />
          <Typography variant="h6" color="text.secondary" gutterBottom>
            {currentFolder ? 'Bu klasör boş' : 'Henüz klasör veya dosya yok'}
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            {currentFolder
              ? 'Dosya yükleyerek başlayın'
              : 'Yeni klasör oluşturun veya dosya yükleyin'}
          </Typography>

          {/* Show buttons in both root and folder views */}
          <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center' }}>
            <Button
              variant="contained"
              startIcon={<CreateNewFolderIcon />}
              onClick={() => setCreateFolderOpen(true)}
              sx={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                color: 'white',
              }}
            >
              Yeni Klasör Oluştur
            </Button>
            <Button
              variant="outlined"
              startIcon={<CloudUploadIcon />}
              onClick={() => setFileUploadOpen(true)}
              sx={{
                borderColor: '#667eea',
                color: '#667eea',
                '&:hover': {
                  borderColor: '#764ba2',
                  backgroundColor: 'rgba(102, 126, 234, 0.04)',
                }
              }}
            >
              Dosya Yükle
            </Button>
          </Box>
          </MotionBox>
        </Box>
      )}

      {/* Grid View - Only show when there's content */}
      {!loading && (folders.length > 0 || files.length > 0) && viewMode === 'grid' && (
        <Box sx={{ flex: 1, overflow: 'hidden' }}>
          {/* Folders Section */}
          {folders.length > 0 && (
            <Box sx={{ mb: 3 }}>
              <Typography
                variant="subtitle2"
                color="text.secondary"
                sx={{ mb: 1.5, fontWeight: 500 }}
              >
                Klasörler
              </Typography>
              <Grid container spacing={2}>
                <AnimatePresence>
                  {folders.map((folder, index) => (
                    <Grid item xs={12} sm={6} md={4} lg={3} key={folder.id}>
                      <MotionBox
                        initial={{ opacity: 0, scale: 0.9 }}
                        animate={{ opacity: 1, scale: 1 }}
                        exit={{ opacity: 0, scale: 0.9 }}
                        transition={{ duration: 0.3, delay: index * 0.05 }}
                      >
                        <FolderCard
                          folder={folder}
                          onOpen={handleFolderOpen}
                          onDelete={handleDeleteFolder}
                          onShare={folder => handleShare(folder, 'folder')}
                        />
                      </MotionBox>
                    </Grid>
                  ))}
                </AnimatePresence>
              </Grid>
            </Box>
          )}

          {/* Files Section */}
          {files.length > 0 && (
            <Box>
              <Typography
                variant="subtitle2"
                color="text.secondary"
                sx={{ mb: 1.5, fontWeight: 500 }}
              >
                Dosyalar
              </Typography>
              <Grid container spacing={2}>
                <AnimatePresence>
                  {files.map((file, index) => (
                    <Grid item xs={12} sm={6} md={4} lg={3} key={file.id}>
                      <MotionBox
                        initial={{ opacity: 0, scale: 0.9 }}
                        animate={{ opacity: 1, scale: 1 }}
                        exit={{ opacity: 0, scale: 0.9 }}
                        transition={{ duration: 0.3, delay: index * 0.05 }}
                      >
                        <FileCard
                          file={file}
                          onDownload={handleDownloadFile}
                          onDelete={handleDeleteFile}
                          onShare={file => handleShare(file, 'file')}
                        />
                      </MotionBox>
                    </Grid>
                  ))}
                </AnimatePresence>
              </Grid>
            </Box>
          )}
        </Box>
      )}

      {/* List View - Only show when there's content */}
      {!loading && (folders.length > 0 || files.length > 0) && viewMode === 'list' && (
        <Box sx={{ flex: 1, overflow: 'hidden' }}>
          <TableContainer sx={{ height: '100%' }}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell padding="checkbox">
                  <Checkbox />
                </TableCell>
                <TableCell>Ad</TableCell>
                <TableCell>Sahip</TableCell>
                <TableCell>Son Değiştirilme</TableCell>
                <TableCell>Dosya Boyutu</TableCell>
                <TableCell></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {/* Folders */}
              {folders.map(folder => (
                <TableRow
                  key={folder.id}
                  hover
                  sx={{ cursor: 'pointer' }}
                  onClick={() => handleFolderOpen(folder)}
                >
                  <TableCell padding="checkbox">
                    <Checkbox />
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                      <FolderIcon sx={{ color: 'primary.main' }} />
                      <Typography variant="body2">{folder.name}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      Ben
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      {formatDate(folder.created_at)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      —
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <IconButton
                      size="small"
                      onClick={e => {
                        e.stopPropagation();
                        handleMenuOpen(e, { ...folder, type: 'folder' });
                      }}
                    >
                      <MoreVertIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}

              {/* Files */}
              {files.map(file => (
                <TableRow key={file.id} hover>
                  <TableCell padding="checkbox">
                    <Checkbox />
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                      <InsertDriveFileIcon sx={{ color: 'text.secondary' }} />
                      <Typography variant="body2">{file.filename}</Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      Ben
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      {formatDate(file.created_at)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      {formatFileSize(file.size)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <IconButton
                      size="small"
                      onClick={e => handleMenuOpen(e, { ...file, type: 'file' })}
                    >
                      <MoreVertIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          </TableContainer>
        </Box>
      )}

      {/* Context Menu */}
      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={handleMenuClose}>
        {selectedItem?.type === 'file' ? (
          [
            <MenuItem
              key="download"
              onClick={() => {
                handleDownloadFile(selectedItem);
                handleMenuClose();
              }}
            >
              <ListItemIcon>
                <DownloadIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>İndir</ListItemText>
            </MenuItem>,
            <MenuItem
              key="delete"
              onClick={() => {
                handleDeleteFile(selectedItem);
                handleMenuClose();
              }}
            >
              <ListItemIcon>
                <DeleteIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Sil</ListItemText>
            </MenuItem>,
          ]
        ) : (
          <MenuItem
            onClick={() => {
              handleDeleteFolder(selectedItem);
              handleMenuClose();
            }}
          >
            <ListItemIcon>
              <DeleteIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>Sil</ListItemText>
          </MenuItem>
        )}
      </Menu>


      {/* Create Folder Dialog */}
      <CreateFolderDialog
        open={createFolderOpen}
        onClose={() => setCreateFolderOpen(false)}
        onSubmit={handleCreateFolder}
      />

      {/* File Upload Dialog */}
      <FileUpload
        open={fileUploadOpen}
        onClose={() => setFileUploadOpen(false)}
        onUploadSuccess={() => {
          setFileUploadOpen(false);
          setRefreshTrigger(prev => prev + 1);
        }}
        userId={user?.id}
        currentFolderId={currentFolder?.id}
      />

      {/* Share Dialog */}
      <ShareDialog
        open={shareDialogOpen}
        onClose={() => {
          setShareDialogOpen(false);
          setShareResource(null);
          setShareResourceType(null);
        }}
        resource={shareResource}
        resourceType={shareResourceType}
      />

    </Box>
  );
});

export default FileExplorerNew;
