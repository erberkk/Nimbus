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
  const [previewDialogOpen, setPreviewDialogOpen] = useState(false);
  const [previewFile, setPreviewFile] = useState(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editFile, setEditFile] = useState(null);
  const [fileUploadMode, setFileUploadMode] = useState('both');

  // Dialog control functions
  const openCreateFolder = useCallback(() => {
    setCreateFolderOpen(true);
  }, []);

  const closeCreateFolder = useCallback(() => {
    setCreateFolderOpen(false);
  }, []);

  const openFileUpload = useCallback((mode = 'both') => {
    setFileUploadMode(mode);
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

  const openPreviewDialog = useCallback(file => {
    setPreviewFile(file);
    setPreviewDialogOpen(true);
  }, []);

  const closePreviewDialog = useCallback(() => {
    setPreviewDialogOpen(false);
    setPreviewFile(null);
  }, []);

  const openEditDialog = useCallback(file => {
    setEditFile(file);
    setEditDialogOpen(true);
  }, []);

  const closeEditDialog = useCallback(() => {
    setEditDialogOpen(false);
    setEditFile(null);
  }, []);

  return {
    // Dialog states
    createFolderOpen,
    fileUploadOpen,
    fileUploadMode,
    shareDialogOpen,
    shareResource,
    shareResourceType,
    previewDialogOpen,
    previewFile,
    editDialogOpen,
    editFile,

    // Dialog control functions
    openCreateFolder,
    closeCreateFolder,
    openFileUpload,
    closeFileUpload,
    openShareDialog,
    closeShareDialog,
    openPreviewDialog,
    closePreviewDialog,
    openEditDialog,
    closeEditDialog,
  };
};
