// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SceneViewer - Read-only left panel (~40% width) in Play mode.
 *
 * Displays the active scene content and/or the session's prep notes.
 * Supports three display modes:
 *
 * - **notes**: Renders prep notes as Markdown with wiki-link entity
 *   navigation.
 * - **scene**: Shows the active scene's title, objective, description,
 *   GM notes, and linked entity chips.
 * - **mixed**: Presents a tabbed interface toggling between the scene
 *   view and the notes view, defaulting to the "Scene" tab.
 *
 * When no scene is available in 'scene' or 'mixed' mode, the component
 * falls back to displaying the prep notes.
 */

import { useState, useMemo } from 'react';
import { Box, Chip, Tab, Tabs, Typography } from '@mui/material';
import MarkdownRenderer from '../MarkdownRenderer';
import type { WikiLinkEntity } from '../MarkdownRenderer';
import type { Scene } from '../../api/scenes';
import type { Entity } from '../../types';

/**
 * Props for the SceneViewer component.
 */
export interface SceneViewerProps {
    /** The currently active scene, or null if no scene is selected. */
    scene: Scene | null;
    /** Session prep notes rendered as Markdown. */
    prepNotes: string;
    /** Display mode controlling which content is shown. */
    mode: 'scene' | 'notes' | 'mixed';
    /** The campaign ID (used for context). */
    campaignId: number;
    /** Callback fired when an entity chip or wiki link is clicked. */
    onEntityClick: (entityId: number) => void;
    /** All entities available for wiki-link matching and chip display. */
    entities: Entity[];
}

/**
 * Renders the prep notes section using MarkdownRenderer with
 * wiki-link entity navigation.
 */
function NotesContent({
    prepNotes,
    wikiLinkEntities,
    onEntityClick,
}: {
    prepNotes: string;
    wikiLinkEntities: WikiLinkEntity[];
    onEntityClick: (entityId: number) => void;
}) {
    return (
        <MarkdownRenderer
            content={prepNotes}
            entities={wikiLinkEntities}
            onEntityNavigate={onEntityClick}
        />
    );
}

/**
 * Renders the active scene details: title, objective, description,
 * GM notes, and linked entity chips.
 */
function SceneContent({
    scene,
    wikiLinkEntities,
    onEntityClick,
    entities,
}: {
    scene: Scene;
    wikiLinkEntities: WikiLinkEntity[];
    onEntityClick: (entityId: number) => void;
    entities: Entity[];
}) {
    /** Entities whose IDs appear in the active scene's entityIds list. */
    const linkedEntities = useMemo(
        () => entities.filter((e) => scene.entityIds.includes(e.id)),
        [entities, scene.entityIds]
    );

    return (
        <>
            <Typography
                variant="h5"
                sx={{ fontFamily: 'Cinzel', mb: 2 }}
            >
                {scene.title}
            </Typography>

            {scene.objective && (
                <Box sx={{ mb: 2 }}>
                    <Typography variant="caption" color="text.secondary">
                        Objective
                    </Typography>
                    <Typography variant="body2">
                        {scene.objective}
                    </Typography>
                </Box>
            )}

            {scene.description && (
                <Box sx={{ mb: 2 }}>
                    <MarkdownRenderer
                        content={scene.description}
                        entities={wikiLinkEntities}
                        onEntityNavigate={onEntityClick}
                    />
                </Box>
            )}

            {scene.gmNotes && (
                <Box sx={{ mb: 2 }}>
                    <Typography
                        variant="caption"
                        color="warning.main"
                    >
                        GM Notes
                    </Typography>
                    <MarkdownRenderer
                        content={scene.gmNotes}
                        entities={wikiLinkEntities}
                        onEntityNavigate={onEntityClick}
                    />
                </Box>
            )}

            {linkedEntities.length > 0 && (
                <Box
                    sx={{
                        display: 'flex',
                        flexWrap: 'wrap',
                        gap: 0.5,
                        mt: 2,
                    }}
                >
                    {linkedEntities.map((entity) => (
                        <Chip
                            key={entity.id}
                            label={entity.name}
                            size="small"
                            onClick={() => onEntityClick(entity.id)}
                        />
                    ))}
                </Box>
            )}
        </>
    );
}

/**
 * SceneViewer renders a read-only view of the active scene content
 * and/or session prep notes in the Play mode left panel.
 */
export function SceneViewer({
    scene,
    prepNotes,
    mode,
    onEntityClick,
    entities,
}: SceneViewerProps) {
    /** Index of the active tab in mixed mode (0 = Scene, 1 = Notes). */
    const [tabIndex, setTabIndex] = useState(0);

    /** Maps campaign entities to the WikiLinkEntity shape for MarkdownRenderer. */
    const wikiLinkEntities: WikiLinkEntity[] = useMemo(
        () =>
            entities.map((e) => ({
                id: e.id,
                name: e.name,
                entityType: e.entityType,
                description: e.description ?? null,
            })),
        [entities]
    );

    /**
     * Renders the notes panel. Used in 'notes' mode and as a
     * fallback when no scene is available.
     */
    const notesPanel = (
        <NotesContent
            prepNotes={prepNotes}
            wikiLinkEntities={wikiLinkEntities}
            onEntityClick={onEntityClick}
        />
    );

    /**
     * Renders the scene panel when a scene is available.
     */
    const scenePanel = scene ? (
        <SceneContent
            scene={scene}
            wikiLinkEntities={wikiLinkEntities}
            onEntityClick={onEntityClick}
            entities={entities}
        />
    ) : null;

    /**
     * Determines which content to render based on the current mode
     * and scene availability.
     */
    const renderContent = () => {
        if (mode === 'notes') {
            return notesPanel;
        }

        if (mode === 'scene') {
            return scene ? scenePanel : notesPanel;
        }

        // mode === 'mixed'
        if (!scene) {
            return notesPanel;
        }

        return (
            <>
                <Tabs
                    value={tabIndex}
                    onChange={(_e, newValue: number) =>
                        setTabIndex(newValue)
                    }
                    sx={{ mb: 2 }}
                >
                    <Tab label="Scene" />
                    <Tab label="Notes" />
                </Tabs>
                {tabIndex === 0 && scenePanel}
                {tabIndex === 1 && notesPanel}
            </>
        );
    };

    return (
        <Box
            sx={{
                height: '100%',
                overflow: 'auto',
                p: 2,
            }}
        >
            {renderContent()}
        </Box>
    );
}

export default SceneViewer;
