// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for relationship type operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { relationshipTypesApi } from '../api/relationshipTypes';
import type { RelationshipType, CreateRelationshipTypeInput } from '../types';

/**
 * Query keys for relationship type queries.
 */
export const relationshipTypeKeys = {
    all: ['relationshipTypes'] as const,
    lists: () => [...relationshipTypeKeys.all, 'list'] as const,
    list: (campaignId: number) =>
        [...relationshipTypeKeys.lists(), campaignId] as const,
};

/**
 * Hook to fetch relationship types for a campaign.
 * Returns both system defaults and campaign-specific custom types.
 */
export function useRelationshipTypes(
    campaignId: number,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: relationshipTypeKeys.list(campaignId),
        queryFn: () => relationshipTypesApi.list(campaignId),
        enabled: options?.enabled ?? !!campaignId,
    });
}

/**
 * Hook to create a custom relationship type.
 */
export function useCreateRelationshipType() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            input,
        }: {
            campaignId: number;
            input: CreateRelationshipTypeInput;
        }) => relationshipTypesApi.create(campaignId, input),
        onSuccess: (data: RelationshipType) => {
            // Invalidate relationship types for this campaign
            queryClient.invalidateQueries({
                queryKey: relationshipTypeKeys.list(data.campaignId ?? 0),
            });
        },
    });
}

/**
 * Hook to delete a custom relationship type.
 */
export function useDeleteRelationshipType() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            typeId,
        }: {
            campaignId: number;
            typeId: number;
        }) => relationshipTypesApi.delete(campaignId, typeId),
        onSuccess: (
            _data: void,
            variables: { campaignId: number; typeId: number }
        ) => {
            // Invalidate relationship types for this campaign
            queryClient.invalidateQueries({
                queryKey: relationshipTypeKeys.list(variables.campaignId),
            });
        },
    });
}
