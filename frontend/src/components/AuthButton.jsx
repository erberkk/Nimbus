import {
  Stack,
  Avatar,
  Typography,
  Button,
  CircularProgress,
  IconButton,
  Menu,
  MenuItem,
} from '@mui/material';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../contexts/AuthContext';
import LogoutIcon from '@mui/icons-material/Logout';
import PersonIcon from '@mui/icons-material/Person';
import SettingsIcon from '@mui/icons-material/Settings';
import { motion } from 'framer-motion';

const MotionButton = motion.create(Button);

const AuthButton = () => {
  const { t } = useTranslation();
  const { user, loading, isAuthenticated, logout } = useAuth();
  const [anchorEl, setAnchorEl] = useState(null);

  const handleMenuOpen = event => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    handleMenuClose();
    logout();
  };

  if (loading) {
    return (
      <Stack direction="row" spacing={2} alignItems="center">
        <CircularProgress size={24} />
      </Stack>
    );
  }

  if (isAuthenticated && user) {
    return (
      <>
        <Stack
          direction="row"
          spacing={2}
          alignItems="center"
          component={motion.div}
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.3 }}
        >
          <Stack direction="row" spacing={1.5} alignItems="center">
            <Avatar
              src={user.avatar}
              alt={user.name}
              sx={{
                width: 40,
                height: 40,
                cursor: 'pointer',
                border: '2px solid',
                borderColor: 'primary.main',
              }}
              onClick={handleMenuOpen}
            />
            <Typography
              variant="body1"
              fontWeight={600}
              sx={{ display: { xs: 'none', sm: 'block' } }}
            >
              {user.name}
            </Typography>
          </Stack>

          <MotionButton
            variant="outlined"
            color="error"
            size="small"
            startIcon={<LogoutIcon />}
            onClick={handleLogout}
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            sx={{
              borderRadius: 2,
              display: { xs: 'none', md: 'flex' },
            }}
          >
            {t('logout')}
          </MotionButton>

          <IconButton
            color="error"
            onClick={handleLogout}
            sx={{ display: { xs: 'flex', md: 'none' } }}
          >
            <LogoutIcon />
          </IconButton>
        </Stack>

        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={handleMenuClose}
          PaperProps={{
            sx: {
              mt: 1.5,
              minWidth: 200,
              borderRadius: 2,
            },
          }}
        >
          <MenuItem onClick={handleMenuClose}>
            <PersonIcon sx={{ mr: 1.5 }} />
            Profil
          </MenuItem>
          <MenuItem onClick={handleMenuClose}>
            <SettingsIcon sx={{ mr: 1.5 }} />
            Ayarlar
          </MenuItem>
          <MenuItem onClick={handleLogout} sx={{ color: 'error.main' }}>
            <LogoutIcon sx={{ mr: 1.5 }} />
            {t('logout')}
          </MenuItem>
        </Menu>
      </>
    );
  }

  return null;
};

export default AuthButton;
