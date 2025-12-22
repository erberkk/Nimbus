import { useTranslation } from 'react-i18next';
import {
  Box,
  Breadcrumbs,
  Link,
  Alert,
  IconButton,
  Toolbar,
} from '@mui/material';
import HomeIcon from '@mui/icons-material/Home';
import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import ViewListIcon from '@mui/icons-material/ViewList';
import ViewModuleIcon from '@mui/icons-material/ViewModule';
import SearchBar from './SearchBar';

const FileExplorerHeader = ({
  selectedMenu,
  navigation,
  uiState,
  fileExplorer,
  onBreadcrumbClick,
  onViewModeChange,
  searchQuery,
  onSearchChange,
}) => {
  const { t } = useTranslation();

  // Get the header title based on selectedMenu
  const getHeaderTitle = () => {
    switch (selectedMenu) {
      case 'shared':
        return t('header.shared') || 'Shared with me';
      case 'recent':
        return t('sidebar.recent') || 'Recent';
      case 'starred':
        return t('sidebar.starred') || 'Starred';
      case 'trash':
        return t('sidebar.trash') || 'Trash';
      case 'home':
      default:
        return 'My Drive';
    }
  };

  const headerTitle = getHeaderTitle();

  return (
    <>
      {/* Top Toolbar */}
      <Toolbar
        sx={{
          px: 2,
          py: 2,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          mb: 2,
          borderRadius: 3,
          background: 'rgba(255, 255, 255, 0.7)',
          backdropFilter: 'blur(10px)',
          backgroundColor: 'rgba(255, 255, 255, 0.7)',
          border: '1px solid rgba(255, 255, 255, 0.5)',
          boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.1)',
          transition: 'all 0.3s ease',
          '&:hover': {
            backgroundColor: 'rgba(255, 255, 255, 0.8)',
            boxShadow: '0 8px 32px 0 rgba(31, 38, 135, 0.15)',
          },
        }}
      >
        {/* Breadcrumb */}
        <Breadcrumbs separator={<NavigateNextIcon fontSize="small" />}>
          {['shared', 'recent', 'starred', 'trash'].includes(selectedMenu)
            ? [
              <Link
                key={selectedMenu}
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
                  fontWeight: 500,
                  transition: 'all 0.3s ease',
                  padding: '4px 8px',
                  borderRadius: 1,
                  '&:hover': {
                    color: 'primary.main',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                  },
                }}
              >
                <HomeIcon fontSize="small" />
                {headerTitle}
              </Link>,
              // Folder path için breadcrumb oluştur
              ...navigation.getCurrentNavState().folderPath.map((folder, index) => (
                <Link
                  key={`folder-${index}`}
                  component="button"
                  variant="body1"
                  onClick={() => onBreadcrumbClick(index)}
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 0.5,
                    color:
                      index === navigation.getCurrentNavState().folderPath.length - 1
                        ? 'text.primary'
                        : 'text.secondary',
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
              )),
            ]
            : [
              <Link
                key="home"
                component="button"
                variant="body1"
                onClick={() => onBreadcrumbClick(-1)}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 0.5,
                  color: navigation.getCurrentNavState().currentFolder
                    ? 'text.secondary'
                    : 'text.primary',
                  textDecoration: 'none',
                  cursor: 'pointer',
                  fontSize: '0.95rem',
                  '&:hover': {
                    color: 'primary.main',
                  },
                }}
              >
                <HomeIcon fontSize="small" />
                {headerTitle}
              </Link>,
              // Path'deki her klasör için breadcrumb oluştur
              ...navigation.getCurrentNavState().folderPath.map((folder, index) => (
                <Link
                  key={`folder-${index}`}
                  component="button"
                  variant="body1"
                  onClick={() => onBreadcrumbClick(index)}
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 0.5,
                    color:
                      index === navigation.getCurrentNavState().folderPath.length - 1
                        ? 'text.primary'
                        : 'text.secondary',
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
              )),
            ]}
        </Breadcrumbs>

        {/* Search Bar */}
        <Box sx={{ flex: 1, display: 'flex', justifyContent: 'center', px: 2 }}>
          <SearchBar
            value={searchQuery}
            onChange={onSearchChange}
          />
        </Box>

        {/* View Mode Toggle */}
        <Box
          sx={{
            display: 'flex',
            gap: 0.5,
            alignItems: 'center',
            background: 'rgba(102, 126, 234, 0.1)',
            padding: '4px',
            borderRadius: 2,
            backdropFilter: 'blur(10px)',
          }}
        >
          <IconButton
            size="small"
            onClick={() => onViewModeChange('list')}
            sx={{
              bgcolor: uiState.viewMode === 'list' ? 'rgba(102, 126, 234, 0.2)' : 'transparent',
              color: uiState.viewMode === 'list' ? 'primary.main' : 'text.secondary',
              transition: 'all 0.3s ease',
              borderRadius: 1,
              '&:hover': {
                bgcolor: 'rgba(102, 126, 234, 0.15)',
              },
            }}
          >
            <ViewListIcon />
          </IconButton>
          <IconButton
            size="small"
            onClick={() => onViewModeChange('grid')}
            sx={{
              bgcolor: uiState.viewMode === 'grid' ? 'rgba(102, 126, 234, 0.2)' : 'transparent',
              color: uiState.viewMode === 'grid' ? 'primary.main' : 'text.secondary',
              transition: 'all 0.3s ease',
              borderRadius: 1,
              '&:hover': {
                bgcolor: 'rgba(102, 126, 234, 0.15)',
              },
            }}
          >
            <ViewModuleIcon />
          </IconButton>
        </Box>
      </Toolbar>

      {/* Error Alert */}
      {fileExplorer.error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => { }}>
          {fileExplorer.error}
        </Alert>
      )}
    </>
  );
};

export default FileExplorerHeader;
