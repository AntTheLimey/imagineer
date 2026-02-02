// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * HomePage - Landing page with smart campaign redirect logic.
 *
 * Handles the root route (/) with the following behavior:
 * 1. If a current campaign is stored and valid, redirect to its overview
 * 2. If no current campaign but user has campaigns, redirect to the latest
 * 3. If no campaigns exist, show the welcome/onboarding screen
 */

import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Box, CircularProgress, Typography } from '@mui/material';
import { useCampaignContext } from '../contexts/CampaignContext';
import NoCampaignSelected from './NoCampaignSelected';

/**
 * HomePage component that handles smart campaign routing.
 *
 * Shows a loading state while determining the appropriate redirect,
 * then either redirects to a campaign overview or shows the welcome screen.
 *
 * @returns The HomePage element
 */
export default function HomePage() {
    const navigate = useNavigate();
    const {
        currentCampaignId,
        campaigns,
        campaignsLoading,
        isInitializing,
        setCurrentCampaignId,
        getLatestCampaign,
    } = useCampaignContext();

    // Track if we've attempted to redirect
    const [redirectAttempted, setRedirectAttempted] = useState(false);

    useEffect(() => {
        // Wait for initialization to complete
        if (isInitializing || campaignsLoading) {
            return;
        }

        // Already attempted redirect, don't try again
        if (redirectAttempted) {
            return;
        }

        // If we have a current campaign ID (validated), redirect to it
        if (currentCampaignId) {
            setRedirectAttempted(true);
            navigate(`/campaigns/${currentCampaignId}/overview`, {
                replace: true,
            });
            return;
        }

        // No current campaign set, check if user has any campaigns
        if (campaigns && campaigns.length > 0) {
            // Get the latest campaign and set it as current
            const latestCampaign = getLatestCampaign();
            if (latestCampaign) {
                setRedirectAttempted(true);
                setCurrentCampaignId(latestCampaign.id);
                navigate(`/campaigns/${latestCampaign.id}/overview`, {
                    replace: true,
                });
                return;
            }
        }

        // No campaigns exist, mark as attempted so we show welcome screen
        setRedirectAttempted(true);
    }, [
        currentCampaignId,
        campaigns,
        campaignsLoading,
        isInitializing,
        redirectAttempted,
        navigate,
        setCurrentCampaignId,
        getLatestCampaign,
    ]);

    // Show loading while initializing or loading campaigns
    if (isInitializing || campaignsLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '60vh',
                    gap: 2,
                }}
            >
                <CircularProgress />
                <Typography variant="body2" color="text.secondary">
                    Loading your campaigns...
                </Typography>
            </Box>
        );
    }

    // Show loading briefly while redirect is in progress
    if (!redirectAttempted) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '60vh',
                    gap: 2,
                }}
            >
                <CircularProgress />
                <Typography variant="body2" color="text.secondary">
                    Loading your campaigns...
                </Typography>
            </Box>
        );
    }

    // If we've attempted redirect and we're still here, show the welcome screen
    // This happens when user has no campaigns
    return <NoCampaignSelected />;
}
