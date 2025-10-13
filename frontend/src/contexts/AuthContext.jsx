import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { api, authApi } from '../services/api';
import { useNavigate } from 'react-router-dom';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [token, setTokenState] = useState(null);
  const navigate = useNavigate();

  const getTokenFromUrl = useCallback(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const tokenFromUrl = urlParams.get('token');

    if (tokenFromUrl) {
      const newUrl = window.location.origin + window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);
      return tokenFromUrl;
    }
    return null;
  }, []);

  const loadProfile = useCallback(async () => {
    try {
      const response = await authApi.getProfile();
      setUser(response.user);
      return response.user;
    } catch (error) {
      console.error('Profil yükleme hatası:', error);
      api.setToken(null);
      setTokenState(null);
      setUser(null);
      return null;
    }
  }, []);

  const login = useCallback(() => {
    authApi.googleLogin();
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout hatası:', error);
    } finally {
      setUser(null);
      setTokenState(null);
      api.setToken(null);
      // Landing page'e yönlendir
      navigate('/', { replace: true });
    }
  }, [navigate]);

  useEffect(() => {
    const initAuth = async () => {
      setLoading(true);

      const tokenFromUrl = getTokenFromUrl();
      let currentToken = tokenFromUrl || localStorage.getItem('nimbus_token');

      if (currentToken) {
        api.setToken(currentToken);
        setTokenState(currentToken);
        await loadProfile();
      }

      setLoading(false);
    };

    initAuth();
  }, []);

  const value = {
    user,
    loading,
    isAuthenticated: !!user,
    login,
    logout,
    token,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
