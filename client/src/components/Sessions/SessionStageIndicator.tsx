// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionStageIndicator - Shows session stage as a colored chip.
 *
 * Displays the current session stage (prep, play, wrap_up) with an
 * appropriate icon and color to provide visual context.
 */

import { Chip, type ChipProps } from '@mui/material';
import {
    EditNote as EditNoteIcon,
    PlayArrow as PlayArrowIcon,
    CheckCircle as CheckCircleIcon,
} from '@mui/icons-material';
import type { SessionStage } from '../../types';

/**
 * Props for the SessionStageIndicator component.
 */
export interface SessionStageIndicatorProps {
    /** The current session stage. */
    stage: SessionStage;
    /** Size variant for the chip. */
    size?: 'small' | 'medium';
}

/**
 * Configuration for each session stage including label, color, and icon.
 */
const stageConfig: Record<
    SessionStage,
    {
        label: string;
        color: ChipProps['color'];
        icon: React.ReactElement;
    }
> = {
    prep: {
        label: 'Prep',
        color: 'info',
        icon: <EditNoteIcon fontSize="small" />,
    },
    play: {
        label: 'Play',
        color: 'success',
        icon: <PlayArrowIcon fontSize="small" />,
    },
    wrap_up: {
        label: 'Wrap-up',
        color: 'warning',
        icon: <CheckCircleIcon fontSize="small" />,
    },
};

/**
 * Component for displaying the current session stage as a colored chip.
 *
 * Shows different colors and icons for each stage:
 * - Prep: Blue/info with EditNote icon
 * - Play: Green/success with PlayArrow icon
 * - Wrap-up: Orange/warning with CheckCircle icon
 *
 * @param props - The component props.
 * @returns A React element containing the stage chip.
 *
 * @example
 * ```tsx
 * <SessionStageIndicator stage="prep" />
 * <SessionStageIndicator stage="play" size="small" />
 * ```
 */
export default function SessionStageIndicator({
    stage,
    size = 'small',
}: SessionStageIndicatorProps) {
    const config = stageConfig[stage];

    return (
        <Chip
            label={config.label}
            color={config.color}
            icon={config.icon}
            size={size}
            variant="filled"
        />
    );
}
