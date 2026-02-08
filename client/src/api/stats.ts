// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Stats API service - dashboard statistics.
 */

import { apiClient } from './client';

/**
 * Dashboard statistics response.
 */
export interface DashboardStats {
    campaignCount: number;
    npcCount: number;
    locationCount: number;
    timelineEventCount: number;
    itemCount: number;
    factionCount: number;
    totalEntityCount: number;
}

/**
 * Campaign-specific statistics response.
 */
export interface CampaignStats {
    entityCounts: Record<string, number>;
    relationshipCount: number;
    timelineEventCount: number;
    sessionCount: number;
    conflictCount: number;
}

/**
 * Stats API service.
 */
export const statsApi = {
    /**
     * Get dashboard statistics (aggregate counts across all campaigns).
     */
    getDashboard(): Promise<DashboardStats> {
        return apiClient.get<DashboardStats>('/stats/dashboard');
    },

    /**
     * Get statistics for a specific campaign.
     */
    getCampaign(campaignId: number): Promise<CampaignStats> {
        return apiClient.get<CampaignStats>(`/campaigns/${campaignId}/stats`);
    },
};

export default statsApi;
