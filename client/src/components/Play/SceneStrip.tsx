// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SceneStrip - Horizontal scrollable bar of scene cards.
 *
 * Displayed above the main Play panels when scenes exist. Each card
 * shows the scene title, a type icon, and a clickable status chip.
 * The active scene is highlighted with a primary-coloured border.
 */

import React from 'react';
import {
    Box,
    Chip,
    IconButton,
    Paper,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Explore as ExploreIcon,
    Flight as FlightIcon,
    GpsFixed as GpsFixedIcon,
    MoreHoriz as MoreHorizIcon,
    People as PeopleIcon,
    Psychology as PsychologyIcon,
    TheaterComedy as TheaterComedyIcon,
    Weekend as WeekendIcon,
} from '@mui/icons-material';
import type { Scene } from '../../api/scenes';

/**
 * Maps scene type strings to their display icon and label.
 */
const SCENE_TYPE_CONFIG: Record<string, { icon: React.ReactElement; label: string }> = {
    combat: { icon: <GpsFixedIcon fontSize="small" />, label: 'Combat' },
    social: { icon: <PeopleIcon fontSize="small" />, label: 'Social' },
    exploration: { icon: <ExploreIcon fontSize="small" />, label: 'Exploration' },
    puzzle: { icon: <PsychologyIcon fontSize="small" />, label: 'Puzzle' },
    roleplay: { icon: <TheaterComedyIcon fontSize="small" />, label: 'Roleplay' },
    travel: { icon: <FlightIcon fontSize="small" />, label: 'Travel' },
    downtime: { icon: <WeekendIcon fontSize="small" />, label: 'Downtime' },
    other: { icon: <MoreHorizIcon fontSize="small" />, label: 'Other' },
};

/**
 * Maps scene status strings to MUI Chip colour variants.
 */
const STATUS_CHIP_COLOR: Record<string, 'default' | 'primary' | 'success' | 'warning'> = {
    planned: 'default',
    active: 'primary',
    completed: 'success',
    skipped: 'warning',
};

/**
 * Props for the SceneStrip component.
 */
export interface SceneStripProps {
    /** Array of scenes to display in the strip. */
    scenes: Scene[];
    /** ID of the currently active scene, or null if none is active. */
    activeSceneId: number | null;
    /** Callback fired when a scene card is clicked. */
    onSceneSelect: (id: number) => void;
    /** Callback fired when a scene's status chip is clicked. */
    onSceneStatusChange: (id: number, status: string) => void;
    /** Callback fired when the add-scene button is clicked. */
    onAddScene: () => void;
}

/**
 * Horizontal scrollable strip of scene cards.
 *
 * Renders a row of compact scene cards that scroll horizontally when
 * the list overflows. Each card displays the scene title, a type icon
 * from SCENE_TYPE_CONFIG, and a status chip. Clicking a card selects
 * the scene; clicking the status chip triggers a status change callback.
 *
 * An add button at the end of the strip allows creating new scenes.
 *
 * @param props - The component props.
 * @returns A React element containing the scene strip, or null when
 *          there are no scenes.
 *
 * @example
 * ```tsx
 * <SceneStrip
 *     scenes={scenes}
 *     activeSceneId={activeId}
 *     onSceneSelect={(id) => setActiveId(id)}
 *     onSceneStatusChange={(id, status) => cycleStatus(id, status)}
 *     onAddScene={() => openNewSceneDialog()}
 * />
 * ```
 */
export function SceneStrip({
    scenes,
    activeSceneId,
    onSceneSelect,
    onSceneStatusChange,
    onAddScene,
}: SceneStripProps) {
    if (scenes.length === 0) {
        return null;
    }

    return (
        <Box
            sx={{
                display: 'flex',
                overflowX: 'auto',
                gap: 1,
                pb: 1,
                borderBottom: 1,
                borderColor: 'divider',
                px: 1,
                py: 1,
                minHeight: 72,
            }}
        >
            {scenes.map((scene) => {
                const isActive = scene.id === activeSceneId;
                const typeConfig = SCENE_TYPE_CONFIG[scene.sceneType] ?? SCENE_TYPE_CONFIG.other;
                const chipColor = STATUS_CHIP_COLOR[scene.status] ?? 'default';

                /**
                 * Handles click on the status chip without propagating
                 * the event to the parent card.
                 */
                const handleChipClick = (e: React.MouseEvent) => {
                    e.stopPropagation();
                    onSceneStatusChange(scene.id, scene.status);
                };

                return (
                    <Paper
                        key={scene.id}
                        variant="outlined"
                        onClick={() => onSceneSelect(scene.id)}
                        sx={{
                            minWidth: 140,
                            flexShrink: 0,
                            cursor: 'pointer',
                            p: 1,
                            border: isActive ? 2 : 1,
                            borderColor: isActive ? 'primary.main' : 'divider',
                            elevation: isActive ? 2 : 0,
                            boxShadow: isActive ? 2 : 0,
                            transition: 'all 0.15s ease-in-out',
                            '&:hover': {
                                boxShadow: 2,
                            },
                        }}
                    >
                        <Typography
                            variant="body2"
                            fontWeight={600}
                            noWrap
                        >
                            {scene.title}
                        </Typography>

                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 0.5,
                                mt: 0.5,
                            }}
                        >
                            <Tooltip title={typeConfig.label}>
                                {typeConfig.icon}
                            </Tooltip>
                            <Chip
                                label={scene.status}
                                size="small"
                                color={chipColor}
                                onClick={handleChipClick}
                            />
                        </Box>
                    </Paper>
                );
            })}

            <Tooltip title="Add scene">
                <IconButton
                    onClick={onAddScene}
                    sx={{
                        flexShrink: 0,
                        alignSelf: 'center',
                    }}
                    aria-label="add scene"
                >
                    <AddIcon />
                </IconButton>
            </Tooltip>
        </Box>
    );
}

export default SceneStrip;
