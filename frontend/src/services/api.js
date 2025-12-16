const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class ApiService {
  constructor() {
    this.token = localStorage.getItem('nimbus_token');
  }

  setToken(token) {
    this.token = token;
    if (token) {
      localStorage.setItem('nimbus_token', token);
    } else {
      localStorage.removeItem('nimbus_token');
    }
  }

  getAuthHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    return headers;
  }

  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    const config = {
      headers: {
        ...this.getAuthHeaders(),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      window.toast?.error('Network error occurred');
      throw error;
    }
  }

  // GET request
  async get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  // POST request
  async post(endpoint, data = {}, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // PUT request
  async put(endpoint, data = {}, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // DELETE request
  async delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }
}

export const api = new ApiService();

// Auth specific methods
export const authApi = {
  // Google OAuth login
  googleLogin: () => {
    window.location.href = `${API_BASE_URL}/auth/google`;
  },

  // Logout
  logout: async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
      window.toast?.error(t('network_error'));
    } finally {
      api.setToken(null);
    }
  },

  // Get user profile
  getProfile: () => api.get('/user/profile'),
};

// File specific methods
export const fileApi = {
  // Get presigned URL for upload
  getUploadPresignedURL: (filename, contentType) => {
    return api.get(
      `/files/upload-url?filename=${encodeURIComponent(filename)}&content_type=${encodeURIComponent(contentType)}`
    );
  },

  // Create file record in MongoDB
  createFile: fileData => {
    return api.post('/files/', fileData);
  },

  // Get presigned URL for download
  getDownloadPresignedURL: filename => {
    return api.get(`/files/download-url?filename=${encodeURIComponent(filename)}`);
  },

  // Get presigned URL for preview
  getPreviewPresignedURL: (fileId, filename) => {
    // Prefer file_id if available (for shared files), fallback to filename
    if (fileId) {
      return api.get(`/files/preview-url?file_id=${encodeURIComponent(fileId)}`);
    }
    return api.get(`/files/preview-url?filename=${encodeURIComponent(filename)}`);
  },

  // List user files
  listFiles: () => {
    return api.get('/files/');
  },

  // Delete file (soft or hard)
  deleteFile: (fileId, permanent = false) => {
    return api.delete(`/files/${fileId}${permanent ? '?permanent=true' : ''}`);
  },

  // Get recent files
  getRecent: () => {
    return api.get('/files/recent');
  },

  // Get starred files
  getStarred: () => {
    return api.get('/files/starred');
  },

  // Get trash files
  getTrash: () => {
    return api.get('/files/trash');
  },

  // Toggle star
  toggleStar: fileId => {
    return api.post(`/files/${fileId}/star`);
  },

  // Restore file
  restoreFile: fileId => {
    return api.post(`/files/${fileId}/restore`);
  },

  // Move file to folder
  moveFile: (fileId, folderId) => {
    return api.post(`/files/${fileId}/move`, { folder_id: folderId });
  },

  // Get OnlyOffice editor config
  getOnlyOfficeConfig: (fileId, mode = 'edit') => {
    return api.get(`/files/onlyoffice-config?file_id=${encodeURIComponent(fileId)}&mode=${mode}`);
  },

  // Get file content as text (for code files)
  getFileContent: async (fileId, cacheBust = false) => {
    const apiInstance = new ApiService();
    let url = `${API_BASE_URL}/files/content?file_id=${encodeURIComponent(fileId)}`;
    if (cacheBust) {
      url += `&_t=${Date.now()}`;
    }
    const headers = apiInstance.getAuthHeaders();
    headers['Accept'] = 'text/plain';
    const response = await fetch(url, { headers });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP error! status: ${response.status}`);
    }
    return await response.text();
  },

  updateFileContent: (fileId, content) => {
    return api.put(`/files/${encodeURIComponent(fileId)}/content`, { content });
  },

  // RAG Document Processing
  processDocument: fileId => {
    return api.post(`/files/${encodeURIComponent(fileId)}/process`, {});
  },

  // Query document with AI
  queryDocument: async (fileId, question) => {
    const response = await fetch(`${API_BASE_URL}/ai/query`, {
      method: 'POST',
      headers: api.getAuthHeaders(),
      body: JSON.stringify({
        file_id: fileId,
        question: question,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Query failed');
    }

    return await response.json();
  },

  // Get conversation history
  getConversationHistory: async fileId => {
    const response = await fetch(
      `${API_BASE_URL}/ai/conversation?file_id=${encodeURIComponent(fileId)}`,
      {
        method: 'GET',
        headers: api.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to fetch conversation history');
    }

    return await response.json();
  },

  // Clear conversation history
  clearConversationHistory: async fileId => {
    const response = await fetch(
      `${API_BASE_URL}/ai/conversation?file_id=${encodeURIComponent(fileId)}`,
      {
        method: 'DELETE',
        headers: api.getAuthHeaders(),
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to clear conversation history');
    }

    return await response.json();
  },
};

// Folder specific methods
export const folderApi = {
  // Create folder
  createFolder: async folderData => {
    const response = await api.post('/folders/', folderData);
    return response.data;
  },

  // Get user folders
  getUserFolders: async () => {
    const response = await api.get('/folders/');
    return response.data;
  },

  // Get root contents (folders + root files)
  getRootContents: async () => {
    const response = await api.get('/folders/root');
    return response;
  },

  // Get storage usage
  getStorageUsage: () => {
    return api.get('/folders/storage');
  },

  // Get folder contents
  getFolderContents: async folderId => {
    if (!folderId) return { files: [], subfolders: [], breadcrumbs: [] };
    const response = await api.get(`/folders/${folderId}`);
    return response;
  },

  // Update folder
  updateFolder: (folderId, updates) => {
    return api.put(`/folders/${folderId}`, updates);
  },

  // Delete folder (soft or hard)
  deleteFolder: async (folderId, permanent = false) => {
    const response = await api.delete(`/folders/${folderId}${permanent ? '?permanent=true' : ''}`);
    return response.data;
  },

  // Get starred folders
  getStarred: () => {
    return api.get('/folders/starred');
  },

  // Get trash folders
  getTrash: () => {
    return api.get('/folders/trash');
  },

  // Toggle star
  toggleStar: folderId => {
    return api.post(`/folders/${folderId}/star`);
  },

  // Restore folder
  restoreFolder: async folderId => {
    const response = await api.post(`/folders/${folderId}/restore`);
    return response.data;
  },

  // Move folder
  moveFolder: async (folderID, targetFolderID) => {
    const response = await api.post(`/folders/${folderID}/move`, { folder_id: targetFolderID });
    return response.data;
  },
};

// Share API (access list management)
export const shareApi = {
  // Get resource shares (access list info)
  getResourceShares: resourceId => {
    return api.get(`/shares/resource/${resourceId}`);
  },

  // Get shared with me
  getSharedWithMe: () => {
    return api.get('/shares/shared-with-me');
  },

  // Get shared folder contents
  getSharedFolderContents: folderId => {
    return api.get(`/shares/shared-folder/${folderId}`);
  },

  // Update access permission
  updateAccessPermission: (resourceId, updateData) => {
    return api.put(`/shares/access/${resourceId}`, updateData);
  },

  // Remove user access
  removeUserAccess: (resourceId, userId) => {
    return api.delete(`/shares/access/${resourceId}/${userId}`);
  },

  // Get resource by public link (automatically adds user to access list with read permission)
  getResourceByPublicLink: publicLink => {
    return api.get(`/shares/public/${publicLink}`);
  },
};

// User API
export const userApi = {
  // Search users
  searchUsers: query => {
    return api.get(`/users/search?q=${encodeURIComponent(query)}`);
  },
};
