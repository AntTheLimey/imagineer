// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for session operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    sessionsApi,
    ListSessionsParams,
    CreateSessionInput,
    UpdateSessionInput,
} from '../api/sessions';
import type { Session } from '../types';

/**
 * Query keys for session-related queries.
 */
export const sessionKeys = {
    all: ['sessions'] as const,
    lists: () => [...sessionKeys.all, 'list'] as const,
    list: (campaignId: string, params?: ListSessionsParams) =>
        [...sessionKeys.lists(), campaignId, params] as const,
    byChapter: (campaignId: string, chapterId: string) =>
        [...sessionKeys.lists(), 'byChapter', campaignId, chapterId] as const,
    details: () => [...sessionKeys.all, 'detail'] as const,
    detail: (campaignId: string, sessionId: string) =>
        [...sessionKeys.details(), campaignId, sessionId] as const,
};

/**
 * Hook to fetch list of sessions for a campaign with optional filtering.
 */
export function useSessions(campaignId: string, params?: ListSessionsParams) {
    return useQuery({
        queryKey: sessionKeys.list(campaignId, params),
        queryFn: () => sessionsApi.list(campaignId, params),
        enabled: !!campaignId,
    });
}

/**
 * Hook to fetch sessions for a specific chapter.
 */
export function useSessionsByChapter(campaignId: string, chapterId: string) {
    return useQuery({
        queryKey: sessionKeys.byChapter(campaignId, chapterId),
        queryFn: () => sessionsApi.listByChapter(campaignId, chapterId),
        enabled: !!campaignId && !!chapterId,
    });
}

/**
 * Hook to fetch a single session by ID.
 */
export function useSession(
    campaignId: string,
    sessionId: string,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: sessionKeys.detail(campaignId, sessionId),
        queryFn: () => sessionsApi.get(campaignId, sessionId),
        enabled: options?.enabled ?? (!!campaignId && !!sessionId),
    });
}

/**
 * Hook to create a new session.
 */
export function useCreateSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            input,
        }: {
            campaignId: string;
            input: CreateSessionInput;
        }) => sessionsApi.create(campaignId, input),
        onSuccess: (data: Session) => {
            // Invalidate session lists for this campaign
            queryClient.invalidateQueries({
                queryKey: sessionKeys.lists(),
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    // Match any list query that includes this campaign ID
                    return key.includes(data.campaignId);
                },
            });
            // If session has a chapter, also invalidate that chapter's session list
            if (data.chapterId) {
                queryClient.invalidateQueries({
                    queryKey: sessionKeys.byChapter(
                        data.campaignId,
                        data.chapterId
                    ),
                });
            }
        },
    });
}

/**
 * Hook to update an existing session.
 */
export function useUpdateSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            sessionId,
            input,
        }: {
            campaignId: string;
            sessionId: string;
            input: UpdateSessionInput;
        }) => sessionsApi.update(campaignId, sessionId, input),
        onSuccess: (data: Session) => {
            // Update the specific session in cache
            queryClient.setQueryData(
                sessionKeys.detail(data.campaignId, data.id),
                data
            );
            // Invalidate session lists for this campaign
            queryClient.invalidateQueries({
                queryKey: sessionKeys.lists(),
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    return key.includes(data.campaignId);
                },
            });
            // If session has a chapter, also invalidate that chapter's session list
            if (data.chapterId) {
                queryClient.invalidateQueries({
                    queryKey: sessionKeys.byChapter(
                        data.campaignId,
                        data.chapterId
                    ),
                });
            }
        },
    });
}

/**
 * Hook to delete a session.
 */
export function useDeleteSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            sessionId,
        }: {
            campaignId: string;
            sessionId: string;
        }) => sessionsApi.delete(campaignId, sessionId),
        onSuccess: (
            _data: void,
            variables: { campaignId: string; sessionId: string }
        ) => {
            // Remove from cache
            queryClient.removeQueries({
                queryKey: sessionKeys.detail(
                    variables.campaignId,
                    variables.sessionId
                ),
            });
            // Invalidate session lists for this campaign
            queryClient.invalidateQueries({
                queryKey: sessionKeys.lists(),
                predicate: (query) => {
                    const key = query.queryKey as unknown[];
                    return key.includes(variables.campaignId);
                },
            });
        },
    });
}
