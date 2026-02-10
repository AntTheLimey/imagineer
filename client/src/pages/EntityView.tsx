// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * EntityView - Read-only entity detail page.
 *
 * Displays entity information in a readable format with navigation to the
 * entity editor. Wiki links within the description and GM notes resolve to
 * other entity view pages, enabling wiki-style browsing of campaign content.
 */

import { useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Chip,
    CircularProgress,
    Container,
    Divider,
    Paper,
    Typography,
} from '@mui/material';
import {
    ArrowBack as ArrowBackIcon,
    Edit as EditIcon,
    HistoryEdu as LogIcon,
    Link as RelationshipIcon,
} from '@mui/icons-material';
import {
    useEntity,
    useEntities,
    useEntityRelationships,
    useEntityLogs,
    useCampaignOwnership,
} from '../hooks';
import { MarkdownRenderer } from '../components/MarkdownRenderer';
import type { WikiLinkEntity } from '../components/MarkdownRenderer';
import type { EntityType } from '../types';

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
const ENTITY_TYPE_COLORS: Record<
    EntityType,
    'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info'
> = {
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
 * Formats a date string for display.
 */
function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    });
}

/**
 * Formats a date string as a short date and time for log timestamps.
 */
function formatDateTime(dateString: string): string {
    return new Date(dateString).toLocaleString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

/**
 * Read-only detail page for viewing a single campaign entity.
 *
 * Loads entity data via the `useEntity` hook and renders the entity name,
 * type, tags, description, GM notes, relationships, and metadata timestamps.
 * Markdown content supports wiki-link navigation to other entity view pages.
 *
 * @returns The React element for the Entity View page
 */
export default function EntityView() {
    const params = useParams<{
        campaignId: string;
        entityId: string;
    }>();
    const campaignId = params.campaignId ? Number(params.campaignId) : undefined;
    const entityId = params.entityId ? Number(params.entityId) : undefined;
    const navigate = useNavigate();

    // Check if current user is the campaign owner (GM)
    const { isOwner: isGM } = useCampaignOwnership(campaignId ?? 0);

    // Fetch the entity
    const {
        data: entity,
        isLoading: entityLoading,
        error: entityError,
    } = useEntity(campaignId ?? 0, entityId ?? 0, {
        enabled: !!campaignId && !!entityId,
    });

    // Fetch all entities for wiki link resolution
    const { data: allEntities } = useEntities({
        campaignId: campaignId ?? 0,
    });

    // Map entities to WikiLinkEntity shape for MarkdownRenderer
    const wikiLinkEntities: WikiLinkEntity[] | undefined = useMemo(
        () =>
            allEntities?.map((e) => ({
                id: e.id,
                name: e.name,
                entityType: e.entityType,
                description: e.description ?? null,
            })),
        [allEntities]
    );

    // Fetch relationships for this entity
    const { data: relationships } = useEntityRelationships(
        campaignId ?? 0,
        entityId ?? 0,
        { enabled: !!campaignId && !!entityId }
    );

    // Fetch log entries for this entity
    const { data: logs } = useEntityLogs(
        campaignId ?? 0,
        entityId ?? 0,
    );

    /**
     * Navigate to another entity's view page when a wiki link is clicked.
     */
    const handleEntityNavigate = useCallback(
        (targetEntityId: number) => {
            if (campaignId) {
                navigate(`/campaigns/${campaignId}/entities/${targetEntityId}`);
            }
        },
        [campaignId, navigate]
    );

    /**
     * Fallback handler for unresolved wiki link clicks.
     */
    const handleEntityClick = useCallback(
        (name: string) => {
            if (campaignId) {
                navigate(
                    `/campaigns/${campaignId}/entities?search=${encodeURIComponent(name)}`
                );
            }
        },
        [campaignId, navigate]
    );

    /**
     * Navigate back using browser history.
     */
    const handleBack = useCallback(() => {
        navigate(-1);
    }, [navigate]);

    /**
     * Navigate to the entity editor.
     */
    const handleEdit = useCallback(() => {
        if (campaignId && entityId) {
            navigate(`/campaigns/${campaignId}/entities/${entityId}/edit`);
        }
    }, [campaignId, entityId, navigate]);

    /**
     * Resolve the name of a related entity from the loaded entity list.
     */
    const resolveEntityName = useCallback(
        (relatedEntityId: number): string => {
            const found = allEntities?.find((e) => e.id === relatedEntityId);
            return found?.name ?? `Entity #${relatedEntityId}`;
        },
        [allEntities]
    );

    // Loading state
    if (entityLoading) {
        return (
            <Container maxWidth="md" sx={{ py: 4 }}>
                <Box
                    sx={{
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center',
                        minHeight: 300,
                    }}
                >
                    <CircularProgress />
                </Box>
            </Container>
        );
    }

    // Error state
    if (entityError || !entity) {
        return (
            <Container maxWidth="md" sx={{ py: 4 }}>
                <Button
                    startIcon={<ArrowBackIcon />}
                    onClick={handleBack}
                    sx={{ mb: 2 }}
                >
                    Back
                </Button>
                <Alert severity="error">
                    {entityError
                        ? 'Failed to load entity. Please try again later.'
                        : 'Entity not found.'}
                </Alert>
            </Container>
        );
    }

    const relationshipCount = relationships?.length ?? 0;

    return (
        <Container maxWidth="md" sx={{ py: 4 }}>
            {/* Navigation buttons */}
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    mb: 3,
                }}
            >
                <Button
                    startIcon={<ArrowBackIcon />}
                    onClick={handleBack}
                >
                    Back
                </Button>
                <Button
                    variant="contained"
                    startIcon={<EditIcon />}
                    onClick={handleEdit}
                >
                    Edit
                </Button>
            </Box>

            <Paper sx={{ p: 3 }}>
                {/* Header: Name, type chip, and tags */}
                <Box sx={{ mb: 3 }}>
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 2,
                            flexWrap: 'wrap',
                            mb: 1,
                        }}
                    >
                        <Typography
                            variant="h4"
                            sx={{ fontFamily: 'Cinzel' }}
                        >
                            {entity.name}
                        </Typography>
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
                        <Box
                            sx={{
                                display: 'flex',
                                gap: 0.5,
                                flexWrap: 'wrap',
                                mt: 1,
                            }}
                        >
                            {entity.tags.map((tag) => (
                                <Chip
                                    key={tag}
                                    label={tag}
                                    size="small"
                                    variant="outlined"
                                />
                            ))}
                        </Box>
                    )}
                </Box>

                <Divider sx={{ mb: 3 }} />

                {/* Description */}
                {entity.description && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            Description
                        </Typography>
                        <Box
                            sx={{
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
                    </Box>
                )}

                {/* GM Notes */}
                {isGM && entity.gmNotes && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            GM Notes
                        </Typography>
                        <Paper
                            variant="outlined"
                            sx={{
                                p: 2,
                                bgcolor: 'warning.light',
                                color: 'warning.contrastText',
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
                        </Paper>
                    </Box>
                )}

                {/* Attributes */}
                {entity.attributes &&
                    Object.keys(entity.attributes).length > 0 && (
                        <Box sx={{ mb: 3 }}>
                            <Typography
                                variant="subtitle2"
                                color="text.secondary"
                                gutterBottom
                            >
                                Attributes
                            </Typography>
                            <Paper variant="outlined" sx={{ p: 2 }}>
                                <pre
                                    style={{
                                        margin: 0,
                                        fontSize: '0.875rem',
                                        overflow: 'auto',
                                    }}
                                >
                                    {JSON.stringify(
                                        entity.attributes,
                                        null,
                                        2
                                    )}
                                </pre>
                            </Paper>
                        </Box>
                    )}

                {/* Relationships */}
                {relationshipCount > 0 && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            Relationships
                        </Typography>
                        <Box
                            sx={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 1,
                            }}
                        >
                            {relationships!.map((rel) => {
                                // The backend view returns relationships
                                // oriented from the queried entity's
                                // perspective. Use targetEntityId (the
                                // "other" entity) for navigation and
                                // displayLabel for the type chip.
                                const relatedId = rel.targetEntityId;
                                const relatedName =
                                    rel.targetEntityName
                                    ?? resolveEntityName(relatedId);
                                const typeLabel =
                                    rel.displayLabel
                                    ?? rel.relationshipType;

                                return (
                                    <Box
                                        key={rel.id}
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 1,
                                        }}
                                    >
                                        <RelationshipIcon
                                            fontSize="small"
                                            color="action"
                                        />
                                        <Chip
                                            label={typeLabel}
                                            size="small"
                                            variant="outlined"
                                        />
                                        <Button
                                            size="small"
                                            sx={{ textTransform: 'none' }}
                                            onClick={() =>
                                                handleEntityNavigate(relatedId)
                                            }
                                        >
                                            {relatedName}
                                        </Button>
                                        {rel.description && (
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                                sx={{
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    whiteSpace: 'nowrap',
                                                }}
                                            >
                                                {rel.description}
                                            </Typography>
                                        )}
                                    </Box>
                                );
                            })}
                        </Box>
                    </Box>
                )}

                {/* Entity Log */}
                {logs && logs.length > 0 && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            Event Log
                        </Typography>
                        <Box
                            sx={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 1,
                            }}
                        >
                            {logs.map((log) => (
                                <Paper
                                    key={log.id}
                                    variant="outlined"
                                    sx={{ p: 2 }}
                                >
                                    <Box
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'flex-start',
                                            gap: 1.5,
                                        }}
                                    >
                                        <LogIcon
                                            fontSize="small"
                                            color="action"
                                            sx={{ mt: 0.25 }}
                                        />
                                        <Box sx={{ flex: 1 }}>
                                            <Typography variant="body2">
                                                {log.content}
                                            </Typography>
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 2,
                                                    mt: 0.5,
                                                }}
                                            >
                                                {log.occurredAt && (
                                                    <Typography
                                                        variant="caption"
                                                        color="text.secondary"
                                                    >
                                                        Occurred: {log.occurredAt}
                                                    </Typography>
                                                )}
                                                <Typography
                                                    variant="caption"
                                                    color="text.secondary"
                                                >
                                                    Logged: {formatDateTime(log.createdAt)}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </Box>
                                </Paper>
                            ))}
                        </Box>
                    </Box>
                )}

                <Divider sx={{ mb: 2 }} />

                {/* Metadata */}
                <Box sx={{ display: 'flex', gap: 4 }}>
                    <Box>
                        <Typography
                            variant="caption"
                            color="text.secondary"
                        >
                            Created
                        </Typography>
                        <Typography variant="body2">
                            {formatDate(entity.createdAt)}
                        </Typography>
                    </Box>
                    <Box>
                        <Typography
                            variant="caption"
                            color="text.secondary"
                        >
                            Updated
                        </Typography>
                        <Typography variant="body2">
                            {formatDate(entity.updatedAt)}
                        </Typography>
                    </Box>
                    <Box>
                        <Typography
                            variant="caption"
                            color="text.secondary"
                        >
                            Version
                        </Typography>
                        <Typography variant="body2">
                            {entity.version}
                        </Typography>
                    </Box>
                </Box>
            </Paper>
        </Container>
    );
}
