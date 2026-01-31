// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for timeline event operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    timelineApi,
    ListTimelineEventsParams,
    CreateTimelineEventInput,
    UpdateTimelineEventInput,
} from '../api';
import type { TimelineEvent } from '../types';

/**
 * Query keys for timeline-related queries.
 */
export const timelineKeys = {
    all: ['timeline'] as const,
    lists: () => [...timelineKeys.all, 'list'] as const,
    list: (params: ListTimelineEventsParams) =>
        [...timelineKeys.lists(), params] as const,
    details: () => [...timelineKeys.all, 'detail'] as const,
    detail: (campaignId: string, eventId: string) =>
        [...timelineKeys.details(), campaignId, eventId] as const,
    forEntity: (campaignId: string, entityId: string) =>
        [...timelineKeys.all, 'entity', campaignId, entityId] as const,
};

/**
 * Hook to fetch list of timeline events for a campaign.
 */
export function useTimelineEvents(params: ListTimelineEventsParams) {
    return useQuery({
        queryKey: timelineKeys.list(params),
        queryFn: () => timelineApi.list(params),
        enabled: !!params.campaignId,
    });
}

/**
 * Hook to fetch a single timeline event by ID.
 */
export function useTimelineEvent(
    campaignId: string,
    eventId: string,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: timelineKeys.detail(campaignId, eventId),
        queryFn: () => timelineApi.get(campaignId, eventId),
        enabled: options?.enabled ?? (!!campaignId && !!eventId),
    });
}

/**
 * Hook to fetch timeline events involving a specific entity.
 */
export function useEntityTimeline(
    campaignId: string,
    entityId: string,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: timelineKeys.forEntity(campaignId, entityId),
        queryFn: () => timelineApi.getForEntity(campaignId, entityId),
        enabled: options?.enabled ?? (!!campaignId && !!entityId),
    });
}

/**
 * Hook to create a new timeline event.
 */
export function useCreateTimelineEvent() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: CreateTimelineEventInput) => timelineApi.create(input),
        onSuccess: (data: TimelineEvent) => {
            // Invalidate timeline lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...timelineKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: string };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
            // Invalidate entity timelines for all involved entities
            data.entityIds.forEach((entityId) => {
                queryClient.invalidateQueries({
                    queryKey: timelineKeys.forEntity(data.campaignId, entityId),
                });
            });
        },
    });
}

/**
 * Hook to update an existing timeline event.
 */
export function useUpdateTimelineEvent() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            eventId,
            input,
        }: {
            campaignId: string;
            eventId: string;
            input: UpdateTimelineEventInput;
        }) => timelineApi.update(campaignId, eventId, input),
        onSuccess: (data: TimelineEvent) => {
            // Update the specific event in cache
            queryClient.setQueryData(
                timelineKeys.detail(data.campaignId, data.id),
                data
            );
            // Invalidate timeline lists
            queryClient.invalidateQueries({
                queryKey: [...timelineKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: string };
                        return params.campaignId === data.campaignId;
                    }
                    return false;
                },
            });
            // Invalidate entity timelines for all involved entities
            data.entityIds.forEach((entityId) => {
                queryClient.invalidateQueries({
                    queryKey: timelineKeys.forEntity(data.campaignId, entityId),
                });
            });
        },
    });
}

/**
 * Hook to delete a timeline event.
 */
export function useDeleteTimelineEvent() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            eventId,
        }: {
            campaignId: string;
            eventId: string;
        }) => timelineApi.delete(campaignId, eventId),
        onSuccess: (_data: void, variables: { campaignId: string; eventId: string }) => {
            // Remove from cache
            queryClient.removeQueries({
                queryKey: timelineKeys.detail(variables.campaignId, variables.eventId),
            });
            // Invalidate timeline lists for this campaign
            queryClient.invalidateQueries({
                queryKey: [...timelineKeys.lists()],
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                        const params = key[2] as { campaignId?: string };
                        return params.campaignId === variables.campaignId;
                    }
                    return false;
                },
            });
        },
    });
}
