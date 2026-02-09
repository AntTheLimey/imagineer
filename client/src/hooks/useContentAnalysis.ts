// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for content analysis - triggering analysis jobs,
 * listing detected items, and resolving them.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { contentAnalysisApi } from '../api/contentAnalysis';
import type {
    ContentAnalysisJob,
    ContentAnalysisItem,
    ResolveAnalysisItemRequest,
    TriggerAnalysisRequest,
    TriggerAnalysisResponse,
    PendingCountResponse,
} from '../api/contentAnalysis';
import { campaignKeys } from './useCampaigns';
import { entityKeys } from './useEntities';
import { chapterKeys } from './useChapters';
import { sessionKeys } from './useSessions';

/**
 * Query keys for content analysis queries.
 */
export const contentAnalysisKeys = {
    all: ['contentAnalysis'] as const,
    jobs: (campaignId: number) =>
        [...contentAnalysisKeys.all, 'jobs', campaignId] as const,
    job: (campaignId: number, jobId: number) =>
        [...contentAnalysisKeys.all, 'job', campaignId, jobId] as const,
    items: (campaignId: number, jobId: number, resolution?: string) =>
        [...contentAnalysisKeys.all, 'items', campaignId, jobId, resolution] as const,
    pendingCount: (campaignId: number, sourceTable?: string, sourceId?: number) =>
        [...contentAnalysisKeys.all, 'pendingCount', campaignId, sourceTable, sourceId] as const,
};

/**
 * Fetch all analysis jobs for a campaign.
 *
 * @param campaignId - The campaign to list jobs for.
 */
export function useAnalysisJobs(campaignId: number) {
    return useQuery<ContentAnalysisJob[]>({
        queryKey: contentAnalysisKeys.jobs(campaignId),
        queryFn: () => contentAnalysisApi.listJobs(campaignId),
        enabled: !!campaignId,
    });
}

/**
 * Fetch a single analysis job by ID.
 *
 * @param campaignId - The campaign the job belongs to.
 * @param jobId - The job ID.
 */
export function useAnalysisJob(campaignId: number, jobId: number) {
    return useQuery<ContentAnalysisJob>({
        queryKey: contentAnalysisKeys.job(campaignId, jobId),
        queryFn: () => contentAnalysisApi.getJob(campaignId, jobId),
        enabled: !!campaignId && !!jobId,
    });
}

/**
 * Fetch items for an analysis job, optionally filtered by resolution
 * status.
 *
 * @param campaignId - The campaign the job belongs to.
 * @param jobId - The job ID.
 * @param resolution - Optional resolution filter (e.g. 'pending').
 */
export function useAnalysisItems(campaignId: number, jobId: number, resolution?: string) {
    return useQuery<ContentAnalysisItem[]>({
        queryKey: contentAnalysisKeys.items(campaignId, jobId, resolution),
        queryFn: () => contentAnalysisApi.listJobItems(campaignId, jobId, resolution),
        enabled: !!campaignId && !!jobId,
    });
}

/**
 * Mutation to resolve a single analysis item. Invalidates content
 * analysis queries and source content caches on success, because the
 * backend may modify source content (e.g., inserting wiki links).
 *
 * @param campaignId - The campaign the item belongs to.
 */
export function useResolveItem(campaignId: number) {
    const queryClient = useQueryClient();
    return useMutation<void, Error, { itemId: number; req: ResolveAnalysisItemRequest }>({
        mutationFn: ({ itemId, req }) =>
            contentAnalysisApi.resolveItem(campaignId, itemId, req),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: contentAnalysisKeys.all,
            });
            // The backend modifies source content (campaigns, entities,
            // chapters, sessions) when applying wiki-link fixes, so
            // invalidate those caches too.
            queryClient.invalidateQueries({
                queryKey: campaignKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: entityKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: chapterKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: sessionKeys.all,
            });
        },
    });
}

/**
 * Mutation to trigger content analysis on a content field. Invalidates
 * all content analysis queries on success.
 *
 * @param campaignId - The campaign to analyse content in.
 */
export function useTriggerAnalysis(campaignId: number) {
    const queryClient = useQueryClient();
    return useMutation<TriggerAnalysisResponse, Error, TriggerAnalysisRequest>({
        mutationFn: (req) =>
            contentAnalysisApi.triggerAnalysis(campaignId, req),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: contentAnalysisKeys.all,
            });
        },
    });
}

/**
 * Mutation to batch-resolve all pending items of a given detection type
 * within a job. Invalidates content analysis and source content caches
 * on success.
 *
 * @param campaignId - The campaign the job belongs to.
 */
export function useBatchResolve(campaignId: number) {
    const queryClient = useQueryClient();
    return useMutation<
        { resolved: number },
        Error,
        { jobId: number; detectionType: string; resolution: string }
    >({
        mutationFn: ({ jobId, detectionType, resolution }) =>
            contentAnalysisApi.batchResolve(campaignId, jobId, {
                detectionType,
                resolution,
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: contentAnalysisKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: campaignKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: entityKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: chapterKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: sessionKeys.all,
            });
        },
    });
}

/**
 * Mutation to revert a previously resolved analysis item back to
 * pending. Invalidates content analysis and source content caches on
 * success.
 *
 * @param campaignId - The campaign the item belongs to.
 */
export function useRevertItem(campaignId: number) {
    const queryClient = useQueryClient();
    return useMutation<{ status: string }, Error, { itemId: number }>({
        mutationFn: ({ itemId }) =>
            contentAnalysisApi.revertItem(campaignId, itemId),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: contentAnalysisKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: campaignKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: entityKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: chapterKeys.all,
            });
            queryClient.invalidateQueries({
                queryKey: sessionKeys.all,
            });
        },
    });
}

/**
 * Fetch the count of pending (unresolved) analysis items for a campaign,
 * optionally scoped to a specific source table and ID.
 *
 * @param campaignId - The campaign to query.
 * @param sourceTable - Optional source table filter.
 * @param sourceId - Optional source ID filter.
 */
export function usePendingAnalysisCount(
    campaignId: number,
    sourceTable?: string,
    sourceId?: number
) {
    return useQuery<PendingCountResponse>({
        queryKey: contentAnalysisKeys.pendingCount(campaignId, sourceTable, sourceId),
        queryFn: () => contentAnalysisApi.getPendingCount(campaignId, sourceTable, sourceId),
        enabled: !!campaignId,
    });
}
