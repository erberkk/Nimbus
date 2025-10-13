import {
  Box,
  Container,
  Typography,
  Button,
  Grid,
  AppBar,
  Toolbar,
  Link as MuiLink,
  IconButton,
  Menu,
  MenuItem,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import CloudIcon from '@mui/icons-material/Cloud';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ShareIcon from '@mui/icons-material/Share';
import SecurityIcon from '@mui/icons-material/Security';
import LanguageIcon from '@mui/icons-material/Language';

const MotionBox = motion.create(Box);
const MotionButton = motion.create(Button);

const LandingPage = () => {
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const [languageMenu, setLanguageMenu] = useState(null);
  const { isAuthenticated, loading } = useAuth();

  // Login olduysa otomatik dashboard'a yönlendir
  useEffect(() => {
    if (!loading && isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, loading, navigate]);

  const handleLoginClick = () => {
    if (isAuthenticated) {
      navigate('/dashboard');
    } else {
      navigate('/login');
    }
  };

  const handleLanguageChange = lng => {
    i18n.changeLanguage(lng);
    setLanguageMenu(null);
    window.toast?.success(t('language_changed', { lng: lng === 'tr' ? 'Türkçe' : 'English' }));
  };

  const scrollToSection = sectionId => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  };

  const features = [
    {
      icon: <CloudUploadIcon sx={{ fontSize: 32 }} />,
      title: t('landing.feature1_title') || 'Easy Upload',
      description:
        t('landing.feature1_desc') ||
        'Drag and drop files or folders. Upload multiple files at once with our intuitive interface.',
      color: '#667eea',
    },
    {
      icon: <ShareIcon sx={{ fontSize: 32 }} />,
      title: t('landing.feature2_title') || 'Smart Sharing',
      description:
        t('landing.feature2_desc') ||
        'Generate secure links to share files and folders. Control permissions and access levels.',
      color: '#4facfe',
    },
    {
      icon: <SecurityIcon sx={{ fontSize: 32 }} />,
      title: t('landing.feature3_title') || 'Secure Storage',
      description:
        t('landing.feature3_desc') ||
        'End-to-end encryption keeps your files safe. Your data is protected with enterprise-grade security.',
      color: '#43e97b',
    },
  ];

  const footerSections = [
    {
      title: t('landing.footer_product') || 'Product',
      links: ['Features', 'Pricing', 'Security'],
    },
    {
      title: t('landing.footer_company') || 'Company',
      links: ['About', 'Blog', 'Careers'],
    },
    {
      title: t('landing.footer_support') || 'Support',
      links: ['Help Center', 'Contact', 'API Docs'],
    },
  ];

  return (
    <Box sx={{ bgcolor: 'background.default' }}>
      {/* Navigation */}
      <AppBar
        position="fixed"
        elevation={0}
        sx={{
          bgcolor: 'background.paper',
          borderBottom: 1,
          borderColor: 'divider',
          borderRadius: 0,
          transition: 'all 0.3s ease',
        }}
      >
        <Container maxWidth="xl">
          <Toolbar disableGutters sx={{ py: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', flexGrow: 1 }}>
              <Box
                sx={{
                  width: 40,
                  height: 40,
                  borderRadius: 2,
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  mr: 1.5,
                }}
              >
                <CloudIcon sx={{ color: 'white', fontSize: 24 }} />
              </Box>
              <Typography
                variant="h5"
                fontWeight={700}
                sx={{
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  backgroundClip: 'text',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}
              >
                {t('nimbus')}
              </Typography>
            </Box>

            <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 2, alignItems: 'center' }}>
              <MuiLink
                component="button"
                onClick={() => scrollToSection('features')}
                sx={{
                  color: 'text.secondary',
                  textDecoration: 'none',
                  fontWeight: 500,
                  cursor: 'pointer',
                  '&:hover': { color: 'primary.main' },
                  transition: 'color 0.3s',
                }}
              >
                {t('landing.nav_features') || 'Features'}
              </MuiLink>
              <MuiLink
                component="button"
                onClick={() => scrollToSection('cta')}
                sx={{
                  color: 'text.secondary',
                  textDecoration: 'none',
                  fontWeight: 500,
                  cursor: 'pointer',
                  '&:hover': { color: 'primary.main' },
                  transition: 'color 0.3s',
                }}
              >
                {t('landing.nav_about') || 'About'}
              </MuiLink>

              {/* Language Selector */}
              <IconButton
                onClick={e => setLanguageMenu(e.currentTarget)}
                sx={{
                  border: 1,
                  borderColor: 'divider',
                  borderRadius: 2,
                  p: 1,
                }}
              >
                <LanguageIcon />
              </IconButton>

              <Menu
                anchorEl={languageMenu}
                open={Boolean(languageMenu)}
                onClose={() => setLanguageMenu(null)}
                PaperProps={{
                  sx: {
                    mt: 1,
                    minWidth: 120,
                    borderRadius: 2,
                  },
                }}
              >
                <MenuItem
                  onClick={() => handleLanguageChange('tr')}
                  selected={i18n.language === 'tr'}
                >
                  🇹🇷 Türkçe
                </MenuItem>
                <MenuItem
                  onClick={() => handleLanguageChange('en')}
                  selected={i18n.language === 'en'}
                >
                  🇺🇸 English
                </MenuItem>
              </Menu>

              <Button
                variant="contained"
                onClick={handleLoginClick}
                sx={{
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  borderRadius: 2,
                  px: 3,
                  py: 1,
                  fontWeight: 600,
                  textTransform: 'none',
                  '&:hover': {
                    background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
                    transform: 'translateY(-2px)',
                    boxShadow: '0 10px 20px rgba(102, 126, 234, 0.3)',
                  },
                  transition: 'all 0.3s ease',
                }}
              >
                {isAuthenticated ? t('dashboard') || 'Dashboard' : t('login')}
              </Button>
            </Box>
          </Toolbar>
        </Container>
      </AppBar>

      {/* Hero Section */}
      <Box
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          minHeight: '100vh',
          pt: 16,
          pb: 8,
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Floating Clouds */}
        <MotionBox
          animate={{
            y: [0, -20, 0],
          }}
          transition={{
            duration: 6,
            repeat: Infinity,
            ease: 'easeInOut',
          }}
          sx={{
            position: 'absolute',
            top: '20%',
            left: '10%',
            opacity: 0.2,
          }}
        >
          <CloudIcon sx={{ fontSize: 64, color: 'white' }} />
        </MotionBox>
        <MotionBox
          animate={{
            y: [0, -20, 0],
          }}
          transition={{
            duration: 6,
            repeat: Infinity,
            ease: 'easeInOut',
            delay: -2,
          }}
          sx={{
            position: 'absolute',
            top: '30%',
            right: '15%',
            opacity: 0.15,
          }}
        >
          <CloudIcon sx={{ fontSize: 80, color: 'white' }} />
        </MotionBox>
        <MotionBox
          animate={{
            y: [0, -20, 0],
          }}
          transition={{
            duration: 6,
            repeat: Infinity,
            ease: 'easeInOut',
            delay: -4,
          }}
          sx={{
            position: 'absolute',
            bottom: '20%',
            left: '25%',
            opacity: 0.25,
          }}
        >
          <CloudIcon sx={{ fontSize: 48, color: 'white' }} />
        </MotionBox>

        <Container
          maxWidth="lg"
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: 'calc(100vh - 64px)',
          }}
        >
          <MotionBox
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            sx={{
              textAlign: 'center',
              position: 'relative',
              zIndex: 1,
              maxWidth: '900px',
              mx: 'auto',
            }}
          >
            <Typography
              variant="h1"
              sx={{
                fontSize: { xs: '2.5rem', sm: '3.5rem', md: '4.5rem', lg: '5.5rem' },
                fontWeight: 700,
                color: 'white',
                mb: 4,
                lineHeight: 1.2,
              }}
            >
              {t('landing.hero_title1') || 'Your Files,'} <br />
              <Box
                component="span"
                sx={{
                  color: '#FBBF24',
                }}
              >
                {t('landing.hero_title2') || 'Anywhere'}
              </Box>
            </Typography>
            <Typography
              variant="h5"
              sx={{
                color: 'rgba(255, 255, 255, 0.9)',
                mb: 6,
                maxWidth: '700px',
                mx: 'auto',
                lineHeight: 1.6,
                fontSize: { xs: '1.1rem', sm: '1.3rem', md: '1.5rem' },
              }}
            >
              {t('landing.hero_subtitle') ||
                'Store, sync, and share your files securely in the cloud. Access your data from any device, anywhere in the world.'}
            </Typography>
            <MotionButton
              variant="contained"
              size="large"
              onClick={handleLoginClick}
              whileHover={{ scale: 1.05, y: -2 }}
              whileTap={{ scale: 0.95 }}
              sx={{
                bgcolor: 'white',
                color: '#667eea',
                px: 6,
                py: 2.5,
                fontSize: { xs: '1rem', sm: '1.1rem' },
                fontWeight: 600,
                borderRadius: 2,
                textTransform: 'none',
                boxShadow: '0 10px 30px rgba(0,0,0,0.2)',
                '&:hover': {
                  bgcolor: 'rgba(255, 255, 255, 0.95)',
                  boxShadow: '0 15px 40px rgba(0,0,0,0.3)',
                },
              }}
            >
              {isAuthenticated
                ? t('landing.cta_dashboard') || 'Go to Dashboard'
                : t('landing.cta_start') || 'Get Started Free'}
            </MotionButton>
          </MotionBox>
        </Container>
      </Box>

      {/* Features Section */}
      <Box id="features" sx={{ py: 12, bgcolor: 'grey.50' }}>
        <Container
          maxWidth="lg"
          sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}
        >
          <MotionBox
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            sx={{ textAlign: 'center', mb: 8, maxWidth: '800px' }}
          >
            <Typography
              variant="h2"
              fontWeight={700}
              sx={{ mb: 3, fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' } }}
            >
              {t('landing.features_title') || 'Powerful Features'}
            </Typography>
            <Typography
              variant="h6"
              color="text.secondary"
              sx={{ fontSize: { xs: '1rem', sm: '1.1rem' } }}
            >
              {t('landing.features_subtitle') || 'Everything you need for modern cloud storage'}
            </Typography>
          </MotionBox>

          <Grid
            container
            spacing={6}
            sx={{ maxWidth: '1100px', width: '100%', justifyContent: 'center' }}
          >
            {features.map((feature, index) => (
              <Grid item xs={4} key={index}>
                <MotionBox
                  initial={{ opacity: 0, y: 30 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.6, delay: index * 0.1 }}
                  whileHover={{
                    y: -3,
                    transition: { duration: 0.2, ease: 'easeOut' },
                  }}
                  sx={{
                    bgcolor: 'white',
                    p: { xs: 1.5, sm: 2 },
                    borderRadius: 3,
                    boxShadow: '0 4px 20px rgba(0,0,0,0.08)',
                    height: '220px',
                    width: '100%',
                    maxWidth: '260px',
                    transition: 'box-shadow 0.3s ease',
                    '&:hover': {
                      boxShadow: '0 10px 40px rgba(0,0,0,0.12)',
                    },
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    textAlign: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Box
                    sx={{
                      width: 40,
                      height: 40,
                      borderRadius: 2,
                      bgcolor: `${feature.color}15`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mb: 1.5,
                      color: feature.color,
                    }}
                  >
                    {feature.icon}
                  </Box>
                  <Typography
                    variant="h6"
                    fontWeight={600}
                    sx={{ mb: 1.5, fontSize: { xs: '1rem', sm: '1.1rem' } }}
                  >
                    {feature.title}
                  </Typography>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ lineHeight: 1.5, fontSize: { xs: '0.8rem', sm: '0.85rem' } }}
                  >
                    {feature.description}
                  </Typography>
                </MotionBox>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* CTA Section */}
      <Box
        id="cta"
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          py: 12,
        }}
      >
        <Container
          maxWidth="md"
          sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}
        >
          <MotionBox
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            sx={{ textAlign: 'center', maxWidth: '600px' }}
          >
            <Typography
              variant="h2"
              fontWeight={700}
              sx={{ color: 'white', mb: 3, fontSize: { xs: '2rem', sm: '2.5rem', md: '3rem' } }}
            >
              {t('landing.cta_title') || 'Ready to Get Started?'}
            </Typography>
            <Typography
              variant="h6"
              sx={{
                color: 'rgba(255, 255, 255, 0.9)',
                mb: 6,
                fontSize: { xs: '1rem', sm: '1.1rem', md: '1.25rem' },
              }}
            >
              {t('landing.cta_subtitle') ||
                'Join thousands of users who trust Nimbus with their files'}
            </Typography>
            <MotionButton
              variant="contained"
              size="large"
              onClick={handleLoginClick}
              whileHover={{ scale: 1.05, y: -2 }}
              whileTap={{ scale: 0.95 }}
              sx={{
                bgcolor: 'white',
                color: '#667eea',
                px: 6,
                py: 2.5,
                fontSize: { xs: '1rem', sm: '1.1rem' },
                fontWeight: 600,
                borderRadius: 2,
                textTransform: 'none',
                boxShadow: '0 10px 30px rgba(0,0,0,0.2)',
                '&:hover': {
                  bgcolor: 'rgba(255, 255, 255, 0.95)',
                  boxShadow: '0 15px 40px rgba(0,0,0,0.3)',
                },
              }}
            >
              {isAuthenticated
                ? t('landing.cta_dashboard') || 'Go to Dashboard'
                : t('landing.cta_button') || 'Start Your Free Trial'}
            </MotionButton>
          </MotionBox>
        </Container>
      </Box>

      {/* Footer */}
      <Box sx={{ bgcolor: 'grey.900', color: 'white', py: 6 }}>
        <Container maxWidth="lg" sx={{ display: 'flex', justifyContent: 'center' }}>
          {/* ANA DEĞİŞİKLİK 1: justifyContent: 'space-between' ekledik.
            Bu, içindeki iki ana grubu (Nimbus ve Linkler Grubu) birbirinden ayıracak.
          */}
          <Grid
            container
            spacing={4}
            sx={{ maxWidth: '1000px', width: '100%', justifyContent: 'space-between' }}
          >
            {/* GRUP 1: Nimbus Sütunu (Bu kısım aynı kalıyor) */}
            <Grid item xs={12} md="auto">
              {' '}
              {/* 'md={2}' yerine 'md="auto"' daha esnek olabilir */}
              <MotionBox
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6 }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Box
                    sx={{
                      width: 32,
                      height: 32,
                      borderRadius: 1.5,
                      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mr: 1.5,
                    }}
                  >
                    <CloudIcon sx={{ color: 'white', fontSize: 20 }} />
                  </Box>
                  <Typography variant="h6" fontWeight={700}>
                    {t('nimbus')}
                  </Typography>
                </Box>
                <Typography
                  variant="body2"
                  sx={{
                    color: 'grey.400',
                    lineHeight: 1.5,
                    fontSize: { xs: '0.8rem', sm: '0.85rem' },
                    wordBreak: 'break-word',
                    maxWidth: '200px',
                    textAlign: 'left',
                  }}
                >
                  {t('landing.footer_desc') ||
                    'Your trusted cloud storage solution for modern teams and individuals.'}
                </Typography>
              </MotionBox>
            </Grid>

            {/*
              GEREKSİZ BOŞLUK SÜTUNU SİLİNDİ: 
              <Grid item xs={false} md={3} sx={{ display: { xs: 'none', md: 'block' } }}></Grid>
              Bu sütuna artık ihtiyacımız yok çünkü justifyContent işimizi görüyor.
            */}

            {/* GRUP 2: Link Sütunları Grubu */}
            <Grid item xs={12} md="auto">
              {/* Bu iç container, link sütunlarını bir arada tutar */}
              <Grid container spacing={4}>
                {footerSections.map((section, index) => (
                  <Grid item xs={12} sm={4} md="auto" key={index}>
                    {' '}
                    {/* md={2.33} yerine 'auto' daha iyi çalışır */}
                    <MotionBox
                      initial={{ opacity: 0, y: 20 }}
                      whileInView={{ opacity: 1, y: 0 }}
                      viewport={{ once: true }}
                      transition={{ duration: 0.6, delay: index * 0.1 }}
                    >
                      <Typography
                        variant="subtitle1"
                        fontWeight={600}
                        sx={{ mb: 2, fontSize: { xs: '0.95rem', sm: '1rem' } }}
                      >
                        {section.title}
                      </Typography>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                        {section.links.map((link, linkIndex) => (
                          <MuiLink
                            key={linkIndex}
                            component="button"
                            sx={{
                              color: 'grey.400',
                              textDecoration: 'none',
                              cursor: 'pointer',
                              fontSize: { xs: '0.8rem', sm: '0.85rem' },
                              textAlign: 'left',
                              '&:hover': { color: 'white' },
                              transition: 'color 0.3s',
                            }}
                          >
                            {link}
                          </MuiLink>
                        ))}
                      </Box>
                    </MotionBox>
                  </Grid>
                ))}
              </Grid>
            </Grid>
          </Grid>
        </Container>

        <Box sx={{ borderTop: 1, borderColor: 'grey.800', mt: 4, pt: 3, textAlign: 'center' }}>
          <Container maxWidth="lg">
            <Typography
              variant="body2"
              sx={{ color: 'grey.500', fontSize: { xs: '0.75rem', sm: '0.8rem' } }}
            >
              © 2025 Nimbus. {t('landing.footer_rights') || 'All rights reserved.'}
            </Typography>
          </Container>
        </Box>
      </Box>
    </Box>
  );
};

export default LandingPage;
