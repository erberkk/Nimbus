import React from 'react';
import {
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Divider,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import FolderIcon from '@mui/icons-material/Folder';

const FileExplorerContextMenu = ({
  menuAnchor,
  selectedItem,
  onClose,
  onDownloadFile,
  onDeleteFile,
  onDeleteFolder,
  onCreateFolder,
  onFileUpload,
}) => {
  // Eğer seçili bir öğe yoksa (boş yere sağ tıklandıysa)
  if (!selectedItem) {
    return (
      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={onClose}>
        <MenuItem
          onClick={() => {
            onCreateFolder();
            onClose();
          }}
        >
          <ListItemIcon>
            <CreateNewFolderIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Yeni Klasör</ListItemText>
        </MenuItem>

        <MenuItem
          onClick={() => {
            onFileUpload();
            onClose();
          }}
        >
          <ListItemIcon>
            <CloudUploadIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Dosya Yükle</ListItemText>
        </MenuItem>

        <MenuItem
          onClick={() => {
            onFileUpload(); // Aynı fonksiyonu kullanabiliriz, dialog'da klasör yükleme seçeneği ekleyeceğiz
            onClose();
          }}
        >
          <ListItemIcon>
            <FolderIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Klasör Yükle</ListItemText>
        </MenuItem>
      </Menu>
    );
  }

  return (
    <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={onClose}>
      {selectedItem?.type === 'file' ? (
        [
          <MenuItem
            key="download"
            onClick={() => {
              onDownloadFile(selectedItem);
              onClose();
            }}
          >
            <ListItemIcon>
              <DownloadIcon fontSize="small" />
            </ListItemIcon>
            <ListItemText>İndir</ListItemText>
          </MenuItem>,
          <Divider key="divider" />,
          <MenuItem
            key="delete"
            onClick={() => {
              onDeleteFile(selectedItem);
              onClose();
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
            onDeleteFolder(selectedItem);
            onClose();
          }}
        >
          <ListItemIcon>
            <DeleteIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>Sil</ListItemText>
        </MenuItem>
      )}
    </Menu>
  );
};

export default FileExplorerContextMenu;
