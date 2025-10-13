import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  List,
  ListItem,
  ListItemText,
  ListItemAvatar,
  Avatar,
  IconButton,
  Select,
  MenuItem,
  Divider,
  CircularProgress,
} from '@mui/material';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import DeleteIcon from '@mui/icons-material/Delete';
import PersonIcon from '@mui/icons-material/Person';
import UserSearch from './UserSearch';
import { shareApi } from '../services/api';
import { useAuth } from '../contexts/AuthContext';

const ShareDialog = ({ open, onClose, resource, resourceType }) => {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);
  const [publicLink, setPublicLink] = useState(null);
  const [copied, setCopied] = useState(false);
  const [userAccessLevel, setUserAccessLevel] = useState('owner'); // owner, write, read

  useEffect(() => {
    if (open && resource) {
      setShares([]);
      setPublicLink(null);
      setCopied(false);
      loadShares();
    }
  }, [open, resource]);

  const loadShares = async () => {
    try {
      setLoading(true);
      const data = await shareApi.getResourceShares(resource.id);
      setShares(data || {});
      
      // Determine user's access level
      if (data && data.access_list) {
        const userAccess = data.access_list.find(access => access.user_id === user.id);
        if (userAccess) {
          setUserAccessLevel(userAccess.access_type); // read or write
        } else if (data.resource_id && user.id === resource.user_id) {
          setUserAccessLevel('owner');
        } else {
          setUserAccessLevel('read'); // Default for shared users
        }
      } else {
        setUserAccessLevel('owner'); // Default for owners
      }
    } catch (error) {
      console.error('Failed to load shares:', error);
      window.toast?.error('Paylaşımlar yüklenemedi');
      setShares({});
      setUserAccessLevel('owner'); // Default fallback
    } finally {
      setLoading(false);
    }
  };


  const handleAddUser = async (selectedUser) => {
    try {
      // Add user to access list (this will update the file/folder's access_list)
      await shareApi.updateAccessPermission(resource.id, {
        user_id: selectedUser.id,
        permission: 'read'
      });

      // Refresh the shares data to get updated access list
      await loadShares();
      window.toast?.success(`${selectedUser.email} ile paylaşıldı`);
    } catch (error) {
      console.error('Failed to share:', error);
      window.toast?.error('Paylaşım başarısız');
    }
  };

  const handlePermissionChange = async (userId, newPermission) => {
    try {
      await shareApi.updateAccessPermission(resource.id, {
        user_id: userId,
        permission: newPermission
      });

      // Refresh the shares data to get updated access list
      await loadShares();
      window.toast?.success('İzin güncellendi');
    } catch (error) {
      console.error('Failed to update permission:', error);
      window.toast?.error('İzin güncellenemedi');
    }
  };

  const handleRemoveUser = async (userId) => {
    try {
      await shareApi.removeUserAccess(resource.id, userId);

      // Refresh the shares data to get updated access list
      await loadShares();
      window.toast?.success('Paylaşım kaldırıldı');
    } catch (error) {
      console.error('Failed to remove share:', error);
      window.toast?.error('Paylaşım kaldırılamadı');
    }
  };


  const handleClose = () => {
    onClose();
  };

  const handleCopyLink = () => {
    if (publicLink) {
      navigator.clipboard.writeText(publicLink);
      setCopied(true);
      window.toast?.success('Link kopyalandı');
      setTimeout(() => setCopied(false), 2000);
    }
  };

  // Get shared users from access_list data
  const getSharedUsers = () => {
    if (!shares) {
      return [];
    }

    // If we have shared_with (user info), use that
    if (shares.shared_with && Array.isArray(shares.shared_with)) {
      return shares.shared_with.map(user => ({
        id: user.id,
        email: user.email,
        name: user.name,
        access_type: getAccessTypeForUser(user.id)
      }));
    }

    // Fallback to access_list if shared_with is not available
    if (shares.access_list && Array.isArray(shares.access_list)) {
      return shares.access_list.map(access => ({
        id: access.user_id,
        email: '', // We don't have email here
        name: access.user_id, // Fallback to ID
        access_type: access.access_type
      }));
    }

    return [];
  };

  // Get access type for a specific user
  const getAccessTypeForUser = (userId) => {
    if (!shares || !shares.access_list) return 'read';

    const accessEntry = shares.access_list.find(access => access.user_id === userId);
    return accessEntry ? accessEntry.access_type : 'read';
  };

  const sharedUsers = getSharedUsers();

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="h6" component="span">
            Paylaş
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {resource?.filename || resource?.name}
          </Typography>
        </Box>
      </DialogTitle>

      <DialogContent>
        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {/* User Search - Only show for owners and write users */}
            {userAccessLevel === 'owner' || userAccessLevel === 'write' ? (
              <Box sx={{ mb: 3 }}>
                <Typography variant="subtitle2" sx={{ mb: 1.5, fontWeight: 600 }}>
                  Kullanıcı Ekle
                </Typography>
                <UserSearch
                  selectedUsers={[]}
                  onSelectUser={handleAddUser}
                  onRemoveUser={() => {}}
                  excludeUserId={user?.id}
                  excludeUserIds={sharedUsers.map(u => u.id)}
                />
              </Box>
            ) : (
              <Box sx={{ mb: 3, p: 2, bgcolor: 'grey.100', borderRadius: 1 }}>
                <Typography variant="body2" color="text.secondary">
                  Sadece görüntüleme yetkiniz var. Bu kaynağı paylaşamaz veya erişimleri düzenleyemezsiniz.
                </Typography>
              </Box>
            )}

            <Divider sx={{ my: 2 }} />

            {/* Users with Access */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle2" sx={{ mb: 1.5, fontWeight: 600 }}>
                Erişimi Olan Kişiler
              </Typography>
              <List dense>
                {/* Owner (Current User) */}
                <ListItem
                  sx={{
                    bgcolor: 'action.hover',
                    borderRadius: 1,
                    mb: 1,
                  }}
                >
                  <ListItemAvatar>
                    <Avatar sx={{ bgcolor: 'primary.main' }}>
                      <PersonIcon />
                    </Avatar>
                  </ListItemAvatar>
                  <ListItemText
                     primary={
                       <Typography variant="body2" fontWeight={600}>
                         {user?.name || user?.email}
                       </Typography>
                     }
                    secondary={
                      <Typography variant="caption" color="text.secondary">
                        Sahip
                      </Typography>
                    }
                  />
                </ListItem>

                {/* Shared Users */}
                {sharedUsers.map((userInfo) => (
                  <ListItem
                    key={userInfo.id}
                    sx={{
                      borderRadius: 1,
                      mb: 0.5,
                      '&:hover': {
                        bgcolor: 'action.hover',
                      },
                    }}
                    secondaryAction={
                      (userAccessLevel === 'owner' || userAccessLevel === 'write') ? (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Select
                            value={userInfo.access_type}
                            onChange={e => handlePermissionChange(userInfo.id, e.target.value)}
                            size="small"
                            sx={{ minWidth: 130 }}
                          >
                            <MenuItem value="read">Görüntüleyen</MenuItem>
                            <MenuItem value="write">Düzenleyen</MenuItem>
                          </Select>
                          <IconButton
                            edge="end"
                            size="small"
                            onClick={() => handleRemoveUser(userInfo.id)}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        </Box>
                      ) : (
                        <Typography variant="body2" color="text.secondary">
                          {userInfo.access_type === 'write' ? 'Düzenleyen' : 'Görüntüleyen'}
                        </Typography>
                      )
                    }
                  >
                    <ListItemAvatar>
                      <Avatar sx={{ bgcolor: 'secondary.main' }}>
                        <PersonIcon />
                      </Avatar>
                    </ListItemAvatar>
                    <ListItemText
                        primary={
                          <Typography variant="body2" fontWeight={500}>
                            {userInfo.name || userInfo.email} {userInfo.id === user?.id && '(Siz)'}
                          </Typography>
                        }
                      secondary={
                        <Typography variant="caption" color="text.secondary">
                          {userInfo.email}
                        </Typography>
                      }
                    />
                  </ListItem>
                ))}
              </List>
            </Box>

            <Divider sx={{ my: 2 }} />

          </>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose}>Kapat</Button>
      </DialogActions>
    </Dialog>
  );
};

export default ShareDialog;
