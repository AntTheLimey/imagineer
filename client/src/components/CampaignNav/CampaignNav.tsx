// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * CampaignNav - Left navigation panel for campaign dashboard.
 *
 * Provides navigation between different campaign management views including
 * settings, entities, sessions, player characters, and import functionality.
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Box,
    Drawer,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    useMediaQuery,
    useTheme,
} from '@mui/material';
import {
    Menu as MenuIcon,
    Settings as SettingsIcon,
    People as PeopleIcon,
    Timeline as TimelineIcon,
    Person as PersonIcon,
    NoteAdd as NoteAddIcon,
    LibraryBooks as LibraryBooksIcon,
} from '@mui/icons-material';

/**
 * Navigation item configuration.
 */
export type CampaignNavItem =
    | 'overview'
    | 'entities'
    | 'sessions'
    | 'player-characters'
    | 'import-notes'
    | 'import-knowledge';

/**
 * Props for the CampaignNav component.
 */
interface CampaignNavProps {
    /** The campaign ID for navigation */
    campaignId: string;
    /** Currently active navigation item */
    activeItem: CampaignNavItem;
    /** Callback when a navigation item is selected */
    onItemSelect: (item: CampaignNavItem) => void;
    /** Whether there are unsaved changes (shows warning on external navigation) */
    hasUnsavedChanges?: boolean;
    /** Callback to check unsaved changes before navigation */
    onCheckUnsavedChanges?: (callback: () => void) => boolean;
}

/**
 * Width of the navigation drawer in pixels.
 */
const NAV_WIDTH = 240;

/**
 * Navigation menu items configuration.
 */
const NAV_ITEMS: Array<{
    id: CampaignNavItem;
    label: string;
    icon: React.ReactNode;
    type: 'internal' | 'route' | 'external';
    path?: string;
}> = [
    {
        id: 'overview',
        label: 'Overview',
        icon: <SettingsIcon />,
        type: 'internal',
    },
    {
        id: 'entities',
        label: 'Manage Entities',
        icon: <PeopleIcon />,
        type: 'route',
        path: '/entities',
    },
    {
        id: 'sessions',
        label: 'Manage Sessions',
        icon: <TimelineIcon />,
        type: 'internal',
    },
    {
        id: 'player-characters',
        label: 'Player Characters',
        icon: <PersonIcon />,
        type: 'internal',
    },
    {
        id: 'import-notes',
        label: 'Import Campaign Notes',
        icon: <NoteAddIcon />,
        type: 'external',
        path: '/import',
    },
    {
        id: 'import-knowledge',
        label: 'Import Knowledge',
        icon: <LibraryBooksIcon />,
        type: 'external',
        path: '/import',
    },
];

/**
 * Left navigation panel for campaign dashboard with responsive drawer behavior.
 *
 * On mobile screens, displays as a collapsible drawer with a hamburger menu.
 * On larger screens, displays as a permanent fixed-width sidebar.
 *
 * @param campaignId - The campaign ID used for constructing navigation paths
 * @param activeItem - The currently selected navigation item
 * @param onItemSelect - Callback invoked when user selects a navigation item
 * @param hasUnsavedChanges - Whether unsaved changes exist (for navigation warnings)
 * @param onCheckUnsavedChanges - Callback to check/warn about unsaved changes
 * @returns A responsive navigation drawer component
 */
export default function CampaignNav({
    campaignId,
    activeItem,
    onItemSelect,
    hasUnsavedChanges = false,
    onCheckUnsavedChanges,
}: CampaignNavProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const [mobileOpen, setMobileOpen] = useState(false);
    const navigate = useNavigate();

    /**
     * Handle navigation item click.
     */
    const handleItemClick = (item: (typeof NAV_ITEMS)[number]) => {
        const performNavigation = () => {
            if (item.type === 'internal') {
                onItemSelect(item.id);
            } else if (item.type === 'route' && item.path) {
                navigate(`/campaigns/${campaignId}${item.path}`);
            } else if (item.type === 'external' && item.path) {
                navigate(item.path);
            }
            setMobileOpen(false);
        };

        // Check for unsaved changes before navigating away
        if (
            hasUnsavedChanges &&
            onCheckUnsavedChanges &&
            (item.type === 'route' || item.type === 'external')
        ) {
            const blocked = onCheckUnsavedChanges(performNavigation);
            if (!blocked) {
                performNavigation();
            }
        } else {
            performNavigation();
        }
    };

    /**
     * Toggle mobile drawer.
     */
    const handleDrawerToggle = () => {
        setMobileOpen(!mobileOpen);
    };

    /**
     * Navigation list content.
     */
    const navContent = (
        <Box
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            <List sx={{ pt: 2 }}>
                {NAV_ITEMS.map((item) => (
                    <ListItem key={item.id} disablePadding>
                        <ListItemButton
                            selected={activeItem === item.id}
                            onClick={() => handleItemClick(item)}
                            sx={{
                                mx: 1,
                                borderRadius: 1,
                                '&.Mui-selected': {
                                    bgcolor: 'action.selected',
                                    '&:hover': {
                                        bgcolor: 'action.selected',
                                    },
                                },
                            }}
                        >
                            <ListItemIcon sx={{ minWidth: 40 }}>
                                {item.icon}
                            </ListItemIcon>
                            <ListItemText
                                primary={item.label}
                                primaryTypographyProps={{
                                    variant: 'body2',
                                    fontWeight: activeItem === item.id ? 600 : 400,
                                }}
                            />
                        </ListItemButton>
                    </ListItem>
                ))}
            </List>
        </Box>
    );

    return (
        <>
            {/* Mobile menu button */}
            {isMobile && (
                <IconButton
                    color="inherit"
                    aria-label="open navigation menu"
                    edge="start"
                    onClick={handleDrawerToggle}
                    sx={{
                        position: 'fixed',
                        left: 16,
                        top: 80,
                        zIndex: theme.zIndex.drawer + 1,
                        bgcolor: 'background.paper',
                        boxShadow: 2,
                        '&:hover': {
                            bgcolor: 'background.paper',
                        },
                    }}
                >
                    <MenuIcon />
                </IconButton>
            )}

            {/* Mobile drawer */}
            <Drawer
                variant="temporary"
                open={mobileOpen}
                onClose={handleDrawerToggle}
                ModalProps={{
                    keepMounted: true, // Better mobile performance
                }}
                sx={{
                    display: { xs: 'block', md: 'none' },
                    '& .MuiDrawer-paper': {
                        boxSizing: 'border-box',
                        width: NAV_WIDTH,
                        top: 64, // Below app bar
                        height: 'calc(100% - 64px)',
                    },
                }}
            >
                {navContent}
            </Drawer>

            {/* Desktop permanent drawer */}
            <Drawer
                variant="permanent"
                sx={{
                    display: { xs: 'none', md: 'block' },
                    width: NAV_WIDTH,
                    flexShrink: 0,
                    '& .MuiDrawer-paper': {
                        width: NAV_WIDTH,
                        boxSizing: 'border-box',
                        position: 'relative',
                        height: '100%',
                        border: 'none',
                        borderRight: 1,
                        borderColor: 'divider',
                    },
                }}
                open
            >
                {navContent}
            </Drawer>
        </>
    );
}

export { NAV_WIDTH };
