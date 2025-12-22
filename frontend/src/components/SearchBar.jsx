import { TextField, InputAdornment, IconButton } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import { useTranslation } from 'react-i18next';

const SearchBar = ({ value, onChange, placeholder }) => {
    const { t } = useTranslation();

    return (
        <TextField
            fullWidth
            size="small"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder || t('search.placeholder') || 'Search files and folders...'}
            InputProps={{
                startAdornment: (
                    <InputAdornment position="start">
                        <SearchIcon sx={{ color: 'text.secondary' }} />
                    </InputAdornment>
                ),
                endAdornment: value && (
                    <InputAdornment position="end">
                        <IconButton
                            size="small"
                            onClick={() => onChange('')}
                            sx={{
                                padding: '4px',
                                '&:hover': {
                                    backgroundColor: 'rgba(0, 0, 0, 0.04)',
                                },
                            }}
                        >
                            <ClearIcon fontSize="small" />
                        </IconButton>
                    </InputAdornment>
                ),
            }}
            sx={{
                maxWidth: 400,
                '& .MuiOutlinedInput-root': {
                    backgroundColor: 'white',
                    borderRadius: 2,
                    transition: 'all 0.3s ease',
                    '&:hover': {
                        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
                    },
                    '&.Mui-focused': {
                        boxShadow: '0 4px 12px rgba(102, 126, 234, 0.2)',
                    },
                },
            }}
        />
    );
};

export default SearchBar;
