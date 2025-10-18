import React from 'react';
import CreateFolderDialog from './CreateFolderDialog';
import FileUpload from './FileUpload';
import ShareDialog from './ShareDialog';

const FileExplorerDialogs = ({
  dialogs,
  uiState,
  navigation,
  user,
  onCreateFolder,
}) => {
  return (
    <>
      {/* Create Folder Dialog */}
      <CreateFolderDialog
        open={dialogs.createFolderOpen}
        onClose={dialogs.closeCreateFolder}
        onSubmit={onCreateFolder}
      />

      {/* File Upload Dialog */}
      <FileUpload
        open={dialogs.fileUploadOpen}
        onClose={dialogs.closeFileUpload}
        onUploadSuccess={() => {
          dialogs.closeFileUpload();
          uiState.triggerRefresh();
        }}
        userId={user?.id}
        currentFolderId={navigation.getCurrentNavState().currentFolder?.id}
      />

      {/* Share Dialog */}
      <ShareDialog
        open={dialogs.shareDialogOpen}
        onClose={dialogs.closeShareDialog}
        resource={dialogs.shareResource}
        resourceType={dialogs.shareResourceType}
      />
    </>
  );
};

export default FileExplorerDialogs;
