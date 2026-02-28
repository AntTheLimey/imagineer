// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterViewPage - Chapter detail page with inline editing.
 *
 * Displays chapter information in a readable format with wiki-link
 * navigation. Supports an inline edit mode with PhaseStrip for
 * Save & Analyze workflow, entity management, and a rich markdown
 * editor for the overview field.
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
    IconButton,
    Paper,
    TextField,
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
    useUpdateChapter,
    useUserSettings,
    useEntities,
    useSessionsByChapter,
    useChapterEntities,
    useChapterRelationships,
    useCreateChapterEntity,
    useDeleteChapterEntity,
} from '../hooks';
import { PhaseStrip } from '../components/PhaseStrip';
import type { PhaseSelection } from '../components/PhaseStrip';
import { MarkdownEditor } from '../components/MarkdownEditor';
import { MarkdownRenderer } from '../components/MarkdownRenderer';
import type { WikiLinkEntity } from '../components/MarkdownRenderer';
import EntityAutocomplete from '../components/EntityAutocomplete';
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
 * Chapter detail page with inline editing and Save & Analyze workflow.
 *
 * Loads chapter data along with associated entities, sessions, and
 * relationships. In read mode, renders the chapter overview with
 * wiki-link navigation, entity groupings by mention type, session
 * listings with stage indicators, relationship visualizations, and
 * metadata. In edit mode, provides inline form fields, a rich
 * markdown editor, entity management, and the PhaseStrip for
 * triggering analysis phases on save.
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

    // Edit mode form data
    const [editTitle, setEditTitle] = useState('');
    const [editOverview, setEditOverview] = useState('');
    const [editSortOrder, setEditSortOrder] = useState(0);

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

    // Update mutation
    const updateChapter = useUpdateChapter();

    // Chapter entity mutations
    const createChapterEntity = useCreateChapterEntity();
    const deleteChapterEntity = useDeleteChapterEntity();

    // LLM availability check
    const { data: userSettings } = useUserSettings();
    const hasLLM = !!userSettings?.contentGenService && !!userSettings?.contentGenApiKey;
    const disabledPhases = !hasLLM
        ? { revise: 'Configure an LLM in Account Settings',
            enrich: 'Configure an LLM in Account Settings' }
        : undefined;

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
     * Enter edit mode and initialise form fields from current chapter data.
     */
    const handleEdit = useCallback(() => {
        if (chapter) {
            setEditTitle(chapter.title);
            setEditOverview(chapter.overview ?? '');
            setEditSortOrder(chapter.sortOrder);
        }
        setIsEditing(true);
    }, [chapter]);

    /**
     * Exit edit mode without saving changes.
     */
    const handleCancelEdit = useCallback(() => {
        setIsEditing(false);
    }, []);

    /**
     * Save changes and optionally trigger analysis phases.
     */
    const handleSave = useCallback(async (phases: PhaseSelection) => {
        if (!campaignId || !chapterId) return;
        if (!editTitle.trim()) return;

        const hasAnyPhase = phases.identify || phases.revise || phases.enrich;
        const phaseKeys: string[] = [];
        if (phases.identify) phaseKeys.push('identify');
        if (phases.revise) phaseKeys.push('revise');
        if (phases.enrich) phaseKeys.push('enrich');

        try {
            const result = await updateChapter.mutateAsync({
                campaignId,
                chapterId,
                input: {
                    title: editTitle.trim(),
                    overview: editOverview.trim() || undefined,
                    sortOrder: editSortOrder,
                },
                options: {
                    analyze: phases.identify || phases.revise,
                    enrich: phases.enrich,
                    phases: phaseKeys.length > 0 ? phaseKeys : undefined,
                },
            });

            // Navigate to analysis wizard if phases selected
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const analysisResult = (result as any)?._analysis;
            if (hasAnyPhase && analysisResult?.jobId) {
                navigate(`/campaigns/${campaignId}/analysis/${analysisResult.jobId}`);
                return;
            }

            setIsEditing(false);
        } catch (error) {
            console.error('Failed to save chapter:', error);
        }
    }, [campaignId, chapterId, editTitle, editOverview, editSortOrder, updateChapter, navigate]);

    /**
     * Add an entity link to the chapter.
     */
    const handleAddEntity = useCallback((entity: { id: number; name: string; entityType: string }) => {
        if (!campaignId || !chapterId) return;
        createChapterEntity.mutate({
            campaignId,
            chapterId,
            input: { entityId: entity.id, mentionType: 'linked' },
        });
    }, [campaignId, chapterId, createChapterEntity]);

    /**
     * Remove an entity link from the chapter.
     */
    const handleRemoveEntity = useCallback((linkId: number) => {
        if (!campaignId || !chapterId) return;
        deleteChapterEntity.mutate({ campaignId, chapterId, linkId });
    }, [campaignId, chapterId, deleteChapterEntity]);

    /**
     * Delete the chapter and navigate back.
     */
    const handleDelete = useCallback(async () => {
        if (!campaignId || !chapterId) return;
        try {
            await deleteChapter.mutateAsync({ campaignId, chapterId });
            navigate(`/campaigns/${campaignId}/sessions`);
        } catch {
            // Error handled by mutation
        }
    }, [campaignId, chapterId, deleteChapter, navigate]);

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
                <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
                    {isEditing ? (
                        <PhaseStrip
                            onSave={handleSave}
                            isDirty={isEditing}
                            isSaving={updateChapter.isPending}
                            disabledPhases={disabledPhases}
                        />
                    ) : (
                        <Button
                            variant="contained"
                            startIcon={<EditIcon />}
                            onClick={handleEdit}
                        >
                            Edit
                        </Button>
                    )}
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

            {/* Cancel banner in edit mode */}
            {isEditing && (
                <Alert severity="info" sx={{ mb: 2 }}>
                    You are in edit mode. Make your changes and use Save &amp; Go to save.
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleCancelEdit}
                        sx={{ ml: 2 }}
                    >
                        Cancel
                    </Button>
                </Alert>
            )}

            {updateChapter.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to save chapter. Please try again.
                </Alert>
            )}

            <Paper sx={{ p: 3 }}>
                {/* Title */}
                <Box sx={{ mb: 3 }}>
                    {isEditing ? (
                        <TextField
                            label="Chapter Title"
                            fullWidth
                            required
                            value={editTitle}
                            onChange={(e) => setEditTitle(e.target.value)}
                        />
                    ) : (
                        <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                            {chapter.title}
                        </Typography>
                    )}
                </Box>

                <Divider sx={{ mb: 3 }} />

                {/* Sort order (edit mode only) */}
                {isEditing && (
                    <Box sx={{ mb: 2 }}>
                        <TextField
                            label="Sort Order"
                            type="number"
                            value={editSortOrder}
                            onChange={(e) => setEditSortOrder(Number(e.target.value))}
                            size="small"
                            sx={{ width: 120 }}
                        />
                    </Box>
                )}

                {/* Overview */}
                <Box sx={{ mb: 3 }}>
                    <Typography
                        variant="subtitle2"
                        color="text.secondary"
                        gutterBottom
                    >
                        Overview
                    </Typography>
                    {isEditing ? (
                        <MarkdownEditor
                            value={editOverview}
                            onChange={setEditOverview}
                            placeholder="Write chapter overview..."
                            minHeight={200}
                            campaignId={campaignId}
                        />
                    ) : (
                        chapter.overview ? (
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
                        )
                    )}
                </Box>

                {/* Entity autocomplete (edit mode only) */}
                {isEditing && campaignId && (
                    <Box sx={{ mb: 2 }}>
                        <EntityAutocomplete
                            campaignId={campaignId}
                            onSelect={handleAddEntity}
                            label="Add entity to chapter..."
                        />
                    </Box>
                )}

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
                                                    {isEditing && (
                                                        <IconButton
                                                            size="small"
                                                            onClick={() => handleRemoveEntity(ce.id)}
                                                            color="error"
                                                        >
                                                            <DeleteIcon fontSize="small" />
                                                        </IconButton>
                                                    )}
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
