// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for entity log management - listing, creating,
 * updating, and deleting log entries for campaign entities.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { entityLogApi } from '../api/entityLog';
import type {
    EntityLog,
    CreateEntityLogInput,
    UpdateEntityLogInput,
} from '../api/entityLog';

/**
 * Query keys for entity log queries.
 */
export const entityLogKeys = {
    all: ['entityLog'] as const,
    list: (campaignId: number, entityId: number) =>
        [...entityLogKeys.all, 'list', campaignId, entityId] as const,
};

/**
 * Fetch all log entries for a specific entity within a campaign.
 *
 * @param campaignId - The campaign the entity belongs to.
 * @param entityId - The entity to list logs for.
 */
export function useEntityLogs(campaignId: number, entityId: number) {
    return useQuery<EntityLog[]>({
        queryKey: entityLogKeys.list(campaignId, entityId),
        queryFn: () => entityLogApi.list(campaignId, entityId),
        enabled: !!campaignId && !!entityId,
    });
}

/**
 * Mutation to create a new log entry for an entity. Invalidates the
 * entity log list on success.
 *
 * @param campaignId - The campaign the entity belongs to.
 * @param entityId - The entity to create a log entry for.
 */
export function useCreateEntityLog(campaignId: number, entityId: number) {
    const queryClient = useQueryClient();
    return useMutation<EntityLog, Error, CreateEntityLogInput>({
        mutationFn: (input) =>
            entityLogApi.create(campaignId, entityId, input),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: entityLogKeys.list(campaignId, entityId),
            });
        },
    });
}

/**
 * Mutation to update an existing entity log entry. Invalidates the
 * entity log list on success.
 *
 * @param campaignId - The campaign the entity belongs to.
 * @param entityId - The entity the log entry belongs to.
 */
export function useUpdateEntityLog(campaignId: number, entityId: number) {
    const queryClient = useQueryClient();
    return useMutation<
        EntityLog,
        Error,
        { logId: number; input: UpdateEntityLogInput }
    >({
        mutationFn: ({ logId, input }) =>
            entityLogApi.update(campaignId, entityId, logId, input),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: entityLogKeys.list(campaignId, entityId),
            });
        },
    });
}

/**
 * Mutation to delete an entity log entry. Invalidates the entity log
 * list on success.
 *
 * @param campaignId - The campaign the entity belongs to.
 * @param entityId - The entity the log entry belongs to.
 */
export function useDeleteEntityLog(campaignId: number, entityId: number) {
    const queryClient = useQueryClient();
    return useMutation<void, Error, number>({
        mutationFn: (logId) =>
            entityLogApi.delete(campaignId, entityId, logId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: entityLogKeys.list(campaignId, entityId),
            });
        },
    });
}
