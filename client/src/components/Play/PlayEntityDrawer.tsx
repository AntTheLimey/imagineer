// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * PlayEntityDrawer - Read-only entity detail panel for Play mode.
 *
 * Displays a pinned entity's full details when selected via a wiki link
 * click or entity chip in the session narrative. Unlike EntityPreviewPanel,
 * this component omits Edit/Delete actions and shows the complete
 * description and GM notes without line clamping.
 */

import { useNavigate } from 'react-router-dom';
import {
    Box,
    Chip,
    Divider,
    List,
    ListItemButton,
    ListItemText,
    Typography,
} from '@mui/material';
import {
    Link as RelationshipIcon,
} from '@mui/icons-material';
import { useEntityRelationships, useEntities } from '../../hooks';
import { MarkdownRenderer } from '../MarkdownRenderer';
import type { WikiLinkEntity } from '../MarkdownRenderer';
import type { Entity, EntityType } from '../../types';

/**
 * Props for the PlayEntityDrawer component.
 */
export interface PlayEntityDrawerProps {
    /** The entity to display, or null when no entity is selected. */
    entity: Entity | null;
    /** The campaign ID for relationship and entity queries. */
    campaignId: number;
}

/**
 * Human-readable labels for entity types.
 */
const ENTITY_TYPE_LABELS: Record<EntityType, string> = {
    npc: 'NPC',
    location: 'Location',
    item: 'Item',
    faction: 'Faction',
    clue: 'Clue',
    creature: 'Creature',
    organization: 'Organization',
    event: 'Event',
    document: 'Document',
    other: 'Other',
};

/**
 * Colour mapping for entity type chips.
 */
const ENTITY_TYPE_COLORS: Record<EntityType, 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info'> = {
    npc: 'primary',
    location: 'success',
    item: 'warning',
    faction: 'secondary',
    clue: 'info',
    creature: 'error',
    organization: 'secondary',
    event: 'info',
    document: 'default',
    other: 'default',
};

/**
 * Read-only entity detail panel for Play mode.
 *
 * When an entity is selected (via wiki link or entity chip), this component
 * renders the entity name, type and source confidence chips, tags, full
 * description, GM notes, and relationships. All content is read-only with
 * no editing controls.
 *
 * When no entity is selected, displays a prompt instructing the user to
 * click an entity or wiki link.
 *
 * @param props - The component props.
 * @returns A React element containing the entity drawer panel.
 */
export function PlayEntityDrawer({
    entity,
    campaignId,
}: PlayEntityDrawerProps) {
    const navigate = useNavigate();

    /**
     * Navigate to entities page filtered by the clicked wiki link name.
     */
    const handleEntityClick = (name: string) => {
        navigate(`/campaigns/${campaignId}/entities?search=${encodeURIComponent(name)}`);
    };

    /**
     * Navigate to the entity view page for a specific entity.
     */
    const handleEntityNavigate = (entityId: number) => {
        navigate(`/campaigns/${campaignId}/entities/${entityId}`);
    };

    /** All campaign entities for wiki link resolution in description/GM notes. */
    const { data: allEntities } = useEntities({
        campaignId,
    });

    /** Campaign entities mapped to WikiLinkEntity shape for MarkdownRenderer. */
    const wikiLinkEntities: WikiLinkEntity[] | undefined = allEntities?.map(
        (e) => ({
            id: e.id,
            name: e.name,
            entityType: e.entityType,
            description: e.description ?? null,
        })
    );

    /** Relationships for the currently displayed entity. */
    const { data: relationships } = useEntityRelationships(
        campaignId,
        entity?.id ?? 0,
        { enabled: !!entity }
    );

    /** When no entity is selected, render the empty-state prompt. */
    if (!entity) {
        return (
            <Box
                sx={{
                    height: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    textAlign: 'center',
                    p: 3,
                }}
            >
                <Typography variant="body1" color="text.secondary">
                    Click an entity or wiki link to view details.
                </Typography>
            </Box>
        );
    }

    /** Number of relationships to display in the header. */
    const relationshipCount = relationships?.length ?? 0;

    return (
        <Box
            sx={{
                height: '100%',
                overflow: 'auto',
                p: 2,
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            {/* Entity name */}
            <Typography
                variant="h6"
                sx={{
                    fontFamily: 'Cinzel',
                    fontWeight: 600,
                    mb: 1,
                    wordBreak: 'break-word',
                }}
            >
                {entity.name}
            </Typography>

            {/* Type and source confidence chips */}
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 2 }}>
                <Chip
                    label={ENTITY_TYPE_LABELS[entity.entityType]}
                    color={ENTITY_TYPE_COLORS[entity.entityType]}
                    size="small"
                />
                <Chip
                    label={entity.sourceConfidence}
                    variant="outlined"
                    size="small"
                    color={
                        entity.sourceConfidence === 'AUTHORITATIVE'
                            ? 'success'
                            : entity.sourceConfidence === 'DRAFT'
                                ? 'warning'
                                : 'default'
                    }
                />
            </Box>

            {/* Tags */}
            {entity.tags && entity.tags.length > 0 && (
                <Box sx={{ mb: 2 }}>
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{ display: 'block', mb: 0.5 }}
                    >
                        Tags
                    </Typography>
                    <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                        {entity.tags.map((tag) => (
                            <Chip key={tag} label={tag} size="small" variant="outlined" />
                        ))}
                    </Box>
                </Box>
            )}

            <Divider sx={{ my: 1 }} />

            {/* Description */}
            <Box sx={{ mb: 2 }}>
                <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ display: 'block', mb: 0.5 }}
                >
                    Description
                </Typography>
                {entity.description ? (
                    <Box
                        component="div"
                        sx={{
                            fontSize: '0.875rem',
                            '& p': { mt: 0, mb: 1 },
                            '& p:last-child': { mb: 0 },
                        }}
                    >
                        <MarkdownRenderer
                            content={entity.description}
                            onEntityClick={handleEntityClick}
                            entities={wikiLinkEntities}
                            onEntityNavigate={handleEntityNavigate}
                        />
                    </Box>
                ) : (
                    <Typography variant="body2">No description</Typography>
                )}
            </Box>

            {/* GM Notes */}
            {entity.gmNotes && (
                <Box sx={{ mb: 2 }}>
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{ display: 'block', mb: 0.5 }}
                    >
                        GM Notes
                    </Typography>
                    <Box
                        component="div"
                        sx={{
                            fontSize: '0.875rem',
                            '& p': { mt: 0, mb: 1 },
                            '& p:last-child': { mb: 0 },
                        }}
                    >
                        <MarkdownRenderer
                            content={entity.gmNotes}
                            onEntityClick={handleEntityClick}
                            entities={wikiLinkEntities}
                            onEntityNavigate={handleEntityNavigate}
                        />
                    </Box>
                </Box>
            )}

            <Divider sx={{ my: 1 }} />

            {/* Relationships */}
            <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                    <RelationshipIcon fontSize="small" color="action" />
                    <Typography variant="body2" color="text.secondary">
                        {relationshipCount} relationship{relationshipCount !== 1 ? 's' : ''}
                    </Typography>
                </Box>
                {relationships && relationships.length > 0 && (
                    <List dense disablePadding>
                        {relationships.map((rel) => {
                            // Determine the "other" entity in the relationship
                            const isSource = rel.sourceEntityId === entity.id;
                            const otherName = isSource
                                ? rel.targetEntityName
                                : rel.sourceEntityName;
                            const otherEntityId = isSource
                                ? rel.targetEntityId
                                : rel.sourceEntityId;
                            const label = rel.displayLabel ?? rel.relationshipType;

                            return (
                                <ListItemButton
                                    key={rel.id}
                                    onClick={() => handleEntityNavigate(otherEntityId)}
                                    sx={{ py: 0.5, px: 1, borderRadius: 1 }}
                                >
                                    <ListItemText
                                        primary={`${label}: ${otherName ?? 'Unknown'}`}
                                        primaryTypographyProps={{
                                            variant: 'body2',
                                        }}
                                    />
                                </ListItemButton>
                            );
                        })}
                    </List>
                )}
            </Box>
        </Box>
    );
}

export default PlayEntityDrawer;
