// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for entity operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    entitiesApi,
    ListEntitiesParams,
    CreateEntityInput,
    UpdateEntityInput,
} from '../api';
import type { AnalysisOptions } from '../api';
import type { Entity } from '../types';

/**
 * Query keys for entity-related queries.
 */
export const entityKeys = {
    all: ['entities'] as const,
    lists: () => [...entityKeys.all, 'list'] as const,
    list: (params: ListEntitiesParams) => [...entityKeys.lists(), params] as const,
    details: () => [...entityKeys.all, 'detail'] as const,
    detail: (campaignId: number, entityId: number) =>
        [...entityKeys.details(), campaignId, entityId] as const,
    similar: (campaignId: number, name: string) =>
        [...entityKeys.all, 'similar', campaignId, name] as const,
};

/**
 * Hook to fetch list of entities for a campaign with optional pagination and filtering.
 */
export function useEntities(params: ListEntitiesParams) {
    return useQuery({
        queryKey: entityKeys.list(params),
        queryFn: () => entitiesApi.list(params),
        enabled: !!params.campaignId,
    });
}

/**
 * Hook to fetch a single entity by ID.
 */
export function useEntity(
    campaignId: number,
    entityId: number,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: entityKeys.detail(campaignId, entityId),
        queryFn: () => entitiesApi.get(campaignId, entityId),
        enabled: options?.enabled ?? (!!campaignId && !!entityId),
    });
}

/**
 * Hook to search for similar entities (for duplicate detection).
 */
export function useSimilarEntities(
    campaignId: number,
    name: string,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: entityKeys.similar(campaignId, name),
        queryFn: () => entitiesApi.searchSimilar(campaignId, name),
        enabled: options?.enabled ?? (!!campaignId && name.length >= 2),
    });
}

/**
 * Hook to create a new entity.
 */
export function useCreateEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: CreateEntityInput) => entitiesApi.create(input),
        onSuccess: (data: Entity) => {
            // Invalidate entity lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...entityKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: number };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
        },
    });
}

/**
 * Hook to update an existing entity.
 */
export function useUpdateEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            entityId,
            input,
            options,
        }: {
            campaignId: number;
            entityId: number;
            input: UpdateEntityInput;
            options?: AnalysisOptions;
        }) => entitiesApi.update(campaignId, entityId, input, options),
        onSuccess: (data: Entity) => {
            // Update the specific entity in cache
            queryClient.setQueryData(
                entityKeys.detail(data.campaignId, data.id),
                data
            );
            // Invalidate entity lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...entityKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: number };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
        },
    });
}

/**
 * Hook to delete an entity.
 */
export function useDeleteEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            entityId,
        }: {
            campaignId: number;
            entityId: number;
        }) => entitiesApi.delete(campaignId, entityId),
        onSuccess: (_data: void, variables: { campaignId: number; entityId: number }) => {
            // Remove from cache
            queryClient.removeQueries({
                queryKey: entityKeys.detail(variables.campaignId, variables.entityId),
            });
            // Invalidate entity lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...entityKeys.lists()],
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
