import React from 'react';
import { Chip } from '@mui/material';
import { getFileExtension, getFileTypeColor } from '../utils/fileTypeUtils';

/**
 * Badge showing file extension with color coding
 */
const FileExtensionBadge = ({ filename, contentType, size = 'small' }) => {
    const extension = getFileExtension(filename);
    const colors = getFileTypeColor(contentType, filename);

    if (!extension) return null;

    return (
        <Chip
            label={extension}
            size={size}
            sx={{
                backgroundColor: colors.light,
                color: colors.dark,
                fontWeight: 600,
                fontSize: size === 'small' ? '0.7rem' : '0.75rem',
                height: size === 'small' ? 20 : 24,
                border: `1px solid ${colors.primary}40`,
                '& .MuiChip-label': {
                    px: 1,
                },
            }}
        />
    );
};

export default FileExtensionBadge;
