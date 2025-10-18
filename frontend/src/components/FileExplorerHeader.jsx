import React from 'react';
import {
  Box,
  Typography,
  Button,
  Breadcrumbs,
  Link,
  Alert,
  IconButton,
  Toolbar,
} from '@mui/material';
import HomeIcon from '@mui/icons-material/Home';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import CreateNewFolderIcon from '@mui/icons-material/CreateNewFolder';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ViewListIcon from '@mui/icons-material/ViewList';
import ViewModuleIcon from '@mui/icons-material/ViewModule';

const FileExplorerHeader = ({
  selectedMenu,
  navigation,
  uiState,
  fileExplorer,
  onBreadcrumbClick,
  onCreateFolder,
  onFileUpload,
  onViewModeChange,
}) => {
  return (
    <>
      {/* Top Toolbar */}
      <Toolbar
        sx={{
          px: 0,
          py: 1.5,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: 1,
          borderColor: 'divider',
          mb: 1,
        }}
      >
        {/* Breadcrumb */}
        <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />}>
          {selectedMenu === 'shared' ? [
            <Link
              key="shared"
              component="button"
              variant="body1"
              onClick={() => onBreadcrumbClick(-1)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 0.5,
                color: 'text.primary',
                textDecoration: 'none',
                cursor: 'pointer',
                fontSize: '0.95rem',
                '&:hover': {
                  color: 'primary.main',
                },
              }}
            >
              <HomeIcon fontSize="small" />
              Paylaşılanlarım
            </Link>,
            // Shared folder path için breadcrumb oluştur
            ...(navigation.getCurrentNavState().folderPath.map((folder, index) => (
              <Link
                key={`shared-folder-${index}`}
                component="button"
                variant="body1"
                onClick={() => onBreadcrumbClick(index)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5,
                  color: index === navigation.getCurrentNavState().folderPath.length - 1 ? 'text.primary' : 'text.secondary',
                  textDecoration: 'none',
                  cursor: 'pointer',
                  fontSize: '0.95rem',
                  '&:hover': {
                    color: 'primary.main',
                  },
                }}
              >
                {folder.name}
              </Link>
            )))
          ] : [
            <Link
              key="home"
              component="button"
              variant="body1"
              onClick={() => onBreadcrumbClick(-1)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 0.5,
                color: navigation.getCurrentNavState().currentFolder ? 'text.secondary' : 'text.primary',
                textDecoration: 'none',
                cursor: 'pointer',
                fontSize: '0.95rem',
                '&:hover': {
                  color: 'primary.main',
                },
              }}
            >
              <HomeIcon fontSize="small" />
              My Drive
            </Link>,
            // Path'deki her klasör için breadcrumb oluştur
            ...(navigation.getCurrentNavState().folderPath.map((folder, index) => (
              <Link
                key={`folder-${index}`}
                component="button"
                variant="body1"
                onClick={() => onBreadcrumbClick(index)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5,
                  color: index === navigation.getCurrentNavState().folderPath.length - 1 ? 'text.primary' : 'text.secondary',
                  textDecoration: 'none',
                  cursor: 'pointer',
                  fontSize: '0.95rem',
                  '&:hover': {
                    color: 'primary.main',
                  },
                }}
              >
                {folder.name}
              </Link>
            )))
          ]}
        </Breadcrumbs>

        {/* Action Buttons */}
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <Button
            variant="outlined"
            startIcon={<CreateNewFolderIcon />}
            onClick={onCreateFolder}
            size="small"
            sx={{ textTransform: 'none' }}
          >
            Klasör Oluştur
          </Button>
          <Button
            variant="outlined"
            startIcon={<CloudUploadIcon />}
            onClick={onFileUpload}
            size="small"
            sx={{ textTransform: 'none' }}
          >
            Dosya Yükle
          </Button>

          {/* View Mode Toggle */}
          <Box sx={{ display: 'flex', gap: 1 }}>
            <IconButton
              size="small"
              onClick={() => onViewModeChange('list')}
              sx={{
                bgcolor: uiState.viewMode === 'list' ? 'action.selected' : 'transparent',
              }}
            >
              <ViewListIcon />
            </IconButton>
            <IconButton
              size="small"
              onClick={() => onViewModeChange('grid')}
              sx={{
                bgcolor: uiState.viewMode === 'grid' ? 'action.selected' : 'transparent',
              }}
            >
              <ViewModuleIcon />
            </IconButton>
            <IconButton size="small">
              {/* More options can be added here */}
            </IconButton>
          </Box>
        </Box>
      </Toolbar>

      {/* Error Alert */}
      {fileExplorer.error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => {}}>
          {fileExplorer.error}
        </Alert>
      )}
    </>
  );
};

export default FileExplorerHeader;
