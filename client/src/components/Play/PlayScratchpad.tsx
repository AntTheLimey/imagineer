// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * PlayScratchpad - A free-form auto-saving textarea for tracking in-game
 * state during a live TTRPG session.
 *
 * This is intentionally plain text (not a MarkdownEditor) for ephemeral
 * tracking of HP, initiative order, conditions, loot, and other transient
 * game state. Auto-save is handled by the parent component's useAutosave
 * hook; this component simply displays the last-saved indicator.
 */

import { useEffect, useState } from 'react';
import { Box, TextField, Typography } from '@mui/material';

/**
 * Props for the PlayScratchpad component.
 */
export interface PlayScratchpadProps {
    /** The current scratchpad text content. */
    value: string;
    /** Callback fired when the user edits the scratchpad text. */
    onChange: (value: string) => void;
    /** ISO 8601 timestamp of the last successful save, or null if unsaved. */
    lastSaved: string | null;
}

/**
 * Formats an ISO 8601 date string as a human-readable relative time.
 */
function formatRelativeTime(isoString: string): string {
    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
    return date.toLocaleDateString();
}

/**
 * PlayScratchpad renders a full-height plain-text textarea with an
 * auto-saved timestamp indicator beneath it.
 */
export function PlayScratchpad({ value, onChange, lastSaved }: PlayScratchpadProps) {
    // Re-render the relative time indicator every 30 seconds so that
    // "just now" naturally transitions to "1 minute ago", etc.
    const [, setTick] = useState(0);

    useEffect(() => {
        if (!lastSaved) return;
        const interval = setInterval(() => setTick((t) => t + 1), 30_000);
        return () => clearInterval(interval);
    }, [lastSaved]);

    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
                p: 2,
            }}
        >
            <TextField
                multiline
                fullWidth
                placeholder="Track HP, initiative, conditions, loot..."
                value={value}
                onChange={(e) => onChange(e.target.value)}
                sx={{ flexGrow: 1 }}
                InputProps={{
                    style: { height: '100%' },
                }}
                inputProps={{
                    style: { height: '100%', alignSelf: 'flex-start' },
                }}
            />
            {lastSaved && (
                <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ mt: 1 }}
                >
                    Auto-saved {formatRelativeTime(lastSaved)}
                </Typography>
            )}
        </Box>
    );
}

export default PlayScratchpad;
