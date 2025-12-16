import React, { useState, useEffect, useCallback } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Breadcrumbs,
    Link,
    Box,
    Typography,
    CircularProgress
} from '@mui/material';
import FolderIcon from '@mui/icons-material/Folder';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import HomeIcon from '@mui/icons-material/Home';
import { folderApi } from '../services/api';

const getDescendantIds = async (folderId, api) => {
    const descendants = new Set([folderId]);
    const toProcess = [folderId];
    
    while (toProcess.length > 0) {
        const currentId = toProcess.pop();
        try {
            const data = await api.getFolderContents(currentId);
            const folders = data.folders || [];
            folders.forEach(f => {
                if (!descendants.has(f.id)) {
                    descendants.add(f.id);
                    toProcess.push(f.id);
                }
            });
        } catch (error) {
            console.error("Descendants load error:", error);
        }
    }
    
    return descendants;
};

const MoveDialog = ({ open, onClose, onMove, item, itemType }) => {
    const [currentFolderId, setCurrentFolderId] = useState(null);
    const [folders, setFolders] = useState([]);
    const [breadcrumbs, setBreadcrumbs] = useState([]);
    const [loading, setLoading] = useState(false);
    const [excludedFolderIds, setExcludedFolderIds] = useState(new Set());

    const loadFolders = useCallback(async (folderId) => {
        setLoading(true);
        try {
            if (folderId) {
                const data = await folderApi.getFolderContents(folderId);
                let validFolders = data.folders || [];
                if (itemType === 'folder') {
                    validFolders = validFolders.filter(f => !excludedFolderIds.has(f.id));
                }
                setFolders(validFolders);
                setBreadcrumbs(data.breadcrumbs || []);
            } else {
                const data = await folderApi.getRootContents();
                let validFolders = data.folders || [];
                if (itemType === 'folder') {
                    validFolders = validFolders.filter(f => !excludedFolderIds.has(f.id));
                }
                setFolders(validFolders);
                setBreadcrumbs([]);
            }
            setCurrentFolderId(folderId);
        } catch (error) {
            console.error("Klasörler yüklenemedi", error);
        } finally {
            setLoading(false);
        }
    }, [excludedFolderIds, itemType]);

    useEffect(() => {
        if (open && itemType === 'folder') {
            const loadDescendants = async () => {
                const descendants = await getDescendantIds(item.id, folderApi);
                setExcludedFolderIds(descendants);
            };
            loadDescendants();
        } else {
            setExcludedFolderIds(new Set());
        }
    }, [open, item, itemType]);

    useEffect(() => {
        if (open) {
            setCurrentFolderId(null);
            if (itemType === 'folder') {
                loadFolders(null);
            } else if (itemType !== 'folder') {
                loadFolders(null);
            }
        }
    }, [open, excludedFolderIds, itemType, loadFolders]);

    const handleFolderClick = (folder) => {
        loadFolders(folder.id);
    };

    const handleBreadcrumbClick = (folderId) => {
        loadFolders(folderId);
    };

    const handleMove = () => {
        onMove(item, currentFolderId, itemType);
        onClose();
    };

    return (
        <Dialog 
            open={open} 
            onClose={onClose} 
            maxWidth="sm" 
            fullWidth
            PaperProps={{
                sx: {
                    background: 'rgba(255, 255, 255, 0.9)',
                    backdropFilter: 'blur(10px)',
                    borderRadius: 3,
                    boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.15)',
                }
            }}
        >
            <DialogTitle sx={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                color: 'white',
                fontWeight: 600,
                fontSize: '1.1rem',
            }}>
                📁 {item?.filename || item?.name}
            </DialogTitle>
            <DialogContent dividers sx={{ 
                minHeight: '300px',
                background: 'rgba(255, 255, 255, 0.5)',
                backdropFilter: 'blur(10px)',
            }}>
                <Box sx={{ mb: 2, mt: 1 }}>
                    <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />} aria-label="breadcrumb">
                        <Link
                            component="button"
                            underline="hover"
                            color="inherit"
                            onClick={() => handleBreadcrumbClick(null)}
                            sx={{ 
                                display: 'flex', 
                                alignItems: 'center',
                                '&:hover': {
                                    textDecoration: 'underline',
                                    color: '#667eea',
                                },
                                cursor: 'pointer',
                                transition: 'all 0.2s ease',
                            }}
                        >
                            <HomeIcon sx={{ mr: 0.5 }} fontSize="inherit" />
                            Ana Dizin
                        </Link>
                        {breadcrumbs.map((crumb) => (
                            <Link
                                key={crumb.id}
                                component="button"
                                underline="hover"
                                color="inherit"
                                onClick={() => handleBreadcrumbClick(crumb.id)}
                                sx={{
                                    '&:hover': {
                                        textDecoration: 'underline',
                                        color: '#667eea',
                                    },
                                    cursor: 'pointer',
                                    transition: 'all 0.2s ease',
                                }}
                            >
                                {crumb.name}
                            </Link>
                        ))}
                    </Breadcrumbs>
                </Box>

                {loading ? (
                    <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
                        <CircularProgress />
                    </Box>
                ) : (
                    <List sx={{ mt: 1 }}>
                        {folders.length === 0 ? (
                            <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', mt: 2 }}>
                                Bu klasörde alt klasör yok.
                            </Typography>
                        ) : (
                            folders.map((folder) => (
                                <ListItem
                                    key={folder.id}
                                    onClick={() => handleFolderClick(folder)}
                                    sx={{
                                        borderRadius: 1.5,
                                        mb: 0.75,
                                        background: 'rgba(255, 255, 255, 0.6)',
                                        border: '1px solid rgba(102, 126, 234, 0.2)',
                                        transition: 'all 0.2s ease',
                                        cursor: 'pointer',
                                        '&:hover': {
                                            background: 'rgba(102, 126, 234, 0.1)',
                                            border: '1px solid rgba(102, 126, 234, 0.5)',
                                            transform: 'translateX(4px)',
                                        }
                                    }}
                                >
                                    <ListItemIcon>
                                        <FolderIcon sx={{ color: '#667eea' }} />
                                    </ListItemIcon>
                                    <ListItemText 
                                        primary={folder.name}
                                        primaryTypographyProps={{
                                            sx: { fontWeight: 500, color: '#333' }
                                        }}
                                    />
                                </ListItem>
                            ))
                        )}
                    </List>
                )}
            </DialogContent>
            <DialogActions sx={{
                background: 'rgba(255, 255, 255, 0.5)',
                backdropFilter: 'blur(10px)',
                p: 2,
                gap: 1,
            }}>
                <Button 
                    onClick={onClose}
                    sx={{
                        textTransform: 'none',
                        fontSize: '0.9rem',
                        '&:hover': {
                            background: 'rgba(0, 0, 0, 0.05)',
                        }
                    }}
                >
                    İptal
                </Button>
                <Button 
                    onClick={handleMove} 
                    variant="contained" 
                    sx={{
                        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                        textTransform: 'none',
                        fontSize: '0.9rem',
                        fontWeight: 600,
                        borderRadius: 1.5,
                        '&:hover': {
                            boxShadow: '0 8px 16px rgba(102, 126, 234, 0.3)',
                        }
                    }}
                >
                    Buraya Taşı
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default MoveDialog;
