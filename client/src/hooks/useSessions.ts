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
    list: (campaignId: number, params?: ListSessionsParams) =>
        [...sessionKeys.lists(), campaignId, params] as const,
    byChapter: (campaignId: number, chapterId: number) =>
        [...sessionKeys.lists(), 'byChapter', campaignId, chapterId] as const,
    details: () => [...sessionKeys.all, 'detail'] as const,
    detail: (campaignId: number, sessionId: number) =>
        [...sessionKeys.details(), campaignId, sessionId] as const,
};

/**
 * Fetches the list of sessions for a campaign, optionally filtered by `params`.
 *
 * @param campaignId - The campaign identifier to fetch sessions for
 * @param params - Optional query parameters to filter or paginate the session list
 * @returns The React Query result containing the list of sessions for the specified campaign
 */
export function useSessions(campaignId: number, params?: ListSessionsParams) {
    return useQuery({
        queryKey: sessionKeys.list(campaignId, params),
        queryFn: () => sessionsApi.list(campaignId, params),
        enabled: !!campaignId,
    });
}

/**
 * Fetches sessions for a specific chapter within a campaign.
 *
 * The query is enabled only when both `campaignId` and `chapterId` are truthy.
 *
 * @returns The query result containing the session list for the specified chapter.
 */
export function useSessionsByChapter(campaignId: number, chapterId: number) {
    return useQuery({
        queryKey: sessionKeys.byChapter(campaignId, chapterId),
        queryFn: () => sessionsApi.listByChapter(campaignId, chapterId),
        enabled: !!campaignId && !!chapterId,
    });
}

/**
 * Fetches a single session for a campaign by session ID.
 *
 * @param campaignId - ID of the campaign that owns the session
 * @param sessionId - ID of the session to fetch
 * @param options - Optional settings for the query
 * @param options.enabled - If provided, controls whether the query is enabled; otherwise the query is enabled when both `campaignId` and `sessionId` are truthy
 * @returns The React Query result for the session, containing the session data when available
 */
export function useSession(
    campaignId: number,
    sessionId: number,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: sessionKeys.detail(campaignId, sessionId),
        queryFn: () => sessionsApi.get(campaignId, sessionId),
        enabled: options?.enabled ?? (!!campaignId && !!sessionId),
    });
}

/**
 * Create a new session and update related session query caches on success.
 *
 * Invalidates cached session lists for the created session's campaign and, if the session has a chapterId, invalidates that campaign+chapter list as well.
 *
 * @returns The React Query mutation object for creating sessions. The mutation expects an object with `campaignId: number` and `input: CreateSessionInput`. */
export function useCreateSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            input,
        }: {
            campaignId: number;
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
 * Create a React Query mutation hook that updates an existing session.
 *
 * Calls the sessions API to update a session. On success it replaces the session detail in the cache,
 * invalidates session list queries for the affected campaign, and if the updated session has a `chapterId`
 * also invalidates that chapter's session list.
 *
 * @returns A mutation hook that accepts an object `{ campaignId, sessionId, input }` to perform the update;
 *          on success the hook updates the session detail cache and invalidates relevant list queries.
 */
export function useUpdateSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            sessionId,
            input,
            options,
        }: {
            campaignId: number;
            sessionId: number;
            input: UpdateSessionInput;
            options?: { analyze?: boolean; enrich?: boolean; phases?: string[] };
        }) => sessionsApi.update(campaignId, sessionId, input, options),
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
 * Provides a mutation hook to delete a session and update session-related React Query cache.
 *
 * On success, removes the deleted session's detail from the cache and invalidates session list queries for the affected campaign.
 *
 * @returns The mutation result configured to delete a session by `{ campaignId, sessionId }`; on success it updates the cache as described above.
 */
export function useDeleteSession() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            sessionId,
        }: {
            campaignId: number;
            sessionId: number;
        }) => sessionsApi.delete(campaignId, sessionId),
        onSuccess: (
            _data: void,
            variables: { campaignId: number; sessionId: number }
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