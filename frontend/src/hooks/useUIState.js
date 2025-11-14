import { useState, useCallback } from 'react';

/**
 * UI state management hook
 * Manages view mode, menu states, and other UI-related states
 */
export const useUIState = () => {
  const [viewMode, setViewMode] = useState('grid'); // 'grid' or 'list'
  const [menuAnchor, setMenuAnchor] = useState(null);
  const [selectedItem, setSelectedItem] = useState(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // View mode functions
  const toggleViewMode = useCallback(() => {
    setViewMode(prev => (prev === 'grid' ? 'list' : 'grid'));
  }, []);

  // Context menu functions
  const openContextMenu = useCallback((event, item) => {
    event.preventDefault();
    setMenuAnchor(event.currentTarget);
    setSelectedItem(item);
  }, []);

  const closeContextMenu = useCallback(() => {
    setMenuAnchor(null);
    setSelectedItem(null);
  }, []);

  // Refresh function
  const triggerRefresh = useCallback(() => {
    setRefreshTrigger(prev => prev + 1);
  }, []);

  return {
    // UI states
    viewMode,
    menuAnchor,
    selectedItem,
    refreshTrigger,

    // UI control functions
    setViewMode,
    toggleViewMode,
    openContextMenu,
    closeContextMenu,
    triggerRefresh,
  };
};
