import { Box } from '@mui/material';
import { useState, useRef, useEffect } from 'react';
import Sidebar from '../components/Sidebar';
import FileExplorerNew from '../components/FileExplorerNew';
import CreateFolderDialog from '../components/CreateFolderDialog';
import FileUpload from '../components/FileUpload';

const HomePage = () => {
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [fileUploadOpen, setFileUploadOpen] = useState(false);
  const [uploadMode, setUploadMode] = useState('file'); // 'file' | 'folder' | 'both'
  
  // Load selectedMenu from localStorage on mount
  const getInitialMenu = () => {
    try {
      const saved = localStorage.getItem('nimbus_selected_menu');
      return saved || 'home';
    } catch (error) {
      return 'home';
    }
  };
  
  const [selectedMenu, setSelectedMenu] = useState(getInitialMenu);
  const [chatFile, setChatFile] = useState(null); // File to open in chat panel
  const fileExplorerRef = useRef();
  
  // Save selectedMenu to localStorage whenever it changes
  useEffect(() => {
    try {
      localStorage.setItem('nimbus_selected_menu', selectedMenu);
    } catch (error) {
      console.error('Error saving selected menu:', error);
    }
  }, [selectedMenu]);

  const handleCreateFolder = () => {
    setCreateFolderOpen(true);
  };

  const handleFileUpload = () => {
    setUploadMode('file');
    setFileUploadOpen(true);
  };

  const handleMenuChange = menuId => {
    setSelectedMenu(menuId);
  };

  return (
    <Box
      sx={{
        display: 'flex',
        bgcolor: 'background.default',
        height: '100vh',
        overflow: 'hidden',
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
      }}
    >
      {/* Sidebar */}
      <Sidebar
        onCreateFolder={handleCreateFolder}
        onFileUpload={handleFileUpload}
        onMenuChange={handleMenuChange}
        selectedMenu={selectedMenu}
        onConversationClick={file => {
          // Open chat panel with the file
          setChatFile(file);
        }}
      />

      {/* Main Content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          ml: '220px',
          mt: '64px',
          height: 'calc(100vh - 64px)',
          overflow: 'hidden',
          p: 2,
        }}
      >
        <FileExplorerNew 
          ref={fileExplorerRef} 
          selectedMenu={selectedMenu}
          chatFile={chatFile}
          onChatFileCleared={() => setChatFile(null)}
        />
      </Box>

      {/* Dialogs */}
      <CreateFolderDialog
        open={createFolderOpen}
        onClose={() => setCreateFolderOpen(false)}
        onSubmit={async folderData => {
          // Use FileExplorerNew's handleCreateFolder function
          if (fileExplorerRef.current?.handleCreateFolder) {
            await fileExplorerRef.current.handleCreateFolder(folderData);
          } else {
            throw new Error('FileExplorer handleCreateFolder not available');
          }
        }}
      />

      <FileUpload
        open={fileUploadOpen}
        onClose={() => setFileUploadOpen(false)}
        onUploadSuccess={() => {
          setFileUploadOpen(false);
        }}
        mode={uploadMode}
      />
    </Box>
  );
};

export default HomePage;
