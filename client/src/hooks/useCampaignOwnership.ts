// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hook for checking campaign ownership (GM status).
 *
 * This hook combines authentication state with campaign data to determine
 * if the current user is the Game Master (owner) of a campaign. This is
 * used to control access to GM-only features like GM Notes.
 */

import { useMemo } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useCampaign } from './useCampaigns';
import type { Campaign } from '../types';

/**
 * Result of the campaign ownership check.
 */
export interface CampaignOwnershipResult {
    /** Whether the current user is the campaign owner (GM). */
    isOwner: boolean;
    /** Whether the ownership check is still loading. */
    isLoading: boolean;
    /** The campaign data if available. */
    campaign: Campaign | undefined;
}

/**
 * Hook to check if the current authenticated user is the owner of a campaign.
 *
 * Returns isOwner as false if:
 * - User is not authenticated
 * - Campaign data is not yet loaded
 * - User ID does not match campaign owner ID
 *
 * @param campaignId - The ID of the campaign to check ownership for
 * @returns CampaignOwnershipResult with isOwner, isLoading, and campaign data
 */
export function useCampaignOwnership(campaignId: string): CampaignOwnershipResult {
    const { user, isAuthenticated } = useAuth();
    const { data: campaign, isLoading: campaignLoading } = useCampaign(campaignId, {
        enabled: !!campaignId,
    });

    const result = useMemo((): CampaignOwnershipResult => {
        // If not authenticated, user cannot be owner
        if (!isAuthenticated || !user) {
            return {
                isOwner: false,
                isLoading: campaignLoading,
                campaign,
            };
        }

        // If campaign is still loading, return loading state
        if (campaignLoading || !campaign) {
            return {
                isOwner: false,
                isLoading: true,
                campaign: undefined,
            };
        }

        // Check if user ID matches campaign owner ID
        const isOwner = campaign.ownerId === user.id;

        return {
            isOwner,
            isLoading: false,
            campaign,
        };
    }, [isAuthenticated, user, campaign, campaignLoading]);

    return result;
}

/**
 * Utility function to check if a user is the owner of a campaign.
 *
 * This is a pure function alternative to the hook, useful for cases where
 * you already have the user and campaign data available.
 *
 * @param userId - The ID of the user to check
 * @param campaign - The campaign to check ownership of
 * @returns true if the user is the campaign owner, false otherwise
 */
export function isCampaignOwner(
    userId: string | undefined,
    campaign: Campaign | undefined
): boolean {
    if (!userId || !campaign || !campaign.ownerId) {
        return false;
    }
    return campaign.ownerId === userId;
}
