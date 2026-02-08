// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * AppShell - Main application layout with persistent left navigation,
 * campaign dropdown selector, and user menu.
 *
 * Provides the outer shell for all authenticated pages with:
 * - Fixed header with logo, campaign selector, and user menu
 * - Persistent left navigation drawer with campaign-specific items
 * - Main content area for page rendering
 */

import { useState, useEffect, useMemo, ReactNode } from 'react';
import {
    AppBar,
    Avatar,
    Box,
    CircularProgress,
    Collapse,
    Divider,
    Drawer,
    FormControl,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    Select,
    SelectChangeEvent,
    Skeleton,
    Toolbar,
    Typography,
    useMediaQuery,
    useTheme,
} from '@mui/material';
import {
    Add as AddIcon,
    Business as OrganizationIcon,
    Category as OtherIcon,
    Dashboard as OverviewIcon,
    Description as DocumentIcon,
    Event as EventIcon,
    ExpandLess,
    ExpandMore,
    FileUpload as ImportIcon,
    Groups as FactionIcon,
    Inventory as ItemIcon,
    List as AllIcon,
    Logout as LogoutIcon,
    Menu as MenuIcon,
    People as EntitiesIcon,
    Person as NpcIcon,
    Pets as CreatureIcon,
    Place as LocationIcon,
    Search as ClueIcon,
    Settings as SettingsIcon,
    EventNote as SessionsIcon,
    Timeline as TimelineIcon,
} from '@mui/icons-material';
import { useNavigate, useLocation, useParams, useSearchParams } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useCampaignContext } from '../contexts/CampaignContext';
import { useEntities } from '../hooks';
import type { EntityType } from '../types';

/**
 * Width of the navigation drawer in pixels.
 */
const DRAWER_WIDTH = 240;

interface AppShellProps {
    children: ReactNode;
}

/**
 * Navigation item configuration.
 */
interface NavItem {
    id: string;
    label: string;
    icon: React.ReactNode;
    path: string;
    requiresCampaign: boolean;
    hasChildren?: boolean;
}

/**
 * Entity type configuration for navigation sub-items.
 */
interface EntityTypeNavConfig {
    type: EntityType | 'all';
    label: string;
    icon: React.ReactNode;
}

/**
 * Entity type navigation items with labels and icons.
 */
const ENTITY_TYPE_NAV_ITEMS: EntityTypeNavConfig[] = [
    { type: 'all', label: 'All Entities', icon: <AllIcon /> },
    { type: 'npc', label: 'NPCs', icon: <NpcIcon /> },
    { type: 'location', label: 'Locations', icon: <LocationIcon /> },
    { type: 'faction', label: 'Factions', icon: <FactionIcon /> },
    { type: 'item', label: 'Items', icon: <ItemIcon /> },
    { type: 'event', label: 'Events', icon: <EventIcon /> },
    { type: 'clue', label: 'Clues', icon: <ClueIcon /> },
    { type: 'creature', label: 'Creatures', icon: <CreatureIcon /> },
    { type: 'organization', label: 'Organizations', icon: <OrganizationIcon /> },
    { type: 'document', label: 'Documents', icon: <DocumentIcon /> },
    { type: 'other', label: 'Other', icon: <OtherIcon /> },
];

/**
 * Navigation menu items for campaign views.
 */
const NAV_ITEMS: NavItem[] = [
    {
        id: 'overview',
        label: 'Overview',
        icon: <OverviewIcon />,
        path: '/campaigns/:id/overview',
        requiresCampaign: true,
    },
    {
        id: 'entities',
        label: 'Entities',
        icon: <EntitiesIcon />,
        path: '/campaigns/:id/entities',
        requiresCampaign: true,
        hasChildren: true,
    },
    {
        id: 'sessions',
        label: 'Sessions',
        icon: <SessionsIcon />,
        path: '/campaigns/:id/sessions',
        requiresCampaign: true,
    },
    {
        id: 'import',
        label: 'Import',
        icon: <ImportIcon />,
        path: '/campaigns/:id/import',
        requiresCampaign: true,
    },
    {
        id: 'timeline',
        label: 'Timeline',
        icon: <TimelineIcon />,
        path: '/campaigns/:id/timeline',
        requiresCampaign: true,
    },
];

/**
 * Derives up to two uppercase initials from a person's name for use as an
 * avatar fallback.
 *
 * @param name - The full name string to derive initials from.
 * @returns Two uppercase characters from the name, or '?' if empty.
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
 * Render the application shell with persistent left navigation,
 * campaign selector, and user menu.
 *
 * The left navigation shows campaign-specific items when a campaign is
 * selected. The campaign dropdown allows switching between campaigns
 * and creating new ones.
 *
 * @param children - The main content to render inside the shell
 * @returns The AppShell element wrapping the provided children
 */
export default function AppShell({ children }: AppShellProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const navigate = useNavigate();
    const location = useLocation();
    const [searchParams] = useSearchParams();
    const { id: routeCampaignIdParam } = useParams<{ id: string }>();
    const routeCampaignId = routeCampaignIdParam ? Number(routeCampaignIdParam) : undefined;

    const { user, logout } = useAuth();
    const {
        currentCampaignId,
        setCurrentCampaignId,
        isLoading: campaignLoading,
        campaigns,
        campaignsLoading,
    } = useCampaignContext();

    // Mobile drawer state
    const [mobileOpen, setMobileOpen] = useState(false);

    // User menu state
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

    // Entities tree expanded state
    const [entitiesExpanded, setEntitiesExpanded] = useState(false);

    // Effective campaign ID: prefer route parameter over stored value
    const effectiveCampaignId = routeCampaignId ?? currentCampaignId;

    // Check if we're on the entities page
    const isOnEntitiesPage = effectiveCampaignId
        ? location.pathname.startsWith(`/campaigns/${effectiveCampaignId}/entities`)
        : false;

    // Auto-expand entities when on the entities page
    useEffect(() => {
        if (isOnEntitiesPage) {
            setEntitiesExpanded(true);
        }
    }, [isOnEntitiesPage]);

    // Fetch entities for counts (only when campaign is selected)
    const { data: entities, isLoading: entitiesLoading } = useEntities({
        campaignId: effectiveCampaignId ?? 0,
        pageSize: 1000,
    });

    // Calculate entity counts by type
    const typeCounts = useMemo(() => {
        const counts: Record<EntityType | 'all', number> = {
            all: 0,
            npc: 0,
            location: 0,
            item: 0,
            faction: 0,
            clue: 0,
            creature: 0,
            organization: 0,
            event: 0,
            document: 0,
            other: 0,
        };

        if (!entities) return counts;

        for (const entity of entities) {
            counts[entity.entityType]++;
            counts.all++;
        }

        return counts;
    }, [entities]);

    // Sync campaign ID from route to context/localStorage
    // When user navigates directly to a campaign URL, update the stored selection
    useEffect(() => {
        if (routeCampaignId && routeCampaignId !== currentCampaignId) {
            // Verify the campaign exists in the user's list before setting
            if (campaigns && campaigns.some(c => c.id === routeCampaignId)) {
                setCurrentCampaignId(routeCampaignId);
            }
        }
    }, [routeCampaignId, currentCampaignId, campaigns, setCurrentCampaignId]);

    const handleDrawerToggle = () => {
        setMobileOpen(!mobileOpen);
    };

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

    const handleSettings = () => {
        handleUserMenuClose();
        navigate('/settings');
    };

    const handleCampaignChange = (event: SelectChangeEvent<string>) => {
        const value = event.target.value;

        if (value === '__create__') {
            // Navigate to create campaign page
            navigate('/campaigns/new');
            return;
        }

        const numericId = value ? Number(value) : null;
        setCurrentCampaignId(numericId);

        // Navigate to campaign overview if a campaign is selected
        if (numericId) {
            navigate(`/campaigns/${numericId}/overview`);
        }
    };

    const handleNavItemClick = (item: NavItem) => {
        if (item.requiresCampaign && effectiveCampaignId) {
            const path = item.path.replace(':id', String(effectiveCampaignId));
            navigate(path);
        }
        setMobileOpen(false);
    };

    const handleLogoClick = () => {
        if (effectiveCampaignId) {
            navigate(`/campaigns/${effectiveCampaignId}/overview`);
        } else {
            navigate('/');
        }
    };

    /**
     * Check if a nav item is currently active.
     */
    const isNavItemActive = (item: NavItem): boolean => {
        if (!effectiveCampaignId) return false;
        const itemPath = item.path.replace(':id', String(effectiveCampaignId));
        // For entities, only mark as active if we're on the base entities path
        // (not viewing a specific entity type via query param)
        if (item.id === 'entities') {
            return location.pathname === itemPath && !searchParams.has('type');
        }
        return location.pathname === itemPath ||
            location.pathname.startsWith(itemPath + '/');
    };

    /**
     * Check if an entity type nav item is currently active.
     */
    const isEntityTypeActive = (type: EntityType | 'all'): boolean => {
        if (!effectiveCampaignId || !isOnEntitiesPage) return false;
        const currentType = searchParams.get('type');
        if (type === 'all') {
            return !currentType;
        }
        return currentType === type;
    };

    /**
     * Handle clicking on entity type sub-item.
     */
    const handleEntityTypeClick = (type: EntityType | 'all') => {
        if (!effectiveCampaignId) return;
        if (type === 'all') {
            navigate(`/campaigns/${effectiveCampaignId}/entities`);
        } else {
            navigate(`/campaigns/${effectiveCampaignId}/entities?type=${type}`);
        }
        setMobileOpen(false);
    };

    /**
     * Handle clicking on the Entities parent item.
     */
    const handleEntitiesClick = () => {
        setEntitiesExpanded(!entitiesExpanded);
    };

    /**
     * Render a standard navigation item.
     */
    const renderNavItem = (item: NavItem) => {
        // Special handling for entities - it has children
        if (item.id === 'entities') {
            return (
                <Box key={item.id}>
                    <ListItem disablePadding>
                        <ListItemButton
                            onClick={handleEntitiesClick}
                            sx={{
                                mx: 1,
                                borderRadius: 1,
                                bgcolor: isOnEntitiesPage ? 'action.selected' : 'transparent',
                                '&:hover': {
                                    bgcolor: isOnEntitiesPage
                                        ? 'action.selected'
                                        : 'action.hover',
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
                                    fontWeight: isOnEntitiesPage ? 600 : 400,
                                }}
                            />
                            {entitiesExpanded ? <ExpandLess /> : <ExpandMore />}
                        </ListItemButton>
                    </ListItem>
                    <Collapse in={entitiesExpanded} timeout="auto" unmountOnExit>
                        <List component="div" disablePadding>
                            {ENTITY_TYPE_NAV_ITEMS.map((entityType) => {
                                const isActive = isEntityTypeActive(entityType.type);
                                const count = typeCounts[entityType.type];

                                return (
                                    <ListItem key={entityType.type} disablePadding>
                                        <ListItemButton
                                            selected={isActive}
                                            onClick={() => handleEntityTypeClick(entityType.type)}
                                            sx={{
                                                pl: 4,
                                                mx: 1,
                                                borderRadius: 1,
                                                '&.Mui-selected': {
                                                    bgcolor: 'primary.main',
                                                    color: 'primary.contrastText',
                                                    '&:hover': {
                                                        bgcolor: 'primary.dark',
                                                    },
                                                    '& .MuiListItemIcon-root': {
                                                        color: 'primary.contrastText',
                                                    },
                                                },
                                            }}
                                        >
                                            <ListItemIcon
                                                sx={{
                                                    minWidth: 36,
                                                    color: isActive ? 'inherit' : 'text.secondary',
                                                }}
                                            >
                                                {entityType.icon}
                                            </ListItemIcon>
                                            <ListItemText
                                                primary={entityType.label}
                                                primaryTypographyProps={{
                                                    variant: 'body2',
                                                    fontWeight: isActive ? 600 : 400,
                                                }}
                                            />
                                            {entitiesLoading ? (
                                                <Skeleton
                                                    variant="circular"
                                                    width={20}
                                                    height={20}
                                                />
                                            ) : (
                                                <Typography
                                                    variant="body2"
                                                    sx={{
                                                        color: isActive ? 'inherit' : 'text.secondary',
                                                        fontWeight: 500,
                                                        minWidth: 20,
                                                        textAlign: 'right',
                                                        fontSize: '0.75rem',
                                                    }}
                                                >
                                                    {count}
                                                </Typography>
                                            )}
                                        </ListItemButton>
                                    </ListItem>
                                );
                            })}
                        </List>
                    </Collapse>
                </Box>
            );
        }

        // Standard nav item
        return (
            <ListItem key={item.id} disablePadding>
                <ListItemButton
                    selected={isNavItemActive(item)}
                    onClick={() => handleNavItemClick(item)}
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
                            fontWeight: isNavItemActive(item) ? 600 : 400,
                        }}
                    />
                </ListItemButton>
            </ListItem>
        );
    };

    /**
     * Navigation drawer content.
     */
    const drawerContent = (
        <Box
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            {/* Spacer for toolbar */}
            <Toolbar />

            {/* Navigation items */}
            <List sx={{ flexGrow: 1, pt: 1 }}>
                {effectiveCampaignId ? (
                    NAV_ITEMS.map((item) => renderNavItem(item))
                ) : (
                    <Box sx={{ p: 2, textAlign: 'center' }}>
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ mb: 1 }}
                        >
                            Select a campaign to get started
                        </Typography>
                    </Box>
                )}
            </List>
        </Box>
    );

    return (
        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
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
                    {/* Mobile menu button */}
                    {isMobile && (
                        <IconButton
                            color="inherit"
                            edge="start"
                            onClick={handleDrawerToggle}
                            sx={{ mr: 1 }}
                        >
                            <MenuIcon />
                        </IconButton>
                    )}

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
                        onClick={handleLogoClick}
                    >
                        Imagineer
                    </Typography>

                    {/* Campaign Dropdown */}
                    <FormControl size="small" sx={{ minWidth: 200, ml: 2 }}>
                        {campaignsLoading || campaignLoading ? (
                            <Skeleton
                                variant="rectangular"
                                width={200}
                                height={40}
                                sx={{ borderRadius: 1 }}
                            />
                        ) : (
                            <Select
                                value={effectiveCampaignId ? String(effectiveCampaignId) : ''}
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
                                    <MenuItem key={campaign.id} value={String(campaign.id)}>
                                        {campaign.name}
                                    </MenuItem>
                                ))}
                                <Divider />
                                <MenuItem value="__create__">
                                    <ListItemIcon sx={{ minWidth: 36 }}>
                                        <AddIcon fontSize="small" />
                                    </ListItemIcon>
                                    Create New Campaign
                                </MenuItem>
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
                                <MenuItem onClick={handleSettings}>
                                    <ListItemIcon>
                                        <SettingsIcon fontSize="small" />
                                    </ListItemIcon>
                                    Account Settings
                                </MenuItem>
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

            {/* Left Navigation Drawer - Mobile */}
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
                        width: DRAWER_WIDTH,
                    },
                }}
            >
                {drawerContent}
            </Drawer>

            {/* Left Navigation Drawer - Desktop */}
            <Drawer
                variant="permanent"
                sx={{
                    display: { xs: 'none', md: 'block' },
                    width: DRAWER_WIDTH,
                    flexShrink: 0,
                    '& .MuiDrawer-paper': {
                        width: DRAWER_WIDTH,
                        boxSizing: 'border-box',
                        borderRight: 1,
                        borderColor: 'divider',
                    },
                }}
                open
            >
                {drawerContent}
            </Drawer>

            {/* Main Content Area */}
            <Box
                component="main"
                sx={{
                    flexGrow: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    minHeight: '100vh',
                    width: { md: `calc(100% - ${DRAWER_WIDTH}px)` },
                }}
            >
                {/* Spacer for fixed AppBar */}
                <Toolbar />

                {/* Content */}
                <Box
                    sx={{
                        flexGrow: 1,
                        p: 3,
                        overflow: 'auto',
                    }}
                >
                    {children}
                </Box>
            </Box>
        </Box>
    );
}

export { DRAWER_WIDTH };
