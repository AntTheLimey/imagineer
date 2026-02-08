// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * CampaignDashboard - Main hub page for campaign management.
 *
 * Provides a left navigation panel with access to campaign settings,
 * entities, sessions, player characters, and import functionality.
 * Uses the FullScreenLayout for consistent header and save actions.
 */

import { useState, useCallback, useMemo, useEffect } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import {
    Alert,
    Box,
    Skeleton,
} from '@mui/material';
import { FullScreenLayout } from '../layouts';
import CampaignNav, {
    NAV_WIDTH,
    type CampaignNavItem,
} from '../components/CampaignNav';
import CampaignSettings, {
    type CampaignSettingsData,
    campaignToFormData,
} from '../components/CampaignSettings';
import { ChapterList, ChapterEditor } from '../components/Chapters';
import { SessionList, SessionEditor } from '../components/Sessions';
import {
    useCampaign,
    useUpdateCampaign,
    useUnsavedChanges,
} from '../hooks';
import { useCampaignContext } from '../contexts/CampaignContext';
import type { Chapter, Session } from '../types';

/**
 * Render a two-column UI that lets the user browse and manage chapters and sessions for a campaign.
 *
 * The left column shows a chapter list with create/edit actions; the right column shows sessions for the selected chapter
 * or an informational prompt when no chapter is selected.
 *
 * @param campaignId - The id of the campaign whose chapters and sessions are shown
 * @param selectedChapterId - The id of the currently selected chapter, if any
 * @param onSelectChapter - Called with a chapter id when the user selects a chapter
 * @param selectedSessionId - The id of the currently selected session, if any
 * @param onSelectSession - Called with a session id when the user selects a session
 * @param onCreateChapter - Called to initiate creating a new chapter
 * @param onEditChapter - Called with a chapter when the user chooses to edit it
 * @param onCreateSession - Called to initiate creating a new session
 * @param onEditSession - Called with a session when the user chooses to edit it
 * @returns The rendered sessions management view as a React element
 */
function SessionsView({
    campaignId,
    selectedChapterId,
    onSelectChapter,
    selectedSessionId,
    onSelectSession,
    onCreateChapter,
    onEditChapter,
    onCreateSession,
    onEditSession,
}: {
    campaignId: number;
    selectedChapterId?: number;
    onSelectChapter: (id: number) => void;
    selectedSessionId?: number;
    onSelectSession: (id: number) => void;
    onCreateChapter: () => void;
    onEditChapter: (chapter: Chapter) => void;
    onCreateSession: () => void;
    onEditSession: (session: Session) => void;
}) {
    return (
        <Box sx={{ display: 'flex', gap: 3, height: '100%' }}>
            {/* Chapters Panel */}
            <Box sx={{ width: 320, flexShrink: 0 }}>
                <ChapterList
                    campaignId={campaignId}
                    selectedChapterId={selectedChapterId}
                    onSelectChapter={onSelectChapter}
                    onCreateChapter={onCreateChapter}
                    onEditChapter={onEditChapter}
                />
            </Box>

            {/* Sessions Panel */}
            <Box sx={{ flexGrow: 1, minWidth: 0 }}>
                {selectedChapterId ? (
                    <SessionList
                        campaignId={campaignId}
                        chapterId={selectedChapterId}
                        selectedSessionId={selectedSessionId}
                        onSelectSession={onSelectSession}
                        onCreateSession={onCreateSession}
                        onEditSession={onEditSession}
                    />
                ) : (
                    <Alert severity="info">
                        Select a chapter to view its sessions, or create a new chapter to get started.
                    </Alert>
                )}
            </Box>
        </Box>
    );
}

/**
 * Placeholder component for Player Characters view.
 */
function PlayerCharactersView() {
    return (
        <Box sx={{ maxWidth: 800, mx: 'auto' }}>
            <Alert severity="info">
                Player Characters management is coming soon. This view will allow
                you to manage PCs associated with your campaign.
            </Alert>
        </Box>
    );
}

/**
 * Render the campaign dashboard for managing campaign settings, sessions, and player characters.
 *
 * Provides a left navigation for switching views, editing campaign settings with dirty-state
 * tracking and save actions, managing chapters and sessions (with editors), and a placeholder
 * for player characters. Protects against navigation when there are unsaved changes.
 *
 * @returns The Campaign Dashboard page component
 */
export default function CampaignDashboard() {
    const { id } = useParams<{ id: string }>();
    const campaignId = id ? Number(id) : undefined;
    const navigate = useNavigate();
    const location = useLocation();

    // Determine initial nav item based on URL path
    const getInitialNavItem = (): CampaignNavItem => {
        if (location.pathname.endsWith('/sessions')) return 'sessions';
        if (location.pathname.endsWith('/player-characters')) return 'player-characters';
        return 'overview';
    };

    // Active navigation item state
    const [activeNavItem, setActiveNavItem] = useState<CampaignNavItem>(getInitialNavItem);

    // Form data state (managed here for proper dirty tracking)
    const [formData, setFormData] = useState<CampaignSettingsData | null>(null);
    const [isFormDirty, setIsFormDirty] = useState(false);

    // Sessions view state
    const [selectedChapterId, setSelectedChapterId] = useState<number | undefined>();
    const [selectedSessionId, setSelectedSessionId] = useState<number | undefined>();
    const [chapterEditorOpen, setChapterEditorOpen] = useState(false);
    const [editingChapterId, setEditingChapterId] = useState<number | undefined>();
    const [sessionEditorOpen, setSessionEditorOpen] = useState(false);
    const [editingSessionId, setEditingSessionId] = useState<number | undefined>();

    // Campaign context for keeping app-wide state updated
    const { setCurrentCampaignId } = useCampaignContext();

    // Fetch campaign data
    const {
        data: campaign,
        isLoading: campaignLoading,
        error: campaignError,
    } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Update campaign mutation
    const updateCampaign = useUpdateCampaign();

    // Unsaved changes protection
    const {
        isDirty,
        setIsDirty,
        clearDirty,
        checkUnsavedChanges,
        ConfirmDialog,
    } = useUnsavedChanges({
        message:
            'You have unsaved changes to this campaign. Are you sure you want to leave?',
    });

    // Initialize form data when campaign loads
    useEffect(() => {
        if (campaign && !formData) {
            setFormData(campaignToFormData(campaign));
        }
    }, [campaign, formData]);

    // Sync dirty state
    useEffect(() => {
        setIsDirty(isFormDirty);
    }, [isFormDirty, setIsDirty]);

    /**
     * Handle form data changes from CampaignSettings.
     */
    const handleFormChange = useCallback(
        (data: CampaignSettingsData, dirty: boolean) => {
            setFormData(data);
            setIsFormDirty(dirty);
        },
        []
    );

    /**
     * Handle navigation item selection.
     */
    const handleNavItemSelect = useCallback((item: CampaignNavItem) => {
        setActiveNavItem(item);
    }, []);

    /**
     * Save campaign settings.
     */
    const handleSave = useCallback(async (): Promise<boolean> => {
        if (!campaignId || !formData) {
            return false;
        }

        if (!formData.name.trim()) {
            // TODO: Show validation error to user
            console.warn('Campaign name is required');
            return false;
        }

        try {
            const updatedCampaign = await updateCampaign.mutateAsync({
                id: campaignId!,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    settings: {
                        genre: formData.genre || undefined,
                        imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                    },
                },
            });

            // Update the campaign context with new data
            setCurrentCampaignId(updatedCampaign.id);

            clearDirty();
            setIsFormDirty(false);
            return true;
        } catch (error) {
            console.error('Failed to save campaign:', error);
            return false;
        }
    }, [campaignId, formData, updateCampaign, setCurrentCampaignId, clearDirty]);

    /**
     * Save and close - return to campaigns list.
     */
    const handleSaveAndClose = useCallback(async () => {
        const saved = await handleSave();
        if (saved) {
            navigate('/campaigns');
        }
    }, [handleSave, navigate]);

    /**
     * Handle back navigation with unsaved changes check.
     */
    const handleBack = useCallback(() => {
        const goBack = () => navigate('/');
        if (!checkUnsavedChanges(goBack)) {
            goBack();
        }
    }, [navigate, checkUnsavedChanges]);

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Home', path: '/' },
            { label: campaign?.name ?? 'Campaign' },
        ],
        [campaign]
    );

    /**
     * Render the main content based on active navigation item.
     */
    const renderContent = () => {
        switch (activeNavItem) {
            case 'overview':
                return (
                    <CampaignSettings
                        campaign={campaign ?? null}
                        isLoading={campaignLoading}
                        error={campaignError}
                        onChange={handleFormChange}
                        formData={formData ?? undefined}
                    />
                );
            case 'sessions':
                return (
                    <>
                        <SessionsView
                            campaignId={campaignId!}
                            selectedChapterId={selectedChapterId}
                            onSelectChapter={setSelectedChapterId}
                            selectedSessionId={selectedSessionId}
                            onSelectSession={setSelectedSessionId}
                            onCreateChapter={() => {
                                setEditingChapterId(undefined);
                                setChapterEditorOpen(true);
                            }}
                            onEditChapter={(chapter) => {
                                setEditingChapterId(chapter.id);
                                setChapterEditorOpen(true);
                            }}
                            onCreateSession={() => {
                                setEditingSessionId(undefined);
                                setSessionEditorOpen(true);
                            }}
                            onEditSession={(session) => {
                                setEditingSessionId(session.id);
                                setSessionEditorOpen(true);
                            }}
                        />

                        {/* Chapter Editor Dialog */}
                        <ChapterEditor
                            campaignId={campaignId!}
                            chapterId={editingChapterId}
                            open={chapterEditorOpen}
                            onClose={() => setChapterEditorOpen(false)}
                            onSave={(chapter) => {
                                setSelectedChapterId(chapter.id);
                                setChapterEditorOpen(false);
                            }}
                        />

                        {/* Session Editor Dialog */}
                        <SessionEditor
                            campaignId={campaignId!}
                            chapterId={selectedChapterId}
                            sessionId={editingSessionId}
                            open={sessionEditorOpen}
                            onClose={() => setSessionEditorOpen(false)}
                            onSave={(session) => {
                                setSelectedSessionId(session.id);
                                setSessionEditorOpen(false);
                            }}
                        />
                    </>
                );
            case 'player-characters':
                return <PlayerCharactersView />;
            default:
                return null;
        }
    };

    // Loading state for initial page load
    if (campaignLoading && !campaign) {
        return (
            <FullScreenLayout
                title="Loading..."
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
            >
                <Box sx={{ display: 'flex', height: '100%' }}>
                    <Box
                        sx={{
                            width: NAV_WIDTH,
                            flexShrink: 0,
                            display: { xs: 'none', md: 'block' },
                        }}
                    >
                        <Skeleton variant="rectangular" height="100%" />
                    </Box>
                    <Box sx={{ flexGrow: 1, p: 3 }}>
                        <Skeleton variant="rectangular" height={56} sx={{ mb: 3 }} />
                        <Skeleton variant="rectangular" height={200} sx={{ mb: 3 }} />
                        <Skeleton variant="rectangular" height={56} />
                    </Box>
                </Box>
            </FullScreenLayout>
        );
    }

    // Error state
    if (campaignError && !campaign) {
        return (
            <FullScreenLayout
                title="Error"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                onBack={handleBack}
            >
                <Alert severity="error" sx={{ maxWidth: 800, mx: 'auto' }}>
                    Failed to load campaign. Please try again later.
                </Alert>
            </FullScreenLayout>
        );
    }

    // Determine if save buttons should be shown based on active view
    const showSaveButtons = activeNavItem === 'overview';

    return (
        <FullScreenLayout
            title={campaign?.name ?? 'Campaign Dashboard'}
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={updateCampaign.isPending}
            onSave={showSaveButtons ? handleSave : undefined}
            onSaveAndClose={showSaveButtons ? handleSaveAndClose : undefined}
            onBack={handleBack}
            showSaveButtons={showSaveButtons}
        >
            <Box
                sx={{
                    display: 'flex',
                    height: '100%',
                    mx: -3, // Counteract padding from FullScreenLayout
                    mt: -3,
                }}
            >
                {/* Left Navigation */}
                {campaignId && (
                    <CampaignNav
                        campaignId={campaignId}
                        activeItem={activeNavItem}
                        onItemSelect={handleNavItemSelect}
                        hasUnsavedChanges={isDirty}
                        onCheckUnsavedChanges={checkUnsavedChanges}
                    />
                )}

                {/* Main Content Area */}
                <Box
                    component="main"
                    sx={{
                        flexGrow: 1,
                        p: 3,
                        overflow: 'auto',
                        ml: { xs: 0, md: 0 }, // No margin needed, drawer handles its own space
                    }}
                >
                    {/* Mutation error alert */}
                    {updateCampaign.error && (
                        <Alert severity="error" sx={{ mb: 3, maxWidth: 800, mx: 'auto' }}>
                            Failed to save campaign settings. Please try again.
                        </Alert>
                    )}

                    {/* Render active view content */}
                    {renderContent()}
                </Box>
            </Box>

            {/* Navigation confirmation dialog */}
            {ConfirmDialog}
        </FullScreenLayout>
    );
}