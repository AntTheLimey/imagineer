// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for relationship operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    relationshipsApi,
    ListRelationshipsParams,
    CreateRelationshipInput,
    UpdateRelationshipInput,
} from '../api';
import type { Relationship } from '../types';

/**
 * Query keys for relationship-related queries.
 */
export const relationshipKeys = {
    all: ['relationships'] as const,
    lists: () => [...relationshipKeys.all, 'list'] as const,
    list: (params: ListRelationshipsParams) =>
        [...relationshipKeys.lists(), params] as const,
    details: () => [...relationshipKeys.all, 'detail'] as const,
    detail: (campaignId: number, relationshipId: number) =>
        [...relationshipKeys.details(), campaignId, relationshipId] as const,
    forEntity: (campaignId: number, entityId: number) =>
        [...relationshipKeys.all, 'entity', campaignId, entityId] as const,
};

/**
 * Hook to fetch list of relationships for a campaign.
 */
export function useRelationships(params: ListRelationshipsParams) {
    return useQuery({
        queryKey: relationshipKeys.list(params),
        queryFn: () => relationshipsApi.list(params),
        enabled: !!params.campaignId,
    });
}

/**
 * Hook to fetch a single relationship by ID.
 */
export function useRelationship(
    campaignId: number,
    relationshipId: number,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: relationshipKeys.detail(campaignId, relationshipId),
        queryFn: () => relationshipsApi.get(campaignId, relationshipId),
        enabled: options?.enabled ?? (!!campaignId && !!relationshipId),
    });
}

/**
 * Hook to fetch all relationships for a specific entity.
 */
export function useEntityRelationships(
    campaignId: number,
    entityId: number,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: relationshipKeys.forEntity(campaignId, entityId),
        queryFn: () => relationshipsApi.getForEntity(campaignId, entityId),
        enabled: options?.enabled ?? (!!campaignId && !!entityId),
    });
}

/**
 * Hook to create a new relationship.
 */
export function useCreateRelationship() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: CreateRelationshipInput) =>
            relationshipsApi.create(input),
        onSuccess: (data: Relationship) => {
            // Invalidate relationship lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...relationshipKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: number };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
            // Invalidate entity relationships for both source and target
            queryClient.invalidateQueries({
                queryKey: relationshipKeys.forEntity(data.campaignId, data.sourceEntityId),
            });
            queryClient.invalidateQueries({
                queryKey: relationshipKeys.forEntity(data.campaignId, data.targetEntityId),
            });
        },
    });
}

/**
 * Hook to update an existing relationship.
 */
export function useUpdateRelationship() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            relationshipId,
            input,
        }: {
            campaignId: number;
            relationshipId: number;
            input: UpdateRelationshipInput;
        }) => relationshipsApi.update(campaignId, relationshipId, input),
        onSuccess: (data: Relationship) => {
            // Update the specific relationship in cache
            queryClient.setQueryData(
                relationshipKeys.detail(data.campaignId, data.id),
                data
            );
            // Invalidate relationship lists
            queryClient.invalidateQueries({
                queryKey: [...relationshipKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: number };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
            // Invalidate entity relationships
            queryClient.invalidateQueries({
                queryKey: relationshipKeys.forEntity(data.campaignId, data.sourceEntityId),
            });
            queryClient.invalidateQueries({
                queryKey: relationshipKeys.forEntity(data.campaignId, data.targetEntityId),
            });
        },
    });
}

/**
 * Hook to delete a relationship.
 */
export function useDeleteRelationship() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            relationshipId,
        }: {
            campaignId: number;
            relationshipId: number;
        }) => relationshipsApi.delete(campaignId, relationshipId),
        onSuccess: (
            _data: void,
            variables: { campaignId: number; relationshipId: number }
        ) => {
            // Remove from cache
            queryClient.removeQueries({
                queryKey: relationshipKeys.detail(
                    variables.campaignId,
                    variables.relationshipId
                ),
            });
            // Invalidate relationship lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...relationshipKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: number };
                        return params.campaignId === variables.campaignId;
                    }
                    return false;
                },
            });
        },
    });
}
