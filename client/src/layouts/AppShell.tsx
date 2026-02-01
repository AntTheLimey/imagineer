// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * AppShell - Outer wrapper with minimal header containing logo, campaign
 * dropdown, and user menu. Provides the top-level structure for the app.
 */

import { useState, ReactNode } from 'react';
import {
    AppBar,
    Avatar,
    Box,
    CircularProgress,
    Divider,
    FormControl,
    IconButton,
    ListItemIcon,
    Menu,
    MenuItem,
    Select,
    SelectChangeEvent,
    Skeleton,
    Toolbar,
    Typography,
} from '@mui/material';
import { Logout as LogoutIcon } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useCampaignContext } from '../contexts/CampaignContext';
import { useCampaigns } from '../hooks';

interface AppShellProps {
    children: ReactNode;
}

/**
 * Derives up to two uppercase initials from a person's name for use as an avatar fallback.
 *
 * @param name - The full name string to derive initials from.
 * @returns `?` if `name` is empty or only whitespace; otherwise two uppercase characters: the first letters of the first two name parts when there are two or more parts, or the first two characters of the trimmed name.
 */
function getInitials(name: string): string {
    if (!name || name.length === 0) {
        return '?';
    }
    const trimmedName = name.trim();
    if (trimmedName.length === 0) {
        return '?';
    }
    const parts = trimmedName.split(' ').filter((part) => part.length > 0);
    if (parts.length >= 2) {
        return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return trimmedName.substring(0, 2).toUpperCase();
}

/**
 * Render the application's outer shell with a fixed header and main content area.
 *
 * The header includes a clickable brand logo, a campaign selector, and a user avatar menu
 * with sign out. Loading states for campaigns are indicated with skeletons and a progress spinner.
 *
 * @param children - The main content to render inside the shell
 * @returns The AppShell element wrapping the provided children
 */
export default function AppShell({ children }: AppShellProps) {
    const navigate = useNavigate();
    const { user, logout } = useAuth();
    const { currentCampaignId, setCurrentCampaignId, isLoading: campaignLoading } =
        useCampaignContext();

    // Fetch all campaigns for the dropdown
    const { data: campaigns, isLoading: campaignsLoading } = useCampaigns();

    // User menu state
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    const handleUserMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const handleUserMenuClose = () => {
        setAnchorEl(null);
    };

    const handleLogout = () => {
        handleUserMenuClose();
        logout();
        navigate('/login');
    };

    const handleCampaignChange = (event: SelectChangeEvent<string>) => {
        const campaignId = event.target.value;
        setCurrentCampaignId(campaignId || null);

        // Navigate to campaign-specific view if a campaign is selected
        if (campaignId) {
            navigate(`/campaigns/${campaignId}/entities`);
        } else {
            navigate('/campaigns');
        }
    };

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
            {/* Header */}
            <AppBar
                position="fixed"
                sx={{
                    zIndex: (theme) => theme.zIndex.drawer + 1,
                    bgcolor: 'background.paper',
                    borderBottom: 1,
                    borderColor: 'divider',
                }}
                elevation={0}
            >
                <Toolbar sx={{ gap: 2 }}>
                    {/* Logo */}
                    <Typography
                        variant="h6"
                        noWrap
                        component="div"
                        sx={{
                            fontFamily: 'Cinzel',
                            color: 'primary.main',
                            cursor: 'pointer',
                        }}
                        onClick={() => navigate('/')}
                    >
                        Imagineer
                    </Typography>

                    {/* Campaign Dropdown */}
                    <FormControl
                        size="small"
                        sx={{ minWidth: 200, ml: 2 }}
                    >
                        {campaignsLoading || campaignLoading ? (
                            <Skeleton
                                variant="rectangular"
                                width={200}
                                height={40}
                                sx={{ borderRadius: 1 }}
                            />
                        ) : (
                            <Select
                                value={currentCampaignId ?? ''}
                                onChange={handleCampaignChange}
                                displayEmpty
                                sx={{
                                    '& .MuiSelect-select': {
                                        py: 1,
                                    },
                                }}
                            >
                                <MenuItem value="">
                                    <em>Select a campaign...</em>
                                </MenuItem>
                                {campaigns?.map((campaign) => (
                                    <MenuItem key={campaign.id} value={campaign.id}>
                                        {campaign.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        )}
                    </FormControl>

                    {/* Spacer */}
                    <Box sx={{ flexGrow: 1 }} />

                    {/* Loading indicator for campaign switch */}
                    {campaignLoading && (
                        <CircularProgress size={24} sx={{ mr: 2 }} />
                    )}

                    {/* User Menu */}
                    {user && (
                        <>
                            <IconButton
                                onClick={handleUserMenuOpen}
                                size="small"
                                aria-controls="user-menu"
                                aria-haspopup="true"
                            >
                                <Avatar
                                    src={user.avatarUrl}
                                    alt={user.name}
                                    sx={{ width: 36, height: 36 }}
                                >
                                    {getInitials(user.name)}
                                </Avatar>
                            </IconButton>
                            <Menu
                                id="user-menu"
                                anchorEl={anchorEl}
                                open={Boolean(anchorEl)}
                                onClose={handleUserMenuClose}
                                anchorOrigin={{
                                    vertical: 'bottom',
                                    horizontal: 'right',
                                }}
                                transformOrigin={{
                                    vertical: 'top',
                                    horizontal: 'right',
                                }}
                            >
                                <MenuItem disabled>
                                    <Box>
                                        <Typography variant="body2">
                                            {user.name}
                                        </Typography>
                                        <Typography
                                            variant="caption"
                                            color="text.secondary"
                                        >
                                            {user.email}
                                        </Typography>
                                    </Box>
                                </MenuItem>
                                <Divider />
                                <MenuItem onClick={handleLogout}>
                                    <ListItemIcon>
                                        <LogoutIcon fontSize="small" />
                                    </ListItemIcon>
                                    Sign out
                                </MenuItem>
                            </Menu>
                        </>
                    )}
                </Toolbar>
            </AppBar>

            {/* Main Content Area */}
            <Box
                component="main"
                sx={{
                    flexGrow: 1,
                    mt: 8, // Account for fixed AppBar
                    display: 'flex',
                    flexDirection: 'column',
                }}
            >
                {children}
            </Box>
        </Box>
    );
}