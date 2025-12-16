import { useTranslation } from 'react-i18next';
import { Menu, MenuItem, ListItemIcon, ListItemText } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import DeleteIcon from '@mui/icons-material/Delete';
import DriveFileMoveIcon from '@mui/icons-material/DriveFileMove';
import ShareIcon from '@mui/icons-material/Share';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import EditIcon from '@mui/icons-material/Edit';
import InfoIcon from '@mui/icons-material/Info';
import StarIcon from '@mui/icons-material/Star';
import StarBorderIcon from '@mui/icons-material/StarBorder';
import RestoreFromTrashIcon from '@mui/icons-material/RestoreFromTrash';
import { isAskableFile, isEditable } from '../utils/fileUtils';

const FileItemMenu = ({
    anchorEl,
    open,
    onClose,
    item,
    itemType,
    onInfo,
    onDownload,
    onEdit,
    onAskNimbus,
    onShare,
    onMove,
    onToggleStar,
    onRestore,
    onDelete,
}) => {
    const { t } = useTranslation();

    const isTrash = (item && !!item.deleted_at) || (item && item.isTrash === true);
    const isStarred = item && !!item.is_starred;

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
        if (onInfo && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="info" onClick={() => handleAction(onInfo, item)}>
                    <ListItemIcon>
                        <InfoIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('info')}</ListItemText>
                </MenuItem>
            );
        }
        if (onDownload && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="download" onClick={() => handleAction(onDownload, item)}>
                    <ListItemIcon>
                        <DownloadIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('download')}</ListItemText>
                </MenuItem>
            );
        }
        if (fileIsEditable && onEdit && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="edit" onClick={() => handleAction(onEdit, item)} sx={{ color: '#667eea' }}>
                    <ListItemIcon>
                        <EditIcon fontSize="small" sx={{ color: '#667eea' }} />
                    </ListItemIcon>
                    <ListItemText>{t('edit')}</ListItemText>
                </MenuItem>
            );
        }
        if (fileIsAskable && onAskNimbus && !isTrash) {
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
        if (onToggleStar && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="star" onClick={() => handleAction(onToggleStar, item)}>
                    <ListItemIcon>
                        {isStarred ? <StarIcon fontSize="small" sx={{ color: '#FFD700' }} /> : <StarBorderIcon fontSize="small" />}
                    </ListItemIcon>
                    <ListItemText>{isStarred ? t('starred.remove') : t('starred.add')}</ListItemText>
                </MenuItem>
            );
        }
        if (onShare && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="share" onClick={() => handleAction(onShare, item)}>
                    <ListItemIcon>
                        <ShareIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('share')}</ListItemText>
                </MenuItem>
            );
        }
        if (onMove && !isTrash) {
            fileMenuItems.push(
                <MenuItem key="move" onClick={() => handleAction(onMove, item, itemType)}>
                    <ListItemIcon>
                        <DriveFileMoveIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('move')}</ListItemText>
                </MenuItem>
            );
        }
        if (onRestore && isTrash) {
            fileMenuItems.push(
                <MenuItem key="restore" onClick={() => handleAction(onRestore, item)}>
                    <ListItemIcon>
                        <RestoreFromTrashIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('restore')}</ListItemText>
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
        if (onToggleStar && !isTrash) {
            folderMenuItems.push(
                <MenuItem key="star" onClick={() => handleAction(onToggleStar, item)}>
                    <ListItemIcon>
                        {isStarred ? <StarIcon fontSize="small" sx={{ color: '#FFD700' }} /> : <StarBorderIcon fontSize="small" />}
                    </ListItemIcon>
                    <ListItemText>{isStarred ? t('starred.remove') : t('starred.add')}</ListItemText>
                </MenuItem>
            );
        }
        if (onShare && !isTrash) {
            folderMenuItems.push(
                <MenuItem key="share" onClick={() => handleAction(onShare, item)}>
                    <ListItemIcon>
                        <ShareIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('share')}</ListItemText>
                </MenuItem>
            );
        }
        if (onMove && !isTrash) {
            folderMenuItems.push(
                <MenuItem key="move" onClick={() => handleAction(onMove, item, itemType)}>
                    <ListItemIcon>
                        <DriveFileMoveIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('move')}</ListItemText>
                </MenuItem>
            );
        }
        if (onRestore && isTrash) {
            folderMenuItems.push(
                <MenuItem key="restore" onClick={() => handleAction(onRestore, item)}>
                    <ListItemIcon>
                        <RestoreFromTrashIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('restore')}</ListItemText>
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
