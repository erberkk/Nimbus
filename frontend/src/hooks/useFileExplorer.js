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
      } else if (selectedMenu === 'recent') {
        const response = await fileApi.getRecent();
        if (response && response.files) {
          setFolders([]);
          setFiles(response.files);
        } else {
          setFolders([]);
          setFiles([]);
        }
      } else if (selectedMenu === 'starred') {
        const currentNav = getCurrentNavState();
        const { currentFolder } = currentNav;

        if (currentFolder) {
          // Star'lanmış bir klasörün içindeyiz, sadece star'lanmış içeriği göster
          const response = await folderApi.getFolderContents(currentFolder.id, true);
          if (!response) {
            setFolders([]);
            setFiles([]);
            return;
          }
          setFolders(response.folders || []);
          setFiles(response.files || []);
        } else {
          // Root seviyede, sadece root seviyedeki star'lanmış klasörleri ve dosyaları göster
          const [filesRes, foldersRes] = await Promise.all([
            fileApi.getStarred(),
            folderApi.getStarred()
          ]);

          setFiles(filesRes?.files || []);
          setFolders(foldersRes?.folders || []);
        }
      } else if (selectedMenu === 'trash') {
        const currentNav = getCurrentNavState();
        const { currentFolder } = currentNav;

        if (currentFolder) {
          const response = await folderApi.getFolderContents(currentFolder.id);

          setFiles(response.files || []);
          setFolders(response.subfolders || []);
        } else {
          const [filesRes, foldersRes] = await Promise.all([
            fileApi.getTrash(),
            folderApi.getTrash()
          ]);

          setFiles(filesRes?.files || []);
          setFolders(foldersRes?.folders || []);
        }
      } else {
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

        if (!response) {
          throw new Error('Klasör oluşturma yanıtı alınamadı');
        }

        if (response.folder) {
          window.toast?.success('Klasör başarıyla oluşturuldu');
          loadContents();
          return response.folder;
        } else {
          // Response var ama folder yok - beklenmeyen durum
          console.error('Unexpected response structure:', response);
          throw new Error('Klasör oluşturma yanıtı beklenmeyen formatta');
        }
      } catch (error) {
        console.error('Klasör oluşturma hatası:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Klasör oluşturulurken hata oluştu';
        window.toast?.error(errorMessage);
        throw error;
      }
    },
    [getCurrentNavState, loadContents]
  );

  const deleteFolder = useCallback(
    async (folderId, permanent = false) => {
      try {
        await folderApi.deleteFolder(folderId, permanent);
        window.toast?.success(permanent ? 'Klasör kalıcı olarak silindi' : 'Klasör çöp kutusuna taşındı');
        loadContents();
      } catch (error) {
        console.error('Klasör silme hatası:', error);
        window.toast?.error('Klasör silinirken hata oluştu');
      }
    },
    [loadContents]
  );

  const deleteFile = useCallback(
    async (fileId, permanent = false) => {
      try {
        await fileApi.deleteFile(fileId, permanent);
        window.toast?.success(permanent ? 'Dosya kalıcı olarak silindi' : 'Dosya çöp kutusuna taşındı');
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

  const toggleStar = useCallback(
    async (item, type) => {
      try {
        let newStatus;
        if (type === 'folder') {
          const response = await folderApi.toggleStar(item.id);
          newStatus = response.is_starred;
        } else {
          const response = await fileApi.toggleStar(item.id);
          newStatus = response.is_starred;
        }
        
        // Show success toast
        const itemName = type === 'folder' ? item.name : item.filename;
        if (newStatus) {
          window.toast?.success(`${itemName} yıldızlara eklendi`);
        } else {
          window.toast?.success(`${itemName} yıldızlardan çıkarıldı`);
        }
        
        loadContents();
      } catch (error) {
        console.error('Yıldızlama hatası:', error);
        window.toast?.error('İşlem başarısız');
      }
    },
    [loadContents]
  );

  const restoreItem = useCallback(
    async (item, type) => {
      try {
        if (type === 'folder') {
          await folderApi.restoreFolder(item.id);
        } else {
          await fileApi.restoreFile(item.id);
        }
        window.toast?.success('Öğe geri yüklendi');
        loadContents();
      } catch (error) {
        console.error('Geri yükleme hatası:', error);
        window.toast?.error('Geri yükleme başarısız');
      }
    },
    [loadContents]
  );

  const permanentDelete = useCallback(
    async (item, type) => {
      if (type === 'folder') {
        await deleteFolder(item.id, true);
      } else {
        await deleteFile(item.id, true);
      }
    },
    [deleteFile, deleteFolder]
  );

  const moveItem = useCallback(
    async (item, targetFolderId, type) => {
      try {
        if (type === 'folder') {
          await folderApi.moveFolder(item.id, targetFolderId);
        } else {
          await fileApi.moveFile(item.id, targetFolderId);
        }
        window.toast?.success('Öğe taşındı');
        loadContents();
      } catch (error) {
        console.error('Taşıma hatası:', error);
        window.toast?.error('Taşıma işlemi başarısız: ' + (error.response?.data?.error || error.message));
        throw error;
      }
    },
    [loadContents]
  );

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
    toggleStar,
    restoreItem,
    permanentDelete,
    moveItem,
  };
};
