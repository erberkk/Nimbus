import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  IconButton,
  TextField,
  Paper,
  Avatar,
  Fade,
  Slide,
  Chip,
} from '@mui/material';
import { motion, AnimatePresence } from 'framer-motion';
import CloseIcon from '@mui/icons-material/Close';
import SendIcon from '@mui/icons-material/Send';
import SmartToyIcon from '@mui/icons-material/SmartToy';
import DescriptionIcon from '@mui/icons-material/Description';

const NimbusChatPanel = ({ isOpen, onClose, file }) => {
  const [messages, setMessages] = useState([]);
  const [inputValue, setInputValue] = useState('');
  const [isTyping, setIsTyping] = useState(false);

  // Ultra ultra smooth renk döngüsü animasyonu
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

  useEffect(() => {
    if (isOpen && file) {
      // Panel açıldığında hoş geldin mesajı
      setMessages([
        {
          id: 1,
          text: `Merhaba! "${file.filename}" dosyası hakkında ne öğrenmek istiyorsun?`,
          isBot: true,
          timestamp: new Date(),
        },
      ]);
    }
  }, [isOpen, file]);

  const handleSendMessage = () => {
    if (!inputValue.trim()) return;

    const newMessage = {
      id: Date.now(),
      text: inputValue,
      isBot: false,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, newMessage]);
    setInputValue('');
    setIsTyping(true);

    // Mock response after 2 seconds
    setTimeout(() => {
      const botResponse = {
        id: Date.now() + 1,
        text: "Bu özellik henüz geliştirilme aşamasında. Yakında dosya içeriğini analiz edip sorularınızı yanıtlayabileceğim! 🤖",
        isBot: true,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, botResponse]);
      setIsTyping(false);
    }, 2000);
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <style>{gradientAnimation}</style>
      <AnimatePresence>
        <motion.div
          initial={{ x: '100%', opacity: 0 }}
          animate={{ x: 0, opacity: 1 }}
          exit={{ x: '100%', opacity: 0 }}
          transition={{
            type: 'spring',
            damping: 25,
            stiffness: 200,
            duration: 0.6,
          }}
          style={{
            position: 'fixed',
            top: '64px',
            right: 0,
            width: '400px',
            height: 'calc(100vh - 64px)',
            zIndex: 1500,
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            boxShadow: '-10px 0 30px rgba(0,0,0,0.3)',
            display: 'flex',
            flexDirection: 'column',
            animation: 'smoothGradientShift 90s linear infinite',
          }}
        >
        {/* Header */}
        <Box
          sx={{
            p: 3,
            borderBottom: '1px solid rgba(255,255,255,0.2)',
            background: 'rgba(255,255,255,0.1)',
            backdropFilter: 'blur(10px)',
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Avatar
                sx={{
                  background: 'linear-gradient(45deg, #667eea, #764ba2)',
                  width: 40,
                  height: 40,
                }}
              >
                <SmartToyIcon />
              </Avatar>
              <Box>
                <Typography variant="h6" sx={{ color: 'white', fontWeight: 600 }}>
                  Nimbus AI
                </Typography>
                <Typography variant="body2" sx={{ color: 'rgba(255,255,255,0.8)' }}>
                  Dosya Asistanı
                </Typography>
              </Box>
            </Box>
            <IconButton
              onClick={onClose}
              sx={{
                color: 'white',
                '&:hover': {
                  background: 'rgba(255,255,255,0.1)',
                },
              }}
            >
              <CloseIcon />
            </IconButton>
          </Box>

          {/* File Info */}
          {file && (
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                p: 2,
                background: 'rgba(255,255,255,0.1)',
                borderRadius: 2,
                border: '1px solid rgba(255,255,255,0.2)',
              }}
            >
              <DescriptionIcon sx={{ color: 'white', fontSize: 24 }} />
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography
                  variant="body2"
                  sx={{
                    color: 'white',
                    fontWeight: 500,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {file.filename}
                </Typography>
                <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.7)' }}>
                  {(file.size / 1024).toFixed(1)} KB
                </Typography>
              </Box>
            </Box>
          )}
        </Box>

        {/* Messages */}
        <Box
          sx={{
            flex: 1,
            overflow: 'auto',
            p: 2,
            display: 'flex',
            flexDirection: 'column',
            gap: 2,
          }}
        >
          {messages.map((message) => (
            <motion.div
              key={message.id}
              initial={{ opacity: 0, y: 20, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              transition={{ duration: 0.3 }}
              style={{
                alignSelf: message.isBot ? 'flex-start' : 'flex-end',
                maxWidth: '80%',
              }}
            >
              <Paper
                sx={{
                  p: 2,
                  background: message.isBot
                    ? 'rgba(255,255,255,0.15)'
                    : 'rgba(255,255,255,0.9)',
                  color: message.isBot ? 'white' : '#333',
                  borderRadius: message.isBot ? '20px 20px 20px 5px' : '20px 20px 5px 20px',
                  backdropFilter: 'blur(10px)',
                  border: '1px solid rgba(255,255,255,0.2)',
                }}
              >
                <Typography variant="body2">{message.text}</Typography>
                <Typography
                  variant="caption"
                  sx={{
                    display: 'block',
                    mt: 1,
                    opacity: 0.7,
                    fontSize: '0.7rem',
                  }}
                >
                  {message.timestamp.toLocaleTimeString('tr-TR', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </Typography>
              </Paper>
            </motion.div>
          ))}

          {isTyping && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              style={{ alignSelf: 'flex-start' }}
            >
              <Paper
                sx={{
                  p: 2,
                  background: 'rgba(255,255,255,0.15)',
                  color: 'white',
                  borderRadius: '20px 20px 20px 5px',
                  backdropFilter: 'blur(10px)',
                  border: '1px solid rgba(255,255,255,0.2)',
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <Typography variant="body2">Nimbus yazıyor</Typography>
                  <Box sx={{ display: 'flex', gap: 0.5 }}>
                    {[0, 1, 2].map((i) => (
                      <motion.div
                        key={i}
                        animate={{ opacity: [0.3, 1, 0.3] }}
                        transition={{
                          duration: 1,
                          repeat: Infinity,
                          delay: i * 0.2,
                        }}
                        style={{
                          width: 6,
                          height: 6,
                          borderRadius: '50%',
                          background: 'white',
                        }}
                      />
                    ))}
                  </Box>
                </Box>
              </Paper>
            </motion.div>
          )}
        </Box>

        {/* Input */}
        <Box
          sx={{
            p: 2,
            borderTop: '1px solid rgba(255,255,255,0.2)',
            background: 'rgba(255,255,255,0.1)',
            backdropFilter: 'blur(10px)',
          }}
        >
          <Box sx={{ display: 'flex', gap: 1 }}>
            <TextField
              fullWidth
              placeholder="Dosya hakkında soru sor..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              variant="outlined"
              size="small"
              sx={{
                '& .MuiOutlinedInput-root': {
                  background: 'rgba(255,255,255,0.9)',
                  borderRadius: '20px',
                  '& fieldset': {
                    border: 'none',
                  },
                  '&:hover fieldset': {
                    border: 'none',
                  },
                  '&.Mui-focused fieldset': {
                    border: '2px solid rgba(255,255,255,0.5)',
                  },
                },
              }}
            />
            <IconButton
              onClick={handleSendMessage}
              disabled={!inputValue.trim()}
              sx={{
                background: 'rgba(255,255,255,0.2)',
                color: 'white',
                '&:hover': {
                  background: 'rgba(255,255,255,0.3)',
                },
                '&:disabled': {
                  background: 'rgba(255,255,255,0.1)',
                  color: 'rgba(255,255,255,0.5)',
                },
              }}
            >
              <SendIcon />
            </IconButton>
          </Box>
        </Box>
       </motion.div>
     </AnimatePresence>
   </>
   );
};

export default NimbusChatPanel;
