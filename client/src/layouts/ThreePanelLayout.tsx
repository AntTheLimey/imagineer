// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ThreePanelLayout - Responsive three-panel layout with collapsible side panels.
 *
 * Desktop: All three panels visible
 * Tablet: Left panel in drawer, center and right visible
 * Mobile: Bottom navigation with panels as full-screen views
 */

import { useState, ReactNode } from 'react';
import {
    Box,
    BottomNavigation,
    BottomNavigationAction,
    Drawer,
    IconButton,
    Paper,
    useTheme,
    useMediaQuery,
} from '@mui/material';
import {
    ChevronLeft as ChevronLeftIcon,
    ChevronRight as ChevronRightIcon,
    Menu as MenuIcon,
    ViewSidebar as LeftPanelIcon,
    Dashboard as CenterIcon,
    Info as RightPanelIcon,
} from '@mui/icons-material';

const LEFT_PANEL_WIDTH = 240;
const RIGHT_PANEL_WIDTH = 320;

interface ThreePanelLayoutProps {
    /** Content for the left panel (navigation/list view) */
    leftPanel?: ReactNode;
    /** Content for the center panel (main content) */
    centerPanel: ReactNode;
    /** Content for the right panel (details/properties) */
    rightPanel?: ReactNode;
    /** Default state for left panel collapse (desktop only) */
    leftPanelCollapsed?: boolean;
    /** Default state for right panel collapse (desktop only) */
    rightPanelCollapsed?: boolean;
    /** Callback when left panel collapse state changes */
    onLeftPanelToggle?: (collapsed: boolean) => void;
    /** Callback when right panel collapse state changes */
    onRightPanelToggle?: (collapsed: boolean) => void;
}

/**
 * ThreePanelLayout component providing a responsive three-panel structure.
 *
 * Features:
 * - Desktop (lg+): All three panels side by side with collapse toggles
 * - Tablet (md): Left panel as drawer, center and right visible
 * - Mobile (sm-): Bottom navigation tabs to switch between panels
 *
 * The layout automatically adapts to screen size using MUI breakpoints.
 *
 * @param leftPanel - Content for the left navigation/list panel
 * @param centerPanel - Content for the main center panel (required)
 * @param rightPanel - Content for the right details/properties panel
 * @param leftPanelCollapsed - Initial collapsed state for left panel
 * @param rightPanelCollapsed - Initial collapsed state for right panel
 * @param onLeftPanelToggle - Callback when left panel toggle changes
 * @param onRightPanelToggle - Callback when right panel toggle changes
 * @returns A responsive three-panel layout
 */
export default function ThreePanelLayout({
    leftPanel,
    centerPanel,
    rightPanel,
    leftPanelCollapsed: initialLeftCollapsed = false,
    rightPanelCollapsed: initialRightCollapsed = false,
    onLeftPanelToggle,
    onRightPanelToggle,
}: ThreePanelLayoutProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
    const isTablet = useMediaQuery(theme.breakpoints.between('sm', 'lg'));
    const isDesktop = useMediaQuery(theme.breakpoints.up('lg'));

    // Panel collapse state (desktop)
    const [leftCollapsed, setLeftCollapsed] = useState(initialLeftCollapsed);
    const [rightCollapsed, setRightCollapsed] = useState(initialRightCollapsed);

    // Drawer state (tablet)
    const [leftDrawerOpen, setLeftDrawerOpen] = useState(false);

    // Mobile navigation state
    const [mobileTab, setMobileTab] = useState<'left' | 'center' | 'right'>('center');

    const handleLeftToggle = () => {
        const newState = !leftCollapsed;
        setLeftCollapsed(newState);
        onLeftPanelToggle?.(newState);
    };

    const handleRightToggle = () => {
        const newState = !rightCollapsed;
        setRightCollapsed(newState);
        onRightPanelToggle?.(newState);
    };

    // Mobile layout with bottom navigation
    if (isMobile) {
        return (
            <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
                {/* Content area */}
                <Box sx={{ flexGrow: 1, overflow: 'auto', p: 2 }}>
                    {mobileTab === 'left' && leftPanel}
                    {mobileTab === 'center' && centerPanel}
                    {mobileTab === 'right' && rightPanel}
                </Box>

                {/* Bottom navigation */}
                <Paper
                    sx={{ position: 'fixed', bottom: 0, left: 0, right: 0 }}
                    elevation={3}
                >
                    <BottomNavigation
                        value={mobileTab}
                        onChange={(_, newValue: 'left' | 'center' | 'right') => setMobileTab(newValue)}
                        showLabels
                    >
                        {leftPanel && (
                            <BottomNavigationAction
                                value="left"
                                label="Menu"
                                icon={<LeftPanelIcon />}
                            />
                        )}
                        <BottomNavigationAction
                            value="center"
                            label="Main"
                            icon={<CenterIcon />}
                        />
                        {rightPanel && (
                            <BottomNavigationAction
                                value="right"
                                label="Details"
                                icon={<RightPanelIcon />}
                            />
                        )}
                    </BottomNavigation>
                </Paper>

                {/* Padding for bottom nav */}
                <Box sx={{ height: 56 }} />
            </Box>
        );
    }

    // Tablet layout with drawer for left panel
    if (isTablet) {
        return (
            <Box sx={{ display: 'flex', height: '100%' }}>
                {/* Left panel drawer */}
                {leftPanel && (
                    <Drawer
                        variant="temporary"
                        open={leftDrawerOpen}
                        onClose={() => setLeftDrawerOpen(false)}
                        ModalProps={{ keepMounted: true }}
                        sx={{
                            '& .MuiDrawer-paper': {
                                width: LEFT_PANEL_WIDTH,
                                boxSizing: 'border-box',
                                top: 64, // Below AppBar
                                height: 'calc(100% - 64px)',
                            },
                        }}
                    >
                        <Box sx={{ p: 2, height: '100%', overflow: 'auto' }}>
                            {leftPanel}
                        </Box>
                    </Drawer>
                )}

                {/* Center panel */}
                <Box
                    sx={{
                        flexGrow: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        overflow: 'hidden',
                    }}
                >
                    {/* Menu button for drawer */}
                    {leftPanel && (
                        <Box sx={{ p: 1 }}>
                            <IconButton
                                onClick={() => setLeftDrawerOpen(true)}
                                size="small"
                            >
                                <MenuIcon />
                            </IconButton>
                        </Box>
                    )}
                    <Box sx={{ flexGrow: 1, overflow: 'auto', p: 2 }}>
                        {centerPanel}
                    </Box>
                </Box>

                {/* Right panel */}
                {rightPanel && (
                    <Box
                        sx={{
                            width: RIGHT_PANEL_WIDTH,
                            borderLeft: 1,
                            borderColor: 'divider',
                            overflow: 'auto',
                            p: 2,
                        }}
                    >
                        {rightPanel}
                    </Box>
                )}
            </Box>
        );
    }

    // Desktop layout with all three panels
    return (
        <Box sx={{ display: 'flex', height: '100%' }}>
            {/* Left panel */}
            {leftPanel && (
                <Box
                    sx={{
                        width: leftCollapsed ? 0 : LEFT_PANEL_WIDTH,
                        flexShrink: 0,
                        borderRight: leftCollapsed ? 0 : 1,
                        borderColor: 'divider',
                        overflow: 'hidden',
                        transition: theme.transitions.create('width', {
                            easing: theme.transitions.easing.sharp,
                            duration: theme.transitions.duration.enteringScreen,
                        }),
                    }}
                >
                    <Box
                        sx={{
                            width: LEFT_PANEL_WIDTH,
                            height: '100%',
                            overflow: 'auto',
                            p: 2,
                        }}
                    >
                        {leftPanel}
                    </Box>
                </Box>
            )}

            {/* Left collapse toggle */}
            {leftPanel && isDesktop && (
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        borderRight: 1,
                        borderColor: 'divider',
                    }}
                >
                    <IconButton
                        onClick={handleLeftToggle}
                        size="small"
                        sx={{ borderRadius: 0, height: '100%', width: 24 }}
                    >
                        {leftCollapsed ? <ChevronRightIcon /> : <ChevronLeftIcon />}
                    </IconButton>
                </Box>
            )}

            {/* Center panel */}
            <Box
                sx={{
                    flexGrow: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    overflow: 'hidden',
                    minWidth: 0, // Prevent flex item from overflowing
                }}
            >
                <Box sx={{ flexGrow: 1, overflow: 'auto', p: 3 }}>
                    {centerPanel}
                </Box>
            </Box>

            {/* Right collapse toggle */}
            {rightPanel && isDesktop && (
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        borderLeft: 1,
                        borderColor: 'divider',
                    }}
                >
                    <IconButton
                        onClick={handleRightToggle}
                        size="small"
                        sx={{ borderRadius: 0, height: '100%', width: 24 }}
                    >
                        {rightCollapsed ? <ChevronLeftIcon /> : <ChevronRightIcon />}
                    </IconButton>
                </Box>
            )}

            {/* Right panel */}
            {rightPanel && (
                <Box
                    sx={{
                        width: rightCollapsed ? 0 : RIGHT_PANEL_WIDTH,
                        flexShrink: 0,
                        borderLeft: rightCollapsed ? 0 : 1,
                        borderColor: 'divider',
                        overflow: 'hidden',
                        transition: theme.transitions.create('width', {
                            easing: theme.transitions.easing.sharp,
                            duration: theme.transitions.duration.enteringScreen,
                        }),
                    }}
                >
                    <Box
                        sx={{
                            width: RIGHT_PANEL_WIDTH,
                            height: '100%',
                            overflow: 'auto',
                            p: 2,
                        }}
                    >
                        {rightPanel}
                    </Box>
                </Box>
            )}
        </Box>
    );
}
