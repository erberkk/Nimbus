import React from 'react';
import { Box } from '@mui/material';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import DescriptionIcon from '@mui/icons-material/Description';
import GridOnIcon from '@mui/icons-material/GridOn';
import SlideshowIcon from '@mui/icons-material/Slideshow';
import ImageIcon from '@mui/icons-material/Image';
import VideoFileIcon from '@mui/icons-material/VideoFile';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import CodeIcon from '@mui/icons-material/Code';
import FolderZipIcon from '@mui/icons-material/FolderZip';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import { getFileTypeIcon, getFileTypeColor } from '../utils/fileTypeUtils';

const iconComponents = {
    PictureAsPdf: PictureAsPdfIcon,
    Description: DescriptionIcon,
    GridOn: GridOnIcon,
    Slideshow: SlideshowIcon,
    Image: ImageIcon,
    VideoFile: VideoFileIcon,
    TextSnippet: TextSnippetIcon,
    Code: CodeIcon,
    FolderZip: FolderZipIcon,
    InsertDriveFile: InsertDriveFileIcon,
};

/**
 * Colored file type icon component
 */
const FileTypeIcon = ({ filename, contentType, size = 24 }) => {
    const iconName = getFileTypeIcon(contentType, filename);
    const colors = getFileTypeColor(contentType, filename);
    const IconComponent = iconComponents[iconName] || InsertDriveFileIcon;

    return (
        <Box
            sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: size + 16,
                height: size + 16,
                borderRadius: 1,
                backgroundColor: colors.light,
                border: `1px solid ${colors.primary}20`,
            }}
        >
            <IconComponent
                sx={{
                    fontSize: size,
                    color: colors.primary,
                }}
            />
        </Box>
    );
};

export default FileTypeIcon;
