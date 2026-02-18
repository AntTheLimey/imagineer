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
 * WrapUpStage - Main wrap-up editing layout for a session.
 *
 * Provides a two-column layout with an editable date and Markdown notes
 * area on the left and a read-only reference panel on the right. An
 * AnalysisBadge banner appears above the columns when pending analysis
 * items exist for the session.
 */

import { Box, Paper, TextField, Typography } from '@mui/material';
import type { Scene } from '../../api/scenes';
import type { Entity } from '../../types';
import { AnalysisBadge } from '../AnalysisBadge';
import MarkdownEditor from '../MarkdownEditor/MarkdownEditor';
import SessionReferencePanel from './SessionReferencePanel';

/**
 * Props for the WrapUpStage component.
 */
export interface WrapUpStageProps {
    /** The campaign this session belongs to. */
    campaignId: number;
    /** The session being wrapped up. */
    sessionId: number;
    /** The actual date the session was played. */
    actualDate: string;
    /** Callback fired when the actual date changes. */
    onActualDateChange: (value: string) => void;
    /** Post-session wrap-up notes (Markdown). */
    actualNotes: string;
    /** Callback fired when the wrap-up notes change. */
    onActualNotesChange: (value: string) => void;
    /** Read-only prep notes for the reference panel. */
    prepNotes: string;
    /** Read-only play notes for the reference panel. */
    playNotes: string;
    /** Scenes associated with this session. */
    scenes: Scene[];
    /** Entities linked to this session for wiki-link resolution. */
    entities: Entity[];
    /** Callback fired when an entity wiki link is clicked. */
    onEntityNavigate?: (entityId: number) => void;
}

/**
 * Two-column wrap-up editor with a date field, Markdown notes, and a
 * collapsible reference sidebar showing prep notes, play notes, and
 * scene summaries from earlier session stages.
 *
 * @param props - The component props.
 * @returns A React element containing the wrap-up editing layout.
 */
export default function WrapUpStage({
    campaignId,
    sessionId,
    actualDate,
    onActualDateChange,
    actualNotes,
    onActualNotesChange,
    prepNotes,
    playNotes,
    scenes,
    entities,
    onEntityNavigate,
}: WrapUpStageProps) {
    return (
        <Box>
            <AnalysisBadge
                variant="banner"
                sourceTable="sessions"
                sourceId={sessionId}
                campaignId={campaignId}
            />

            <Box sx={{ display: 'flex', gap: 3, mt: 2 }}>
                {/* Left column - editable wrap-up fields */}
                <Paper sx={{ p: 3, flex: '1 1 60%', minWidth: 0 }}>
                    <TextField
                        label="Actual Date"
                        type="date"
                        fullWidth
                        InputLabelProps={{ shrink: true }}
                        value={actualDate}
                        onChange={(e) => onActualDateChange(e.target.value)}
                        sx={{ mb: 3 }}
                    />
                    <Typography variant="subtitle2" sx={{ mb: 1 }}>
                        Actual Notes
                    </Typography>
                    <MarkdownEditor
                        value={actualNotes}
                        onChange={onActualNotesChange}
                        placeholder="Write your session wrap-up notes here..."
                        minHeight={300}
                        campaignId={campaignId}
                    />
                </Paper>

                {/* Right column - read-only reference material */}
                <Box sx={{ flex: '1 1 40%', minWidth: 280 }}>
                    <SessionReferencePanel
                        prepNotes={prepNotes}
                        playNotes={playNotes}
                        scenes={scenes}
                        entities={entities}
                        onEntityNavigate={onEntityNavigate}
                    />
                </Box>
            </Box>
        </Box>
    );
}
