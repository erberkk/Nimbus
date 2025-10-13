import { useState, useEffect, useCallback } from 'react';
import { api, authApi } from '../services/api';

export const useAuth = () => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [token, setTokenState] = useState(null);

  // Token'ı URL'den al (OAuth callback için)
  const getTokenFromUrl = useCallback(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const tokenFromUrl = urlParams.get('token');

    if (tokenFromUrl) {
      // URL'yi temizle
      const newUrl = window.location.origin + window.location.pathname;
      window.history.replaceState({}, document.title, newUrl);

      return tokenFromUrl;
    }
    return null;
  }, []);

  // Kullanıcı profilini yükle
  const loadProfile = useCallback(async authToken => {
    try {
      const response = await authApi.getProfile();
      setUser(response.user);
      return response.user;
    } catch (error) {
      console.error('Profil yükleme hatası:', error);
      window.toast?.error(t('unauthorized'));
      // Token geçersiz, temizle
      api.setToken(null);
      setTokenState(null);
      return null;
    }
  }, []);

  // Login işlemi
  const login = useCallback(() => {
    authApi.googleLogin();
  }, []);

  // Logout işlemi
  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout hatası:', error);
      window.toast?.error(t('network_error'));
    } finally {
      setUser(null);
      setTokenState(null);
      api.setToken(null);
    }
  }, []);

  // İlk yükleme ve token kontrolü
  useEffect(() => {
    const initAuth = async () => {
      setLoading(true);

      // URL'den token al
      const tokenFromUrl = getTokenFromUrl();

      // LocalStorage veya URL'den token al
      let currentToken = tokenFromUrl || localStorage.getItem('nimbus_token');

      if (currentToken) {
        api.setToken(currentToken);
        setTokenState(currentToken);

        // Kullanıcı profilini yükle
        await loadProfile(currentToken);
      }

      setLoading(false);
    };

    initAuth();
  }, [getTokenFromUrl, loadProfile]);

  return {
    user,
    loading,
    isAuthenticated: !!user,
    login,
    logout,
    token,
  };
};
