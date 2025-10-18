import { useState, useCallback } from 'react';

/**
 * Navigation state management hook
 * Manages separate navigation states for different menu types (home, shared)
 */
export const useNavigation = (selectedMenu = 'home') => {
  const [navigationState, setNavigationState] = useState({
    shared: { currentFolder: null, folderPath: [] },
    home: { currentFolder: null, folderPath: [] }
  });

  // Helper functions to get current navigation state
  const getCurrentNavState = useCallback(() => {
    return navigationState[selectedMenu] || navigationState.home;
  }, [navigationState, selectedMenu]);

  const updateNavState = useCallback((updates) => {
    setNavigationState(prev => ({
      ...prev,
      [selectedMenu]: { ...prev[selectedMenu], ...updates }
    }));
  }, [selectedMenu]);

  // Navigation functions
  const handleFolderOpen = useCallback((folder) => {
    const currentNav = getCurrentNavState();
    const newPath = [...currentNav.folderPath, {
      id: folder.id,
      name: folder.name
    }];
    
    updateNavState({
      currentFolder: folder,
      folderPath: newPath
    });
  }, [getCurrentNavState, updateNavState]);

  const handleBackToRoot = useCallback(() => {
    updateNavState({
      currentFolder: null,
      folderPath: []
    });
  }, [updateNavState]);

  const handleBreadcrumbClick = useCallback((index) => {
    const currentNav = getCurrentNavState();
    
    if (index === -1) {
      // Root'a dön
      handleBackToRoot();
    } else if (index < currentNav.folderPath.length - 1) {
      // Belirli bir klasöre dön
      const newPath = currentNav.folderPath.slice(0, index + 1);
      const targetFolder = newPath[newPath.length - 1];
      
      updateNavState({
        currentFolder: targetFolder,
        folderPath: newPath
      });
    }
  }, [getCurrentNavState, handleBackToRoot, updateNavState]);

  return {
    navigationState,
    getCurrentNavState,
    updateNavState,
    handleFolderOpen,
    handleBackToRoot,
    handleBreadcrumbClick
  };
};
