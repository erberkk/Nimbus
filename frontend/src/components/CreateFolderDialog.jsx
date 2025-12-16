import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Box,
  ToggleButtonGroup,
  ToggleButton,
} from '@mui/material';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import FolderIcon from '@mui/icons-material/Folder';

const FOLDER_COLORS = [
  '#667eea',
  '#764ba2',
  '#f093fb',
  '#4facfe',
  '#43e97b',
  '#fa709a',
  '#fee140',
  '#30cfd0',
];

const CreateFolderDialog = ({ open, onClose, onSubmit }) => {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [color, setColor] = useState(FOLDER_COLORS[0]);

  const handleSubmit = () => {
    if (name.trim()) {
      onSubmit({ name: name.trim(), color });
      handleClose();
    }
  };

  const handleClose = () => {
    setName('');
    setColor(FOLDER_COLORS[0]);
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      PaperProps={{
        sx: {
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(102, 126, 234, 0.2)',
          boxShadow: '0 12px 48px 0 rgba(31, 38, 135, 0.25)',
          borderRadius: 3,
        },
      }}
    >
      <DialogTitle
        sx={{
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          color: 'white',
          fontWeight: 600,
          py: 2,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <FolderIcon />
        {t('folder.create')}
      </DialogTitle>
      <DialogContent sx={{ pt: 3 }}>
        <Box sx={{ mt: 2 }}>
          <TextField
            autoFocus
            fullWidth
            label={t('folder.name')}
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyPress={e => {
              if (e.key === 'Enter') {
                handleSubmit();
              }
            }}
            placeholder={t('folder.name_placeholder')}
            sx={{ mb: 3 }}
          />

          <Box sx={{ mb: 2 }}>
            <Box sx={{ mb: 1.5, fontWeight: 600, fontSize: '0.9rem' }}>{t('folder.color')}</Box>
            <Box sx={{ display: 'flex', gap: 1.5, flexWrap: 'wrap' }}>
              {FOLDER_COLORS.map(c => (
                <Box
                  key={c}
                  onClick={() => setColor(c)}
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: 2,
                    background: `linear-gradient(135deg, ${c} 0%, ${c}dd 100%)`,
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    border: color === c ? '3px solid' : '2px solid transparent',
                    borderColor: color === c ? 'primary.main' : 'transparent',
                    transition: 'all 0.2s ease',
                    '&:hover': {
                      transform: 'scale(1.1)',
                      boxShadow: '0px 4px 12px rgba(0,0,0,0.15)',
                    },
                  }}
                >
                  <FolderIcon sx={{ color: 'white', fontSize: 28 }} />
                </Box>
              ))}
            </Box>
          </Box>

          <Box
            sx={{
              mt: 3,
              p: 2.5,
              borderRadius: 2,
              background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.08) 100%)',
              border: '1px solid rgba(102, 126, 234, 0.15)',
              display: 'flex',
              alignItems: 'center',
              gap: 2,
              transition: 'all 0.3s ease',
            }}
          >
            <Box
              sx={{
                width: 60,
                height: 60,
                borderRadius: 2,
                background: `linear-gradient(135deg, ${color} 0%, ${color}dd 100%)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <FolderIcon sx={{ fontSize: 36, color: 'white' }} />
            </Box>
            <Box>
              <Box sx={{ fontWeight: 600, fontSize: '1.1rem' }}>{name || t('folder.new')}</Box>
              <Box sx={{ fontSize: '0.85rem', color: 'text.secondary' }}>
                {t('folder.items_zero')}
              </Box>
            </Box>
          </Box>
        </Box>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 3, gap: 1.5 }}>
        <Button
          onClick={handleClose}
          variant="outlined"
          sx={{
            borderColor: 'rgba(102, 126, 234, 0.3)',
            color: 'text.secondary',
            '&:hover': {
              borderColor: 'rgba(102, 126, 234, 0.5)',
              backgroundColor: 'rgba(102, 126, 234, 0.05)',
            },
            transition: 'all 0.3s ease',
          }}
        >
          {t('cancel')}
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={!name.trim()}
          sx={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            color: 'white',
            fontWeight: 600,
            px: 3,
            '&:hover': {
              background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
              transform: 'translateY(-2px)',
              boxShadow: '0 8px 24px rgba(102, 126, 234, 0.4)',
            },
            '&:disabled': {
              background: 'rgba(0, 0, 0, 0.12)',
              color: 'rgba(0, 0, 0, 0.26)',
            },
            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          }}
        >
          Oluştur
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default CreateFolderDialog;
