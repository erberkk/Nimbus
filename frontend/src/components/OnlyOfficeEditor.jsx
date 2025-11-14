import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  CircularProgress,
  IconButton,
  Slide,
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import { fileApi } from '../services/api';

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

// OnlyOffice Document Server URL (from environment or default)
const ONLYOFFICE_SERVER_URL = import.meta.env.VITE_ONLYOFFICE_SERVER_URL || 'http://localhost:5000';

const OnlyOfficeEditor = ({ open, onClose, file, onSave, mode = 'edit' }) => {
  const { t } = useTranslation();
  const editorRef = useRef(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [config, setConfig] = useState(null);
  const [scriptLoaded, setScriptLoaded] = useState(false);
  const editorInstanceRef = useRef(null);

  useEffect(() => {
    if (!open) {
      setScriptLoaded(false);
      return;
    }

    const scriptId = 'onlyoffice-editor-script';

    if (document.getElementById(scriptId) && window.DocsAPI && window.DocsAPI.DocEditor) {
      setScriptLoaded(true);
      setLoading(false);
      return;
    }

    if (window.DocsAPI && window.DocsAPI.DocEditor) {
      setScriptLoaded(true);
      setLoading(false);
      return;
    }

    const script = document.createElement('script');
    script.id = scriptId;
    script.src = `${ONLYOFFICE_SERVER_URL}/web-apps/apps/api/documents/api.js`;
    script.async = true;

    script.onload = () => {
      setTimeout(() => {
        if (window.DocsAPI && window.DocsAPI.DocEditor) {
          setScriptLoaded(true);
          setLoading(false);
        } else {
          setError(t('onlyoffice.api_error'));
          setLoading(false);
          window.toast?.error(t('onlyoffice.api_error'));
        }
      }, 100);
    };

    script.onerror = () => {
      setError(t('onlyoffice.connection_error'));
      setLoading(false);
      window.toast?.error(t('onlyoffice.connection_failed'));
    };

    document.head.appendChild(script);
  }, [open]);

  useEffect(() => {
    if (!open || !file || !scriptLoaded) {
      return;
    }

    if (!window.DocsAPI || !window.DocsAPI.DocEditor) {
      return;
    }

    let isMounted = true;

    const loadConfig = async () => {
      try {
        if (!isMounted) return;
        setLoading(true);
        setError(null);

        const configData = await fileApi.getOnlyOfficeConfig(file.id, mode);

        if (!isMounted) return;

        if (!configData || !configData.document || !configData.document.url) {
          throw new Error('Geçersiz OnlyOffice config: document URL bulunamadı');
        }

        setConfig(configData);
        setLoading(false);
      } catch (err) {
        if (!isMounted) return;
        setError(t('onlyoffice.config_error'));
        window.toast?.error(t('onlyoffice.start_error'));
        setLoading(false);
      }
    };

    loadConfig();

    return () => {
      isMounted = false;
    };
  }, [open, file?.id, scriptLoaded, mode]);

  useEffect(() => {
    if (!config || loading || error || editorInstanceRef.current) {
      return;
    }

    let retryCount = 0;
    const maxRetries = 20;

    const checkAndInit = () => {
      if (editorInstanceRef.current) {
        return;
      }

      if (editorRef.current && window.DocsAPI && window.DocsAPI.DocEditor) {
        initializeEditor(config);
      } else if (retryCount < maxRetries) {
        retryCount++;
        setTimeout(checkAndInit, 100);
      } else {
        setError('Editor başlatılamadı: Container veya API hazır değil');
        setLoading(false);
      }
    };

    const timeoutId = setTimeout(checkAndInit, 100);

    return () => clearTimeout(timeoutId);
  }, [config, loading, error]);

  const initializeEditor = editorConfig => {
    if (!editorRef.current || !window.DocsAPI || !window.DocsAPI.DocEditor) {
      setError(t('onlyoffice.editor_error'));
      setLoading(false);
      return;
    }

    try {
      if (editorInstanceRef.current) {
        try {
          if (editorInstanceRef.current.destroyEditor) {
            editorInstanceRef.current.destroyEditor();
          }
        } catch (err) {
          // Ignore
        }
        editorInstanceRef.current = null;
      }

      const editorContainerId =
        editorRef.current.id || `onlyoffice-editor-container-${file?.id || 'default'}`;
      const editor = new window.DocsAPI.DocEditor(editorContainerId, editorConfig);

      if (!editor) {
        throw new Error('Editor instance oluşturulamadı');
      }

      editorInstanceRef.current = editor;

      let eventSetupAttempts = 0;
      const maxEventSetupAttempts = 5;

      const setupEvents = () => {
        eventSetupAttempts++;

        if (editor && editor.events && typeof editor.events.on === 'function') {
          try {
            editor.events.on('onDocumentReady', () => {
              setLoading(false);
            });

            editor.events.on('onError', event => {
              setError('Editor hatası oluştu');
              window.toast?.error(t('onlyoffice.edit_error'));
            });

            editor.events.on('onDocumentStateChange', () => {
              // Document modified
            });
          } catch (err) {
            setLoading(false);
          }
        } else if (eventSetupAttempts < maxEventSetupAttempts) {
          setTimeout(setupEvents, 200);
        } else {
          setLoading(false);
        }
      };

      setupEvents();
    } catch (err) {
      setError(`Editor başlatılamadı: ${err.message || 'Bilinmeyen hata'}`);
      window.toast?.error(t('onlyoffice.init_error'));
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!open && editorInstanceRef.current) {
      try {
        if (typeof editorInstanceRef.current.destroyEditor === 'function') {
          editorInstanceRef.current.destroyEditor();
        }
        editorInstanceRef.current = null;
      } catch (err) {
        // Ignore
      }
      setConfig(null);
      setError(null);
      setLoading(true);
    }
  }, [open]);

  const handleClose = () => {
    if (editorInstanceRef.current) {
      try {
        if (typeof editorInstanceRef.current.destroyEditor === 'function') {
          editorInstanceRef.current.destroyEditor();
        }
        editorInstanceRef.current = null;
      } catch (err) {
        // Ignore
      }
    }
    setConfig(null);
    setError(null);
    setLoading(true);
    setScriptLoaded(false);
    onClose();
  };

  if (!file) return null;

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth={false}
      fullWidth
      TransitionComponent={Transition}
      PaperProps={{
        sx: {
          borderRadius: 2,
          maxHeight: '95vh',
          height: '95vh',
          width: '95vw',
          maxWidth: '95vw',
        },
      }}
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6" noWrap title={file.filename}>
            {file.filename} -{' '}
            {mode === 'edit' ? t('onlyoffice.edit_mode') : t('onlyoffice.preview_mode')}
          </Typography>
          <IconButton onClick={handleClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent
        dividers
        sx={{ p: 0, position: 'relative', height: 'calc(95vh - 120px)', overflow: 'hidden' }}
      >
        {loading && (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              gap: 2,
            }}
          >
            <CircularProgress size={48} />
            <Typography variant="body2" color="text.secondary">
              {t('onlyoffice.loading')}
            </Typography>
          </Box>
        )}

        {error && (
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
              gap: 2,
              p: 3,
            }}
          >
            <Typography variant="h6" color="error">
              Hata
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {error}
            </Typography>
          </Box>
        )}

        {/* Editor container - always render when dialog is open to ensure ref is available */}
        {!error && (
          <Box
            id={`onlyoffice-editor-container-${file?.id || 'default'}`}
            ref={editorRef}
            sx={{
              width: '100%',
              height: '100%',
              minHeight: '600px',
              display: loading ? 'none' : 'block',
            }}
          />
        )}
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={handleClose} variant="outlined">
          Kapat
        </Button>
        {mode === 'edit' && onSave && (
          <Button
            onClick={() => {
              if (onSave) {
                onSave(file);
              }
              handleClose();
            }}
            variant="contained"
            sx={{
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              '&:hover': {
                background: 'linear-gradient(135deg, #5568d3 0%, #6a4190 100%)',
              },
            }}
          >
            Kaydet
          </Button>
        )}
      </DialogActions>
    </Dialog>
  );
};

export default OnlyOfficeEditor;
