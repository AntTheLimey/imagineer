// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionStageNav - Navigation between session stages.
 *
 * Shows three stages as tabs/buttons:
 * - Prep (icon: EditNote)
 * - Play (icon: PlayArrow)
 * - Wrap-up (icon: CheckCircle)
 *
 * User can freely navigate between stages.
 */

import {
    ToggleButton,
    ToggleButtonGroup,
    Tooltip,
    type ToggleButtonGroupProps,
} from '@mui/material';
import {
    EditNote as EditNoteIcon,
    PlayArrow as PlayArrowIcon,
    CheckCircle as CheckCircleIcon,
} from '@mui/icons-material';
import type { SessionStage } from '../../types';

/**
 * Props for the SessionStageNav component.
 */
export interface SessionStageNavProps {
    /** The currently selected stage. */
    currentStage: SessionStage;
    /** Callback fired when the stage changes. */
    onStageChange: (stage: SessionStage) => void;
    /** Whether the navigation is disabled. */
    disabled?: boolean;
    /** Size variant for the buttons. */
    size?: ToggleButtonGroupProps['size'];
    /** Orientation of the button group. */
    orientation?: ToggleButtonGroupProps['orientation'];
}

/**
 * Configuration for each session stage.
 */
interface StageConfig {
    value: SessionStage;
    label: string;
    description: string;
    icon: React.ReactElement;
}

/**
 * Stage definitions with labels, descriptions, and icons.
 */
const stages: StageConfig[] = [
    {
        value: 'prep',
        label: 'Prep',
        description: 'Prepare for the session: notes, scenes, NPCs',
        icon: <EditNoteIcon />,
    },
    {
        value: 'play',
        label: 'Play',
        description: 'Session in progress: track events and decisions',
        icon: <PlayArrowIcon />,
    },
    {
        value: 'wrap_up',
        label: 'Wrap-up',
        description: 'Post-session: record outcomes and consequences',
        icon: <CheckCircleIcon />,
    },
];

/**
 * Navigation component for switching between session stages.
 *
 * Provides a toggle button group allowing users to navigate between
 * the three session workflow stages: Prep, Play, and Wrap-up.
 *
 * @param props - The component props.
 * @returns A React element containing the stage navigation.
 *
 * @example
 * ```tsx
 * <SessionStageNav
 *     currentStage={session.stage}
 *     onStageChange={handleStageChange}
 * />
 * ```
 */
export default function SessionStageNav({
    currentStage,
    onStageChange,
    disabled = false,
    size = 'medium',
    orientation = 'horizontal',
}: SessionStageNavProps) {
    /**
     * Handles stage selection change.
     */
    const handleChange = (
        _event: React.MouseEvent<HTMLElement>,
        newStage: SessionStage | null
    ) => {
        // Prevent deselection - always require a stage to be selected
        if (newStage !== null) {
            onStageChange(newStage);
        }
    };

    return (
        <ToggleButtonGroup
            value={currentStage}
            exclusive
            onChange={handleChange}
            aria-label="session stage navigation"
            size={size}
            orientation={orientation}
            disabled={disabled}
            sx={{
                '& .MuiToggleButton-root': {
                    px: 2,
                    gap: 1,
                },
            }}
        >
            {stages.map((stage) => (
                <Tooltip
                    key={stage.value}
                    title={stage.description}
                    placement={orientation === 'vertical' ? 'right' : 'bottom'}
                >
                    <ToggleButton
                        value={stage.value}
                        aria-label={stage.label}
                        sx={{
                            '&.Mui-selected': {
                                backgroundColor:
                                    stage.value === 'prep'
                                        ? 'info.main'
                                        : stage.value === 'play'
                                          ? 'success.main'
                                          : 'warning.main',
                                color: 'white',
                                '&:hover': {
                                    backgroundColor:
                                        stage.value === 'prep'
                                            ? 'info.dark'
                                            : stage.value === 'play'
                                              ? 'success.dark'
                                              : 'warning.dark',
                                },
                            },
                        }}
                    >
                        {stage.icon}
                        {stage.label}
                    </ToggleButton>
                </Tooltip>
            ))}
        </ToggleButtonGroup>
    );
}
