import { useState, useEffect } from 'react';
import {
  Autocomplete,
  TextField,
  Chip,
  Box,
  CircularProgress,
  Avatar,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import PersonIcon from '@mui/icons-material/Person';
import { userApi } from '../services/api';

const UserSearch = ({ selectedUsers, onSelectUser, onRemoveUser, excludeUserId, excludeUserIds = [] }) => {
  const { t } = useTranslation();
  const [inputValue, setInputValue] = useState('');
  const [options, setOptions] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (inputValue.length < 2) {
      setOptions([]);
      return;
    }

    const timeoutId = setTimeout(async () => {
      try {
        setLoading(true);
        const users = await userApi.searchUsers(inputValue);
        // Filter out excluded users (owner + already shared users)
        const excludedIds = [excludeUserId, ...excludeUserIds].filter(Boolean);
        const filteredUsers = users.filter(user => !excludedIds.includes(user.id));
        setOptions(filteredUsers);
      } catch (error) {
        console.error('User search failed:', error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [inputValue, excludeUserId, excludeUserIds]);

  return (
    <Box>
      <Autocomplete
        multiple
        freeSolo
        options={options}
        value={selectedUsers}
        inputValue={inputValue}
        onInputChange={(e, value) => setInputValue(value)}
        onChange={(e, value) => {
          if (value.length > selectedUsers.length) {
            const newUser = value[value.length - 1];
            if (typeof newUser === 'object') {
              onSelectUser(newUser);
            }
          }
        }}
        getOptionLabel={option => (typeof option === 'string' ? option : option.email)}
        isOptionEqualToValue={(option, value) => option.id === value.id}
        loading={loading}
        renderInput={params => (
          <TextField
            {...params}
            placeholder="Email ile kullanıcı ara..."
            variant="outlined"
            InputProps={{
              ...params.InputProps,
              endAdornment: (
                <>
                  {loading ? <CircularProgress color="inherit" size={20} /> : null}
                  {params.InputProps.endAdornment}
                </>
              ),
            }}
          />
        )}
        renderOption={(props, option) => {
          const { key, ...otherProps } = props;
          return (
            <Box component="li" key={key} {...otherProps} sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'primary.main' }}>
                <PersonIcon fontSize="small" />
              </Avatar>
              <Box>
                <Typography variant="body2">{option.name || option.email}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {option.email}
                </Typography>
              </Box>
            </Box>
          );
        }}
        renderTags={(value, getTagProps) =>
          value.map((option, index) => (
            <Chip
              {...getTagProps({ index })}
              key={option.id}
              label={option.email}
              onDelete={() => onRemoveUser(option.id)}
              avatar={
                <Avatar sx={{ width: 24, height: 24, bgcolor: 'primary.main' }}>
                  <PersonIcon fontSize="small" />
                </Avatar>
              }
            />
          ))
        }
        noOptionsText={
          inputValue.length < 2 ? 'En az 2 karakter girin' : 'Kullanıcı bulunamadı'
        }
      />
    </Box>
  );
};

export default UserSearch;

