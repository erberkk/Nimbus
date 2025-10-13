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

  // List user files
  listFiles: () => {
    return api.get('/files/');
  },

  // Delete file
  deleteFile: fileId => {
    return api.delete(`/files/${fileId}`);
  },

  // Move file to folder
  moveFile: (fileId, folderId) => {
    return api.post(`/files/${fileId}/move`, { folder_id: folderId });
  },
};

// Folder specific methods
export const folderApi = {
  // Create folder
  createFolder: folderData => {
    return api.post('/folders/', folderData);
  },

  // Get user folders
  getUserFolders: () => {
    return api.get('/folders/');
  },

  // Get root contents (folders + root files)
  getRootContents: () => {
    return api.get('/folders/root');
  },

  // Get storage usage
  getStorageUsage: () => {
    return api.get('/folders/storage');
  },

  // Get folder contents
  getFolderContents: folderId => {
    return api.get(`/folders/${folderId}`);
  },

  // Update folder
  updateFolder: (folderId, updates) => {
    return api.put(`/folders/${folderId}`, updates);
  },

  // Delete folder
  deleteFolder: folderId => {
    return api.delete(`/folders/${folderId}`);
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
