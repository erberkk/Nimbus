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

const FileExplorerNew = forwardRef(({ selectedMenu = 'home' }, ref) => {
  const { t } = useTranslation();
  const { user } = useAuth();

  // Custom hooks
  const navigation = useNavigation(selectedMenu);
  const fileExplorer = useFileExplorer(selectedMenu, navigation.getCurrentNavState);
  const dialogs = useDialogs();
  const uiState = useUIState();

  useImperativeHandle(ref, () => ({
    handleCreateFolder: fileExplorer.createFolder,
  }));

  useEffect(() => {
    fileExplorer.loadContents();
  }, [fileExplorer.loadContents, uiState.refreshTrigger]);

  const handleFolderOpen = folder => {
    // For shared folders, mark them as shared for proper navigation
    const updatedFolder = folder.isShared ? { ...folder, isShared: true } : folder;
    navigation.handleFolderOpen(updatedFolder);
  };

  const handleBackToRoot = () => {
    navigation.handleBackToRoot();
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
      />

      <FileExplorerDialogs
        dialogs={dialogs}
        uiState={uiState}
        navigation={navigation}
        user={user}
        onCreateFolder={handleCreateFolder}
        onDownloadFile={handleDownloadFile}
      />
    </Box>
  );
});

export default FileExplorerNew;
