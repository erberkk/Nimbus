import React, { useState, useEffect, useCallback } from 'react';
import { Snackbar, Alert, Slide, IconButton } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import { motion } from 'framer-motion';

const ToastProvider = ({ children }) => {
  const [toast, setToast] = useState({
    open: false,
    message: '',
    severity: 'info',
    duration: 4000,
  });

  const showToast = useCallback((message, severity = 'info', duration = 4000) => {
    setToast({
      open: true,
      message,
      severity,
      duration,
    });
  }, []);

  const hideToast = useCallback(() => {
    setToast(prev => ({ ...prev, open: false }));
  }, []);

  // Global toast fonksiyonlarını window'a ekle
  useEffect(() => {
    window.showToast = showToast;
    window.toast = {
      success: (message, duration) => showToast(message, 'success', duration),
      error: (message, duration) => showToast(message, 'error', duration),
      warning: (message, duration) => showToast(message, 'warning', duration),
      info: (message, duration) => showToast(message, 'info', duration),
    };

    return () => {
      delete window.showToast;
      delete window.toast;
    };
  }, [showToast]);

  const handleClose = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
    hideToast();
  };

  return (
    <>
      {children}
      <Snackbar
        open={toast.open}
        autoHideDuration={toast.duration}
        onClose={handleClose}
        TransitionComponent={props => <Slide {...props} direction="left" />}
        anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
        sx={{
          mt: 8, // Navbar'ın altında
          mr: 2,
        }}
      >
        <Alert
          onClose={handleClose}
          severity={toast.severity}
          variant="filled"
          sx={{
            minWidth: '300px',
            maxWidth: '500px',
            borderRadius: 2,
            boxShadow: '0px 8px 24px rgba(0,0,0,0.12)',
            '& .MuiAlert-icon': {
              fontSize: 24,
            },
            '& .MuiAlert-message': {
              fontSize: '0.95rem',
              fontWeight: 500,
            },
          }}
          action={
            <IconButton size="small" aria-label="close" color="inherit" onClick={handleClose}>
              <CloseIcon fontSize="small" />
            </IconButton>
          }
        >
          {toast.message}
        </Alert>
      </Snackbar>
    </>
  );
};

export default ToastProvider;
