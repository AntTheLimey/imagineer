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
import { useParams, useNavigate } from 'react-router-dom';
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
import {
    useCampaign,
    useUpdateCampaign,
    useUnsavedChanges,
} from '../hooks';
import { useCampaignContext } from '../contexts/CampaignContext';

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
 * Campaign Dashboard page serving as the main hub for campaign management.
 *
 * Features a left navigation panel for switching between views:
 * - Overview (campaign settings)
 * - Manage Entities (navigates to entities page)
 * - Manage Sessions (navigates to timeline page)
 * - Player Characters
 * - Import Campaign Notes
 * - Import Knowledge
 *
 * The settings view supports editing campaign name, description, game system,
 * genre, and image style prompt with dirty state tracking and save functionality.
 *
 * @returns The Campaign Dashboard page component
 */
export default function CampaignDashboard() {
    const { id: campaignId } = useParams<{ id: string }>();
    const navigate = useNavigate();

    // Active navigation item state
    const [activeNavItem, setActiveNavItem] = useState<CampaignNavItem>('overview');

    // Form data state (managed here for proper dirty tracking)
    const [formData, setFormData] = useState<CampaignSettingsData | null>(null);
    const [isFormDirty, setIsFormDirty] = useState(false);

    // Campaign context for keeping app-wide state updated
    const { setCurrentCampaignId } = useCampaignContext();

    // Fetch campaign data
    const {
        data: campaign,
        isLoading: campaignLoading,
        error: campaignError,
    } = useCampaign(campaignId ?? '', {
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
                id: campaignId,
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
        const goBack = () => navigate('/campaigns');
        if (!checkUnsavedChanges(goBack)) {
            goBack();
        }
    }, [navigate, checkUnsavedChanges]);

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Campaigns', path: '/campaigns' },
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
