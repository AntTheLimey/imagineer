// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterEditorPage - Full-screen chapter editing page.
 *
 * Provides a focused editing experience for creating and editing chapters
 * with a two-column layout: form on the left, entity panel on the right.
 * Includes autosave, draft recovery, and unsaved changes protection.
 */

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    AlertTitle,
    Box,
    Link as MuiLink,
    Paper,
    Skeleton,
    TextField,
    Typography,
} from '@mui/material';
import { Link } from 'react-router-dom';
import { FullScreenLayout } from '../layouts';
import { MarkdownEditor } from '../components/MarkdownEditor';
import {
    useChapter,
    useChapters,
    useCreateChapter,
    useUpdateChapter,
    useCampaign,
    useDraft,
    useAutosave,
    useUnsavedChanges,
} from '../hooks';
import { useChapterEntities, useCreateChapterEntity, useDeleteChapterEntity } from '../hooks/useChapterEntities';
import { useUserSettings } from '../hooks/useUserSettings';
import { useEntities } from '../hooks/useEntities';
import type { Entity } from '../types';

/**
 * Form data structure for chapter editing.
 */
interface ChapterFormData {
    title: string;
    overview: string;
    sortOrder: number;
}

/**
 * Default form values for new chapters.
 */
const DEFAULT_FORM_DATA: ChapterFormData = {
    title: '',
    overview: '',
    sortOrder: 0,
};

/**
 * Full-screen editor for creating and editing chapters with entity association panel.
 *
 * @returns The React element for the Chapter Editor page
 */
export default function ChapterEditorPage() {
    const { campaignId, chapterId } = useParams<{
        campaignId: string;
        chapterId?: string;
    }>();
    const navigate = useNavigate();

    const isNewChapter = !chapterId || chapterId === 'new';
    const draftKey = isNewChapter
        ? `chapter-new-${campaignId}`
        : `chapter-${chapterId}`;

    // Fetch campaign for breadcrumbs
    const { data: campaign } = useCampaign(campaignId ?? '', {
        enabled: !!campaignId,
    });

    // Fetch chapters to determine next sort order
    const { data: chapters } = useChapters(campaignId ?? '');

    // Fetch existing chapter if editing
    const {
        data: existingChapter,
        isLoading: chapterLoading,
        error: chapterError,
    } = useChapter(campaignId ?? '', chapterId ?? '', {
        enabled: !isNewChapter && !!campaignId && !!chapterId,
    });

    // Fetch chapter entities for the entity panel
    const { data: chapterEntities } = useChapterEntities(campaignId ?? '', chapterId ?? '');

    // Fetch all campaign entities for the entity selector
    const { data: allEntities } = useEntities(campaignId ?? '');

    // Check if embedding service is configured
    const { data: userSettings } = useUserSettings();
    const isEmbeddingConfigured = !!(
        userSettings?.embeddingService && userSettings?.embeddingApiKey
    );

    // Form state
    const [formData, setFormData] = useState<ChapterFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<
        Partial<Record<keyof ChapterFormData, string>>
    >({});

    // Pending entity links for new chapters
    const [pendingEntityLinks, setPendingEntityLinks] = useState<string[]>([]);

    // Draft management
    const { getDraft, deleteDraft } = useDraft();
    const [showDraftRecovery, setShowDraftRecovery] = useState(false);

    // Unsaved changes protection
    const { isDirty, setIsDirty, clearDirty, checkUnsavedChanges, ConfirmDialog } =
        useUnsavedChanges({
            message:
                'You have unsaved changes to this chapter. Are you sure you want to leave?',
        });

    // Track the last hydrated chapter ID to detect route changes
    const lastHydratedChapterIdRef = useRef<string | undefined>(undefined);

    // Autosave
    const { lastSaved } = useAutosave({
        data: formData,
        key: draftKey,
        enabled: isDirty,
    });

    // Mutations
    const createChapter = useCreateChapter();
    const updateChapter = useUpdateChapter();
    const createChapterEntity = useCreateChapterEntity();
    const deleteChapterEntity = useDeleteChapterEntity();

    const isSaving = createChapter.isPending || updateChapter.isPending;

    // Calculate next sort order for new chapters
    const nextSortOrder = useMemo(() => {
        if (!chapters || chapters.length === 0) return 0;
        return Math.max(...chapters.map(c => c.sortOrder)) + 1;
    }, [chapters]);

    // Initialize form data from existing chapter or check for draft
    useEffect(() => {
        if (!isNewChapter && existingChapter) {
            const isNewChapterRoute = existingChapter.id !== lastHydratedChapterIdRef.current;

            if (!isNewChapterRoute && isDirty) {
                return;
            }

            if (isNewChapterRoute && isDirty) {
                clearDirty();
            }

            setFormData({
                title: existingChapter.title,
                overview: existingChapter.overview ?? '',
                sortOrder: existingChapter.sortOrder,
            });

            lastHydratedChapterIdRef.current = existingChapter.id;
        } else if (isNewChapter) {
            lastHydratedChapterIdRef.current = undefined;

            // Set default sort order for new chapters
            setFormData(prev => ({
                ...prev,
                sortOrder: nextSortOrder,
            }));

            // Check for existing draft
            const draft = getDraft<ChapterFormData>(draftKey);
            if (draft) {
                setShowDraftRecovery(true);
            }
        }
    }, [existingChapter, isNewChapter, getDraft, draftKey, isDirty, clearDirty, nextSortOrder]);

    /**
     * Recover draft data.
     */
    const handleRecoverDraft = useCallback(() => {
        const draft = getDraft<ChapterFormData>(draftKey);
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
        <K extends keyof ChapterFormData>(field: K, value: ChapterFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            setIsDirty(true);
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));
        },
        [setIsDirty]
    );

    /**
     * Validate the form.
     */
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof ChapterFormData, string>> = {};

        if (!formData.title.trim()) {
            errors.title = 'Title is required';
        }

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, [formData]);

    /**
     * Save the chapter.
     */
    const handleSave = useCallback(async (): Promise<boolean> => {
        if (!validateForm() || !campaignId) {
            return false;
        }

        try {
            if (isNewChapter) {
                const newChapter = await createChapter.mutateAsync({
                    campaignId,
                    input: {
                        title: formData.title,
                        overview: formData.overview || undefined,
                        sortOrder: formData.sortOrder,
                    },
                });

                // Create pending entity links using Promise.allSettled
                if (pendingEntityLinks.length > 0) {
                    const linkResults = await Promise.allSettled(
                        pendingEntityLinks.map((entityId) =>
                            createChapterEntity.mutateAsync({
                                campaignId,
                                chapterId: newChapter.id,
                                input: { entityId },
                            })
                        )
                    );
                    const failedLinks = linkResults.filter((r) => r.status === 'rejected');
                    if (failedLinks.length > 0) {
                        console.warn(`Failed to create ${failedLinks.length} entity link(s)`);
                    }
                }

                deleteDraft(draftKey);
                clearDirty();

                // Navigate to edit the newly created chapter
                navigate(`/campaigns/${campaignId}/chapters/${newChapter.id}/edit`, {
                    replace: true,
                });
            } else if (chapterId) {
                await updateChapter.mutateAsync({
                    campaignId,
                    chapterId,
                    input: {
                        title: formData.title,
                        overview: formData.overview || undefined,
                        sortOrder: formData.sortOrder,
                    },
                });

                deleteDraft(draftKey);
                clearDirty();
            }

            return true;
        } catch (error) {
            console.error('Failed to save chapter:', error);
            return false;
        }
    }, [
        validateForm,
        campaignId,
        isNewChapter,
        formData,
        createChapter,
        updateChapter,
        chapterId,
        pendingEntityLinks,
        createChapterEntity,
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
            navigate(`/campaigns/${campaignId}/sessions`);
        }
    }, [handleSave, navigate, campaignId]);

    /**
     * Handle back navigation with unsaved changes check.
     */
    const handleBack = useCallback(() => {
        const goBack = () => navigate(`/campaigns/${campaignId}/sessions`);
        if (!checkUnsavedChanges(goBack)) {
            goBack();
        }
    }, [navigate, campaignId, checkUnsavedChanges]);

    /**
     * Link an entity to this chapter.
     */
    const handleLinkEntity = useCallback(
        async (entityId: string) => {
            try {
                if (isNewChapter) {
                    // Add to pending links for new chapters
                    setPendingEntityLinks((prev) => [...prev, entityId]);
                } else if (campaignId && chapterId) {
                    // Create the link immediately for existing chapters
                    await createChapterEntity.mutateAsync({
                        campaignId,
                        chapterId,
                        input: { entityId },
                    });
                }
            } catch (error) {
                console.error('Failed to link entity:', error);
            }
        },
        [isNewChapter, campaignId, chapterId, createChapterEntity]
    );

    /**
     * Unlink an entity from this chapter.
     */
    const handleUnlinkEntity = useCallback(
        async (linkId: string, entityId: string) => {
            try {
                if (isNewChapter) {
                    // Remove from pending links for new chapters
                    setPendingEntityLinks((prev) => prev.filter((id) => id !== entityId));
                } else if (campaignId && chapterId) {
                    // Delete the link immediately for existing chapters
                    await deleteChapterEntity.mutateAsync({
                        campaignId,
                        chapterId,
                        linkId,
                    });
                }
            } catch (error) {
                console.error('Failed to unlink entity:', error);
            }
        },
        [isNewChapter, campaignId, chapterId, deleteChapterEntity]
    );

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Campaigns', path: '/campaigns' },
            {
                label: campaign?.name ?? 'Campaign',
                path: `/campaigns/${campaignId}`,
            },
            {
                label: 'Sessions',
                path: `/campaigns/${campaignId}/sessions`,
            },
            { label: isNewChapter ? 'New Chapter' : (existingChapter?.title ?? 'Edit Chapter') },
        ],
        [campaign, campaignId, isNewChapter, existingChapter]
    );

    // Get linked entities for display
    const linkedEntities = useMemo(() => {
        if (isNewChapter) {
            // For new chapters, show entities from pending links
            return pendingEntityLinks
                .map((id) => allEntities?.find((e) => e.id === id))
                .filter((e): e is Entity => !!e);
        }
        // For existing chapters, show entities from chapter_entities
        return chapterEntities?.map((ce) => ce.entity).filter((e): e is Entity => !!e) ?? [];
    }, [isNewChapter, pendingEntityLinks, allEntities, chapterEntities]);

    // Get available entities (not yet linked)
    const availableEntities = useMemo(() => {
        const linkedIds = new Set(
            isNewChapter
                ? pendingEntityLinks
                : chapterEntities?.map((ce) => ce.entityId) ?? []
        );
        return allEntities?.filter((e) => !linkedIds.has(e.id)) ?? [];
    }, [isNewChapter, pendingEntityLinks, chapterEntities, allEntities]);

    // Loading state
    if (!isNewChapter && chapterLoading) {
        return (
            <FullScreenLayout
                title="Loading..."
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                backPath={`/campaigns/${campaignId}/sessions`}
            >
                <Box sx={{ display: 'flex', gap: 3 }}>
                    <Box sx={{ flex: 1 }}>
                        <Skeleton variant="rectangular" height={56} sx={{ mb: 2 }} />
                        <Skeleton variant="rectangular" height={300} sx={{ mb: 2 }} />
                        <Skeleton variant="rectangular" height={56} />
                    </Box>
                    <Box sx={{ width: 320, flexShrink: 0 }}>
                        <Skeleton variant="rectangular" height={400} />
                    </Box>
                </Box>
            </FullScreenLayout>
        );
    }

    // Error state
    if (chapterError) {
        return (
            <FullScreenLayout
                title="Error"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                backPath={`/campaigns/${campaignId}/sessions`}
            >
                <Alert severity="error">
                    Failed to load chapter. The chapter may not exist or you may not
                    have permission to view it.
                </Alert>
            </FullScreenLayout>
        );
    }

    return (
        <FullScreenLayout
            title={isNewChapter ? 'New Chapter' : 'Edit Chapter'}
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={isSaving}
            onSave={handleSave}
            onSaveAndClose={handleSaveAndClose}
            onBack={handleBack}
            lastSaved={lastSaved}
        >
            {/* Draft recovery alert */}
            {showDraftRecovery && (
                <Alert
                    severity="info"
                    sx={{ mb: 3 }}
                    action={
                        <Box sx={{ display: 'flex', gap: 1 }}>
                            <MuiLink
                                component="button"
                                variant="body2"
                                onClick={handleRecoverDraft}
                            >
                                Recover
                            </MuiLink>
                            <MuiLink
                                component="button"
                                variant="body2"
                                onClick={handleDiscardDraft}
                            >
                                Discard
                            </MuiLink>
                        </Box>
                    }
                >
                    You have an unsaved draft from a previous session. Would you like
                    to recover it?
                </Alert>
            )}

            {/* Two-column layout */}
            <Box sx={{ display: 'flex', gap: 3, height: '100%' }}>
                {/* Left column - Form */}
                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Paper sx={{ p: 3 }}>
                        {/* Title */}
                        <TextField
                            fullWidth
                            label="Title"
                            value={formData.title}
                            onChange={(e) => updateField('title', e.target.value)}
                            error={!!formErrors.title}
                            helperText={formErrors.title}
                            sx={{ mb: 3 }}
                            autoFocus={isNewChapter}
                        />

                        {/* Overview */}
                        <Typography variant="subtitle2" sx={{ mb: 1 }}>
                            Overview
                        </Typography>
                        <MarkdownEditor
                            value={formData.overview}
                            onChange={(value) => updateField('overview', value)}
                            placeholder="Describe this chapter's story arc, themes, and key events..."
                            minHeight={300}
                        />

                        {/* Sort Order */}
                        <TextField
                            fullWidth
                            label="Sort Order"
                            type="number"
                            value={formData.sortOrder}
                            onChange={(e) =>
                                updateField('sortOrder', parseInt(e.target.value, 10) || 0)
                            }
                            sx={{ mt: 3 }}
                            helperText="Controls the display order of chapters (lower numbers appear first)"
                        />
                    </Paper>
                </Box>

                {/* Right column - Entity Panel */}
                <Box sx={{ width: 320, flexShrink: 0 }}>
                    <Paper sx={{ p: 2, height: '100%' }}>
                        <Typography variant="h6" gutterBottom>
                            Linked Entities
                        </Typography>

                        {/* Embedding config warning */}
                        {!isEmbeddingConfigured && (
                            <Alert severity="warning" sx={{ mb: 2 }}>
                                <AlertTitle>AI Features Unavailable</AlertTitle>
                                Configure your embedding API key in{' '}
                                <MuiLink component={Link} to="/settings">
                                    Account Settings
                                </MuiLink>{' '}
                                to enable AI-powered entity detection.
                            </Alert>
                        )}

                        {/* Linked entities list */}
                        {linkedEntities.length > 0 ? (
                            <Box sx={{ mb: 2 }}>
                                {linkedEntities.map((entity) => {
                                    const link = chapterEntities?.find(
                                        (ce) => ce.entityId === entity.id
                                    );
                                    return (
                                        <Box
                                            key={entity.id}
                                            sx={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'space-between',
                                                py: 1,
                                                borderBottom: 1,
                                                borderColor: 'divider',
                                            }}
                                        >
                                            <Box>
                                                <Typography variant="body2">
                                                    {entity.name}
                                                </Typography>
                                                <Typography
                                                    variant="caption"
                                                    color="text.secondary"
                                                >
                                                    {entity.entityType}
                                                </Typography>
                                            </Box>
                                            <MuiLink
                                                component="button"
                                                variant="caption"
                                                onClick={() =>
                                                    handleUnlinkEntity(
                                                        link?.id ?? '',
                                                        entity.id
                                                    )
                                                }
                                                sx={{ color: 'error.main' }}
                                            >
                                                Remove
                                            </MuiLink>
                                        </Box>
                                    );
                                })}
                            </Box>
                        ) : (
                            <Typography
                                variant="body2"
                                color="text.secondary"
                                sx={{ mb: 2 }}
                            >
                                No entities linked yet.
                            </Typography>
                        )}

                        {/* Quick entity selector */}
                        {availableEntities.length > 0 && (
                            <Box>
                                <Typography variant="subtitle2" sx={{ mb: 1 }}>
                                    Link an Entity
                                </Typography>
                                {availableEntities.slice(0, 5).map((entity) => (
                                    <Box
                                        key={entity.id}
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'space-between',
                                            py: 0.5,
                                        }}
                                    >
                                        <Typography variant="body2" noWrap sx={{ flex: 1 }}>
                                            {entity.name}
                                        </Typography>
                                        <MuiLink
                                            component="button"
                                            variant="caption"
                                            onClick={() => handleLinkEntity(entity.id)}
                                        >
                                            Link
                                        </MuiLink>
                                    </Box>
                                ))}
                                {availableEntities.length > 5 && (
                                    <Typography
                                        variant="caption"
                                        color="text.secondary"
                                    >
                                        +{availableEntities.length - 5} more entities
                                        available
                                    </Typography>
                                )}
                            </Box>
                        )}
                    </Paper>
                </Box>
            </Box>

            {/* Navigation confirmation dialog */}
            {ConfirmDialog}
        </FullScreenLayout>
    );
}
