// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionList - Displays sessions for a chapter with stage indicators.
 *
 * Features:
 * - List sessions in chronological order
 * - Show session title, date, stage indicator
 * - Visual indicator for current stage (prep/play/wrap_up)
 * - Click to select session
 * - Add new session button
 */

import type { MouseEvent } from 'react';
import { useState, useCallback } from 'react';
import {
    Alert,
    Box,
    Button,
    CircularProgress,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
} from '@mui/icons-material';
import { useSessionsByChapter, useDeleteSession } from '../../hooks/useSessions';
import type { Session } from '../../types';
import SessionStageIndicator from './SessionStageIndicator';

/**
 * Props for the SessionList component.
 */
export interface SessionListProps {
    /** The campaign ID. */
    campaignId: number;
    /** The chapter ID to list sessions for. */
    chapterId: number;
    /** The currently selected session ID. */
    selectedSessionId?: number;
    /** Callback fired when a session is selected. */
    onSelectSession: (sessionId: number) => void;
    /** Optional callback to open the create session dialog. */
    onCreateSession?: () => void;
    /** Optional callback to open the edit session dialog. */
    onEditSession?: (session: Session) => void;
}

/**
 * Props for the SessionListItem component.
 */
interface SessionListItemProps {
    /** The session to display. */
    session: Session;
    /** Whether this session is selected. */
    isSelected: boolean;
    /** Callback fired when the session is clicked. */
    onSelect: () => void;
    /** Callback fired when the edit button is clicked. */
    onEdit: () => void;
    /** Callback fired when the delete button is clicked. */
    onDelete: () => void;
    /** Whether a delete operation is in progress. */
    isDeleting: boolean;
}

/**
 * Format a date string into a short month/day representation (e.g., "Feb 3").
 *
 * @param dateString - An optional date string parseable by `Date`. If omitted or falsy, the function returns an empty string.
 * @returns The formatted date as "Mon D" (short month and numeric day), or an empty string when `dateString` is falsy.
 */
function formatDate(dateString?: string): string {
    if (!dateString) return '';

    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
    });
}

/**
 * Gets a human-readable title for a session.
 *
 * Prefers the session's explicit `title`; if absent, uses `Session #{sessionNumber}` when
 * `sessionNumber` is present; otherwise returns `"Untitled Session"`.
 *
 * @param session - The session object to derive the title from
 * @returns The display title for the session
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
 * Renders a single session row with title, stage indicator, date, and edit/delete actions.
 *
 * @param session - The session to display
 * @param isSelected - Whether the session is currently selected (affects styling)
 * @param onSelect - Called when the item is clicked to select the session
 * @param onEdit - Called when the edit action is triggered
 * @param onDelete - Called when the delete action is triggered
 * @param isDeleting - If true, disables the delete button and shows a progress indicator
 * @returns The list item element representing the session
 */
function SessionListItem({
    session,
    isSelected,
    onSelect,
    onEdit,
    onDelete,
    isDeleting,
}: SessionListItemProps) {
    const handleEditClick = (event: MouseEvent) => {
        event.stopPropagation();
        onEdit();
    };

    const handleDeleteClick = (event: MouseEvent) => {
        event.stopPropagation();
        onDelete();
    };

    const displayDate = formatDate(session.actualDate || session.plannedDate);
    const title = getSessionTitle(session);

    return (
        <ListItem
            disablePadding
            secondaryAction={
                <Box sx={{ display: 'flex', gap: 0.5 }}>
                    <Tooltip title="Edit session">
                        <IconButton
                            edge="end"
                            size="small"
                            onClick={handleEditClick}
                            aria-label="edit session"
                        >
                            <EditIcon fontSize="small" />
                        </IconButton>
                    </Tooltip>
                    <Tooltip title="Delete session">
                        <IconButton
                            edge="end"
                            size="small"
                            onClick={handleDeleteClick}
                            disabled={isDeleting}
                            aria-label="delete session"
                        >
                            {isDeleting ? (
                                <CircularProgress size={18} />
                            ) : (
                                <DeleteIcon fontSize="small" />
                            )}
                        </IconButton>
                    </Tooltip>
                </Box>
            }
        >
            <ListItemButton
                selected={isSelected}
                onClick={onSelect}
                sx={{ pr: 10 }}
            >
                <ListItemText
                    primary={
                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 1,
                            }}
                        >
                            <Typography
                                variant="body1"
                                component="span"
                                noWrap
                                fontWeight={isSelected ? 600 : 400}
                                sx={{ flexGrow: 1, minWidth: 0 }}
                            >
                                {title}
                            </Typography>
                            <SessionStageIndicator
                                stage={session.stage}
                                size="small"
                            />
                        </Box>
                    }
                    secondary={displayDate}
                    primaryTypographyProps={{
                        component: 'div',
                    }}
                />
            </ListItemButton>
        </ListItem>
    );
}

/**
 * Component for displaying and managing a list of sessions for a chapter.
 *
 * Provides session selection, stage indicators, and actions for
 * creating, editing, and deleting sessions.
 *
 * @param props - The component props.
 * @returns A React element containing the session list.
 *
 * @example
 * ```tsx
 * <SessionList
 *     campaignId={campaignId}
 *     chapterId={chapterId}
 *     selectedSessionId={selectedSessionId}
 *     onSelectSession={handleSelectSession}
 *     onCreateSession={() => setEditorOpen(true)}
 *     onEditSession={handleEditSession}
 * />
 * ```
 */
export default function SessionList({
    campaignId,
    chapterId,
    selectedSessionId,
    onSelectSession,
    onCreateSession,
    onEditSession,
}: SessionListProps) {
    const [deleteError, setDeleteError] = useState<string | null>(null);

    const {
        data: sessions,
        isLoading,
        error: fetchError,
    } = useSessionsByChapter(campaignId, chapterId);

    const deleteSessionMutation = useDeleteSession();

    /**
     * Handles deleting a session with confirmation.
     */
    const handleDeleteSession = useCallback(
        async (session: Session) => {
            const title = getSessionTitle(session);
            const confirmed = window.confirm(
                `Are you sure you want to delete "${title}"? ` +
                    'This action cannot be undone.'
            );

            if (!confirmed) return;

            setDeleteError(null);

            try {
                await deleteSessionMutation.mutateAsync({
                    campaignId,
                    sessionId: session.id,
                });
            } catch (err) {
                const message =
                    err instanceof Error
                        ? err.message
                        : 'Failed to delete session';
                setDeleteError(message);
            }
        },
        [campaignId, deleteSessionMutation]
    );

    // Sort sessions by planned/actual date, then by session number
    const sortedSessions = sessions
        ? [...sessions].sort((a, b) => {
              const dateA = a.actualDate || a.plannedDate || '';
              const dateB = b.actualDate || b.plannedDate || '';
              if (dateA !== dateB) {
                  return dateA.localeCompare(dateB);
              }
              return (a.sessionNumber ?? 0) - (b.sessionNumber ?? 0);
          })
        : [];

    // Loading state
    if (isLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    py: 4,
                }}
            >
                <CircularProgress size={24} />
            </Box>
        );
    }

    // Error state
    if (fetchError) {
        return (
            <Alert severity="error" sx={{ m: 2 }}>
                Failed to load sessions: {fetchError.message}
            </Alert>
        );
    }

    return (
        <Box>
            {/* Header with add button */}
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    px: 2,
                    py: 1,
                    borderBottom: 1,
                    borderColor: 'divider',
                }}
            >
                <Typography variant="subtitle1" fontWeight={500}>
                    Sessions
                </Typography>
                {onCreateSession && (
                    <Button
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={onCreateSession}
                    >
                        Add
                    </Button>
                )}
            </Box>

            {/* Delete error alert */}
            {deleteError && (
                <Alert
                    severity="error"
                    sx={{ m: 1 }}
                    onClose={() => setDeleteError(null)}
                >
                    {deleteError}
                </Alert>
            )}

            {/* Session list */}
            {sortedSessions.length > 0 ? (
                <List disablePadding>
                    {sortedSessions.map((session) => (
                        <SessionListItem
                            key={session.id}
                            session={session}
                            isSelected={selectedSessionId === session.id}
                            onSelect={() => onSelectSession(session.id)}
                            onEdit={() => onEditSession?.(session)}
                            onDelete={() => handleDeleteSession(session)}
                            isDeleting={
                                deleteSessionMutation.isPending &&
                                deleteSessionMutation.variables?.sessionId ===
                                    session.id
                            }
                        />
                    ))}
                </List>
            ) : (
                <Box sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                        No sessions yet. Click &quot;Add&quot; to create your
                        first session.
                    </Typography>
                </Box>
            )}
        </Box>
    );
}