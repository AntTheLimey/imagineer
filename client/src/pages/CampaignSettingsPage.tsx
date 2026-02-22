/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * CampaignSettingsPage - Thin route wrapper that fetches campaign
 * data and renders the CampaignSettings component.
 */

import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, Typography } from '@mui/material';
import CampaignSettings from '../components/CampaignSettings';
import { useCampaign } from '../hooks';

// CampaignSettings manages its own save state internally.
const handleChange = () => {};

export default function CampaignSettingsPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const campaignId = id ? Number(id) : undefined;
    const {
        data: campaign,
        isLoading,
        error,
    } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Safety net: if the campaign query errors (e.g. 404 after
    // deletion), redirect to the campaigns list instead of showing
    // a flash of "Failed to load campaign".
    useEffect(() => {
        if (error) {
            navigate('/campaigns', { replace: true });
        }
    }, [error, navigate]);

    return (
        <Box>
            <Typography
                variant="h4"
                sx={{ fontFamily: 'Cinzel', mb: 3 }}
            >
                Campaign Settings
            </Typography>
            <CampaignSettings
                campaign={campaign ?? null}
                isLoading={isLoading}
                error={error}
                onChange={handleChange}
                hideDescription
            />
        </Box>
    );
}
