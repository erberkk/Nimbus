import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Alert,
  Chip,
} from '@mui/material';
import { useAuth } from '../contexts/AuthContext';
import { shareApi } from '../services/api';
import FolderIcon from '@mui/icons-material/Folder';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import PersonIcon from '@mui/icons-material/Person';

const SharePage = () => {
  const { publicLink } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated, user } = useAuth();
  const [resource, setResource] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!publicLink) {
      setError('Geçersiz paylaşım bağlantısı');
      setLoading(false);
      return;
    }

    if (!isAuthenticated) {
      // Kullanıcı giriş yapmamışsa login sayfasına yönlendir
      navigate('/login', {
        state: {
          from: `/share/${publicLink}`,
          message: 'Bu dosyayı görüntülemek için giriş yapmanız gerekiyor.',
        },
      });
      return;
    }

    loadResource();
  }, [publicLink, isAuthenticated, navigate]);

  const loadResource = async () => {
    try {
      setLoading(true);
      const data = await shareApi.getResourceByPublicLink(publicLink);
      setResource(data);
    } catch (error) {
      console.error('Failed to load shared resource:', error);
      setError('Paylaşılan kaynak yüklenemedi veya erişim izniniz yok');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          minHeight: '100vh',
          bgcolor: 'background.default',
        }}
      >
        <CircularProgress size={60} />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ p: 3, maxWidth: 600, mx: 'auto', mt: 4 }}>
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
        <Button variant="contained" onClick={() => navigate('/dashboard')}>
          Ana Sayfaya Dön
        </Button>
      </Box>
    );
  }

  if (!resource) {
    return (
      <Box sx={{ p: 3, maxWidth: 600, mx: 'auto', mt: 4 }}>
        <Alert severity="warning">Kaynak bulunamadı veya erişim izniniz yok</Alert>
        <Button variant="contained" onClick={() => navigate('/dashboard')} sx={{ mt: 2 }}>
          Ana Sayfaya Dön
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3, maxWidth: 800, mx: 'auto' }}>
      <Typography variant="h4" gutterBottom>
        Paylaşılan {resource.resource_type === 'folder' ? 'Klasör' : 'Dosya'}
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            {resource.resource_type === 'folder' ? (
              <FolderIcon sx={{ fontSize: 40, mr: 2, color: 'primary.main' }} />
            ) : (
              <InsertDriveFileIcon sx={{ fontSize: 40, mr: 2, color: 'secondary.main' }} />
            )}
            <Box sx={{ flex: 1 }}>
              <Typography variant="h5" gutterBottom>
                {resource.resource.name || resource.resource.filename}
              </Typography>
              {resource.resource_type === 'file' && (
                <Typography variant="body2" color="text.secondary">
                  Boyut: {(resource.resource.size / 1024 / 1024).toFixed(2)} MB
                </Typography>
              )}
            </Box>
          </Box>

          <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
            <Chip
              icon={<PersonIcon />}
              label={`Sahibi: ${user?.name || 'Bilinmiyor'}`}
              variant="outlined"
            />
            <Chip
              label={`Erişim: ${resource.resource.access_list?.find(a => a.user_id === user?.id)?.access_type === 'write' ? 'Düzenleme' : 'Görüntüleme'}`}
              color="primary"
              variant="outlined"
            />
          </Box>

          <Typography variant="body2" color="text.secondary">
            Bu {resource.resource_type === 'folder' ? 'klasör' : 'dosya'} sizinle paylaşıldı.
            İçeriği görüntülemek için "Görüntüle" butonuna tıklayın.
          </Typography>

          <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
            <Button
              variant="contained"
              onClick={() =>
                navigate(
                  `/dashboard?resource=${resource.resource.id}&type=${resource.resource_type}`
                )
              }
            >
              Görüntüle
            </Button>
            <Button variant="outlined" onClick={() => navigate('/dashboard')}>
              Ana Sayfaya Dön
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

export default SharePage;
