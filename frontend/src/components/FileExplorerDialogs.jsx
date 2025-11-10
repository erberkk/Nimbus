import React from 'react';
import CreateFolderDialog from './CreateFolderDialog';
import FileUpload from './FileUpload';
import ShareDialog from './ShareDialog';
import FilePreviewDialog from './FilePreviewDialog';

const FileExplorerDialogs = ({
  dialogs,
  uiState,
  navigation,
  user,
  onCreateFolder,
  onDownloadFile,
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

      {/* File Preview Dialog */}
      <FilePreviewDialog
        open={dialogs.previewDialogOpen}
        onClose={dialogs.closePreviewDialog}
        file={dialogs.previewFile}
        onDownload={onDownloadFile}
      />
    </>
  );
};

export default FileExplorerDialogs;
