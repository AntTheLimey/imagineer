// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * FullScreenLayout - Full-screen editing layout with header containing
 * back button, breadcrumbs, title, and action buttons.
 */

import { ReactNode } from 'react';
import {
    AppBar,
    Box,
    Breadcrumbs,
    Button,
    Chip,
    CircularProgress,
    IconButton,
    Link,
    Toolbar,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    ArrowBack as BackIcon,
    Save as SaveIcon,
    Done as SaveCloseIcon,
} from '@mui/icons-material';
import { useNavigate, Link as RouterLink } from 'react-router-dom';

/**
 * Breadcrumb item definition.
 */
interface BreadcrumbItem {
    /** Display label for the breadcrumb */
    label: string;
    /** URL path for the breadcrumb link (optional for last item) */
    path?: string;
}

interface FullScreenLayoutProps {
    /** Page title displayed in the header */
    title: string;
    /** Subtitle or additional context (optional) */
    subtitle?: string;
    /** Breadcrumb navigation items */
    breadcrumbs?: BreadcrumbItem[];
    /** Main content to render */
    children: ReactNode;
    /** Whether the form has unsaved changes */
    isDirty?: boolean;
    /** Whether a save operation is in progress */
    isSaving?: boolean;
    /** Handler for the Save button */
    onSave?: () => void;
    /** Handler for the Save & Close button */
    onSaveAndClose?: () => void;
    /** Handler for the Back/Cancel button (defaults to navigate -1) */
    onBack?: () => void;
    /** Path to navigate to when back button is clicked (alternative to onBack) */
    backPath?: string;
    /** Custom action buttons to display in the header */
    actions?: ReactNode;
    /** Whether to show the save buttons (default: true) */
    showSaveButtons?: boolean;
}

/**
 * FullScreenLayout component for editing views.
 *
 * Features:
 * - Fixed header with back button, breadcrumbs, and title
 * - Save and Save & Close action buttons
 * - Dirty state indicator showing unsaved changes
 * - Full-height content area that scrolls
 * - Clean, focused editing experience
 *
 * @param title - The page title
 * @param subtitle - Optional subtitle text
 * @param breadcrumbs - Array of breadcrumb items for navigation
 * @param children - The main editor content
 * @param isDirty - Whether there are unsaved changes
 * @param isSaving - Whether save is in progress
 * @param onSave - Handler for Save button
 * @param onSaveAndClose - Handler for Save & Close button
 * @param onBack - Handler for Back button
 * @param backPath - Alternative path for back navigation
 * @param actions - Custom action buttons
 * @param showSaveButtons - Whether to show save buttons
 * @returns A full-screen editing layout
 */
export default function FullScreenLayout({
    title,
    subtitle,
    breadcrumbs = [],
    children,
    isDirty = false,
    isSaving = false,
    onSave,
    onSaveAndClose,
    onBack,
    backPath,
    actions,
    showSaveButtons = true,
}: FullScreenLayoutProps) {
    const navigate = useNavigate();

    const handleBack = () => {
        if (onBack) {
            onBack();
        } else if (backPath) {
            navigate(backPath);
        } else {
            navigate(-1);
        }
    };

    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                height: '100vh',
                bgcolor: 'background.default',
            }}
        >
            {/* Header */}
            <AppBar
                position="static"
                sx={{
                    bgcolor: 'background.paper',
                    borderBottom: 1,
                    borderColor: 'divider',
                }}
                elevation={0}
            >
                <Toolbar sx={{ gap: 2 }}>
                    {/* Back button */}
                    <Tooltip title="Go back">
                        <IconButton
                            edge="start"
                            onClick={handleBack}
                            aria-label="go back"
                        >
                            <BackIcon />
                        </IconButton>
                    </Tooltip>

                    {/* Title section */}
                    <Box sx={{ flexGrow: 1, minWidth: 0 }}>
                        {/* Breadcrumbs */}
                        {breadcrumbs.length > 0 && (
                            <Breadcrumbs
                                sx={{
                                    '& .MuiBreadcrumbs-ol': {
                                        flexWrap: 'nowrap',
                                    },
                                    mb: 0.5,
                                }}
                            >
                                {breadcrumbs.map((crumb, index) => {
                                    const isLast = index === breadcrumbs.length - 1;
                                    return isLast || !crumb.path ? (
                                        <Typography
                                            key={crumb.label}
                                            variant="caption"
                                            color="text.secondary"
                                            noWrap
                                        >
                                            {crumb.label}
                                        </Typography>
                                    ) : (
                                        <Link
                                            key={crumb.label}
                                            component={RouterLink}
                                            to={crumb.path}
                                            variant="caption"
                                            color="text.secondary"
                                            underline="hover"
                                            noWrap
                                        >
                                            {crumb.label}
                                        </Link>
                                    );
                                })}
                            </Breadcrumbs>
                        )}

                        {/* Title and subtitle */}
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography
                                variant="h6"
                                component="h1"
                                sx={{
                                    fontFamily: 'Cinzel',
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis',
                                    whiteSpace: 'nowrap',
                                }}
                            >
                                {title}
                            </Typography>
                            {subtitle && (
                                <Typography
                                    variant="body2"
                                    color="text.secondary"
                                    sx={{
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap',
                                    }}
                                >
                                    {subtitle}
                                </Typography>
                            )}
                            {/* Dirty state indicator */}
                            {isDirty && (
                                <Chip
                                    label="Unsaved"
                                    size="small"
                                    color="warning"
                                    sx={{ ml: 1 }}
                                />
                            )}
                        </Box>
                    </Box>

                    {/* Custom actions */}
                    {actions}

                    {/* Save buttons */}
                    {showSaveButtons && (
                        <Box sx={{ display: 'flex', gap: 1 }}>
                            {onSave && (
                                <Button
                                    variant="outlined"
                                    startIcon={
                                        isSaving ? (
                                            <CircularProgress size={16} />
                                        ) : (
                                            <SaveIcon />
                                        )
                                    }
                                    onClick={onSave}
                                    disabled={isSaving || !isDirty}
                                >
                                    Save
                                </Button>
                            )}
                            {onSaveAndClose && (
                                <Button
                                    variant="contained"
                                    startIcon={
                                        isSaving ? (
                                            <CircularProgress size={16} color="inherit" />
                                        ) : (
                                            <SaveCloseIcon />
                                        )
                                    }
                                    onClick={onSaveAndClose}
                                    disabled={isSaving}
                                >
                                    Save & Close
                                </Button>
                            )}
                        </Box>
                    )}
                </Toolbar>
            </AppBar>

            {/* Content area */}
            <Box
                component="main"
                sx={{
                    flexGrow: 1,
                    overflow: 'auto',
                    p: 3,
                    bgcolor: 'background.default',
                }}
            >
                {children}
            </Box>
        </Box>
    );
}
