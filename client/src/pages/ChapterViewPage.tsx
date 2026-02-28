// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterViewPage - Read-only chapter detail page.
 *
 * Displays chapter information in a readable format with navigation to the
 * chapter editor. Wiki links within the overview resolve to entity view
 * pages, enabling wiki-style browsing of campaign content.
 */

import { useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Chip,
    CircularProgress,
    Container,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Paper,
    Typography,
} from '@mui/material';
import {
    ArrowBack as ArrowBackIcon,
    Edit as EditIcon,
    Delete as DeleteIcon,
    Link as RelationshipIcon,
} from '@mui/icons-material';
import {
    useChapter,
    useDeleteChapter,
    useEntities,
    useSessionsByChapter,
    useChapterEntities,
    useChapterRelationships,
} from '../hooks';
import { MarkdownRenderer } from '../components/MarkdownRenderer';
import type { WikiLinkEntity } from '../components/MarkdownRenderer';
import SessionStageIndicator from '../components/Sessions/SessionStageIndicator';
import type {
    EntityType,
    ChapterEntityMentionType,
} from '../types';

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
 * Human-readable labels for mention types.
 */
const MENTION_TYPE_LABELS: Record<ChapterEntityMentionType, string> = {
    featured: 'Featured',
    linked: 'Linked',
    mentioned: 'Mentioned',
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
 * Read-only detail page for viewing a single campaign chapter.
 *
 * Loads chapter data along with associated entities, sessions, and
 * relationships. Renders the chapter overview with wiki-link navigation,
 * entity groupings by mention type, session listings with stage indicators,
 * relationship visualizations, and metadata.
 *
 * @returns The React element for the Chapter View page
 */
export default function ChapterViewPage() {
    const params = useParams<{ campaignId: string; chapterId: string }>();
    const campaignId = params.campaignId ? Number(params.campaignId) : undefined;
    const chapterId = params.chapterId ? Number(params.chapterId) : undefined;
    const navigate = useNavigate();

    // Local UI state
    const [isEditing, setIsEditing] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    // Fetch chapter data
    const {
        data: chapter,
        isLoading,
        error,
    } = useChapter(campaignId ?? 0, chapterId ?? 0);

    // Fetch chapter entities
    const { data: chapterEntities } = useChapterEntities(
        campaignId ?? 0,
        chapterId ?? 0
    );

    // Fetch sessions for this chapter
    const { data: sessions } = useSessionsByChapter(
        campaignId ?? 0,
        chapterId ?? 0
    );

    // Fetch relationships involving chapter entities
    const { data: relationships } = useChapterRelationships(
        campaignId ?? 0,
        chapterId ?? 0
    );

    // Fetch all entities for wiki link resolution
    const { data: allEntities } = useEntities({
        campaignId: campaignId ?? 0,
    });

    // Delete mutation
    const deleteChapter = useDeleteChapter();

    // Group chapter entities by mention type
    const groupedEntities = useMemo(() => {
        if (!chapterEntities) return { featured: [], linked: [], mentioned: [] };
        return {
            featured: chapterEntities.filter(ce => ce.mentionType === 'featured'),
            linked: chapterEntities.filter(ce => ce.mentionType === 'linked'),
            mentioned: chapterEntities.filter(ce => ce.mentionType === 'mentioned'),
        };
    }, [chapterEntities]);

    // Build a set of entity IDs in this chapter for relationship styling
    const chapterEntityIds = useMemo(() => {
        if (!chapterEntities) return new Set<number>();
        return new Set(chapterEntities.map(ce => ce.entityId));
    }, [chapterEntities]);

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
     * Navigate back to the Sessions page.
     */
    const handleBack = useCallback(() => {
        if (campaignId) {
            navigate(`/campaigns/${campaignId}/sessions`);
        }
    }, [campaignId, navigate]);

    /**
     * Enter edit mode (placeholder for future implementation).
     */
    const handleEdit = useCallback(() => {
        setIsEditing(true);
    }, []);

    /**
     * Delete the chapter and navigate back.
     */
    const handleDelete = async () => {
        if (!campaignId || !chapterId) return;
        try {
            await deleteChapter.mutateAsync({ campaignId, chapterId });
            navigate(`/campaigns/${campaignId}/sessions`);
        } catch {
            // Error handled by mutation
        }
    };

    // Suppress unused variable warning — edit mode is set but not yet
    // consumed because the editing UI will be added in a future commit.
    void isEditing;

    // Loading state
    if (isLoading) {
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
    if (error || !chapter) {
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
                    {error
                        ? 'Failed to load chapter. Please try again later.'
                        : 'Chapter not found.'}
                </Alert>
            </Container>
        );
    }

    const sessionCount = sessions?.length ?? 0;
    const relationshipCount = relationships?.length ?? 0;

    return (
        <Container maxWidth="md" sx={{ py: 4 }}>
            {/* Header bar */}
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
                <Box sx={{ display: 'flex', gap: 1 }}>
                    <Button
                        variant="contained"
                        startIcon={<EditIcon />}
                        onClick={handleEdit}
                    >
                        Edit
                    </Button>
                    <Button
                        variant="outlined"
                        color="error"
                        startIcon={<DeleteIcon />}
                        onClick={() => setDeleteDialogOpen(true)}
                    >
                        Delete
                    </Button>
                </Box>
            </Box>

            <Paper sx={{ p: 3 }}>
                {/* Title */}
                <Box sx={{ mb: 3 }}>
                    <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                        {chapter.title}
                    </Typography>
                </Box>

                <Divider sx={{ mb: 3 }} />

                {/* Overview */}
                <Box sx={{ mb: 3 }}>
                    <Typography
                        variant="subtitle2"
                        color="text.secondary"
                        gutterBottom
                    >
                        Overview
                    </Typography>
                    {chapter.overview ? (
                        <Box
                            sx={{
                                '& p': { mt: 0, mb: 1 },
                                '& p:last-child': { mb: 0 },
                            }}
                        >
                            <MarkdownRenderer
                                content={chapter.overview}
                                onEntityClick={handleEntityClick}
                                entities={wikiLinkEntities}
                                onEntityNavigate={handleEntityNavigate}
                            />
                        </Box>
                    ) : (
                        <Typography color="text.secondary">
                            No overview yet.
                        </Typography>
                    )}
                </Box>

                {/* Entities */}
                {chapterEntities && chapterEntities.length > 0 && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            Entities
                        </Typography>
                        <Box
                            sx={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 2,
                            }}
                        >
                            {(
                                ['featured', 'linked', 'mentioned'] as const
                            ).map((mentionType) => {
                                const entities = groupedEntities[mentionType];
                                if (entities.length === 0) return null;
                                return (
                                    <Box key={mentionType}>
                                        <Typography
                                            variant="body2"
                                            color="text.secondary"
                                            sx={{ mb: 0.5 }}
                                        >
                                            {MENTION_TYPE_LABELS[mentionType]}
                                        </Typography>
                                        <Box
                                            sx={{
                                                display: 'flex',
                                                flexDirection: 'column',
                                                gap: 0.5,
                                            }}
                                        >
                                            {entities.map((ce) => (
                                                <Box
                                                    key={ce.id}
                                                    sx={{
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        gap: 1,
                                                    }}
                                                >
                                                    <Button
                                                        size="small"
                                                        sx={{
                                                            textTransform: 'none',
                                                        }}
                                                        onClick={() => {
                                                            if (campaignId && ce.entity) {
                                                                navigate(
                                                                    `/campaigns/${campaignId}/entities/${ce.entity.id}`
                                                                );
                                                            }
                                                        }}
                                                    >
                                                        {ce.entity?.name ??
                                                            `Entity #${ce.entityId}`}
                                                    </Button>
                                                    {ce.entity && (
                                                        <Chip
                                                            label={
                                                                ENTITY_TYPE_LABELS[
                                                                    ce.entity.entityType
                                                                ]
                                                            }
                                                            color={
                                                                ENTITY_TYPE_COLORS[
                                                                    ce.entity.entityType
                                                                ]
                                                            }
                                                            size="small"
                                                        />
                                                    )}
                                                    <Chip
                                                        label={
                                                            MENTION_TYPE_LABELS[
                                                                ce.mentionType
                                                            ]
                                                        }
                                                        size="small"
                                                        variant="outlined"
                                                    />
                                                </Box>
                                            ))}
                                        </Box>
                                    </Box>
                                );
                            })}
                        </Box>
                    </Box>
                )}

                {/* Sessions */}
                {sessionCount > 0 && (
                    <Box sx={{ mb: 3 }}>
                        <Typography
                            variant="subtitle2"
                            color="text.secondary"
                            gutterBottom
                        >
                            Sessions ({sessionCount})
                        </Typography>
                        <Box
                            sx={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 1,
                            }}
                        >
                            {sessions!.map((session) => (
                                <Box
                                    key={session.id}
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 1,
                                    }}
                                >
                                    <Button
                                        size="small"
                                        sx={{ textTransform: 'none' }}
                                        onClick={() => {
                                            if (campaignId) {
                                                navigate(
                                                    `/campaigns/${campaignId}/sessions/${session.id}`
                                                );
                                            }
                                        }}
                                    >
                                        {session.title ??
                                            `Session #${session.sessionNumber ?? session.id}`}
                                    </Button>
                                    {session.plannedDate && (
                                        <Typography
                                            variant="body2"
                                            color="text.secondary"
                                        >
                                            {formatDate(session.plannedDate)}
                                        </Typography>
                                    )}
                                    <SessionStageIndicator stage={session.stage} />
                                </Box>
                            ))}
                        </Box>
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
                                const sourceIsExternal = !chapterEntityIds.has(
                                    rel.sourceEntityId
                                );
                                const targetIsExternal = !chapterEntityIds.has(
                                    rel.targetEntityId
                                );
                                const typeLabel =
                                    rel.displayLabel ?? rel.relationshipType;

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
                                        <Typography
                                            variant="body2"
                                            sx={{
                                                ...(sourceIsExternal && {
                                                    color: 'text.secondary',
                                                    fontStyle: 'italic',
                                                }),
                                            }}
                                        >
                                            {rel.sourceEntityName ??
                                                `Entity #${rel.sourceEntityId}`}
                                        </Typography>
                                        <Chip
                                            label={typeLabel}
                                            size="small"
                                            variant="outlined"
                                        />
                                        <Typography
                                            variant="body2"
                                            sx={{
                                                ...(targetIsExternal && {
                                                    color: 'text.secondary',
                                                    fontStyle: 'italic',
                                                }),
                                            }}
                                        >
                                            {rel.targetEntityName ??
                                                `Entity #${rel.targetEntityId}`}
                                        </Typography>
                                    </Box>
                                );
                            })}
                        </Box>
                    </Box>
                )}

                <Divider sx={{ mb: 2 }} />

                {/* Metadata */}
                <Box sx={{ display: 'flex', gap: 4 }}>
                    <Box>
                        <Typography variant="caption" color="text.secondary">
                            Sort Order
                        </Typography>
                        <Typography variant="body2">
                            {chapter.sortOrder}
                        </Typography>
                    </Box>
                    <Box>
                        <Typography variant="caption" color="text.secondary">
                            Created
                        </Typography>
                        <Typography variant="body2">
                            {formatDate(chapter.createdAt)}
                        </Typography>
                    </Box>
                    <Box>
                        <Typography variant="caption" color="text.secondary">
                            Updated
                        </Typography>
                        <Typography variant="body2">
                            {formatDate(chapter.updatedAt)}
                        </Typography>
                    </Box>
                </Box>
            </Paper>

            {/* Delete confirmation dialog */}
            <Dialog
                open={deleteDialogOpen}
                onClose={() => setDeleteDialogOpen(false)}
            >
                <DialogTitle>Delete Chapter</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Are you sure you want to delete &quot;{chapter.title}&quot;?
                        This action cannot be undone.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setDeleteDialogOpen(false)}>
                        Cancel
                    </Button>
                    <Button
                        onClick={handleDelete}
                        color="error"
                        variant="contained"
                    >
                        Delete
                    </Button>
                </DialogActions>
            </Dialog>
        </Container>
    );
}
