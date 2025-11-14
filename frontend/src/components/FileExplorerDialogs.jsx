import React from 'react';
import CreateFolderDialog from './CreateFolderDialog';
import FileUpload from './FileUpload';
import ShareDialog from './ShareDialog';
import FilePreviewDialog from './FilePreviewDialog';
import OnlyOfficeEditor from './OnlyOfficeEditor';
import { isCodeFile } from '../utils/fileUtils';

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

      {/* OnlyOffice Editor Dialog - Only for Office documents */}
      {dialogs.editFile &&
        !isCodeFile(dialogs.editFile.content_type, dialogs.editFile.filename) && (
          <OnlyOfficeEditor
            open={dialogs.editDialogOpen}
            onClose={dialogs.closeEditDialog}
            file={dialogs.editFile}
            onSave={() => {
              // Trigger refresh after save
              uiState.triggerRefresh();
            }}
          />
        )}

      {/* Code files use FilePreviewDialog for editing */}
      {dialogs.editFile && isCodeFile(dialogs.editFile.content_type, dialogs.editFile.filename) && (
        <FilePreviewDialog
          open={dialogs.editDialogOpen}
          onClose={dialogs.closeEditDialog}
          file={dialogs.editFile}
          onDownload={onDownloadFile}
          onSave={() => {
            // Trigger refresh after save
            uiState.triggerRefresh();
          }}
        />
      )}
    </>
  );
};

export default FileExplorerDialogs;
