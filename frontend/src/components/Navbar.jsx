import {
  AppBar,
  Toolbar,
  Typography,
  Box,
  Container,
  Stack,
  IconButton,
  Menu,
  MenuItem,
} from '@mui/material';
import { motion } from 'framer-motion';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import LanguageIcon from '@mui/icons-material/Language';
import AuthButton from './AuthButton';

const MotionAppBar = motion.create(AppBar);

const Navbar = () => {
  const { t, i18n } = useTranslation();
  const [languageMenu, setLanguageMenu] = useState(null);

  const handleLanguageChange = lng => {
    i18n.changeLanguage(lng);
    setLanguageMenu(null);
    // Toast ile dil değişikliğini bildir
    window.toast?.success(t('language_changed', { lng: lng === 'tr' ? 'Türkçe' : 'English' }));
  };

  return (
    <MotionAppBar
      position="sticky"
      color="default"
      elevation={0}
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
      sx={{
        bgcolor: 'background.paper',
        borderBottom: 1,
        borderColor: 'divider',
        borderBottomLeftRadius: 0,
        borderBottomRightRadius: 0,
      }}
    >
      <Container maxWidth="xl">
        <Toolbar disableGutters sx={{ py: 1 }}>
          <Stack direction="row" spacing={1.5} alignItems="center" sx={{ flexGrow: 1 }}>
            <Box
              sx={{
                width: 56,
                height: 56,
                borderRadius: 2,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                overflow: 'hidden',
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
          </Stack>

          <Stack direction="row" spacing={1} alignItems="center">
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
              onClick={e => e.stopPropagation()}
              PaperProps={{
                sx: {
                  mt: 1,
                  minWidth: 120,
                  borderRadius: 2,
                  zIndex: 1300, // Ensure menu is above other elements
                },
              }}
            >
              <MenuItem
                onClick={e => {
                  e.stopPropagation();
                  handleLanguageChange('tr');
                }}
                selected={i18n.language === 'tr'}
              >
                🇹🇷 Türkçe
              </MenuItem>
              <MenuItem
                onClick={e => {
                  e.stopPropagation();
                  handleLanguageChange('en');
                }}
                selected={i18n.language === 'en'}
              >
                🇺🇸 English
              </MenuItem>
            </Menu>

            <AuthButton />
          </Stack>
        </Toolbar>
      </Container>
    </MotionAppBar>
  );
};

export default Navbar;
