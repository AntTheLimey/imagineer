// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionEditorPage - Full-screen session editing page.
 *
 * Provides a focused editing experience for creating and editing sessions
 * with a two-column layout: form on the left, scene panel on the right.
 * Includes stage tabs, autosave, draft recovery, and unsaved changes
 * protection.
 */

import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Chip,
    FormControl,
    IconButton,
    InputLabel,
    Link as MuiLink,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Snackbar,
    Tab,
    Tabs,
    TextField,
    Typography,
} from '@mui/material';
import {
    Edit as EditIcon,
    Delete as DeleteIcon,
    Upload as UploadIcon,
} from '@mui/icons-material';
import { FullScreenLayout } from '../layouts';
import { SaveSplitButton } from '../components/SaveSplitButton';
import type { SaveMode } from '../components/SaveSplitButton';
import { MarkdownEditor } from '../components/MarkdownEditor';
import {
    SceneStrip,
    SceneViewer,
    PlayEntityDrawer,
    PlayScratchpad,
    PlayEntitySidebar,
} from '../components/Play';
import { ImportNotesDialog } from '../components/Sessions';
import {
    useCampaign,
    useChapters,
    useEntities,
    useDraft,
    useAutosave,
    useUnsavedChanges,
} from '../hooks';
import { useSession, useCreateSession, useUpdateSession } from '../hooks/useSessions';
import { useScenes, useCreateScene, useUpdateScene, useDeleteScene } from '../hooks/useScenes';
import type { SessionStage } from '../types';
import type { Entity } from '../types';

/**
 * Form data structure for session editing.
 */
interface SessionFormData {
    title: string;
    chapterId: number | null;
    sessionNumber: number | null;
    plannedDate: string;
    stage: SessionStage;
    prepNotes: string;
    actualNotes: string;
    playNotes: string;
}

/**
 * Default form values for new sessions.
 */
const DEFAULT_FORM_DATA: SessionFormData = {
    title: '',
    chapterId: null,
    sessionNumber: null,
    plannedDate: '',
    stage: 'prep',
    prepNotes: '',
    actualNotes: '',
    playNotes: '',
};

/**
 * Stage tab definitions for the session workflow.
 */
const stageTabs: { value: SessionStage; label: string }[] = [
    { value: 'prep', label: 'Prep' },
    { value: 'play', label: 'Play' },
    { value: 'wrap_up', label: 'Wrap-up' },
    { value: 'completed', label: 'Completed' },
];

/**
 * Scene type options for the scene editor form.
 */
const sceneTypeOptions = [
    'combat',
    'exploration',
    'social',
    'puzzle',
    'roleplay',
    'travel',
    'downtime',
    'other',
];

/**
 * Props for the SceneEditForm component.
 */
interface SceneEditFormProps {
    sceneForm: { title: string; sceneType: string; description: string };
    setSceneForm: React.Dispatch<React.SetStateAction<{ title: string; sceneType: string; description: string }>>;
    sceneTypeOptions: string[];
    handleCancelScene: () => void;
    handleSaveScene: () => void;
}

/**
 * Inline scene edit/add form component.
 *
 * Extracted as a standalone component so that the "add scene" and
 * "edit scene" usages receive distinct React element identities,
 * preventing accidental state sharing between the two instances.
 */
function SceneEditForm({ sceneForm, setSceneForm, sceneTypeOptions: typeOptions, handleCancelScene, handleSaveScene }: SceneEditFormProps) {
    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
            <TextField
                size="small"
                label="Scene Title"
                value={sceneForm.title}
                onChange={(e) => setSceneForm((prev) => ({ ...prev, title: e.target.value }))}
                autoFocus
            />
            <FormControl size="small">
                <InputLabel>Scene Type</InputLabel>
                <Select
                    value={sceneForm.sceneType}
                    label="Scene Type"
                    onChange={(e) => setSceneForm((prev) => ({ ...prev, sceneType: e.target.value }))}
                >
                    {typeOptions.map((type) => (
                        <MenuItem key={type} value={type}>
                            {type.charAt(0).toUpperCase() + type.slice(1)}
                        </MenuItem>
                    ))}
                </Select>
            </FormControl>
            <TextField
                size="small"
                label="Description"
                value={sceneForm.description}
                onChange={(e) => setSceneForm((prev) => ({ ...prev, description: e.target.value }))}
                multiline
                rows={2}
            />
            <Box sx={{ display: 'flex', gap: 1, justifyContent: 'flex-end' }}>
                <Button size="small" onClick={handleCancelScene}>
                    Cancel
                </Button>
                <Button
                    size="small"
                    variant="contained"
                    onClick={handleSaveScene}
                    disabled={!sceneForm.title.trim()}
                >
                    Save
                </Button>
            </Box>
        </Box>
    );
}

/**
 * Produce a human-friendly label for how long ago an ISO timestamp occurred.
 *
 * @param isoString - An ISO 8601 timestamp string
 * @returns A relative time description
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
 * Full-screen editor for creating and editing sessions with scene
 * management sidebar.
 *
 * @returns The React element for the Session Editor page
 */
export default function SessionEditorPage() {
    const params = useParams<{ campaignId: string; sessionId?: string }>();
    const campaignId = params.campaignId ? Number(params.campaignId) : undefined;
    const sessionId = params.sessionId ? Number(params.sessionId) : undefined;
    const navigate = useNavigate();

    const isNewSession = !sessionId;
    const draftKey = isNewSession
        ? `session-new-${campaignId}`
        : `session-${sessionId}`;

    // Fetch campaign for breadcrumbs
    const { data: campaign } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Fetch chapters for the chapter selector dropdown
    const { data: chapters } = useChapters(campaignId ?? 0);

    // Fetch existing session if editing
    const {
        data: session,
        isLoading: sessionLoading,
        error: sessionError,
    } = useSession(campaignId ?? 0, sessionId ?? 0, {
        enabled: !isNewSession && !!campaignId && !!sessionId,
    });

    // Fetch scenes for the scene sidebar
    const { data: scenes } = useScenes(campaignId ?? 0, sessionId ?? 0);

    // Scene mutations
    const createScene = useCreateScene(campaignId ?? 0, sessionId ?? 0);
    const updateScene = useUpdateScene(campaignId ?? 0, sessionId ?? 0);
    const deleteScene = useDeleteScene(campaignId ?? 0, sessionId ?? 0);

    /** All campaign entities for Play mode entity sidebar and drawer. */
    const { data: entities } = useEntities({ campaignId: campaignId ?? 0 });

    // Form state
    const [formData, setFormData] = useState<SessionFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<
        Partial<Record<keyof SessionFormData, string>>
    >({});

    // Scene editing state
    const [addingScene, setAddingScene] = useState(false);
    const [editingSceneId, setEditingSceneId] = useState<number | null>(null);
    const [sceneForm, setSceneForm] = useState({ title: '', sceneType: 'other', description: '' });

    /** ID of the currently active scene in Play mode, or null if none. */
    const [activeSceneId, setActiveSceneId] = useState<number | null>(null);
    /** The entity currently displayed in the PlayEntityDrawer. */
    const [selectedEntity, setSelectedEntity] = useState<Entity | null>(null);
    /** Whether the PlayEntitySidebar is in its narrow collapsed state. */
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    /** Whether the ImportNotesDialog is open. */
    const [importDialogOpen, setImportDialogOpen] = useState(false);

    // Draft management
    const { getDraft, deleteDraft } = useDraft();
    const [showDraftRecovery, setShowDraftRecovery] = useState(false);

    // Unsaved changes protection
    const { isDirty, setIsDirty, clearDirty, checkUnsavedChanges, ConfirmDialog } =
        useUnsavedChanges({
            message:
                'You have unsaved changes to this session. Are you sure you want to leave?',
        });

    // Track the last hydrated session ID to detect route changes
    const lastHydratedSessionIdRef = useRef<number | undefined>(undefined);

    // Autosave
    const { lastSaved } = useAutosave({
        data: formData,
        key: draftKey,
        enabled: isDirty,
    });

    /** The Scene object for the currently active scene, or null. */
    const activeScene = useMemo(
        () => scenes?.find((s) => s.id === activeSceneId) ?? null,
        [scenes, activeSceneId]
    );
    /** Display mode for SceneViewer: 'notes' when no scenes, 'scene' when a scene is active, 'mixed' otherwise. */
    const playMode: 'scene' | 'notes' | 'mixed' = !scenes?.length
        ? 'notes'
        : activeSceneId
            ? 'scene'
            : 'notes';

    /** Auto-select the first active scene (or first scene) when entering Play. */
    useEffect(() => {
        if (formData.stage === 'play' && scenes?.length && !activeSceneId) {
            const first = scenes.find((s) => s.status === 'active');
            setActiveSceneId(first?.id ?? scenes[0].id);
        }
    }, [formData.stage, scenes, activeSceneId]);

    // Mutations
    const createSession = useCreateSession();
    const updateSession = useUpdateSession();

    const isSaving = createSession.isPending || updateSession.isPending;

    // Analysis snackbar state
    const [analysisSnackbar, setAnalysisSnackbar] = useState<{
        open: boolean;
        jobId: number;
        count: number;
        message?: string;
    }>({ open: false, jobId: 0, count: 0 });

    // Initialize form data from existing session or check for draft
    useEffect(() => {
        if (!isNewSession && session) {
            const isNewRoute = session.id !== lastHydratedSessionIdRef.current;

            if (!isNewRoute && isDirty) {
                return;
            }

            if (isNewRoute && isDirty) {
                clearDirty();
            }

            setFormData({
                title: session.title ?? '',
                chapterId: session.chapterId ?? null,
                sessionNumber: session.sessionNumber ?? null,
                plannedDate: session.plannedDate ?? '',
                stage: session.stage,
                prepNotes: session.prepNotes ?? '',
                actualNotes: session.actualNotes ?? '',
                playNotes: session.playNotes ?? '',
            });

            lastHydratedSessionIdRef.current = session.id;
        } else if (isNewSession) {
            lastHydratedSessionIdRef.current = undefined;

            // Check for existing draft
            const draft = getDraft<SessionFormData>(draftKey);
            if (draft) {
                setShowDraftRecovery(true);
            }
        }
    }, [session, isNewSession, getDraft, draftKey, isDirty, clearDirty]);

    /**
     * Recover draft data.
     */
    const handleRecoverDraft = useCallback(() => {
        const draft = getDraft<SessionFormData>(draftKey);
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
        <K extends keyof SessionFormData>(field: K, value: SessionFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            setIsDirty(true);
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));
        },
        [setIsDirty]
    );

    /**
     * Validate the form before saving.
     */
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof SessionFormData, string>> = {};

        // Session title is optional, so no required field validation
        // but we could add validation rules here in the future

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, []);

    /**
     * Save the session with an optional save mode.
     *
     * @param mode - The save mode controlling analysis behavior
     */
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const handleSave = useCallback(async (_mode: SaveMode = 'analyze'): Promise<boolean> => {
        if (!validateForm() || !campaignId) {
            return false;
        }

        try {
            if (isNewSession) {
                const newSession = await createSession.mutateAsync({
                    campaignId,
                    input: {
                        title: formData.title || undefined,
                        chapterId: formData.chapterId || undefined,
                        sessionNumber: formData.sessionNumber || undefined,
                        plannedDate: formData.plannedDate || undefined,
                        stage: formData.stage,
                        prepNotes: formData.prepNotes || undefined,
                    },
                });

                deleteDraft(draftKey);
                clearDirty();

                navigate(`/campaigns/${campaignId}/sessions/${newSession.id}/edit`, {
                    replace: true,
                });
            } else if (sessionId) {
                await updateSession.mutateAsync({
                    campaignId,
                    sessionId,
                    input: {
                        title: formData.title || undefined,
                        chapterId: formData.chapterId || undefined,
                        sessionNumber: formData.sessionNumber || undefined,
                        plannedDate: formData.plannedDate || undefined,
                        stage: formData.stage,
                        prepNotes: formData.prepNotes || undefined,
                        actualNotes: formData.actualNotes || undefined,
                        playNotes: formData.playNotes || undefined,
                    },
                });

                deleteDraft(draftKey);
                clearDirty();
            }

            return true;
        } catch (error) {
            console.error('Failed to save session:', error);
            return false;
        }
    }, [
        validateForm,
        campaignId,
        isNewSession,
        formData,
        createSession,
        updateSession,
        sessionId,
        deleteDraft,
        draftKey,
        clearDirty,
        navigate,
    ]);

    /**
     * Cycle a scene's status: planned -> active -> completed -> skipped -> planned.
     */
    const handleSceneStatusChange = useCallback(
        async (id: number, status: string) => {
            const cycle: Record<string, string> = {
                planned: 'active',
                active: 'completed',
                completed: 'skipped',
                skipped: 'planned',
            };
            const next = cycle[status] ?? 'planned';
            try {
                await updateScene.mutateAsync({ sceneId: id, input: { status: next } });
            } catch (error) {
                console.error('Failed to update scene status:', error);
            }
        },
        [updateScene]
    );

    /**
     * Handle entity selection from wiki links, chips, or sidebar.
     */
    const handleEntitySelect = useCallback(
        (entityOrId: Entity | number) => {
            if (typeof entityOrId === 'number') {
                const entity = entities?.find((e) => e.id === entityOrId);
                setSelectedEntity(entity ?? null);
            } else {
                setSelectedEntity(entityOrId);
            }
        },
        [entities]
    );

    /**
     * Import notes into actualNotes field.
     */
    const handleImportNotes = useCallback(
        (content: string, mode: 'append' | 'replace') => {
            if (mode === 'replace') {
                updateField('actualNotes', content);
            } else {
                const separator = formData.actualNotes ? '\n\n---\n**Imported notes:**\n' : '';
                updateField('actualNotes', formData.actualNotes + separator + content);
            }
            setImportDialogOpen(false);
        },
        [formData.actualNotes, updateField]
    );

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
     * Handle adding a new scene.
     */
    const handleAddScene = useCallback(() => {
        setSceneForm({ title: '', sceneType: 'other', description: '' });
        setEditingSceneId(null);
        setAddingScene(true);
    }, []);

    /**
     * Save a new or edited scene.
     */
    const handleSaveScene = useCallback(async () => {
        if (!sceneForm.title.trim()) return;

        try {
            if (editingSceneId !== null) {
                await updateScene.mutateAsync({
                    sceneId: editingSceneId,
                    input: {
                        title: sceneForm.title,
                        sceneType: sceneForm.sceneType,
                        description: sceneForm.description || undefined,
                    },
                });
                setEditingSceneId(null);
            } else {
                await createScene.mutateAsync({
                    title: sceneForm.title,
                    sceneType: sceneForm.sceneType,
                    description: sceneForm.description || undefined,
                });
                setAddingScene(false);
            }
            setSceneForm({ title: '', sceneType: 'other', description: '' });
        } catch (error) {
            console.error('Failed to save scene:', error);
        }
    }, [sceneForm, editingSceneId, createScene, updateScene]);

    /**
     * Cancel scene editing/adding.
     */
    const handleCancelScene = useCallback(() => {
        setAddingScene(false);
        setEditingSceneId(null);
        setSceneForm({ title: '', sceneType: 'other', description: '' });
    }, []);

    /**
     * Start editing a scene inline.
     */
    const handleEditScene = useCallback((sceneId: number) => {
        const scene = scenes?.find((s) => s.id === sceneId);
        if (scene) {
            setSceneForm({
                title: scene.title,
                sceneType: scene.sceneType,
                description: scene.description ?? '',
            });
            setEditingSceneId(sceneId);
            setAddingScene(false);
        }
    }, [scenes]);

    /**
     * Delete a scene.
     */
    const handleDeleteScene = useCallback(async (sceneId: number) => {
        try {
            await deleteScene.mutateAsync(sceneId);
        } catch (error) {
            console.error('Failed to delete scene:', error);
        }
    }, [deleteScene]);

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
            { label: isNewSession ? 'New Session' : (session?.title ?? 'Edit Session') },
        ],
        [campaign, campaignId, isNewSession, session]
    );

    // Loading state
    if (!isNewSession && sessionLoading) {
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
    if (sessionError) {
        return (
            <FullScreenLayout
                title="Error"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                backPath={`/campaigns/${campaignId}/sessions`}
            >
                <Alert severity="error">
                    Failed to load session. The session may not exist or you may not
                    have permission to view it.
                </Alert>
            </FullScreenLayout>
        );
    }

    return (
        <FullScreenLayout
            title={isNewSession ? 'New Session' : 'Edit Session'}
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={isSaving}
            onBack={handleBack}
            subtitle={lastSaved ? `Auto-saved ${formatRelativeTime(lastSaved)}` : undefined}
            actions={
                formData.stage === 'play' ? (
                    <Button
                        startIcon={<UploadIcon />}
                        variant="outlined"
                        size="small"
                        onClick={() => setImportDialogOpen(true)}
                    >
                        Import Notes
                    </Button>
                ) : undefined
            }
            renderSaveButtons={() => (
                <SaveSplitButton
                    onSave={handleSave}
                    isDirty={isDirty}
                    isSaving={isSaving}
                    defaultMode={formData.stage === 'play' ? 'save' : undefined}
                />
            )}
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

            {/* Stage tabs */}
            <Tabs
                value={formData.stage}
                onChange={(_e, newValue: SessionStage) => updateField('stage', newValue)}
                sx={{ mb: 3, borderBottom: 1, borderColor: 'divider' }}
            >
                {stageTabs.map((tab) => (
                    <Tab key={tab.value} value={tab.value} label={tab.label} />
                ))}
            </Tabs>

            {/* Prep stage content */}
            {formData.stage === 'prep' ? (
                <Box sx={{ display: 'flex', gap: 3, height: '100%' }}>
                    {/* Left column - Form (~65%) */}
                    <Box sx={{ flex: '1 1 65%', minWidth: 0 }}>
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
                                autoFocus={isNewSession}
                            />

                            {/* Chapter selector */}
                            <FormControl fullWidth sx={{ mb: 3 }}>
                                <InputLabel>Chapter</InputLabel>
                                <Select
                                    value={formData.chapterId ?? ''}
                                    label="Chapter"
                                    onChange={(e) => {
                                        const val = e.target.value;
                                        updateField('chapterId', val === '' ? null : Number(val));
                                    }}
                                >
                                    <MenuItem value="">
                                        <em>None</em>
                                    </MenuItem>
                                    {chapters?.map((chapter) => (
                                        <MenuItem key={chapter.id} value={chapter.id}>
                                            {chapter.title}
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>

                            {/* Session Number */}
                            <TextField
                                fullWidth
                                label="Session Number"
                                type="number"
                                value={formData.sessionNumber ?? ''}
                                onChange={(e) => {
                                    const val = e.target.value;
                                    if (val === '') {
                                        updateField('sessionNumber', null);
                                        return;
                                    }
                                    const parsed = parseInt(val, 10);
                                    if (!Number.isNaN(parsed)) {
                                        updateField('sessionNumber', parsed);
                                    }
                                }}
                                sx={{ mb: 3 }}
                            />

                            {/* Planned Date */}
                            <TextField
                                fullWidth
                                label="Planned Date"
                                type="date"
                                value={formData.plannedDate}
                                onChange={(e) => updateField('plannedDate', e.target.value)}
                                InputLabelProps={{ shrink: true }}
                                sx={{ mb: 3 }}
                            />

                            {/* Prep Notes */}
                            <Typography variant="subtitle2" sx={{ mb: 1 }}>
                                Prep Notes
                            </Typography>
                            <MarkdownEditor
                                value={formData.prepNotes}
                                onChange={(value) => updateField('prepNotes', value)}
                                placeholder="Write your session prep notes here..."
                                minHeight={300}
                                campaignId={campaignId}
                            />
                        </Paper>
                    </Box>

                    {/* Right column - Scene sidebar (~35%) */}
                    <Box sx={{ flex: '1 1 35%', flexShrink: 0, minWidth: 280, maxWidth: 420 }}>
                        <Paper sx={{ p: 2, height: '100%' }}>
                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                                <Typography variant="h6">
                                    Scenes {scenes && scenes.length > 0 && <Chip label={scenes.length} size="small" sx={{ ml: 1 }} />}
                                </Typography>
                                <Button size="small" variant="outlined" onClick={handleAddScene} disabled={isNewSession}>
                                    Add Scene
                                </Button>
                            </Box>

                            {/* Add scene form */}
                            {addingScene && (
                                <Paper variant="outlined" sx={{ p: 1.5, mb: 1 }}>
                                    <SceneEditForm
                                        sceneForm={sceneForm}
                                        setSceneForm={setSceneForm}
                                        sceneTypeOptions={sceneTypeOptions}
                                        handleCancelScene={handleCancelScene}
                                        handleSaveScene={handleSaveScene}
                                    />
                                </Paper>
                            )}

                            {/* Scene list */}
                            {isNewSession ? (
                                <Typography variant="body2" color="text.secondary">
                                    Save the session first, then add scenes.
                                </Typography>
                            ) : scenes && scenes.length > 0 ? (
                                scenes.map((scene) => (
                                    <Paper key={scene.id} variant="outlined" sx={{ p: 1.5, mb: 1 }}>
                                        {editingSceneId === scene.id ? (
                                            <SceneEditForm
                                                sceneForm={sceneForm}
                                                setSceneForm={setSceneForm}
                                                sceneTypeOptions={sceneTypeOptions}
                                                handleCancelScene={handleCancelScene}
                                                handleSaveScene={handleSaveScene}
                                            />
                                        ) : (
                                            <Box>
                                                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                                    <Typography variant="subtitle2">{scene.title}</Typography>
                                                    <Box>
                                                        <IconButton size="small" onClick={() => handleEditScene(scene.id)}>
                                                            <EditIcon fontSize="small" />
                                                        </IconButton>
                                                        <IconButton size="small" onClick={() => handleDeleteScene(scene.id)} color="error">
                                                            <DeleteIcon fontSize="small" />
                                                        </IconButton>
                                                    </Box>
                                                </Box>
                                                <Box sx={{ display: 'flex', gap: 0.5, mt: 0.5 }}>
                                                    <Chip label={scene.sceneType} size="small" variant="outlined" />
                                                    <Chip label={scene.status} size="small" color={scene.status === 'completed' ? 'success' : 'default'} />
                                                </Box>
                                                {scene.description && (
                                                    <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                                                        {scene.description.substring(0, 100)}{scene.description.length > 100 ? '...' : ''}
                                                    </Typography>
                                                )}
                                            </Box>
                                        )}
                                    </Paper>
                                ))
                            ) : (
                                <Typography variant="body2" color="text.secondary">
                                    No scenes yet. Add scenes to plan your session structure.
                                </Typography>
                            )}
                        </Paper>
                    </Box>
                </Box>
            ) : formData.stage === 'play' ? (
                /* Play mode layout */
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        height: 'calc(100vh - 64px - 48px)',
                        overflow: 'hidden',
                        mx: -3,
                        mt: -3,
                        mb: -3,
                    }}
                >
                    {/* Scene strip */}
                    {scenes && scenes.length > 0 && (
                        <SceneStrip
                            scenes={scenes}
                            activeSceneId={activeSceneId}
                            onSceneSelect={setActiveSceneId}
                            onSceneStatusChange={handleSceneStatusChange}
                            onAddScene={() => {}}
                        />
                    )}

                    {/* Main panels */}
                    <Box sx={{ display: 'flex', flexGrow: 1, overflow: 'hidden' }}>
                        {/* Entity sidebar */}
                        <PlayEntitySidebar
                            sceneEntityIds={activeScene?.entityIds ?? []}
                            allEntities={entities ?? []}
                            onEntitySelect={handleEntitySelect}
                            collapsed={sidebarCollapsed}
                            onToggle={() => setSidebarCollapsed((prev) => !prev)}
                        />

                        {/* Scene/Notes viewer (~40%) */}
                        <Box
                            sx={{
                                flex: '4 1 0',
                                overflow: 'auto',
                                borderRight: 1,
                                borderColor: 'divider',
                            }}
                        >
                            <SceneViewer
                                scene={activeScene}
                                prepNotes={formData.prepNotes}
                                mode={playMode}
                                onEntityClick={handleEntitySelect}
                                entities={entities ?? []}
                            />
                        </Box>

                        {/* Entity drawer (~30%) */}
                        <Box
                            sx={{
                                flex: '3 1 0',
                                overflow: 'auto',
                                borderRight: 1,
                                borderColor: 'divider',
                            }}
                        >
                            <PlayEntityDrawer
                                entity={selectedEntity}
                                campaignId={campaignId ?? 0}
                            />
                        </Box>

                        {/* Scratchpad (~30%) */}
                        <Box
                            sx={{
                                flex: '3 1 0',
                                overflow: 'auto',
                            }}
                        >
                            <PlayScratchpad
                                value={formData.playNotes}
                                onChange={(value) => updateField('playNotes', value)}
                                lastSaved={lastSaved}
                            />
                        </Box>
                    </Box>
                </Box>
            ) : (
                /* Wrap-up / Completed stage placeholder */
                <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', py: 8 }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                        {stageTabs.find((t) => t.value === formData.stage)?.label} Stage
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                        Coming in Phase 3
                    </Typography>
                </Box>
            )}

            {/* Import Notes dialog */}
            {formData.stage === 'play' && (
                <ImportNotesDialog
                    open={importDialogOpen}
                    onClose={() => setImportDialogOpen(false)}
                    onImport={handleImportNotes}
                />
            )}

            {/* Analysis results snackbar */}
            <Snackbar
                open={analysisSnackbar.open}
                autoHideDuration={analysisSnackbar.count === 0 ? 4000 : 10000}
                onClose={() => setAnalysisSnackbar((prev) => ({ ...prev, open: false }))}
            >
                {analysisSnackbar.count === 0 ? (
                    <Alert
                        severity="success"
                        onClose={() => setAnalysisSnackbar((prev) => ({ ...prev, open: false }))}
                    >
                        {analysisSnackbar.message ?? 'Analysis complete: no issues found.'}
                    </Alert>
                ) : (
                    <Alert
                        severity="warning"
                        onClose={() => setAnalysisSnackbar((prev) => ({ ...prev, open: false }))}
                        action={
                            <Button
                                color="inherit"
                                size="small"
                                onClick={() => {
                                    setAnalysisSnackbar((prev) => ({ ...prev, open: false }));
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

            {/* Navigation confirmation dialog */}
            {ConfirmDialog}
        </FullScreenLayout>
    );
}
