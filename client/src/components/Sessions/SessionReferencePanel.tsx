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
 * SessionReferencePanel - Collapsible reference sidebar for session views.
 *
 * Provides three Accordion sections for prep notes, play notes, and a
 * scene summary. Prep notes render Markdown with wiki-link support;
 * play notes use plain-text pre-wrap formatting; the scene summary
 * delegates to SceneSummaryList.
 */

import {
    Accordion,
    AccordionDetails,
    AccordionSummary,
    Box,
    Chip,
    Typography,
} from '@mui/material';
import { ExpandMore as ExpandMoreIcon } from '@mui/icons-material';
import type { Scene } from '../../api/scenes';
import type { Entity } from '../../types';
import MarkdownRenderer from '../MarkdownRenderer/MarkdownRenderer';
import SceneSummaryList from './SceneSummaryList';

/**
 * Props for the SessionReferencePanel component.
 */
export interface SessionReferencePanelProps {
    /** Markdown-formatted prep notes for the session. */
    prepNotes: string;
    /** Plain-text play notes captured during the session. */
    playNotes: string;
    /** Scenes belonging to the session. */
    scenes: Scene[];
    /** Campaign entities available for wiki-link matching. */
    entities: Entity[];
    /** Callback fired when an entity wiki link is clicked. */
    onEntityNavigate?: (entityId: number) => void;
    /** Whether each accordion section starts expanded. */
    defaultExpanded?: boolean;
}

/**
 * Counts the number of words in a string.
 *
 * Splits on whitespace and filters out empty segments to produce an
 * accurate count regardless of leading, trailing, or repeated spaces.
 */
function wordCount(text: string): number {
    return text.split(/\s+/).filter(Boolean).length;
}

/**
 * Collapsible reference sidebar with prep notes, play notes, and
 * scene summary sections.
 *
 * @param props - The component props.
 * @returns A React element containing the three accordion sections.
 *
 * @example
 * ```tsx
 * <SessionReferencePanel
 *     prepNotes={session.prepNotes ?? ''}
 *     playNotes={session.playNotes ?? ''}
 *     scenes={scenes}
 *     entities={entities}
 *     onEntityNavigate={(id) => navigate(`/entities/${id}`)}
 * />
 * ```
 */
export default function SessionReferencePanel({
    prepNotes,
    playNotes,
    scenes,
    entities,
    onEntityNavigate,
    defaultExpanded = true,
}: SessionReferencePanelProps) {
    const prepWordCount = prepNotes ? wordCount(prepNotes) : 0;
    const playWordCount = playNotes ? wordCount(playNotes) : 0;

    return (
        <Box>
            {/* Prep Notes */}
            <Accordion defaultExpanded={defaultExpanded}>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography sx={{ flexGrow: 1 }}>Prep Notes</Typography>
                    {prepWordCount > 0 && (
                        <Chip
                            label={`${prepWordCount} words`}
                            size="small"
                            sx={{ mr: 1 }}
                        />
                    )}
                </AccordionSummary>
                <AccordionDetails>
                    <MarkdownRenderer
                        content={prepNotes}
                        entities={entities}
                        onEntityNavigate={onEntityNavigate}
                    />
                </AccordionDetails>
            </Accordion>

            {/* Play Notes */}
            <Accordion defaultExpanded={defaultExpanded}>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography sx={{ flexGrow: 1 }}>Play Notes</Typography>
                    {playWordCount > 0 && (
                        <Chip
                            label={`${playWordCount} words`}
                            size="small"
                            sx={{ mr: 1 }}
                        />
                    )}
                </AccordionSummary>
                <AccordionDetails>
                    <Typography
                        variant="body2"
                        sx={{ whiteSpace: 'pre-wrap' }}
                    >
                        {playNotes}
                    </Typography>
                </AccordionDetails>
            </Accordion>

            {/* Scene Summary */}
            <Accordion defaultExpanded={defaultExpanded}>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography sx={{ flexGrow: 1 }}>
                        Scene Summary
                    </Typography>
                    <Chip
                        label={`${scenes.length} scene${scenes.length !== 1 ? 's' : ''}`}
                        size="small"
                        sx={{ mr: 1 }}
                    />
                </AccordionSummary>
                <AccordionDetails>
                    <SceneSummaryList scenes={scenes} />
                </AccordionDetails>
            </Accordion>
        </Box>
    );
}
