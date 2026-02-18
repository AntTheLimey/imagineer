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
 * Includes server-backed draft persistence, autosave, and draft recovery.
 */

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Paper,
    Skeleton,
    Snackbar,
    TextField,
    Typography,
} from '@mui/material';
import { FullScreenLayout } from '../layouts';
import { SaveSplitButton } from '../components/SaveSplitButton';
import type { SaveMode } from '../components/SaveSplitButton';
import { MarkdownEditor } from '../components/MarkdownEditor';
import {
    useChapter,
    useChapters,
    useCreateChapter,
    useUpdateChapter,
    useCampaign,
    useServerDraft,
} from '../hooks';
import { useChapterEntities } from '../hooks/useChapterEntities';

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
 * Full-screen editor for creating and editing chapters with entity association panel.
 *
 * @returns The React element for the Chapter Editor page
 */
export default function ChapterEditorPage() {
    const params = useParams<{
        campaignId: string;
        chapterId?: string;
    }>();
    const campaignId = params.campaignId ? Number(params.campaignId) : undefined;
    const chapterId = params.chapterId && params.chapterId !== 'new' ? Number(params.chapterId) : undefined;
    const navigate = useNavigate();

    const isNewChapter = !chapterId;

    // Fetch campaign for breadcrumbs
    const { data: campaign } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Fetch chapters to determine next sort order
    const { data: chapters } = useChapters(campaignId ?? 0);

    // Fetch existing chapter if editing
    const {
        data: existingChapter,
        isLoading: chapterLoading,
        error: chapterError,
    } = useChapter(campaignId ?? 0, chapterId ?? 0, {
        enabled: !isNewChapter && !!campaignId && !!chapterId,
    });

    // Fetch chapter entities for the entity panel
    const { data: chapterEntities } = useChapterEntities(campaignId ?? 0, chapterId ?? 0);

    // Form state
    const [formData, setFormData] = useState<ChapterFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<
        Partial<Record<keyof ChapterFormData, string>>
    >({});

    // Calculate next sort order for new chapters
    const nextSortOrder = useMemo(() => {
        if (!chapters || chapters.length === 0) return 0;
        return Math.max(...chapters.map(c => c.sortOrder)) + 1;
    }, [chapters]);

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
    } = useServerDraft<ChapterFormData>({
        campaignId: campaignId ?? 0,
        sourceTable: 'chapters',
        sourceId: chapterId ?? 0,
        isNew: isNewChapter,
        committedData: existingChapter ? {
            title: existingChapter.title,
            overview: existingChapter.overview ?? '',
            sortOrder: existingChapter.sortOrder,
        } : {
            ...DEFAULT_FORM_DATA,
            sortOrder: nextSortOrder,
        },
        currentData: formData,
        enabled: !!campaignId,
    });

    const [showDraftRecovery, setShowDraftRecovery] = useState(false);

    // Track the last hydrated chapter ID to detect route changes
    const lastHydratedChapterIdRef = useRef<number | undefined>(undefined);

    // Mutations
    const createChapter = useCreateChapter();
    const updateChapter = useUpdateChapter();

    const isSaving = createChapter.isPending || updateChapter.isPending;

    // Analysis snackbar state
    const [analysisSnackbar, setAnalysisSnackbar] = useState<{
        open: boolean;
        jobId: number;
        count: number;
        message?: string;
    }>({ open: false, jobId: 0, count: 0 });

    // Initialize form data from existing chapter
    useEffect(() => {
        if (!isNewChapter && existingChapter) {
            const isNewChapterRoute = existingChapter.id !== lastHydratedChapterIdRef.current;
            if (!isNewChapterRoute) return;

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
        }
    }, [existingChapter, isNewChapter, nextSortOrder]);

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
     * The user continues editing with the auto-applied draft data.
     */
    const handleDismissDraftBanner = useCallback(() => {
        setShowDraftRecovery(false);
    }, []);

    /**
     * Ditch the draft: delete it from the server and restore
     * the committed data (or defaults for a new chapter).
     */
    const handleDitchDraft = useCallback(() => {
        deleteDraftFromServer();
        // Restore committed data
        if (existingChapter) {
            setFormData({
                title: existingChapter.title,
                overview: existingChapter.overview ?? '',
                sortOrder: existingChapter.sortOrder,
            });
        } else {
            setFormData(DEFAULT_FORM_DATA);
        }
        setShowDraftRecovery(false);
    }, [deleteDraftFromServer, existingChapter]);

    /**
     * Update a form field. Dirty state is tracked automatically by
     * useServerDraft via deep comparison of currentData vs committedData.
     */
    const updateField = useCallback(
        <K extends keyof ChapterFormData>(field: K, value: ChapterFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));
        },
        []
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
     * Save the chapter with an optional save mode that controls whether
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
            if (isNewChapter) {
                const newChapter = await createChapter.mutateAsync({
                    campaignId,
                    input: {
                        title: formData.title,
                        overview: formData.overview || undefined,
                        sortOrder: formData.sortOrder,
                    },
                });

                // If analysis/enrich requested, trigger it on the new chapter
                if (mode !== 'save') {
                    const result = await updateChapter.mutateAsync({
                        campaignId,
                        chapterId: newChapter.id,
                        input: {
                            title: formData.title,
                            overview: formData.overview || undefined,
                            sortOrder: formData.sortOrder,
                        },
                        options: {
                            analyze: true,
                            enrich: mode === 'enrich',
                        },
                    });

                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    const analysisResult = (result as any)?._analysis;
                    if (analysisResult) {
                        await deleteDraftFromServer();
                        if (mode === 'enrich' && analysisResult.jobId) {
                            navigate(`/campaigns/${campaignId}/analysis/${analysisResult.jobId}`, {
                                replace: true,
                            });
                            return true;
                        }
                        // Navigate to edit, show snackbar from there
                        navigate(`/campaigns/${campaignId}/chapters/${newChapter.id}/edit`, {
                            replace: true,
                        });
                        return true;
                    }
                }

                await deleteDraftFromServer();

                // Navigate to edit the newly created chapter
                navigate(`/campaigns/${campaignId}/chapters/${newChapter.id}/edit`, {
                    replace: true,
                });
            } else if (chapterId) {
                const result = await updateChapter.mutateAsync({
                    campaignId,
                    chapterId,
                    input: {
                        title: formData.title,
                        overview: formData.overview || undefined,
                        sortOrder: formData.sortOrder,
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

                await deleteDraftFromServer();
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
        deleteDraftFromServer,
        navigate,
    ]);

    /**
     * Handle back navigation. If there are unsaved changes, auto-save
     * a draft to the server before navigating away.
     */
    const handleBack = useCallback(async () => {
        if (isDirty) {
            await saveDraftToServer();
        }
        navigate(`/campaigns/${campaignId}/sessions`);
    }, [isDirty, saveDraftToServer, navigate, campaignId]);

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
                    {!isNewChapter && (
                        <Box sx={{ width: 320, flexShrink: 0 }}>
                            <Skeleton variant="rectangular" height={400} />
                        </Box>
                    )}
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
            onBack={handleBack}
            subtitle={lastSaved ? `Auto-saved ${formatRelativeTime(lastSaved)}` : undefined}
            renderSaveButtons={() => (
                <SaveSplitButton
                    onSave={handleSave}
                    isDirty={isDirty}
                    isSaving={isSaving}
                />
            )}
        >
            {/* Draft banner — auto-applied, dismissable */}
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
                    {draftUpdatedAt ? ` — last saved ${formatRelativeTime(draftUpdatedAt)}` : ''}
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
                            campaignId={campaignId}
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

                {/* Right column - Linked Entities (read-only, existing chapters only) */}
                {!isNewChapter && (
                    <Box sx={{ width: 320, flexShrink: 0 }}>
                        <Paper sx={{ p: 2, height: '100%' }}>
                            <Typography variant="h6" gutterBottom>
                                Linked Entities
                            </Typography>
                            {chapterEntities && chapterEntities.length > 0 ? (
                                <Box sx={{ maxHeight: 400, overflow: 'auto' }}>
                                    {chapterEntities.map((ce) => (
                                        <Box
                                            key={ce.id}
                                            sx={{
                                                py: 1,
                                                borderBottom: 1,
                                                borderColor: 'divider',
                                            }}
                                        >
                                            <Typography variant="body2">
                                                {ce.entity?.name ?? `Entity #${ce.entityId}`}
                                            </Typography>
                                            <Typography
                                                variant="caption"
                                                color="text.secondary"
                                            >
                                                {ce.entity?.entityType}
                                            </Typography>
                                        </Box>
                                    ))}
                                </Box>
                            ) : (
                                <Typography
                                    variant="body2"
                                    color="text.secondary"
                                >
                                    No entities linked yet. Save with analysis to detect entities.
                                </Typography>
                            )}
                        </Paper>
                    </Box>
                )}
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
