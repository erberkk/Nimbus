import { Box, Card, CardContent, Typography, Button, Container, Stack } from '@mui/material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useEffect } from 'react';
import GoogleIcon from '@mui/icons-material/Google';
import SecurityIcon from '@mui/icons-material/Security';
import SpeedIcon from '@mui/icons-material/Speed';
import { authApi } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

const MotionCard = motion.create(Card);
const MotionBox = motion.create(Box);

const LoginPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isAuthenticated, loading } = useAuth();

  // Zaten login olduysa dashboard'a yönlendir
  useEffect(() => {
    if (!loading && isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, loading, navigate]);

  const handleGoogleLogin = () => {
    authApi.googleLogin();
  };

  const features = [
    {
      icon: (
        <img
          src="/nimbus_logo.png"
          alt="Cloud Storage"
          style={{ width: 56, height: 56, objectFit: 'contain' }}
        />
      ),
      title: t('cloud_storage'),
      description: t('cloud_storage_desc'),
    },
    {
      icon: <SecurityIcon sx={{ fontSize: 40, color: '#FBBF24' }} />,
      title: t('secure'),
      description: t('secure_desc'),
    },
    {
      icon: <SpeedIcon sx={{ fontSize: 40, color: '#FBBF24' }} />,
      title: t('fast_access'),
      description: t('fast_access_desc'),
    },
  ];

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Animated background elements */}
      <MotionBox
        sx={{
          position: 'absolute',
          width: 400,
          height: 400,
          borderRadius: '50%',
          background: 'rgba(255,255,255,0.1)',
          top: -100,
          right: -100,
        }}
        animate={{
          scale: [1, 1.2, 1],
          rotate: [0, 180, 360],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: 'linear',
        }}
      />

      <MotionBox
        sx={{
          position: 'absolute',
          width: 300,
          height: 300,
          borderRadius: '50%',
          background: 'rgba(255,255,255,0.08)',
          bottom: -50,
          left: -50,
        }}
        animate={{
          scale: [1, 1.3, 1],
          rotate: [360, 180, 0],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          ease: 'linear',
        }}
      />

      <Container maxWidth="lg">
        <Stack
          direction={{ xs: 'column', md: 'row' }}
          spacing={4}
          alignItems="center"
          justifyContent="center"
        >
          {/* Left side - Branding */}
          <MotionBox
            initial={{ opacity: 0, x: -50 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6 }}
            sx={{ flex: 1, color: 'white', textAlign: { xs: 'center', md: 'left' } }}
          >
            <Stack spacing={3}>
              <Box>
                <Typography variant="h2" fontWeight={700} gutterBottom>
                  Nimbus
                </Typography>
                <Typography variant="h5" sx={{ opacity: 0.95 }}>
                  Modern Bulut Depolama Platformu
                </Typography>
              </Box>

              <Typography variant="body1" sx={{ opacity: 0.9, maxWidth: 500 }}>
                Dosyalarınızı güvenle saklayın, kolayca paylaşın ve her yerden erişin. Dropbox
                benzeri modern bir deneyim.
              </Typography>

              <Stack spacing={2} sx={{ mt: 4 }}>
                {features.map((feature, index) => (
                  <MotionBox
                    key={index}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.5, delay: 0.2 + index * 0.1 }}
                  >
                    <Stack direction="row" spacing={2} alignItems="center">
                      {feature.icon}
                      <Box>
                        <Typography variant="h6" fontWeight={600}>
                          {feature.title}
                        </Typography>
                        <Typography variant="body2" sx={{ opacity: 0.8 }}>
                          {feature.description}
                        </Typography>
                      </Box>
                    </Stack>
                  </MotionBox>
                ))}
              </Stack>
            </Stack>
          </MotionBox>

          {/* Right side - Login Card */}
          <MotionCard
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.3 }}
            sx={{
              width: { xs: '100%', sm: 450 },
              p: 2,
            }}
          >
            <CardContent sx={{ p: 4 }}>
              <Stack spacing={3} alignItems="center">
                <Box
                  sx={{
                    width: 120,
                    height: 120,
                    borderRadius: '50%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    mb: 2,
                    overflow: 'hidden',
                    p: 1,
                  }}
                >
                  <img
                    src="/nimbus_logo.png"
                    alt="Nimbus Logo"
                    style={{
                      width: '100%',
                      height: '100%',
                      objectFit: 'contain',
                    }}
                  />
                </Box>

                <Box textAlign="center">
                  <Typography variant="h4" fontWeight={700} gutterBottom>
                    Hoş Geldiniz
                  </Typography>
                  <Typography variant="body1" color="text.secondary">
                    Devam etmek için Google hesabınızla giriş yapın
                  </Typography>
                </Box>

                <Button
                  variant="contained"
                  size="large"
                  fullWidth
                  startIcon={<GoogleIcon />}
                  onClick={handleGoogleLogin}
                  sx={{
                    mt: 2,
                    py: 1.5,
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    '&:hover': {
                      background: 'linear-gradient(135deg, #5568d3 0%, #6a3f8f 100%)',
                    },
                  }}
                >
                  Google ile Giriş Yap
                </Button>

                <Stack
                  direction="row"
                  spacing={1}
                  sx={{ mt: 3, flexWrap: 'wrap', justifyContent: 'center' }}
                >
                  <Typography variant="body2" color="text.secondary">
                    Giriş yaparak
                  </Typography>
                  <Typography
                    variant="body2"
                    color="primary"
                    sx={{ cursor: 'pointer', fontWeight: 600 }}
                  >
                    Kullanım Koşulları
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    ve
                  </Typography>
                  <Typography
                    variant="body2"
                    color="primary"
                    sx={{ cursor: 'pointer', fontWeight: 600 }}
                  >
                    Gizlilik Politikası
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    kabul etmiş olursunuz.
                  </Typography>
                </Stack>
              </Stack>
            </CardContent>
          </MotionCard>
        </Stack>
      </Container>
    </Box>
  );
};

export default LoginPage;
