// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * EntityEditor - Full-screen entity editing page.
 *
 * Provides a focused editing experience for creating and editing entities
 * with autosave, draft recovery, and unsaved changes protection.
 */

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Autocomplete,
    Box,
    Button,
    Chip,
    Divider,
    FormControl,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Snackbar,
    TextField,
    Typography,
} from '@mui/material';
import { FullScreenLayout } from '../layouts';
import { SaveSplitButton } from '../components/SaveSplitButton';
import type { SaveMode } from '../components/SaveSplitButton';
import { MarkdownEditor } from '../components/MarkdownEditor';
import { RelationshipEditor, PendingRelationship } from '../components/RelationshipEditor';
import EntityLogSection from '../components/EntityLog/EntityLogSection';
import {
    useEntity,
    useCreateEntity,
    useUpdateEntity,
    useSimilarEntities,
    useCampaign,
    useServerDraft,
    useCampaignOwnership,
} from '../hooks';
import { useCreateRelationship } from '../hooks/useRelationships';
import type { EntityType, SourceConfidence } from '../types';

/**
 * All available entity types.
 */
const ENTITY_TYPES: EntityType[] = [
    'npc',
    'location',
    'item',
    'faction',
    'clue',
    'creature',
    'organization',
    'event',
    'document',
    'other',
];

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
 * Source confidence options.
 */
const SOURCE_CONFIDENCE_OPTIONS: SourceConfidence[] = [
    'DRAFT',
    'AUTHORITATIVE',
    'SUPERSEDED',
];

/**
 * Form data structure for entity editing.
 */
interface EntityFormData {
    name: string;
    entityType: EntityType;
    description: string;
    tags: string[];
    attributes: string;
    gmNotes: string;
    sourceConfidence: SourceConfidence;
}

/**
 * Default form values for new entities.
 */
const DEFAULT_FORM_DATA: EntityFormData = {
    name: '',
    entityType: 'npc',
    description: '',
    tags: [],
    attributes: '{}',
    gmNotes: '',
    sourceConfidence: 'DRAFT',
};

/**
 * Determine whether a string is parseable as JSON.
 *
 * @returns `true` if `str` contains valid JSON, `false` otherwise.
 */
function isValidJson(str: string): boolean {
    try {
        JSON.parse(str);
        return true;
    } catch {
        return false;
    }
}

/**
 * Produce a human-friendly label for how long ago an ISO timestamp occurred.
 *
 * @param isoString - An ISO 8601 timestamp string (e.g., "2026-02-01T12:34:56Z")
 * @returns `just now` for times less than 1 minute ago, `N minutes ago` for times less than 1 hour, `N hours ago` for times less than 24 hours, or the locale-formatted date for older timestamps
 */
function formatRelativeTime(isoString: string): string {
    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);

    if (diffMins < 1) {
        return 'just now';
    }
    if (diffMins < 60) {
        return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
    }
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) {
        return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
    }
    return date.toLocaleDateString();
}

/**
 * Full-screen editor for creating and editing campaign entities with autosave, draft recovery, validation, and unsaved-change protection.
 *
 * Renders the entity form (name, type, description, tags, JSON attributes, GM notes, source confidence), manages drafts and autosave, detects similar entities for duplicates, handles create/update mutations and relationship creation, and prevents unintended navigation when there are unsaved changes.
 *
 * @returns The React element for the Entity Editor page
 */
export default function EntityEditor() {
    const params = useParams<{
        campaignId: string;
        entityId?: string;
    }>();
    const campaignId = params.campaignId ? Number(params.campaignId) : undefined;
    const entityId = params.entityId && params.entityId !== 'new' ? Number(params.entityId) : undefined;
    const navigate = useNavigate();

    const isNewEntity = !entityId;

    // Fetch campaign for breadcrumbs
    const { data: campaign } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Check if current user is the campaign owner (GM)
    const { isOwner: isGM } = useCampaignOwnership(campaignId ?? 0);

    // Fetch existing entity if editing
    const {
        data: existingEntity,
        isLoading: entityLoading,
        error: entityError,
    } = useEntity(campaignId ?? 0, entityId ?? 0, {
        enabled: !isNewEntity && !!campaignId && !!entityId,
    });

    // Form state
    const [formData, setFormData] = useState<EntityFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<
        Partial<Record<keyof EntityFormData, string>>
    >({});

    // Server-side draft management (replaces useDraft + useAutosave + useUnsavedChanges)
    const {
        serverDraft,
        hasDraft: hasServerDraft,
        saveDraftToServer,
        deleteDraftFromServer,
        isDirty,
        lastSaved,
        isLoading: draftLoading,
        draftUpdatedAt,
    } = useServerDraft<EntityFormData>({
        campaignId: campaignId ?? 0,
        sourceTable: 'entities',
        sourceId: entityId ?? 0,
        isNew: isNewEntity,
        committedData: existingEntity ? {
            name: existingEntity.name,
            entityType: existingEntity.entityType,
            description: existingEntity.description ?? '',
            tags: existingEntity.tags ?? [],
            attributes: JSON.stringify(existingEntity.attributes ?? {}, null, 2),
            gmNotes: existingEntity.gmNotes ?? '',
            sourceConfidence: existingEntity.sourceConfidence,
        } : undefined,
        currentData: formData,
        serverVersion: existingEntity?.version,
        enabled: !!campaignId,
    });

    const [showDraftRecovery, setShowDraftRecovery] = useState(false);

    // Track the last hydrated entity ID to detect route changes
    const lastHydratedEntityIdRef = useRef<number | undefined>(undefined);

    // Similar entities for duplicate detection (new entities only)
    const { data: similarEntities } = useSimilarEntities(
        campaignId ?? 0,
        formData.name,
        { enabled: isNewEntity && formData.name.length >= 2 }
    );

    // State for pending relationships (for new entities)
    const [pendingRelationships, setPendingRelationships] = useState<PendingRelationship[]>([]);

    // Mutations
    const createEntity = useCreateEntity();
    const updateEntity = useUpdateEntity();
    const createRelationship = useCreateRelationship();

    const isSaving = createEntity.isPending || updateEntity.isPending;

    // Analysis snackbar state
    const [analysisSnackbar, setAnalysisSnackbar] = useState<{
        open: boolean;
        jobId: number;
        count: number;
        message?: string;
    }>({ open: false, jobId: 0, count: 0 });

    // Initialize form data from existing entity
    useEffect(() => {
        if (!isNewEntity && existingEntity) {
            const isNewEntityRoute = existingEntity.id !== lastHydratedEntityIdRef.current;
            if (!isNewEntityRoute) return;

            setFormData({
                name: existingEntity.name,
                entityType: existingEntity.entityType,
                description: existingEntity.description ?? '',
                tags: existingEntity.tags ?? [],
                attributes: JSON.stringify(existingEntity.attributes ?? {}, null, 2),
                gmNotes: existingEntity.gmNotes ?? '',
                sourceConfidence: existingEntity.sourceConfidence,
            });

            lastHydratedEntityIdRef.current = existingEntity.id;
        } else if (isNewEntity) {
            lastHydratedEntityIdRef.current = undefined;
        }
    }, [existingEntity, isNewEntity]);

    // Auto-apply server draft on load: show draft data with a dismissable
    // banner. Uses a ref to ensure the draft is only applied once on
    // initial load; after the user ditches the draft, a brief stale
    // reference to serverDraft will not re-apply it.
    const draftAppliedRef = useRef(false);
    useEffect(() => {
        if (!draftLoading && !draftAppliedRef.current) {
            draftAppliedRef.current = true;
            if (hasServerDraft && serverDraft !== null) {
                setFormData(serverDraft);
                setShowDraftRecovery(true);
            }
        }
    }, [draftLoading, hasServerDraft, serverDraft]);

    /**
     * Dismiss the draft banner without discarding the draft.
     */
    const handleDismissDraftBanner = useCallback(() => {
        setShowDraftRecovery(false);
    }, []);

    /**
     * Ditch the draft: delete from the server and restore committed data.
     */
    const handleDitchDraft = useCallback(() => {
        // Delete draft from server
        deleteDraftFromServer();
        // Restore committed data
        if (existingEntity) {
            setFormData({
                name: existingEntity.name,
                entityType: existingEntity.entityType,
                description: existingEntity.description ?? '',
                tags: existingEntity.tags ?? [],
                attributes: JSON.stringify(existingEntity.attributes ?? {}, null, 2),
                gmNotes: existingEntity.gmNotes ?? '',
                sourceConfidence: existingEntity.sourceConfidence,
            });
        } else {
            setFormData(DEFAULT_FORM_DATA);
        }
        setShowDraftRecovery(false);
    }, [deleteDraftFromServer, existingEntity]);

    /**
     * Update a form field. Dirty state is computed automatically by
     * useServerDraft from currentData vs committedData.
     */
    const updateField = useCallback(
        <K extends keyof EntityFormData>(field: K, value: EntityFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            // Clear error for this field
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));
        },
        []
    );

    /**
     * Validate the form.
     */
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof EntityFormData, string>> = {};

        if (!formData.name.trim()) {
            errors.name = 'Name is required';
        }

        if (!formData.entityType) {
            errors.entityType = 'Type is required';
        }

        if (!isValidJson(formData.attributes)) {
            errors.attributes = 'Invalid JSON format';
        }

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, [formData]);

    /**
     * Save the entity with an optional save mode that controls whether
     * analysis and enrichment pipelines are triggered after saving.
     *
     * @param mode - The save mode: "save" (no analysis), "analyze"
     *   (phase 1 only), or "enrich" (phase 1 + phase 2). Defaults to
     *   "analyze" to preserve the existing auto-analyze behavior.
     */
    const handleSave = useCallback(async (mode: SaveMode = 'analyze'): Promise<boolean> => {
        if (!validateForm() || !campaignId) {
            return false;
        }

        try {
            if (isNewEntity) {
                const newEntity = await createEntity.mutateAsync({
                    campaignId,
                    name: formData.name.trim(),
                    entityType: formData.entityType,
                    description: formData.description.trim() || undefined,
                    tags: formData.tags.length > 0 ? formData.tags : undefined,
                    attributes: JSON.parse(formData.attributes),
                    gmNotes: formData.gmNotes.trim() || undefined,
                    sourceConfidence: formData.sourceConfidence,
                });

                // Create pending relationships after entity is created using Promise.allSettled
                if (pendingRelationships.length > 0) {
                    const relationshipResults = await Promise.allSettled(
                        pendingRelationships.map((rel) => {
                            // When direction is incoming (isReversed),
                            // swap source and target so the stored row
                            // is always in the canonical forward
                            // direction.
                            const srcId = rel.isReversed
                                ? rel.targetEntityId
                                : newEntity.id;
                            const tgtId = rel.isReversed
                                ? newEntity.id
                                : rel.targetEntityId;

                            return createRelationship.mutateAsync({
                                campaignId,
                                sourceEntityId: srcId,
                                targetEntityId: tgtId,
                                relationshipTypeId: rel.relationshipTypeId!,
                                description: rel.description,
                            });
                        }
                        )
                    );
                    const failures = relationshipResults.filter(
                        (r) => r.status === 'rejected'
                    );
                    if (failures.length > 0) {
                        console.warn(`${failures.length} relationship(s) failed to create`);
                    }
                }

                // Clear pending relationships
                setPendingRelationships([]);

                // Clean up draft
                await deleteDraftFromServer();

                // Navigate to edit the newly created entity
                navigate(`/campaigns/${campaignId}/entities/${newEntity.id}/edit`, {
                    replace: true,
                });
            } else if (entityId) {
                const result = await updateEntity.mutateAsync({
                    campaignId,
                    entityId,
                    input: {
                        name: formData.name.trim(),
                        description: formData.description.trim() || undefined,
                        tags: formData.tags,
                        attributes: JSON.parse(formData.attributes),
                        gmNotes: formData.gmNotes.trim() || undefined,
                        sourceConfidence: formData.sourceConfidence,
                    },
                    options: {
                        analyze: mode !== 'save',
                        enrich: mode === 'enrich',
                    },
                });

                // Check for analysis results
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                const analysisResult = (result as any)?._analysis;
                if (mode !== 'save' && analysisResult) {
                    if (mode === 'enrich' && analysisResult.jobId) {
                        // Go directly to triage — no snackbar
                        navigate(`/campaigns/${campaignId}/analysis/${analysisResult.jobId}`);
                    } else if (analysisResult.pendingCount > 0) {
                        setAnalysisSnackbar({
                            open: true,
                            jobId: analysisResult.jobId,
                            count: analysisResult.pendingCount,
                        });
                    } else {
                        setAnalysisSnackbar({
                            open: true,
                            jobId: analysisResult.jobId,
                            count: 0,
                            message: 'Analysis complete: no issues found.',
                        });
                    }
                }

                // Clean up draft
                await deleteDraftFromServer();
            }

            return true;
        } catch (error) {
            console.error('Failed to save entity:', error);
            return false;
        }
    }, [
        validateForm,
        campaignId,
        isNewEntity,
        entityId,
        formData,
        pendingRelationships,
        createEntity,
        createRelationship,
        updateEntity,
        deleteDraftFromServer,
        navigate,
    ]);

    /**
     * Handle back navigation. Auto-saves the draft if there are unsaved
     * changes, then navigates back in browser history.
     */
    const handleBack = useCallback(async () => {
        if (isDirty) {
            await saveDraftToServer();
        }
        navigate(-1);
    }, [isDirty, saveDraftToServer, navigate]);

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Home', path: '/' },
            {
                label: campaign?.name ?? 'Campaign',
                path: `/campaigns/${campaignId}/entities`,
            },
            { label: 'Entities', path: `/campaigns/${campaignId}/entities` },
            { label: isNewEntity ? 'New Entity' : formData.name || 'Edit Entity' },
        ],
        [campaign, campaignId, isNewEntity, formData.name]
    );

    // Loading state
    if (!isNewEntity && entityLoading) {
        return (
            <FullScreenLayout
                title="Loading..."
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
            >
                <Box sx={{ maxWidth: 800, mx: 'auto' }}>
                    <Skeleton variant="rectangular" height={56} sx={{ mb: 2 }} />
                    <Skeleton variant="rectangular" height={120} sx={{ mb: 2 }} />
                    <Skeleton variant="rectangular" height={56} sx={{ mb: 2 }} />
                    <Skeleton variant="rectangular" height={160} />
                </Box>
            </FullScreenLayout>
        );
    }

    // Error state
    if (entityError) {
        return (
            <FullScreenLayout
                title="Error"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                onBack={handleBack}
            >
                <Alert severity="error">
                    Failed to load entity. Please try again later.
                </Alert>
            </FullScreenLayout>
        );
    }

    return (
        <FullScreenLayout
            title={isNewEntity ? 'New Entity' : `Edit: ${existingEntity?.name ?? ''}`}
            subtitle={lastSaved ? `Auto-saved ${formatRelativeTime(lastSaved)}` : undefined}
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={isSaving}
            onBack={handleBack}
            renderSaveButtons={() => (
                <SaveSplitButton
                    onSave={handleSave}
                    isDirty={isDirty}
                    isSaving={isSaving}
                />
            )}
        >
            <Box sx={{ maxWidth: 800, mx: 'auto' }}>
                {/* Draft banner -- auto-applied, dismissable */}
                {showDraftRecovery && (
                    <Alert
                        severity="warning"
                        variant="filled"
                        sx={{ mb: 3 }}
                        onClose={handleDismissDraftBanner}
                        action={
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <Button
                                    size="small"
                                    color="inherit"
                                    variant="outlined"
                                    onClick={handleDitchDraft}
                                    sx={{ whiteSpace: 'nowrap' }}
                                >
                                    Ditch Draft
                                </Button>
                            </Box>
                        }
                    >
                        Showing draft
                        {draftUpdatedAt ? ` \u2014 last saved ${formatRelativeTime(draftUpdatedAt)}` : ''}
                    </Alert>
                )}

                {/* Similar entities warning */}
                {isNewEntity && similarEntities && similarEntities.length > 0 && (
                    <Alert severity="warning" sx={{ mb: 3 }}>
                        Similar entities found: {similarEntities.map((e) => e.name).join(', ')}.
                        Please verify this is not a duplicate.
                    </Alert>
                )}

                {/* Mutation error alerts */}
                {createEntity.error && (
                    <Alert severity="error" sx={{ mb: 3 }}>
                        Failed to create entity. Please try again.
                    </Alert>
                )}
                {updateEntity.error && (
                    <Alert severity="error" sx={{ mb: 3 }}>
                        Failed to update entity. Please try again.
                    </Alert>
                )}

                {/* Form */}
                <Paper sx={{ p: 3 }}>
                    <Box sx={{ display: 'flex', gap: 2, mb: 3 }}>
                        <TextField
                            autoFocus
                            label="Name"
                            fullWidth
                            required
                            value={formData.name}
                            onChange={(e) => updateField('name', e.target.value)}
                            error={!!formErrors.name}
                            helperText={formErrors.name}
                        />
                        <FormControl
                            sx={{ minWidth: 200 }}
                            error={!!formErrors.entityType}
                            disabled={!isNewEntity}
                        >
                            <InputLabel required>Type</InputLabel>
                            <Select
                                value={formData.entityType}
                                label="Type"
                                onChange={(e) =>
                                    updateField('entityType', e.target.value as EntityType)
                                }
                            >
                                {ENTITY_TYPES.map((type) => (
                                    <MenuItem key={type} value={type}>
                                        {ENTITY_TYPE_LABELS[type]}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Box>

                    <Box sx={{ mb: 3 }}>
                        <MarkdownEditor
                            label="Description"
                            value={formData.description}
                            onChange={(md) => updateField('description', md)}
                            placeholder="Describe this entity..."
                            error={!!formErrors.description}
                            helperText={formErrors.description}
                            campaignId={campaignId}
                        />
                    </Box>

                    <Autocomplete
                        multiple
                        freeSolo
                        autoSelect
                        options={[]}
                        value={formData.tags}
                        onChange={(_event, newValue) =>
                            updateField('tags', newValue as string[])
                        }
                        onBlur={(event) => {
                            const inputValue = (event.target as HTMLInputElement).value?.trim();
                            if (inputValue && !formData.tags.includes(inputValue)) {
                                updateField('tags', [...formData.tags, inputValue]);
                            }
                        }}
                        renderTags={(value, getTagProps) =>
                            value.map((option, index) => (
                                <Chip
                                    label={option}
                                    size="small"
                                    {...getTagProps({ index })}
                                    key={option}
                                />
                            ))
                        }
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Tags"
                                placeholder="Type and press Enter to add tags"
                                sx={{ mb: 3 }}
                            />
                        )}
                    />

                    <TextField
                        label="Attributes (JSON)"
                        fullWidth
                        multiline
                        rows={6}
                        value={formData.attributes}
                        onChange={(e) => updateField('attributes', e.target.value)}
                        error={!!formErrors.attributes}
                        helperText={
                            formErrors.attributes ??
                            'Game system specific attributes in JSON format'
                        }
                        sx={{ mb: 3 }}
                        InputProps={{
                            sx: { fontFamily: 'monospace', fontSize: '0.875rem' },
                        }}
                    />

                    {isGM && (
                        <TextField
                            label="GM Notes"
                            fullWidth
                            multiline
                            rows={4}
                            value={formData.gmNotes}
                            onChange={(e) => updateField('gmNotes', e.target.value)}
                            helperText="Private notes visible only to the Game Master"
                            sx={{ mb: 3 }}
                        />
                    )}

                    {/* Relationships section */}
                    <Box sx={{ mb: 3 }}>
                        <Typography variant="subtitle1" gutterBottom fontWeight="medium">
                            Relationships
                        </Typography>
                        {campaignId && (
                            <RelationshipEditor
                                campaignId={campaignId}
                                entityId={isNewEntity ? undefined : entityId}
                                onPendingRelationshipsChange={setPendingRelationships}
                            />
                        )}
                    </Box>

                    {/* Entity Log section */}
                    {!isNewEntity && entityId && campaignId && (
                        <Box sx={{ mt: 3 }}>
                            <Divider sx={{ mb: 2 }} />
                            <EntityLogSection
                                campaignId={campaignId}
                                entityId={entityId}
                            />
                        </Box>
                    )}

                    <FormControl fullWidth>
                        <InputLabel>Source Confidence</InputLabel>
                        <Select
                            value={formData.sourceConfidence}
                            label="Source Confidence"
                            onChange={(e) =>
                                updateField(
                                    'sourceConfidence',
                                    e.target.value as SourceConfidence
                                )
                            }
                        >
                            {SOURCE_CONFIDENCE_OPTIONS.map((opt) => (
                                <MenuItem key={opt} value={opt}>
                                    {opt}
                                </MenuItem>
                            ))}
                        </Select>
                        <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
                            DRAFT for unverified info, AUTHORITATIVE for confirmed canon,
                            SUPERSEDED for outdated info
                        </Typography>
                    </FormControl>
                </Paper>
            </Box>

            {/* Analysis results snackbar */}
            <Snackbar
                open={analysisSnackbar.open}
                autoHideDuration={analysisSnackbar.count === 0 ? 4000 : 10000}
                onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
            >
                {analysisSnackbar.count === 0 ? (
                    <Alert
                        severity="success"
                        onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
                    >
                        {analysisSnackbar.message ?? 'Analysis complete: no issues found.'}
                    </Alert>
                ) : (
                    <Alert
                        severity="warning"
                        onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
                        action={
                            <Button
                                color="inherit"
                                size="small"
                                onClick={() => {
                                    setAnalysisSnackbar(prev => ({ ...prev, open: false }));
                                    navigate(`/campaigns/${campaignId}/analysis/${analysisSnackbar.jobId}`);
                                }}
                            >
                                Review Now
                            </Button>
                        }
                    >
                        {`Analysis found ${analysisSnackbar.count} item${analysisSnackbar.count === 1 ? '' : 's'} to review`}
                    </Alert>
                )}
            </Snackbar>

        </FullScreenLayout>
    );
}