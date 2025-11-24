import { useState, useCallback } from 'react';
import { folderApi, fileApi, shareApi, api } from '../services/api';

/**
 * File explorer data management hook
 * Manages folders, files, loading states and file operations
 */
export const useFileExplorer = (selectedMenu, getCurrentNavState) => {
  const [folders, setFolders] = useState([]);
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadContents = useCallback(async () => {
    try {
      setLoading(true);
      setError('');

      const currentNav = getCurrentNavState();
      const { currentFolder } = currentNav;

      if (selectedMenu === 'shared') {
        if (currentFolder && currentFolder.isShared) {
          // Load shared folder contents
          const response = await shareApi.getSharedFolderContents(currentFolder.id);
          if (!response) {
            setFolders([]);
            setFiles([]);
            return;
          }

          const folders = response.folders || [];
          const files = response.files || [];

          setFolders(
            folders.map(item => ({
              ...item.resource,
              item_count: item.resource.ItemCount || item.resource.item_count || 0,
              access_type: item.access_type,
              owner: item.owner,
              isShared: true,
            }))
          );
          setFiles(
            files.map(item => ({
              ...item.resource,
              access_type: item.access_type,
              owner: item.owner,
              isShared: true,
            }))
          );
        } else {
          // Load shared files and folders at root level
          const response = await shareApi.getSharedWithMe();
          if (!response) {
            setFolders([]);
            setFiles([]);
            return;
          }

          const sharedFiles = response.filter(item => item.resource_type === 'file');
          const sharedFolders = response.filter(item => item.resource_type === 'folder');

          const foldersWithCount = sharedFolders.map(item => ({
            ...item.resource,
            item_count: item.resource.ItemCount || item.resource.item_count || 0,
            access_type: item.access_type,
            owner: item.owner,
            isShared: true,
          }));

          setFolders(foldersWithCount);
          setFiles(
            sharedFiles.map(item => ({
              ...item.resource,
              access_type: item.access_type,
              owner: item.owner,
              isShared: true,
            }))
          );
        }
      } else {
        // Load home folder contents
        if (currentFolder) {
          const response = await folderApi.getFolderContents(currentFolder.id);
          if (!response) {
            setFolders([]);
            setFiles([]);
            return;
          }
          setFolders(response.folders || []);
          setFiles(response.files || []);
        } else {
          const response = await folderApi.getRootContents();
          if (!response) {
            setFolders([]);
            setFiles([]);
            return;
          }
          setFolders(response.folders || []);
          setFiles(response.files || []);
        }
      }
    } catch (err) {
      console.error('İçerik yükleme hatası:', err);
      setError('İçerik yüklenirken hata oluştu');
      window.toast?.error('İçerik yüklenirken hata oluştu');
    } finally {
      setLoading(false);
    }
  }, [selectedMenu, getCurrentNavState]);

  const createFolder = useCallback(
    async folderData => {
      try {
        const currentNav = getCurrentNavState();
        const parentId = currentNav.currentFolder?.id || null;

        const response = await folderApi.createFolder({
          ...folderData,
          folder_id: parentId,
        });

        if (response && response.folder) {
          window.toast?.success('Klasör başarıyla oluşturuldu');
          loadContents(); // Refresh the list
          return response.folder;
        }
      } catch (error) {
        console.error('Klasör oluşturma hatası:', error);
        window.toast?.error('Klasör oluşturulurken hata oluştu');
        throw error;
      }
    },
    [getCurrentNavState, loadContents]
  );

  const deleteFolder = useCallback(
    async folderId => {
      try {
        await folderApi.deleteFolder(folderId);
        window.toast?.success('Klasör başarıyla silindi');
        loadContents(); // Refresh the list
      } catch (error) {
        console.error('Klasör silme hatası:', error);
        window.toast?.error('Klasör silinirken hata oluştu');
      }
    },
    [loadContents]
  );

  const deleteFile = useCallback(
    async fileId => {
      try {
        await fileApi.deleteFile(fileId);
        window.toast?.success('Dosya başarıyla silindi');
        loadContents(); // Refresh the list
      } catch (error) {
        console.error('Dosya silme hatası:', error);
        window.toast?.error('Dosya silinirken hata oluştu');
      }
    },
    [loadContents]
  );

  const downloadFile = useCallback(async file => {
    try {
      // Log the file object to debug
      console.log('Downloading file:', file);

      if (!file.id) {
        throw new Error('File ID not available');
      }

      // Get presigned URL using file ID (backend will look up minio_path)
      const response = await api.get(`/files/download-url?file_id=${encodeURIComponent(file.id)}`);

      if (!response || !response.presigned_url) {
        throw new Error('Failed to get download URL');
      }

      // Fetch the file from the presigned URL
      const fileResponse = await fetch(response.presigned_url);
      if (!fileResponse.ok) {
        throw new Error('Failed to download file');
      }

      const blob = await fileResponse.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = file.filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      window.toast?.success('Dosya indiriliyor...');
    } catch (error) {
      console.error('Dosya indirme hatası:', error);
      window.toast?.error('Dosya indirilirken hata oluştu');
    }
  }, []);

  return {
    folders,
    files,
    loading,
    error,
    loadContents,
    createFolder,
    deleteFolder,
    deleteFile,
    downloadFile,
  };
};
