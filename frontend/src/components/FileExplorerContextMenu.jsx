import React from 'react';
import {
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';

const FileExplorerContextMenu = ({
  menuAnchor,
  selectedItem,
  onClose,
  onDownloadFile,
  onDeleteFile,
  onDeleteFolder,
}) => {
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
