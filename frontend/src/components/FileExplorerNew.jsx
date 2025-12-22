import { useEffect, forwardRef, useImperativeHandle } from 'react';
import { useTranslation } from 'react-i18next';
import { Box } from '@mui/material';
import { useAuth } from '../contexts/AuthContext';
import { useNavigation } from '../hooks/useNavigation';
import { useFileExplorer } from '../hooks/useFileExplorer';
import { useDialogs } from '../hooks/useDialogs';
import { useUIState } from '../hooks/useUIState';
import FileExplorerHeader from './FileExplorerHeader';
import FileExplorerContent from './FileExplorerContent';
import FileExplorerContextMenu from './FileExplorerContextMenu';
import FileExplorerDialogs from './FileExplorerDialogs';
import MoveDialog from './MoveDialog';
import React, { useState } from 'react';

const FileExplorerNew = forwardRef(({ selectedMenu = 'home', chatFile, onChatFileCleared }, ref) => {
  const { t } = useTranslation();
  const { user } = useAuth();

  // Custom hooks
  const navigation = useNavigation(selectedMenu);
  const fileExplorer = useFileExplorer(selectedMenu, navigation.getCurrentNavState);
  const dialogs = useDialogs();
  const uiState = useUIState();
  const [moveDialogOpen, setMoveDialogOpen] = useState(false);
  const [itemToMove, setItemToMove] = useState(null);
  const [itemToMoveType, setItemToMoveType] = useState('file');
  const [searchQuery, setSearchQuery] = useState('');

  useImperativeHandle(ref, () => ({
    handleCreateFolder: fileExplorer.createFolder,
  }));

  useEffect(() => {
    fileExplorer.loadContents();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedMenu, uiState.refreshTrigger]);

  useEffect(() => {
    fileExplorer.loadContents();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigation.navigationState[selectedMenu]]);

  const handleFolderOpen = folder => {
    const updatedFolder = folder.isShared ? { ...folder, isShared: true } : folder;
    navigation.handleFolderOpen(updatedFolder);
  };

  const handleBreadcrumbClick = index => {
    navigation.handleBreadcrumbClick(index);
  };

  const handleCreateFolder = async folderData => {
    await fileExplorer.createFolder(folderData);
  };

  const handleDeleteFolder = async folder => {
    if (!window.confirm(t('folder.delete_confirm', { name: folder.name }))) {
      return;
    }
    await fileExplorer.deleteFolder(folder.id);
  };

  const handleDownloadFile = async file => {
    await fileExplorer.downloadFile(file);
  };

  const handleDeleteFile = async file => {
    if (!window.confirm(t('file.delete_confirm'))) {
      return;
    }
    await fileExplorer.deleteFile(file.id);
  };

  const handleUploadSuccess = () => {
    uiState.triggerRefresh();
  };

  const handleMenuOpen = (event, item) => {
    uiState.openContextMenu(event, item);
  };

  const handleMenuClose = () => {
    uiState.closeContextMenu();
  };

  const handleShare = (resource, type) => {
    dialogs.openShareDialog(resource, type);
  };


  const handleOpenMoveDialog = (item, type) => {
    setItemToMove(item);
    setItemToMoveType(type);
    setMoveDialogOpen(true);
  };

  const handleMoveItem = async (item, targetFolderId, type) => {
    await fileExplorer.moveItem(item, targetFolderId, type);
  };

  const handleSearchChange = (query) => {
    setSearchQuery(query);
  };

  return (
    <Box sx={{ height: '100%', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
      <FileExplorerHeader
        selectedMenu={selectedMenu}
        navigation={navigation}
        uiState={uiState}
        fileExplorer={fileExplorer}
        onBreadcrumbClick={handleBreadcrumbClick}
        onCreateFolder={dialogs.openCreateFolder}
        onFileUpload={dialogs.openFileUpload}
        onViewModeChange={uiState.setViewMode}
        searchQuery={searchQuery}
        onSearchChange={handleSearchChange}
      />

      <FileExplorerContent
        fileExplorer={fileExplorer}
        uiState={uiState}
        navigation={navigation}
        onCreateFolder={dialogs.openCreateFolder}
        onFileUpload={dialogs.openFileUpload}
        onFolderOpen={handleFolderOpen}
        onFileDownload={handleDownloadFile}
        onFolderDelete={handleDeleteFolder}
        onFileDelete={handleDeleteFile}
        onShare={handleShare}
        onMenuOpen={handleMenuOpen}
        onUploadSuccess={handleUploadSuccess}
        onPreview={dialogs.openPreviewDialog}
        onEdit={dialogs.openEditDialog}
        onMove={handleOpenMoveDialog}
        onToggleStar={fileExplorer.toggleStar}
        onRestore={fileExplorer.restoreItem}
        chatFile={chatFile}
        onChatFileCleared={onChatFileCleared}
        searchQuery={searchQuery}
      />

      <FileExplorerContextMenu
        menuAnchor={uiState.menuAnchor}
        selectedItem={uiState.selectedItem}
        onClose={handleMenuClose}
        onDownloadFile={handleDownloadFile}
        onDeleteFile={handleDeleteFile}
        onDeleteFolder={handleDeleteFolder}
        onCreateFolder={dialogs.openCreateFolder}
        onFileUpload={dialogs.openFileUpload}
        onMove={handleOpenMoveDialog}
        onToggleStar={fileExplorer.toggleStar}
        onRestore={fileExplorer.restoreItem}
      />

      <FileExplorerDialogs
        dialogs={dialogs}
        uiState={uiState}
        navigation={navigation}
        user={user}
        onCreateFolder={handleCreateFolder}
        onDownloadFile={handleDownloadFile}
      />

      <MoveDialog
        open={moveDialogOpen}
        onClose={() => setMoveDialogOpen(false)}
        onMove={handleMoveItem}
        item={itemToMove}
        itemType={itemToMoveType}
      />
    </Box>
  );
});

export default FileExplorerNew;
