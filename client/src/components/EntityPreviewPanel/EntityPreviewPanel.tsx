// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import {
    Box,
    Button,
    Chip,
    Divider,
    Typography,
} from '@mui/material';
import {
    Edit as EditIcon,
    Delete as DeleteIcon,
    Link as RelationshipIcon,
} from '@mui/icons-material';
import { useEntityRelationships } from '../../hooks';
import type { Entity, EntityType } from '../../types';

/**
 * Props for the EntityPreviewPanel component.
 */
export interface EntityPreviewPanelProps {
    /** The entity to preview, or null if no entity is selected. */
    entity: Entity | null;
    /** The campaign ID for relationship queries. */
    campaignId: string;
    /** Callback fired when the Edit button is clicked. */
    onEdit: () => void;
    /** Callback fired when the Delete button is clicked. */
    onDelete: () => void;
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
 * Maximum length for the description preview.
 */
const DESCRIPTION_PREVIEW_LENGTH = 200;

/**
 * Formats a date string for display.
 */
function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

/**
 * Truncates a string to a maximum length, adding ellipsis if truncated.
 */
function truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength).trim() + '...';
}

/**
 * Right panel component for previewing a selected entity.
 *
 * Displays entity details including name, type, tags, description preview,
 * dates, and relationship count. Provides quick action buttons for editing
 * and deleting the entity.
 *
 * When no entity is selected, displays a placeholder message prompting the
 * user to select an entity.
 *
 * @param props - The component props.
 * @returns A React element containing the entity preview panel.
 */
export default function EntityPreviewPanel({
    entity,
    campaignId,
    onEdit,
    onDelete,
}: EntityPreviewPanelProps) {
    // Fetch relationships for the selected entity
    const { data: relationships } = useEntityRelationships(
        campaignId,
        entity?.id ?? '',
        { enabled: !!entity }
    );

    // Empty state when no entity is selected
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
                    Select an entity to preview
                </Typography>
            </Box>
        );
    }

    const relationshipCount = relationships?.length ?? 0;

    return (
        <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
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

            {/* Description preview */}
            <Box sx={{ mb: 2, flexGrow: 1, overflow: 'auto' }}>
                <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ display: 'block', mb: 0.5 }}
                >
                    Description
                </Typography>
                <Typography variant="body2">
                    {entity.description
                        ? truncateText(entity.description, DESCRIPTION_PREVIEW_LENGTH)
                        : 'No description'}
                </Typography>
            </Box>

            <Divider sx={{ my: 1 }} />

            {/* Relationship count */}
            <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <RelationshipIcon fontSize="small" color="action" />
                    <Typography variant="body2" color="text.secondary">
                        {relationshipCount} relationship{relationshipCount !== 1 ? 's' : ''}
                    </Typography>
                </Box>
            </Box>

            {/* Dates */}
            <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                    <Typography variant="caption" color="text.secondary">
                        Created
                    </Typography>
                    <Typography variant="caption">
                        {formatDate(entity.createdAt)}
                    </Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography variant="caption" color="text.secondary">
                        Updated
                    </Typography>
                    <Typography variant="caption">
                        {formatDate(entity.updatedAt)}
                    </Typography>
                </Box>
            </Box>

            <Divider sx={{ my: 1 }} />

            {/* Action buttons */}
            <Box sx={{ display: 'flex', gap: 1, mt: 'auto', pt: 1 }}>
                <Button
                    variant="contained"
                    startIcon={<EditIcon />}
                    onClick={onEdit}
                    size="small"
                    fullWidth
                >
                    Edit
                </Button>
                <Button
                    variant="outlined"
                    color="error"
                    startIcon={<DeleteIcon />}
                    onClick={onDelete}
                    size="small"
                    fullWidth
                >
                    Delete
                </Button>
            </Box>
        </Box>
    );
}
