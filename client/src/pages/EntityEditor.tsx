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
    Chip,
    FormControl,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    TextField,
    Typography,
} from '@mui/material';
import { FullScreenLayout } from '../layouts';
import { RichTextEditor } from '../components/RichTextEditor';
import {
    useEntity,
    useCreateEntity,
    useUpdateEntity,
    useSimilarEntities,
    useCampaign,
    useDraft,
    useAutosave,
    useUnsavedChanges,
    useCampaignOwnership,
} from '../hooks';
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
 * Full-screen page for creating and editing campaign entities with autosave, draft recovery, validation, and unsaved-change protection.
 *
 * Renders a form for entity fields (name, type, description, tags, JSON attributes, GM notes, source confidence), shows draft-recovery and similar-entity warnings when applicable, handles create/update mutations, and protects navigation when there are unsaved changes.
 *
 * @returns The React element for the Entity Editor page used to create or edit campaign entities.
 */
export default function EntityEditor() {
    const { campaignId, entityId } = useParams<{
        campaignId: string;
        entityId?: string;
    }>();
    const navigate = useNavigate();

    const isNewEntity = !entityId || entityId === 'new';
    const draftKey = isNewEntity
        ? `entity-new-${campaignId}`
        : `entity-${entityId}`;

    // Fetch campaign for breadcrumbs
    const { data: campaign } = useCampaign(campaignId ?? '', {
        enabled: !!campaignId,
    });

    // Check if current user is the campaign owner (GM)
    const { isOwner: isGM } = useCampaignOwnership(campaignId ?? '');

    // Fetch existing entity if editing
    const {
        data: existingEntity,
        isLoading: entityLoading,
        error: entityError,
    } = useEntity(campaignId ?? '', entityId ?? '', {
        enabled: !isNewEntity && !!campaignId && !!entityId,
    });

    // Form state
    const [formData, setFormData] = useState<EntityFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<
        Partial<Record<keyof EntityFormData, string>>
    >({});

    // Draft management
    const { getDraft, deleteDraft } = useDraft();
    const [showDraftRecovery, setShowDraftRecovery] = useState(false);

    // Unsaved changes protection
    const { isDirty, setIsDirty, clearDirty, checkUnsavedChanges, ConfirmDialog } =
        useUnsavedChanges({
            message:
                'You have unsaved changes to this entity. Are you sure you want to leave?',
        });

    // Track the last hydrated entity ID to detect route changes
    const lastHydratedEntityIdRef = useRef<string | undefined>(undefined);

    // Autosave
    const { lastSaved } = useAutosave({
        data: formData,
        key: draftKey,
        enabled: isDirty,
        serverVersion: existingEntity?.version,
    });

    // Similar entities for duplicate detection (new entities only)
    const { data: similarEntities } = useSimilarEntities(
        campaignId ?? '',
        formData.name,
        { enabled: isNewEntity && formData.name.length >= 2 }
    );

    // Mutations
    const createEntity = useCreateEntity();
    const updateEntity = useUpdateEntity();

    const isSaving = createEntity.isPending || updateEntity.isPending;

    // Initialize form data from existing entity or check for draft
    useEffect(() => {
        if (!isNewEntity && existingEntity) {
            // Check if this is a different entity than what we last hydrated
            const isNewEntityRoute = existingEntity.id !== lastHydratedEntityIdRef.current;

            // Skip hydration only if same entity AND user has in-progress edits
            if (!isNewEntityRoute && isDirty) {
                return;
            }

            // If navigating to a different entity, clear dirty state
            if (isNewEntityRoute && isDirty) {
                clearDirty();
            }

            setFormData({
                name: existingEntity.name,
                entityType: existingEntity.entityType,
                description: existingEntity.description ?? '',
                tags: existingEntity.tags ?? [],
                attributes: JSON.stringify(existingEntity.attributes ?? {}, null, 2),
                gmNotes: existingEntity.gmNotes ?? '',
                sourceConfidence: existingEntity.sourceConfidence,
            });

            // Update the ref to track this entity
            lastHydratedEntityIdRef.current = existingEntity.id;
        } else if (isNewEntity) {
            // Reset the ref for new entity routes
            lastHydratedEntityIdRef.current = undefined;
            // Check for existing draft
            const draft = getDraft<EntityFormData>(draftKey);
            if (draft) {
                setShowDraftRecovery(true);
            }
        }
    }, [existingEntity, isNewEntity, getDraft, draftKey, isDirty, clearDirty]);

    /**
     * Recover draft data.
     */
    const handleRecoverDraft = useCallback(() => {
        const draft = getDraft<EntityFormData>(draftKey);
        if (draft) {
            setFormData(draft.data);
            setIsDirty(true);
        }
        setShowDraftRecovery(false);
    }, [getDraft, draftKey, setIsDirty]);

    /**
     * Discard draft data.
     */
    const handleDiscardDraft = useCallback(() => {
        deleteDraft(draftKey);
        setShowDraftRecovery(false);
    }, [deleteDraft, draftKey]);

    /**
     * Update a form field and mark as dirty.
     */
    const updateField = useCallback(
        <K extends keyof EntityFormData>(field: K, value: EntityFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            setIsDirty(true);
            // Clear error for this field
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));
        },
        [setIsDirty]
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
     * Save the entity.
     */
    const handleSave = useCallback(async (): Promise<boolean> => {
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

                // Clean up draft
                deleteDraft(draftKey);
                clearDirty();

                // Navigate to edit the newly created entity
                navigate(`/campaigns/${campaignId}/entities/${newEntity.id}/edit`, {
                    replace: true,
                });
            } else if (entityId) {
                await updateEntity.mutateAsync({
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
                });

                // Clean up draft
                deleteDraft(draftKey);
                clearDirty();
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
        createEntity,
        updateEntity,
        deleteDraft,
        draftKey,
        clearDirty,
        navigate,
    ]);

    /**
     * Save and close.
     */
    const handleSaveAndClose = useCallback(async () => {
        const saved = await handleSave();
        if (saved) {
            navigate(`/campaigns/${campaignId}/entities`);
        }
    }, [handleSave, navigate, campaignId]);

    /**
     * Handle back navigation with unsaved changes check.
     */
    const handleBack = useCallback(() => {
        const goBack = () => navigate(`/campaigns/${campaignId}/entities`);
        // If there are unsaved changes, checkUnsavedChanges will show the dialog
        // and return true. If the user confirms, it will call goBack.
        // If there are no unsaved changes, it returns false and we navigate directly.
        if (!checkUnsavedChanges(goBack)) {
            goBack();
        }
    }, [navigate, campaignId, checkUnsavedChanges]);

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Campaigns', path: '/campaigns' },
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
            onSave={handleSave}
            onSaveAndClose={handleSaveAndClose}
            onBack={handleBack}
        >
            <Box sx={{ maxWidth: 800, mx: 'auto' }}>
                {/* Draft recovery alert */}
                {showDraftRecovery && (
                    <Alert
                        severity="info"
                        sx={{ mb: 3 }}
                        action={
                            <Box sx={{ display: 'flex', gap: 1 }}>
                                <Chip
                                    label="Recover"
                                    color="primary"
                                    onClick={handleRecoverDraft}
                                    size="small"
                                />
                                <Chip
                                    label="Discard"
                                    variant="outlined"
                                    onClick={handleDiscardDraft}
                                    size="small"
                                />
                            </Box>
                        }
                    >
                        You have an unsaved draft from a previous session. Would you like
                        to recover it?
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
                        <RichTextEditor
                            label="Description"
                            value={formData.description}
                            onChange={(html) => updateField('description', html)}
                            placeholder="Describe this entity..."
                            error={!!formErrors.description}
                            helperText={formErrors.description}
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

            {/* Navigation confirmation dialog */}
            {ConfirmDialog}
        </FullScreenLayout>
    );
}