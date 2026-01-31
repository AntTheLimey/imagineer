// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for dashboard and campaign statistics.
 */

import { useQuery } from '@tanstack/react-query';
import { statsApi } from '../api';

/**
 * Query keys for statistics-related queries.
 */
export const statsKeys = {
    all: ['stats'] as const,
    dashboard: () => [...statsKeys.all, 'dashboard'] as const,
    campaign: (campaignId: string) => [...statsKeys.all, 'campaign', campaignId] as const,
};

/**
 * Hook to fetch dashboard statistics (aggregate counts across all campaigns).
 */
export function useDashboardStats() {
    return useQuery({
        queryKey: statsKeys.dashboard(),
        queryFn: () => statsApi.getDashboard(),
        // Refresh stats periodically
        staleTime: 60 * 1000, // 1 minute
    });
}

/**
 * Hook to fetch statistics for a specific campaign.
 */
export function useCampaignStats(campaignId: string, options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: statsKeys.campaign(campaignId),
        queryFn: () => statsApi.getCampaign(campaignId),
        enabled: options?.enabled ?? !!campaignId,
        staleTime: 60 * 1000,
    });
}
