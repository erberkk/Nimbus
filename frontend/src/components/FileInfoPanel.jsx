import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  IconButton,
  Paper,
  Divider,
  Chip,
  Avatar,
  Button,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import CloseIcon from '@mui/icons-material/Close';
import InfoIcon from '@mui/icons-material/Info';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import CalendarTodayIcon from '@mui/icons-material/CalendarToday';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import PersonIcon from '@mui/icons-material/Person';
import StorageIcon from '@mui/icons-material/Storage';
import ShareIcon from '@mui/icons-material/Share';
import LockIcon from '@mui/icons-material/Lock';
import LockOpenIcon from '@mui/icons-material/LockOpen';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import { formatFileSize, formatDate, formatContentType } from '../utils/fileUtils';
import { shareApi } from '../services/api';

const FileInfoPanel = ({ isOpen, onClose, file }) => {
  const [copied, setCopied] = useState(false);
  const [publicLink, setPublicLink] = useState(null);
  const [loadingLink, setLoadingLink] = useState(false);

  useEffect(() => {
    if (isOpen && file) {
      setPublicLink(file.public_link || null);
      if (file.id && !file.public_link) {
        const loadPublicLink = async () => {
          try {
            setLoadingLink(true);
            const data = await shareApi.getResourceShares(file.id);
            if (data && data.public_link) {
              setPublicLink(data.public_link);
            }
          } catch (error) {
            // Silent fail - public link may not exist
          } finally {
            setLoadingLink(false);
          }
        };
        loadPublicLink();
      }
    }
  }, [isOpen, file]);

  const handleCopyPublicLink = async () => {
    if (!publicLink) return;

    try {
      await navigator.clipboard.writeText(`${window.location.origin}/share/${publicLink}`);
      setCopied(true);
      window.toast?.success('Bağlantı kopyalandı');
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      window.toast?.error('Bağlantı kopyalanamadı');
    }
  };

  if (!file) return null;

  const gradientAnimation = `
    @keyframes smoothGradientShift {
      0% { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
      3.125% { background: linear-gradient(140deg, #6671ea 0%, #764ca2 100%); }
      6.25% { background: linear-gradient(145deg, #6772eb 0%, #774da3 100%); }
      9.375% { background: linear-gradient(150deg, #6773eb 0%, #774ea4 100%); }
      12.5% { background: linear-gradient(155deg, #6874ec 0%, #784fa5 100%); }
      15.625% { background: linear-gradient(160deg, #6875ec 0%, #7850a6 100%); }
      18.75% { background: linear-gradient(165deg, #6976ed 0%, #7951a7 100%); }
      21.875% { background: linear-gradient(170deg, #6977ed 0%, #7952a8 100%); }
      25% { background: linear-gradient(175deg, #6a78ee 0%, #7a53a9 100%); }
      28.125% { background: linear-gradient(180deg, #6a79ee 0%, #7a54aa 100%); }
      31.25% { background: linear-gradient(185deg, #6b7aef 0%, #7b55ab 100%); }
      34.375% { background: linear-gradient(190deg, #6b7bef 0%, #7b56ac 100%); }
      37.5% { background: linear-gradient(195deg, #6c7cf0 0%, #7c57ad 100%); }
      40.625% { background: linear-gradient(200deg, #6c7df0 0%, #7c58ae 100%); }
      43.75% { background: linear-gradient(205deg, #6d7ef1 0%, #7d59af 100%); }
      46.875% { background: linear-gradient(210deg, #6d7ff1 0%, #7d5ab0 100%); }
      50% { background: linear-gradient(215deg, #6e80f2 0%, #7e5bb1 100%); }
      53.125% { background: linear-gradient(220deg, #6e81f2 0%, #7e5cb2 100%); }
      56.25% { background: linear-gradient(225deg, #6f82f3 0%, #7f5db3 100%); }
      59.375% { background: linear-gradient(230deg, #6f83f3 0%, #7f5eb4 100%); }
      62.5% { background: linear-gradient(235deg, #7084f4 0%, #805fb5 100%); }
      65.625% { background: linear-gradient(240deg, #7085f4 0%, #8060b6 100%); }
      68.75% { background: linear-gradient(245deg, #7186f5 0%, #8161b7 100%); }
      71.875% { background: linear-gradient(250deg, #7187f5 0%, #8162b8 100%); }
      75% { background: linear-gradient(255deg, #7288f6 0%, #8263b9 100%); }
      78.125% { background: linear-gradient(260deg, #7289f6 0%, #8264ba 100%); }
      81.25% { background: linear-gradient(265deg, #738af7 0%, #8365bb 100%); }
      84.375% { background: linear-gradient(270deg, #738bf7 0%, #8366bc 100%); }
      87.5% { background: linear-gradient(275deg, #748cf8 0%, #8467bd 100%); }
      90.625% { background: linear-gradient(280deg, #748df8 0%, #8468be 100%); }
      93.75% { background: linear-gradient(285deg, #758ef9 0%, #8569bf 100%); }
      96.875% { background: linear-gradient(290deg, #758ff9 0%, #856ac0 100%); }
      100% { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
    }
  `;

  const InfoRow = ({ icon, label, value, chip }) => (
    <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 2, py: 1.5 }}>
      <Box
        sx={{
          width: 40,
          height: 40,
          borderRadius: '50%',
          backgroundColor: 'rgba(255, 255, 255, 0.2)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0,
          mt: 0.5,
        }}
      >
        {icon}
      </Box>
      <Box sx={{ flex: 1, minWidth: 0 }}>
        <Typography variant="caption" sx={{ color: 'rgba(255, 255, 255, 0.7)', display: 'block', mb: 0.5 }}>
          {label}
        </Typography>
        {chip ? (
          <Chip
            label={value}
            size="small"
            sx={{
              backgroundColor: 'rgba(255, 255, 255, 0.2)',
              color: 'white',
              border: '1px solid rgba(255, 255, 255, 0.3)',
              '& .MuiChip-label': {
                fontSize: '0.75rem',
              },
            }}
          />
        ) : (
          <Typography 
            variant="body2" 
            sx={{ 
              color: 'white', 
              fontWeight: 500,
              wordBreak: 'break-word',
            }}
          >
            {value}
          </Typography>
        )}
      </Box>
    </Box>
  );

  return (
    <>
      <style>{gradientAnimation}</style>
      <AnimatePresence>
        {isOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={onClose}
              style={{
                position: 'fixed',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundColor: 'rgba(0, 0, 0, 0.5)',
                zIndex: 1300,
              }}
            />
            <motion.div
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              style={{
                position: 'fixed',
                top: '64px',
                right: 0,
                height: 'calc(100vh - 64px)',
                width: '420px',
                maxWidth: '90vw',
                zIndex: 1301,
                pointerEvents: 'auto',
              }}
            >
              <Paper
                sx={{
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  borderRadius: 0,
                  background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                  animation: 'smoothGradientShift 90s linear infinite',
                  boxShadow: '-4px 0px 20px rgba(0, 0, 0, 0.3)',
                }}
              >
                <Box
                  sx={{
                    p: 3,
                    pb: 2,
                    borderBottom: '1px solid rgba(255, 255, 255, 0.2)',
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                      <Box
                        sx={{
                          width: 48,
                          height: 48,
                          borderRadius: '12px',
                          backgroundColor: 'rgba(255, 255, 255, 0.2)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                        }}
                      >
                        <InfoIcon sx={{ color: 'white', fontSize: 28 }} />
                      </Box>
                      <Typography variant="h6" sx={{ color: 'white', fontWeight: 600 }}>
                        Dosya Bilgileri
                      </Typography>
                    </Box>
                    <IconButton
                      onClick={onClose}
                      sx={{
                        color: 'white',
                        '&:hover': {
                          backgroundColor: 'rgba(255, 255, 255, 0.2)',
                        },
                      }}
                    >
                      <CloseIcon />
                    </IconButton>
                  </Box>

                  <Box
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 2,
                      p: 2,
                      borderRadius: 2,
                      backgroundColor: 'rgba(255, 255, 255, 0.1)',
                      backdropFilter: 'blur(10px)',
                    }}
                  >
                    <Box
                      sx={{
                        width: 56,
                        height: 56,
                        borderRadius: '12px',
                        backgroundColor: 'rgba(255, 255, 255, 0.2)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        flexShrink: 0,
                      }}
                    >
                      <InsertDriveFileIcon sx={{ color: 'white', fontSize: 32 }} />
                    </Box>
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography
                        variant="body1"
                        sx={{
                          color: 'white',
                          fontWeight: 600,
                          mb: 0.5,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                        title={file.filename}
                      >
                        {file.filename}
                      </Typography>
                      <Chip
                        label={formatContentType(file.content_type, file.filename)}
                        size="small"
                        sx={{
                          backgroundColor: 'rgba(255, 255, 255, 0.2)',
                          color: 'white',
                          border: '1px solid rgba(255, 255, 255, 0.3)',
                          fontSize: '0.7rem',
                        }}
                      />
                    </Box>
                  </Box>
                </Box>

                <Box
                  sx={{
                    flex: 1,
                    overflowY: 'auto',
                    p: 3,
                    '&::-webkit-scrollbar': {
                      width: '8px',
                    },
                    '&::-webkit-scrollbar-track': {
                      background: 'rgba(255, 255, 255, 0.1)',
                    },
                    '&::-webkit-scrollbar-thumb': {
                      background: 'rgba(255, 255, 255, 0.3)',
                      borderRadius: '4px',
                      '&:hover': {
                        background: 'rgba(255, 255, 255, 0.4)',
                      },
                    },
                  }}
                >
                  <Typography
                    variant="overline"
                    sx={{
                      color: 'rgba(255, 255, 255, 0.7)',
                      fontWeight: 600,
                      letterSpacing: 1,
                      display: 'block',
                      mb: 2,
                    }}
                  >
                    GENEL BİLGİLER
                  </Typography>

                  <InfoRow
                    icon={<StorageIcon sx={{ color: 'white', fontSize: 20 }} />}
                    label="Dosya Boyutu"
                    value={formatFileSize(file.size)}
                  />

                  <InfoRow
                    icon={<CalendarTodayIcon sx={{ color: 'white', fontSize: 20 }} />}
                    label="Oluşturulma Tarihi"
                    value={formatDate(file.created_at)}
                  />

                  <InfoRow
                    icon={<AccessTimeIcon sx={{ color: 'white', fontSize: 20 }} />}
                    label="Son Güncelleme"
                    value={formatDate(file.updated_at || file.created_at)}
                  />

                  {file.content_type && (
                    <InfoRow
                      icon={<InsertDriveFileIcon sx={{ color: 'white', fontSize: 20 }} />}
                      label="Dosya Türü"
                      value={formatContentType(file.content_type, file.filename)}
                      chip
                    />
                  )}

                  <Divider sx={{ my: 2, borderColor: 'rgba(255, 255, 255, 0.2)' }} />

                  <Typography
                    variant="overline"
                    sx={{
                      color: 'rgba(255, 255, 255, 0.7)',
                      fontWeight: 600,
                      letterSpacing: 1,
                      display: 'block',
                      mb: 2,
                    }}
                  >
                    ERİŞİM BİLGİLERİ
                  </Typography>

                  {file.isShared && file.owner ? (
                    <>
                      <InfoRow
                        icon={
                          file.owner.avatar ? (
                            <Avatar
                              src={file.owner.avatar}
                              sx={{ width: 24, height: 24 }}
                            />
                          ) : (
                            <PersonIcon sx={{ color: 'white', fontSize: 20 }} />
                          )
                        }
                        label="Sahibi"
                        value={file.owner.name || file.owner.email || 'Bilinmiyor'}
                      />

                      <InfoRow
                        icon={
                          file.access_type === 'read' ? (
                            <LockIcon sx={{ color: 'white', fontSize: 20 }} />
                          ) : (
                            <LockOpenIcon sx={{ color: 'white', fontSize: 20 }} />
                          )
                        }
                        label="Erişim Seviyesi"
                        value={file.access_type === 'read' ? 'Görüntüleme' : 'Düzenleme'}
                        chip
                      />
                    </>
                  ) : (
                    <InfoRow
                      icon={<PersonIcon sx={{ color: 'white', fontSize: 20 }} />}
                      label="Sahibi"
                      value="Siz"
                    />
                  )}

                  {publicLink && (
                    <>
                      <Divider sx={{ my: 2, borderColor: 'rgba(255, 255, 255, 0.2)' }} />

                      <Typography
                        variant="overline"
                        sx={{
                          color: 'rgba(255, 255, 255, 0.7)',
                          fontWeight: 600,
                          letterSpacing: 1,
                          display: 'block',
                          mb: 2,
                        }}
                      >
                        PAYLAŞIM LİNKİ
                      </Typography>

                      <Box
                        sx={{
                          p: 2,
                          borderRadius: 2,
                          backgroundColor: 'rgba(255, 255, 255, 0.1)',
                          backdropFilter: 'blur(10px)',
                          border: '1px solid rgba(255, 255, 255, 0.2)',
                        }}
                      >
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1.5 }}>
                          <ShareIcon sx={{ color: 'white', fontSize: 20 }} />
                          <Typography
                            variant="caption"
                            sx={{ color: 'rgba(255, 255, 255, 0.7)', fontWeight: 500 }}
                          >
                            Public Bağlantı
                          </Typography>
                        </Box>
                        <Typography
                          variant="body2"
                          sx={{
                            color: 'white',
                            fontFamily: 'monospace',
                            fontSize: '0.75rem',
                            wordBreak: 'break-all',
                            mb: 1.5,
                            p: 1,
                            backgroundColor: 'rgba(0, 0, 0, 0.2)',
                            borderRadius: 1,
                          }}
                        >
                          {`${window.location.origin}/share/${publicLink}`}
                        </Typography>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<ContentCopyIcon />}
                          onClick={handleCopyPublicLink}
                          disabled={copied || loadingLink}
                          sx={{
                            color: 'white',
                            borderColor: 'rgba(255, 255, 255, 0.3)',
                            '&:hover': {
                              borderColor: 'rgba(255, 255, 255, 0.5)',
                              backgroundColor: 'rgba(255, 255, 255, 0.1)',
                            },
                            '&:disabled': {
                              color: 'rgba(255, 255, 255, 0.5)',
                              borderColor: 'rgba(255, 255, 255, 0.2)',
                            },
                          }}
                        >
                          {copied ? 'Kopyalandı!' : 'Kopyala'}
                        </Button>
                      </Box>
                    </>
                  )}
                </Box>
              </Paper>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
};

export default FileInfoPanel;

