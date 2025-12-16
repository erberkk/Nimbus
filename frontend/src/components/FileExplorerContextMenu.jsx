import React from 'react';
import { useTranslation } from 'react-i18next';
import { Menu, MenuItem, ListItemIcon, ListItemText, Divider } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import FolderIcon from '@mui/icons-material/Folder';
import RestoreFromTrashIcon from '@mui/icons-material/RestoreFromTrash';
import StarBorderIcon from '@mui/icons-material/StarBorder';
import StarIcon from '@mui/icons-material/Star';
import DriveFileRenameOutlineIcon from '@mui/icons-material/DriveFileRenameOutline';

const FileExplorerContextMenu = ({
  menuAnchor,
  selectedItem,
  onClose,
  onDownloadFile,
  onDeleteFile,
  onDeleteFolder,
  onCreateFolder,
  onFileUpload,
  onToggleStar,
  onRestore,
  onPermanentDelete,
  onMove,
}) => {
  const { t } = useTranslation();
  if (!selectedItem) {
    return (
      <Menu
        anchorEl={menuAnchor}
        open={Boolean(menuAnchor)}
        onClose={onClose}
        PaperProps={{
          sx: {
            background: 'rgba(255, 255, 255, 0.95)',
            backdropFilter: 'blur(20px)',
            border: '1px solid rgba(102, 126, 234, 0.2)',
            boxShadow: '0 12px 48px 0 rgba(31, 38, 135, 0.25)',
            borderRadius: 3,
            mt: 0.5,
            minWidth: 220,
            overflow: 'hidden',
          },
        }}
        transformOrigin={{ horizontal: 'left', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'left', vertical: 'bottom' }}
      >
        <MenuItem
          onClick={() => {
            onCreateFolder();
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon
            sx={{
              minWidth: 40,
              color: '#667eea',
            }}
          >
            <CreateNewFolderIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText
            primary={t('folder.new_folder')}
            primaryTypographyProps={{
              fontWeight: 500,
              fontSize: '0.95rem',
            }}
          />
        </MenuItem>

        <MenuItem
          onClick={() => {
            onFileUpload('file');
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon
            sx={{
              minWidth: 40,
              color: '#4facfe',
            }}
          >
            <CloudUploadIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText
            primary={t('folder.upload_file')}
            primaryTypographyProps={{
              fontWeight: 500,
              fontSize: '0.95rem',
            }}
          />
        </MenuItem>

        <MenuItem
          onClick={() => {
            onFileUpload('folder');
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon
            sx={{
              minWidth: 40,
              color: '#764ba2',
            }}
          >
            <FolderIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText
            primary={t('folder.upload_folder')}
            primaryTypographyProps={{
              fontWeight: 500,
              fontSize: '0.95rem',
            }}
          />
        </MenuItem>
      </Menu>
    );
  }

  const isTrashItem = selectedItem.deleted_at != null;

  return (
    <Menu
      anchorEl={menuAnchor}
      open={Boolean(menuAnchor)}
      onClose={onClose}
      PaperProps={{
        sx: {
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(102, 126, 234, 0.2)',
          boxShadow: '0 12px 48px 0 rgba(31, 38, 135, 0.25)',
          borderRadius: 3,
          mt: 0.5,
          minWidth: 220,
          overflow: 'hidden',
        },
      }}
      transformOrigin={{ horizontal: 'left', vertical: 'top' }}
      anchorOrigin={{ horizontal: 'left', vertical: 'bottom' }}
    >
      {isTrashItem && (
        <MenuItem
          onClick={() => {
            onRestore(selectedItem);
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon>
            <RestoreFromTrashIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>{t('restore') || 'Restore'}</ListItemText>
        </MenuItem>
      )}

      {!isTrashItem && (
        <MenuItem
          onClick={() => {
            onToggleStar(selectedItem);
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon>
            {selectedItem.is_starred ? <StarIcon fontSize="small" /> : <StarBorderIcon fontSize="small" />}
          </ListItemIcon>
          <ListItemText>{selectedItem.is_starred ? t('unstar') || 'Unstar' : t('star') || 'Star'}</ListItemText>
        </MenuItem>
      )}

      {selectedItem?.type === 'file' && !isTrashItem && (
        <MenuItem
          key="download"
          onClick={() => {
            onDownloadFile(selectedItem);
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon>
            <DownloadIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>{t('download')}</ListItemText>
        </MenuItem>
      )}

      {!isTrashItem && (
        <MenuItem
          onClick={() => {
            onMove(selectedItem, selectedItem?.type);
            onClose();
          }}
          sx={{
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
            py: 1.5,
            px: 2,
            '&:hover': {
              backgroundColor: 'rgba(102, 126, 234, 0.12)',
              transform: 'translateX(4px)',
            },
          }}
        >
          <ListItemIcon>
            <DriveFileRenameOutlineIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>{t('move') || 'Taşı'}</ListItemText>
        </MenuItem>
      )}

      <Divider key="divider" sx={{ my: 0.5 }} />
      
      {/* Permanent Delete */}
      <MenuItem
        key="delete"
        onClick={() => {
          if (isTrashItem) {
            onPermanentDelete(selectedItem);
          } else {
            selectedItem.type === 'file' ? onDeleteFile(selectedItem) : onDeleteFolder(selectedItem);
          }
          onClose();
        }}
        sx={{
          transition: 'all 0.2s ease',
          color: 'error.main',
          '&:hover': {
            backgroundColor: 'rgba(244, 67, 54, 0.1)',
            paddingLeft: '20px',
          },
        }}
      >
        <ListItemIcon sx={{ color: 'error.main' }}>
          <DeleteIcon fontSize="small" />
        </ListItemIcon>
        <ListItemText>{isTrashItem ? (t('delete_permanently') || 'Delete Permanently') : t('delete')}</ListItemText>
      </MenuItem>
    </Menu>
  );
};

export default FileExplorerContextMenu;
