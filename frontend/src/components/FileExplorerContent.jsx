import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  Button,
  CircularProgress,
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Checkbox,
  IconButton,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import FolderIcon from '@mui/icons-material/Folder';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import FolderCard from './FolderCard';
import FileCard from './FileCard';
import NimbusChatPanel from './NimbusChatPanel';
import FileInfoPanel from './FileInfoPanel';
import { isPreviewable, formatFileSize, formatDate } from '../utils/fileUtils';

const MotionBox = motion.create(Box);

const FileExplorerContent = ({
  fileExplorer,
  uiState,
  navigation,
  onCreateFolder,
  onFileUpload,
  onFolderOpen,
  onFileDownload,
  onFolderDelete,
  onFileDelete,
  onShare,
  onMenuOpen,
  onUploadSuccess,
  onPreview,
  onEdit,
}) => {
  const { t } = useTranslation();
  // Chat panel state
  const [chatPanelOpen, setChatPanelOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState(null);

  // File info panel state
  const [infoPanelOpen, setInfoPanelOpen] = useState(false);
  const [selectedFileForInfo, setSelectedFileForInfo] = useState(null);

  // Nimbus'a Sor fonksiyonu
  const handleAskNimbus = file => {
    setSelectedFile(file);
    setChatPanelOpen(true);
  };

  // Chat paneli kapatma
  const handleCloseChatPanel = () => {
    setChatPanelOpen(false);
    setSelectedFile(null);
  };

  // File info panel açma
  const handleFileInfo = file => {
    setSelectedFileForInfo(file);
    setInfoPanelOpen(true);
  };

  // File info panel kapatma
  const handleCloseInfoPanel = () => {
    setInfoPanelOpen(false);
    setSelectedFileForInfo(null);
  };

  // Direkt dosya yükleme fonksiyonu
  const handleDirectFileUpload = async files => {
    try {
      window.toast?.info(t('folder.uploading', { count: files.length }));

      for (const file of files) {
        await uploadSingleFile(file);
      }

      window.toast?.success(t('folder.upload_success', { count: files.length }));

      // Yükleme tamamlandıktan sonra içeriği yenile
      if (onUploadSuccess) {
        onUploadSuccess();
      } else {
        window.location.reload(); // Fallback
      }
    } catch (error) {
      window.toast?.error(t('folder.upload_error', { error: error.message }));
    }
  };

  // Tek dosya yükleme fonksiyonu
  const uploadSingleFile = async file => {
    const { fileApi } = await import('../services/api');

    // Presigned URL al
    const presignedResponse = await fileApi.getUploadPresignedURL(file.name, file.type);
    const { presigned_url, minio_path } = presignedResponse;

    // Dosyayı MinIO'ya yükle
    const uploadResponse = await fetch(presigned_url, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type || 'application/octet-stream',
      },
    });

    if (!uploadResponse.ok) {
      throw new Error(`Upload failed: ${uploadResponse.statusText}`);
    }

    // MongoDB'ye dosya kaydı oluştur
    const currentFolder = navigation.getCurrentNavState().currentFolder;
    const fileData = {
      filename: file.name,
      size: file.size,
      content_type: file.type,
      minio_path: minio_path,
    };

    if (currentFolder && currentFolder.id) {
      fileData.folder_id = currentFolder.id;
    }

    const result = await fileApi.createFile(fileData);
  };
  if (fileExplorer.loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <CircularProgress size={60} />
      </Box>
    );
  }

  return (
    <Box
      sx={{ position: 'relative', height: '100%', width: '100%' }}
      onContextMenu={e => {
        e.preventDefault();
        onMenuOpen(e, null);
      }}
      onDragEnter={e => {
        e.preventDefault();
        e.stopPropagation();
        const overlay = document.getElementById('drag-overlay');
        if (overlay) overlay.style.display = 'flex';
      }}
      onDragLeave={e => {
        e.preventDefault();
        e.stopPropagation();
        if (!e.currentTarget.contains(e.relatedTarget)) {
          const overlay = document.getElementById('drag-overlay');
          if (overlay) overlay.style.display = 'none';
        }
      }}
      onDragOver={e => {
        e.preventDefault();
        e.stopPropagation();
      }}
      onDrop={async e => {
        e.preventDefault();
        e.stopPropagation();
        const overlay = document.getElementById('drag-overlay');
        if (overlay) overlay.style.display = 'none';

        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
          const files = Array.from(e.dataTransfer.files);
          await handleDirectFileUpload(files);
        }
      }}
    >
      {/* Drag & Drop Overlay */}
      <Box
        sx={{
          position: 'absolute',
          bottom: 30,
          left: '50%',
          transform: 'translateX(-50%)',
          display: 'none',
          zIndex: 1000,
          minWidth: 480,
          maxWidth: 600,
          '@keyframes pulse': {
            '0%': {
              transform: 'translateX(-50%) scale(1)',
            },
            '50%': {
              transform: 'translateX(-50%) scale(1.02)',
            },
            '100%': {
              transform: 'translateX(-50%) scale(1)',
            },
          },
          '@keyframes float': {
            '0%, 100%': {
              transform: 'translateX(-50%) translateY(0px)',
            },
            '50%': {
              transform: 'translateX(-50%) translateY(-3px)',
            },
          },
          '@keyframes shimmer': {
            '0%': {
              transform: 'translateX(-100%)',
            },
            '100%': {
              transform: 'translateX(100%)',
            },
          },
          animation: 'pulse 3s ease-in-out infinite, float 2s ease-in-out infinite',
        }}
        id="drag-overlay"
      >
        <Box
          sx={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            color: 'white',
            borderRadius: 3,
            px: 4,
            py: 2,
            textAlign: 'center',
            boxShadow: '0 20px 60px rgba(102, 126, 234, 0.4), 0 10px 30px rgba(0,0,0,0.2)',
            backdropFilter: 'blur(20px)',
            position: 'relative',
            overflow: 'hidden',
            '&::after': {
              content: '""',
              position: 'absolute',
              top: 0,
              left: '-100%',
              width: '100%',
              height: '100%',
              background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.3), transparent)',
              animation: 'shimmer 3s infinite',
            },
          }}
        >
          <Box
            sx={{
              position: 'relative',
              zIndex: 1,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 2,
            }}
          >
            <Box
              sx={{
                width: 48,
                height: 48,
                borderRadius: '50%',
                bgcolor: 'rgba(255,255,255,0.2)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              <CloudUploadIcon sx={{ fontSize: 24, color: 'white' }} />
            </Box>
            <Box sx={{ textAlign: 'left' }}>
              <Typography
                variant="h6"
                sx={{
                  fontWeight: 600,
                  mb: 0.5,
                  fontSize: '1rem',
                }}
              >
                {t('folder.drag_drop')}
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  opacity: 0.9,
                  fontSize: '0.8rem',
                }}
              >
                {t('folder.drag_drop_hint')}
              </Typography>
            </Box>
          </Box>
        </Box>
      </Box>

      {/* Empty State - Always show when no content */}
      {fileExplorer.folders.length === 0 && fileExplorer.files.length === 0 && (
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: 400,
            mt: 4,
          }}
        >
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
              {navigation.getCurrentNavState().currentFolder
                ? t('folder.empty')
                : t('folder.no_items')}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              {t('folder.empty_hint')}
            </Typography>
          </MotionBox>
        </Box>
      )}

      {/* Grid View - Only show when there's content */}
      {fileExplorer.folders.length > 0 || fileExplorer.files.length > 0 ? (
        uiState.viewMode === 'grid' ? (
          <Box sx={{ flex: 1, overflow: 'hidden' }}>
            {/* Folders Section */}
            {fileExplorer.folders.length > 0 && (
              <Box sx={{ mb: 3 }}>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                  sx={{ mb: 1.5, fontWeight: 500 }}
                >
                  {t('folder.title')}
                </Typography>
                <Grid container spacing={2}>
                  <AnimatePresence>
                    {fileExplorer.folders.map((folder, index) => (
                      <Grid item xs={12} sm={6} md={4} lg={3} key={folder.id}>
                        <MotionBox
                          initial={{ opacity: 0, scale: 0.9 }}
                          animate={{ opacity: 1, scale: 1 }}
                          exit={{ opacity: 0, scale: 0.9 }}
                          transition={{ duration: 0.3, delay: index * 0.05 }}
                        >
                          <FolderCard
                            folder={folder}
                            onOpen={onFolderOpen}
                            onDelete={onFolderDelete}
                            onShare={folder => onShare(folder, 'folder')}
                          />
                        </MotionBox>
                      </Grid>
                    ))}
                  </AnimatePresence>
                </Grid>
              </Box>
            )}

            {/* Files Section */}
            {fileExplorer.files.length > 0 && (
              <Box>
                <Typography
                  variant="subtitle2"
                  color="text.secondary"
                  sx={{ mb: 1.5, fontWeight: 500 }}
                >
                  {t('folder.files_title')}
                </Typography>
                <Grid container spacing={2}>
                  <AnimatePresence>
                    {fileExplorer.files.map((file, index) => (
                      <Grid item xs={12} sm={6} md={4} lg={3} key={file.id}>
                        <MotionBox
                          initial={{ opacity: 0, scale: 0.9 }}
                          animate={{ opacity: 1, scale: 1 }}
                          exit={{ opacity: 0, scale: 0.9 }}
                          transition={{ duration: 0.3, delay: index * 0.05 }}
                        >
                          <FileCard
                            file={file}
                            onDownload={onFileDownload}
                            onDelete={onFileDelete}
                            onShare={file => onShare(file, 'file')}
                            onAskNimbus={handleAskNimbus}
                            onPreview={onPreview}
                            onEdit={onEdit}
                            onInfo={handleFileInfo}
                          />
                        </MotionBox>
                      </Grid>
                    ))}
                  </AnimatePresence>
                </Grid>
              </Box>
            )}
          </Box>
        ) : (
          /* List View - Only show when there's content */
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
                    <TableCell>Boyut</TableCell>
                    <TableCell>Değiştirilme</TableCell>
                    <TableCell padding="checkbox"></TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {/* Folders */}
                  {fileExplorer.folders.map(folder => (
                    <TableRow
                      key={folder.id}
                      hover
                      sx={{ cursor: 'pointer' }}
                      onClick={() => onFolderOpen(folder)}
                      onContextMenu={e => onMenuOpen(e, { ...folder, type: 'folder' })}
                    >
                      <TableCell padding="checkbox">
                        <Checkbox />
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <FolderIcon sx={{ color: folder.color || '#1976d2' }} />
                          <Typography variant="body2">{folder.name}</Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {folder.owner?.name || 'Bilinmiyor'}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {folder.item_count
                            ? t('folder.items', { count: folder.item_count })
                            : t('folder.items_zero')}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {formatDate(folder.updated_at)}
                        </Typography>
                      </TableCell>
                      <TableCell padding="checkbox">
                        <IconButton
                          size="small"
                          onClick={e => {
                            e.stopPropagation();
                            onMenuOpen(e, { ...folder, type: 'folder' });
                          }}
                        >
                          <MoreVertIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}

                  {/* Files */}
                  {fileExplorer.files.map(file => {
                    const fileIsPreviewable = isPreviewable(file?.content_type, file?.filename);

                    return (
                      <TableRow
                        key={file.id}
                        hover
                        onClick={() => {
                          if (fileIsPreviewable && onPreview) {
                            onPreview(file);
                          }
                        }}
                        sx={{
                          cursor: fileIsPreviewable ? 'pointer' : 'default',
                        }}
                      >
                        <TableCell padding="checkbox" onClick={e => e.stopPropagation()}>
                          <Checkbox />
                        </TableCell>
                        <TableCell>
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <InsertDriveFileIcon />
                            <Typography variant="body2">{file.filename}</Typography>
                          </Box>
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary">
                            {file.owner?.name || 'Bilinmiyor'}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary">
                            {formatFileSize(file.size)}
                          </Typography>
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2" color="text.secondary">
                            {formatDate(file.updated_at)}
                          </Typography>
                        </TableCell>
                        <TableCell padding="checkbox" onClick={e => e.stopPropagation()}>
                          <IconButton
                            size="small"
                            onClick={e => {
                              e.stopPropagation();
                              onMenuOpen(e, { ...file, type: 'file' });
                            }}
                          >
                            <MoreVertIcon />
                          </IconButton>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        )
      ) : null}

      {/* Nimbus Chat Panel */}
      <NimbusChatPanel isOpen={chatPanelOpen} onClose={handleCloseChatPanel} file={selectedFile} />

      {/* File Info Panel */}
      <FileInfoPanel
        isOpen={infoPanelOpen}
        onClose={handleCloseInfoPanel}
        file={selectedFileForInfo}
      />
    </Box>
  );
};

export default FileExplorerContent;
