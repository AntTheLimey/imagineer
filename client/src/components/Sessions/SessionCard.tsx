// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionCard - Compact session display for lists.
 *
 * Displays session information in a card format with title, date,
 * stage indicator, and action buttons for edit/delete.
 */

import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    Chip,
    IconButton,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Delete as DeleteIcon,
    Edit as EditIcon,
} from '@mui/icons-material';
import type { Session } from '../../types';
import SessionStageIndicator from './SessionStageIndicator';

/**
 * Props for the SessionCard component.
 */
export interface SessionCardProps {
    /** The session to display. */
    session: Session;
    /** Whether this card is currently selected. */
    isSelected: boolean;
    /** Callback fired when the card is clicked. */
    onClick: () => void;
    /** Callback fired when the edit button is clicked. */
    onEdit: () => void;
    /** Callback fired when the delete button is clicked. */
    onDelete: () => void;
}

/**
 * Format a parseable date string into a short, locale-aware date for display (e.g., "Feb 2, 2026").
 *
 * @param dateString - A string accepted by the Date constructor; when omitted, returns an empty string
 * @returns The localized date formatted with a short month, numeric day, and numeric year, or an empty string if `dateString` is falsy
 */
function formatDate(dateString?: string): string {
    if (!dateString) return '';

    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
    });
}

/**
 * Derives a human-readable title for a session.
 *
 * @param session - The session object to derive the title from.
 * @returns The session's `title` if present, otherwise `Session #<number>` when `sessionNumber` is defined, otherwise `Untitled Session`.
 */
function getSessionTitle(session: Session): string {
    if (session.title) {
        return session.title;
    }
    if (session.sessionNumber !== undefined && session.sessionNumber !== null) {
        return `Session #${session.sessionNumber}`;
    }
    return 'Untitled Session';
}

/**
 * Compact card component for displaying session summary.
 *
 * Displays the session title, date, stage indicator, and status
 * with hover effects and action buttons for editing and deleting.
 *
 * @param props - The component props.
 * @returns A React element containing the session card.
 *
 * @example
 * ```tsx
 * <SessionCard
 *     session={session}
 *     isSelected={selectedSessionId === session.id}
 *     onClick={() => handleSelectSession(session.id)}
 *     onEdit={() => handleEditSession(session)}
 *     onDelete={() => handleDeleteSession(session)}
 * />
 * ```
 */
export default function SessionCard({
    session,
    isSelected,
    onClick,
    onEdit,
    onDelete,
}: SessionCardProps) {
    /**
     * Handles edit button click without triggering card click.
     */
    const handleEditClick = (event: React.MouseEvent) => {
        event.stopPropagation();
        onEdit();
    };

    /**
     * Handles delete button click without triggering card click.
     */
    const handleDeleteClick = (event: React.MouseEvent) => {
        event.stopPropagation();
        onDelete();
    };

    const displayDate = formatDate(session.actualDate || session.plannedDate);
    const title = getSessionTitle(session);

    return (
        <Card
            elevation={isSelected ? 4 : 1}
            sx={{
                position: 'relative',
                transition: 'all 0.2s ease-in-out',
                border: isSelected ? 2 : 1,
                borderColor: isSelected ? 'primary.main' : 'divider',
                '&:hover': {
                    boxShadow: 3,
                    transform: 'translateY(-2px)',
                    '& .session-card-actions': {
                        opacity: 1,
                    },
                },
            }}
        >
            {/* Action buttons - shown on hover */}
            <Box
                className="session-card-actions"
                sx={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    display: 'flex',
                    gap: 0.5,
                    opacity: 0,
                    transition: 'opacity 0.2s ease-in-out',
                    zIndex: 1,
                    backgroundColor: 'background.paper',
                    borderRadius: 1,
                    boxShadow: 1,
                }}
            >
                <Tooltip title="Edit session">
                    <IconButton
                        size="small"
                        onClick={handleEditClick}
                        aria-label="edit session"
                    >
                        <EditIcon fontSize="small" />
                    </IconButton>
                </Tooltip>
                <Tooltip title="Delete session">
                    <IconButton
                        size="small"
                        onClick={handleDeleteClick}
                        aria-label="delete session"
                    >
                        <DeleteIcon fontSize="small" />
                    </IconButton>
                </Tooltip>
            </Box>

            <CardActionArea onClick={onClick}>
                <CardContent>
                    {/* Session title */}
                    <Typography
                        variant="subtitle1"
                        component="h3"
                        fontWeight={isSelected ? 600 : 500}
                        noWrap
                        sx={{ pr: 8 }}
                    >
                        {title}
                    </Typography>

                    {/* Date and stage row */}
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                            mt: 1,
                            flexWrap: 'wrap',
                        }}
                    >
                        {displayDate && (
                            <Typography
                                variant="body2"
                                color="text.secondary"
                            >
                                {displayDate}
                            </Typography>
                        )}
                        <SessionStageIndicator stage={session.stage} />
                        {session.status !== 'PLANNED' && (
                            <Chip
                                label={session.status === 'COMPLETED' ? 'Done' : 'Skipped'}
                                size="small"
                                variant="outlined"
                                color={session.status === 'COMPLETED' ? 'success' : 'default'}
                            />
                        )}
                    </Box>
                </CardContent>
            </CardActionArea>
        </Card>
    );
}