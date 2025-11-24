import React from 'react';
import { useTranslation } from 'react-i18next';
import { Menu, MenuItem, ListItemIcon, ListItemText } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMove';
import ShareIcon from '@mui/icons-material/Share';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import EditIcon from '@mui/icons-material/Edit';
import InfoIcon from '@mui/icons-material/Info';
import { isAskableFile, isEditable } from '../utils/fileUtils';

/**
 * Unified menu component for file operations
 * Used by both card view (FileCard) and list view
 */
const FileItemMenu = ({
    anchorEl,
    open,
    onClose,
    item, // file or folder object
    itemType, // 'file' or 'folder'
    onInfo,
    onDownload,
    onEdit,
    onAskNimbus,
    onShare,
    onMove,
    onDelete,
}) => {
    const { t } = useTranslation();

    const handleAction = (action, ...args) => {
        if (action) {
            action(...args);
        }
        onClose();
    };

    // For files, determine which options to show
    const fileIsAskable = itemType === 'file' && isAskableFile(item?.content_type, item?.filename);
    const fileIsEditable = itemType === 'file' && isEditable(item?.content_type, item?.filename);

    // Build menu items as arrays instead of fragments
    const fileMenuItems = [];
    const folderMenuItems = [];

    if (itemType === 'file') {
        if (onInfo) {
            fileMenuItems.push(
                <MenuItem key="info" onClick={() => handleAction(onInfo, item)}>
                    <ListItemIcon>
                        <InfoIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('info')}</ListItemText>
                </MenuItem>
            );
        }
        if (onDownload) {
            fileMenuItems.push(
                <MenuItem key="download" onClick={() => handleAction(onDownload, item)}>
                    <ListItemIcon>
                        <DownloadIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('download')}</ListItemText>
                </MenuItem>
            );
        }
        if (fileIsEditable && onEdit) {
            fileMenuItems.push(
                <MenuItem key="edit" onClick={() => handleAction(onEdit, item)} sx={{ color: '#667eea' }}>
                    <ListItemIcon>
                        <EditIcon fontSize="small" sx={{ color: '#667eea' }} />
                    </ListItemIcon>
                    <ListItemText>{t('edit')}</ListItemText>
                </MenuItem>
            );
        }
        if (fileIsAskable && onAskNimbus) {
            fileMenuItems.push(
                <MenuItem
                    key="ask-nimbus"
                    onClick={() => handleAction(onAskNimbus, item)}
                    disabled={item?.processing_status !== 'completed'}
                    sx={{ color: '#667eea' }}
                >
                    <ListItemIcon>
                        <SmartToyIcon fontSize="small" sx={{ color: '#667eea' }} />
                    </ListItemIcon>
                    <ListItemText>
                        {item?.processing_status === 'processing'
                            ? t('ai.ask_nimbus_processing')
                            : item?.processing_status === 'failed'
                                ? t('ai.ask_nimbus_error')
                                : t('ai.ask_nimbus')}
                    </ListItemText>
                </MenuItem>
            );
        }
        if (onShare) {
            fileMenuItems.push(
                <MenuItem key="share" onClick={() => handleAction(onShare, item)}>
                    <ListItemIcon>
                        <ShareIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('share')}</ListItemText>
                </MenuItem>
            );
        }
        if (onMove) {
            fileMenuItems.push(
                <MenuItem key="move" onClick={() => handleAction(onMove, item)}>
                    <ListItemIcon>
                        <DriveFileMoveIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('move')}</ListItemText>
                </MenuItem>
            );
        }
        if (onDelete) {
            fileMenuItems.push(
                <MenuItem key="delete" onClick={() => handleAction(onDelete, item)} sx={{ color: 'error.main' }}>
                    <ListItemIcon>
                        <DeleteIcon fontSize="small" sx={{ color: 'error.main' }} />
                    </ListItemIcon>
                    <ListItemText>{t('delete')}</ListItemText>
                </MenuItem>
            );
        }
    } else {
        // Folder menu items
        if (onShare) {
            folderMenuItems.push(
                <MenuItem key="share" onClick={() => handleAction(onShare, item)}>
                    <ListItemIcon>
                        <ShareIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('share')}</ListItemText>
                </MenuItem>
            );
        }
        if (onMove) {
            folderMenuItems.push(
                <MenuItem key="move" onClick={() => handleAction(onMove, item)}>
                    <ListItemIcon>
                        <DriveFileMoveIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('move')}</ListItemText>
                </MenuItem>
            );
        }
        if (onDelete) {
            folderMenuItems.push(
                <MenuItem key="delete" onClick={() => handleAction(onDelete, item)} sx={{ color: 'error.main' }}>
                    <ListItemIcon>
                        <DeleteIcon fontSize="small" sx={{ color: 'error.main' }} />
                    </ListItemIcon>
                    <ListItemText>{t('delete')}</ListItemText>
                </MenuItem>
            );
        }
    }

    return (
        <Menu anchorEl={anchorEl} open={open} onClose={onClose} onClick={e => e.stopPropagation()}>
            {itemType === 'file' ? fileMenuItems : folderMenuItems}
        </Menu>
    );
};

export default FileItemMenu;
