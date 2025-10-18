import { useState, useCallback } from 'react';

/**
 * Dialog state management hook
 * Manages all dialog states and their open/close functions
 */
export const useDialogs = () => {
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [fileUploadOpen, setFileUploadOpen] = useState(false);
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const [shareResource, setShareResource] = useState(null);
  const [shareResourceType, setShareResourceType] = useState(null);

  // Dialog control functions
  const openCreateFolder = useCallback(() => {
    setCreateFolderOpen(true);
  }, []);

  const closeCreateFolder = useCallback(() => {
    setCreateFolderOpen(false);
  }, []);

  const openFileUpload = useCallback(() => {
    setFileUploadOpen(true);
  }, []);

  const closeFileUpload = useCallback(() => {
    setFileUploadOpen(false);
  }, []);

  const openShareDialog = useCallback((resource, resourceType) => {
    setShareResource(resource);
    setShareResourceType(resourceType);
    setShareDialogOpen(true);
  }, []);

  const closeShareDialog = useCallback(() => {
    setShareDialogOpen(false);
    setShareResource(null);
    setShareResourceType(null);
  }, []);

  return {
    // Dialog states
    createFolderOpen,
    fileUploadOpen,
    shareDialogOpen,
    shareResource,
    shareResourceType,
    
    // Dialog control functions
    openCreateFolder,
    closeCreateFolder,
    openFileUpload,
    closeFileUpload,
    openShareDialog,
    closeShareDialog
  };
};
