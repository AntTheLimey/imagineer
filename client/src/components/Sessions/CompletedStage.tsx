/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * CompletedStage - Read-only view of a completed session.
 *
 * Displays session details, notes, scenes, and entities in a two-column
 * layout with an informational alert and a button to reopen the session.
 */

import {
    Alert,
    Box,
    Button,
    Chip,
    Divider,
    Paper,
    Typography,
} from '@mui/material';
import { LockOpen as LockOpenIcon } from '@mui/icons-material';
import type { Scene } from '../../api/scenes';
import type { Entity, SessionStatus } from '../../types';
import { MarkdownRenderer } from '../MarkdownRenderer';
import SceneSummaryList from './SceneSummaryList';

/**
 * Props for the CompletedStage component.
 */
export interface CompletedStageProps {
    /** The campaign this session belongs to. */
    campaignId: number;
    /** The session title. */
    title: string;
    /** The session number, if assigned. */
    sessionNumber: number | null;
    /** The chapter name, if the session belongs to a chapter. */
    chapterName: string | null;
    /** The planned date for the session. */
    plannedDate: string;
    /** The actual date the session was played. */
    actualDate: string;
    /** The session status (COMPLETED or SKIPPED). */
    status: SessionStatus;
    /** Post-session notes written during wrap-up. */
    actualNotes: string;
    /** Prep notes written before the session. */
    prepNotes: string;
    /** Freeform play notes captured during the session. */
    playNotes: string;
    /** Scenes associated with this session. */
    scenes: Scene[];
    /** Entities linked to this session for wiki-link resolution. */
    entities: Entity[];
    /** Callback fired when a wiki-link entity is clicked. */
    onEntityNavigate?: (entityId: number) => void;
    /** Callback fired when the GM wants to reopen this session. */
    onReopen: () => void;
}

/**
 * Returns a MUI Chip color based on session status.
 */
function statusChipColor(
    status: SessionStatus
): 'success' | 'warning' | 'default' {
    switch (status) {
        case 'COMPLETED':
            return 'success';
        case 'SKIPPED':
            return 'warning';
        default:
            return 'default';
    }
}

/**
 * Formats a date string for display. Returns an em dash if empty.
 */
function formatDate(date: string): string {
    if (!date) return '\u2014';
    return date.split('T')[0];
}

/**
 * Builds the subtitle line from session number and chapter name.
 */
function buildSubtitle(
    sessionNumber: number | null,
    chapterName: string | null
): string | null {
    const parts: string[] = [];
    if (sessionNumber != null) {
        parts.push(`Session #${sessionNumber}`);
    }
    if (chapterName) {
        parts.push(`Chapter: ${chapterName}`);
    }
    return parts.length > 0 ? parts.join(' | ') : null;
}

/**
 * Read-only view for a completed (or skipped) session.
 *
 * All content is displayed as static text with no editing controls.
 * The only interactive element is the Reopen button in the info alert.
 *
 * @param props - The component props.
 * @returns A React element containing the completed session view.
 */
export default function CompletedStage({
    title,
    sessionNumber,
    chapterName,
    plannedDate,
    actualDate,
    status,
    actualNotes,
    prepNotes,
    playNotes,
    scenes,
    entities,
    onEntityNavigate,
    onReopen,
}: CompletedStageProps) {
    const subtitle = buildSubtitle(sessionNumber, chapterName);

    return (
        <Box>
            <Alert
                severity="info"
                sx={{ mb: 3 }}
                action={
                    <Button
                        startIcon={<LockOpenIcon />}
                        size="small"
                        onClick={onReopen}
                    >
                        Reopen Session
                    </Button>
                }
            >
                This session is completed. It is now read-only.
            </Alert>

            <Box sx={{ display: 'flex', gap: 3 }}>
                {/* Left column - main session details */}
                <Paper sx={{ p: 3, flex: '1 1 60%' }}>
                    <Typography variant="h5">{title}</Typography>
                    {subtitle && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ mb: 2 }}
                        >
                            {subtitle}
                        </Typography>
                    )}
                    <Box sx={{ display: 'flex', gap: 2 }}>
                        <Typography variant="body2" color="text.secondary">
                            Planned: {formatDate(plannedDate)}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            Actual: {formatDate(actualDate)}
                        </Typography>
                    </Box>
                    <Chip
                        label={status}
                        color={statusChipColor(status)}
                        size="small"
                        sx={{ mb: 2, mt: 1 }}
                    />
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="h6" sx={{ mb: 1 }}>
                        Session Notes
                    </Typography>
                    {actualNotes ? (
                        <MarkdownRenderer
                            content={actualNotes}
                            entities={entities}
                            onEntityNavigate={onEntityNavigate}
                        />
                    ) : (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ fontStyle: 'italic' }}
                        >
                            No session notes recorded.
                        </Typography>
                    )}
                </Paper>

                {/* Right column - prep notes, play notes, scenes */}
                <Paper sx={{ p: 3, flex: '1 1 40%', minWidth: 280 }}>
                    <Typography variant="h6" sx={{ mb: 1 }}>
                        Prep Notes
                    </Typography>
                    {prepNotes ? (
                        <MarkdownRenderer
                            content={prepNotes}
                            entities={entities}
                            onEntityNavigate={onEntityNavigate}
                        />
                    ) : (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ fontStyle: 'italic' }}
                        >
                            No prep notes.
                        </Typography>
                    )}
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="h6" sx={{ mb: 1 }}>
                        Play Notes
                    </Typography>
                    {playNotes ? (
                        <Typography
                            variant="body2"
                            sx={{ whiteSpace: 'pre-wrap' }}
                        >
                            {playNotes}
                        </Typography>
                    ) : (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ fontStyle: 'italic' }}
                        >
                            No play notes.
                        </Typography>
                    )}
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="h6" sx={{ mb: 1 }}>
                        Scenes
                    </Typography>
                    <SceneSummaryList scenes={scenes} />
                </Paper>
            </Box>
        </Box>
    );
}
