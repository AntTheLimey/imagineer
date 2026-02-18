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
 * SceneSummaryList - Read-only list of scenes with status chips.
 *
 * Displays an aggregate summary line followed by each scene's title,
 * type, and status. Used in session detail views where editing is
 * not required.
 */

import { useMemo } from 'react';
import { Box, Chip, Typography } from '@mui/material';
import type { Scene } from '../../api/scenes';

/**
 * Props for the SceneSummaryList component.
 */
export interface SceneSummaryListProps {
    /** The scenes to display. */
    scenes: Scene[];
}

/** Maps scene status to a MUI Chip color. */
const statusColor: Record<string, 'success' | 'info' | 'default'> = {
    completed: 'success',
    active: 'info',
};

/** Statuses rendered with the outlined chip variant. */
const outlinedStatuses = new Set(['skipped', 'planned']);

/**
 * Builds a human-readable aggregate summary such as
 * "4 scenes: 2 completed, 1 skipped".
 */
function buildSummary(scenes: Scene[]): string {
    const total = scenes.length;
    const counts: Record<string, number> = {};
    for (const s of scenes) {
        counts[s.status] = (counts[s.status] || 0) + 1;
    }

    const parts: string[] = [];
    if (counts.completed) parts.push(`${counts.completed} completed`);
    if (counts.skipped) parts.push(`${counts.skipped} skipped`);
    if (counts.active) parts.push(`${counts.active} active`);
    if (counts.planned) parts.push(`${counts.planned} planned`);

    const suffix = parts.length > 0 ? `: ${parts.join(', ')}` : '';
    return `${total} scene${total !== 1 ? 's' : ''}${suffix}`;
}

/**
 * Read-only scene list with status chips and an aggregate summary.
 *
 * @param props - The component props.
 * @returns A React element containing the scene summary list.
 */
export default function SceneSummaryList({ scenes }: SceneSummaryListProps) {
    const summary = useMemo(() => buildSummary(scenes), [scenes]);

    if (scenes.length === 0) {
        return (
            <Typography variant="body2" color="text.secondary">
                No scenes in this session.
            </Typography>
        );
    }

    return (
        <Box>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                {summary}
            </Typography>

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.75 }}>
                {scenes.map((scene) => (
                    <Box
                        key={scene.id}
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                        }}
                    >
                        <Typography variant="body2" sx={{ flexShrink: 0 }}>
                            {scene.title}
                        </Typography>
                        <Chip
                            label={scene.sceneType}
                            size="small"
                            variant="outlined"
                        />
                        <Chip
                            label={scene.status}
                            size="small"
                            color={statusColor[scene.status] ?? 'default'}
                            variant={
                                outlinedStatuses.has(scene.status)
                                    ? 'outlined'
                                    : 'filled'
                            }
                        />
                    </Box>
                ))}
            </Box>
        </Box>
    );
}
