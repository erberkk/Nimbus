import { useState, useCallback } from 'react';

/**
 * Navigation state management hook
 * Manages separate navigation states for different menu types (home, shared)
 */
export const useNavigation = (selectedMenu = 'home') => {
  const [navigationState, setNavigationState] = useState({
    home: { currentFolder: null, folderPath: [] },
    shared: { currentFolder: null, folderPath: [] },
    trash: { currentFolder: null, folderPath: [] },
    recent: { currentFolder: null, folderPath: [] },
    starred: { currentFolder: null, folderPath: [] },
  });

  // Helper functions to get current navigation state
  const getCurrentNavState = useCallback(() => {
    return navigationState[selectedMenu] || navigationState.home;
  }, [navigationState, selectedMenu]);

  const updateNavState = useCallback(
    updates => {
      setNavigationState(prev => ({
        ...prev,
        [selectedMenu]: { ...prev[selectedMenu], ...updates },
      }));
    },
    [selectedMenu]
  );

  // Navigation functions
  const handleFolderOpen = useCallback(
    folder => {
      const currentNav = getCurrentNavState();
      const newPath = [
        ...currentNav.folderPath,
        {
          id: folder.id,
          name: folder.name,
        },
      ];

      updateNavState({
        currentFolder: folder,
        folderPath: newPath,
      });
    },
    [getCurrentNavState, updateNavState]
  );

  const handleBackToRoot = useCallback(() => {
    updateNavState({
      currentFolder: null,
      folderPath: [],
    });
  }, [updateNavState]);

  const handleBreadcrumbClick = useCallback(
    index => {
      const currentNav = getCurrentNavState();

      if (index === -1) {
        handleBackToRoot();
      } else if (index >= 0 && index < currentNav.folderPath.length) {
        const targetFolder = currentNav.folderPath[index];
        const newPath = currentNav.folderPath.slice(0, index + 1);

        updateNavState({
          currentFolder: targetFolder,
          folderPath: newPath,
        });
      }
    },
    [getCurrentNavState, handleBackToRoot, updateNavState]
  );

  return {
    navigationState,
    getCurrentNavState,
    updateNavState,
    handleFolderOpen,
    handleBackToRoot,
    handleBreadcrumbClick,
  };
};
