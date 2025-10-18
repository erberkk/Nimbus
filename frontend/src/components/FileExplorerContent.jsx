import React from 'react';
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
  formatFileSize,
  formatDate,
}) => {
  if (fileExplorer.loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 300 }}>
        <CircularProgress size={60} />
      </Box>
    );
  }

  return (
    <>
      {/* Empty State - Always show when no content */}
      {fileExplorer.folders.length === 0 && fileExplorer.files.length === 0 && (
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
              {navigation.getCurrentNavState().currentFolder ? 'Bu klasör boş' : 'Henüz klasör veya dosya yok'}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              {navigation.getCurrentNavState().currentFolder
                ? 'Dosya yükleyerek başlayın'
                : 'Yeni klasör oluşturun veya dosya yükleyin'}
            </Typography>

            {/* Show buttons in both root and folder views */}
            <Box sx={{ display: 'flex', gap: 2, justifyContent: 'center' }}>
              <Button
                variant="contained"
                startIcon={<CreateNewFolderIcon />}
                onClick={onCreateFolder}
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
                onClick={onFileUpload}
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
                  Klasörler
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
                  Dosyalar
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
                      onContextMenu={(e) => onMenuOpen(e, { ...folder, type: 'folder' })}
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
                          {folder.item_count || 0} öğe
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
                          onClick={(e) => {
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
                  {fileExplorer.files.map(file => (
                    <TableRow key={file.id} hover>
                      <TableCell padding="checkbox">
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
                      <TableCell padding="checkbox">
                        <IconButton
                          size="small"
                          onClick={(e) => onMenuOpen(e, { ...file, type: 'file' })}
                        >
                          <MoreVertIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        )
      ) : null}
    </>
  );
};

export default FileExplorerContent;
