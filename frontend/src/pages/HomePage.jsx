import { Box } from '@mui/material';
import { useState, useRef } from 'react';
import Sidebar from '../components/Sidebar';
import FileExplorerNew from '../components/FileExplorerNew';
import CreateFolderDialog from '../components/CreateFolderDialog';
import FileUpload from '../components/FileUpload';

const HomePage = () => {
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [fileUploadOpen, setFileUploadOpen] = useState(false);
  const [selectedMenu, setSelectedMenu] = useState('home');
  const fileExplorerRef = useRef();

  const handleCreateFolder = () => {
    setCreateFolderOpen(true);
  };

  const handleFileUpload = () => {
    setFileUploadOpen(true);
  };

  const handleMenuChange = (menuId) => {
    setSelectedMenu(menuId);
  };

  return (
    <Box sx={{ 
      display: 'flex', 
      bgcolor: 'background.default', 
      height: '100vh', 
      overflow: 'hidden',
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0
    }}>
      {/* Sidebar */}
      <Sidebar 
        onCreateFolder={handleCreateFolder} 
        onFileUpload={handleFileUpload}
        onMenuChange={handleMenuChange}
        selectedMenu={selectedMenu}
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
        <FileExplorerNew ref={fileExplorerRef} selectedMenu={selectedMenu} />
      </Box>

      {/* Dialogs */}
      <CreateFolderDialog
        open={createFolderOpen}
        onClose={() => setCreateFolderOpen(false)}
        onSubmit={async (folderData) => {
          // Use FileExplorerNew's handleCreateFolder function
          if (fileExplorerRef.current?.handleCreateFolder) {
            try {
              await fileExplorerRef.current.handleCreateFolder(folderData);
            } catch (error) {
              console.error('Folder creation failed:', error);
            }
          }
          setCreateFolderOpen(false);
        }}
      />

      <FileUpload
        open={fileUploadOpen}
        onClose={() => setFileUploadOpen(false)}
        onUploadSuccess={() => {
          console.log('File uploaded successfully');
          setFileUploadOpen(false);
        }}
      />
    </Box>
  );
};

export default HomePage;
