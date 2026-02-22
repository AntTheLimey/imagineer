// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for campaign operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    campaignsApi,
    ListCampaignsParams,
    CreateCampaignInput,
    UpdateCampaignInput,
} from '../api';
import type { Campaign } from '../types';

/**
 * Query keys for campaign-related queries.
 */
export const campaignKeys = {
    all: ['campaigns'] as const,
    lists: () => [...campaignKeys.all, 'list'] as const,
    list: (params?: ListCampaignsParams) => [...campaignKeys.lists(), params] as const,
    details: () => [...campaignKeys.all, 'detail'] as const,
    detail: (id: number) => [...campaignKeys.details(), id] as const,
};

/**
 * Hook to fetch list of campaigns with optional pagination and filtering.
 */
export function useCampaigns(params?: ListCampaignsParams) {
    return useQuery({
        queryKey: campaignKeys.list(params),
        queryFn: () => campaignsApi.list(params),
    });
}

/**
 * Hook to fetch a single campaign by ID.
 */
export function useCampaign(id: number, options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: campaignKeys.detail(id),
        queryFn: () => campaignsApi.get(id),
        enabled: options?.enabled ?? !!id,
    });
}

/**
 * Hook to create a new campaign.
 */
export function useCreateCampaign() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: CreateCampaignInput) => campaignsApi.create(input),
        onSuccess: () => {
            // Invalidate campaign lists to refetch
            queryClient.invalidateQueries({ queryKey: campaignKeys.lists() });
        },
    });
}

/**
 * Hook to update an existing campaign.
 */
export function useUpdateCampaign() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            id,
            input,
            options,
        }: {
            id: number;
            input: UpdateCampaignInput;
            options?: { analyze?: boolean; enrich?: boolean; phases?: string[] };
        }) => campaignsApi.update(id, input, options),
        onSuccess: (data: Campaign) => {
            // Update the specific campaign in cache
            queryClient.setQueryData(campaignKeys.detail(data.id), data);
            // Invalidate lists to refetch
            queryClient.invalidateQueries({ queryKey: campaignKeys.lists() });
        },
    });
}

/**
 * Hook to delete a campaign.
 */
export function useDeleteCampaign() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (id: number) => campaignsApi.delete(id),
        onSuccess: (_data: void, id: number) => {
            // Remove from cache
            queryClient.removeQueries({ queryKey: campaignKeys.detail(id) });
            // Invalidate lists to refetch
            queryClient.invalidateQueries({ queryKey: campaignKeys.lists() });
        },
    });
}
