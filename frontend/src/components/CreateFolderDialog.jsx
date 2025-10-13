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
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Yeni Klasör Oluştur</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2 }}>
          <TextField
            autoFocus
            fullWidth
            label="Klasör Adı"
            value={name}
            onChange={e => setName(e.target.value)}
            onKeyPress={e => {
              if (e.key === 'Enter') {
                handleSubmit();
              }
            }}
            placeholder="Belgeler, Resimler, Projeler..."
            sx={{ mb: 3 }}
          />

          <Box sx={{ mb: 2 }}>
            <Box sx={{ mb: 1.5, fontWeight: 600, fontSize: '0.9rem' }}>Klasör Rengi</Box>
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
              p: 2,
              borderRadius: 2,
              backgroundColor: 'grey.100',
              display: 'flex',
              alignItems: 'center',
              gap: 2,
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
              <Box sx={{ fontWeight: 600, fontSize: '1.1rem' }}>{name || 'Yeni Klasör'}</Box>
              <Box sx={{ fontSize: '0.85rem', color: 'text.secondary' }}>0 öğe</Box>
            </Box>
          </Box>
        </Box>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={handleClose} color="inherit">
          {t('cancel')}
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={!name.trim()}
          sx={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            color: 'white',
            '&:hover': {
              background: 'linear-gradient(135deg, #5568d3 0%, #653a8b 100%)',
              transform: 'translateY(-2px)',
              boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)',
            },
            transition: 'all 0.3s ease',
          }}
        >
          Oluştur
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default CreateFolderDialog;
