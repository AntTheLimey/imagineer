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
 * DiffView - Side-by-side diff view for comparing original and revised
 * content.
 *
 * Displays two panels: the original content on the left and the revised
 * content on the right. Changed lines are highlighted with subtle
 * background colours (red-tinted for removals, green-tinted for
 * additions). The revised panel can optionally be made editable,
 * replacing the line-by-line diff with a multiline TextField.
 */

import { useMemo, useCallback } from 'react';
import { Box, Grid, Paper, TextField, Typography } from '@mui/material';

/**
 * Props for the DiffView component.
 */
export interface DiffViewProps {
    /** The original content before revision. */
    original: string;
    /** The revised content after revision. */
    revised: string;
    /** Callback when the user edits the revised content. */
    onEdit?: (newContent: string) => void;
    /** Whether the revised panel is editable. */
    editable?: boolean;
}

/**
 * A single line-pair from the diff comparison.
 */
interface DiffLine {
    origLine: string;
    revLine: string;
    changed: boolean;
}

/**
 * Compare two texts line-by-line and return an array of diff results.
 * Lines that differ between original and revised are marked as changed.
 *
 * @param original - The original text.
 * @param revised - The revised text.
 * @returns An array of line-pair objects with a changed flag.
 */
// eslint-disable-next-line react-refresh/only-export-components
export function diffLines(original: string, revised: string): DiffLine[] {
    const origLines = original.split('\n');
    const revLines = revised.split('\n');
    const maxLen = Math.max(origLines.length, revLines.length);

    const result: DiffLine[] = [];

    for (let i = 0; i < maxLen; i++) {
        const o = origLines[i] ?? '';
        const r = revLines[i] ?? '';
        result.push({ origLine: o, revLine: r, changed: o !== r });
    }
    return result;
}

/** Background colour for changed lines on the original (removal) side. */
const REMOVAL_BG = 'rgba(255, 0, 0, 0.08)';

/** Background colour for changed lines on the revised (addition) side. */
const ADDITION_BG = 'rgba(0, 128, 0, 0.08)';

/**
 * Shared typography styles for monospaced content display.
 */
const monoStyles = {
    fontFamily: 'monospace',
    whiteSpace: 'pre-wrap',
    fontSize: '0.875rem',
    lineHeight: 1.6,
} as const;

/**
 * Display a side-by-side diff of original and revised content.
 *
 * When `editable` is true and `onEdit` is provided, the revised panel
 * becomes a multiline TextField so the user can make further edits.
 * Otherwise both panels show a line-by-line diff with changed lines
 * highlighted.
 *
 * @param props - The diff view configuration.
 * @returns A Grid layout with two side-by-side panels.
 */
export function DiffView({
    original,
    revised,
    onEdit,
    editable = false,
}: DiffViewProps) {
    const lines = useMemo(() => diffLines(original, revised), [original, revised]);
    const isEditing = editable && onEdit;

    const handleRevisionChange = useCallback(
        (e: React.ChangeEvent<HTMLTextAreaElement>) => {
            onEdit?.(e.target.value);
        },
        [onEdit],
    );

    return (
        <Grid container spacing={2}>
            {/* Original panel */}
            <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" gutterBottom>
                    Original
                </Typography>
                <Paper
                    variant="outlined"
                    role="region"
                    aria-label="Original content"
                    sx={{
                        p: 1.5,
                        bgcolor: 'grey.100',
                        overflow: 'auto',
                        maxHeight: 400,
                    }}
                >
                    {lines.map((line, idx) => (
                        <Box
                            key={idx}
                            data-testid={
                                line.changed
                                    ? 'diff-line-removed'
                                    : 'diff-line-unchanged'
                            }
                            sx={{
                                bgcolor: line.changed
                                    ? REMOVAL_BG
                                    : 'transparent',
                                px: 0.5,
                            }}
                        >
                            <Typography
                                component="span"
                                sx={monoStyles}
                            >
                                {line.origLine || '\u00A0'}
                            </Typography>
                        </Box>
                    ))}
                </Paper>
            </Grid>

            {/* Revised panel */}
            <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" gutterBottom>
                    Revised
                </Typography>
                {isEditing ? (
                    <TextField
                        multiline
                        fullWidth
                        value={revised}
                        onChange={handleRevisionChange}
                        InputProps={{
                            sx: {
                                ...monoStyles,
                                maxHeight: 400,
                                overflow: 'auto',
                            },
                        }}
                        inputProps={{
                            'aria-label': 'Edit revised content',
                        }}
                    />
                ) : (
                    <Paper
                        variant="outlined"
                        role="region"
                        aria-label="Revised content"
                        sx={{
                            p: 1.5,
                            overflow: 'auto',
                            maxHeight: 400,
                        }}
                    >
                        {lines.map((line, idx) => (
                            <Box
                                key={idx}
                                data-testid={
                                    line.changed
                                        ? 'diff-line-added'
                                        : 'diff-line-unchanged'
                                }
                                sx={{
                                    bgcolor: line.changed
                                        ? ADDITION_BG
                                        : 'transparent',
                                    px: 0.5,
                                }}
                            >
                                <Typography
                                    component="span"
                                    sx={monoStyles}
                                >
                                    {line.revLine || '\u00A0'}
                                </Typography>
                            </Box>
                        ))}
                    </Paper>
                )}
            </Grid>
        </Grid>
    );
}

export default DiffView;
