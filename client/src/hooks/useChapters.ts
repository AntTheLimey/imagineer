// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for chapter operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    chaptersApi,
    CreateChapterInput,
    UpdateChapterInput,
} from '../api/chapters';
import type { Chapter } from '../types';

/**
 * Query keys for chapter-related queries.
 */
export const chapterKeys = {
    all: ['chapters'] as const,
    lists: () => [...chapterKeys.all, 'list'] as const,
    list: (campaignId: string) => [...chapterKeys.lists(), campaignId] as const,
    details: () => [...chapterKeys.all, 'detail'] as const,
    detail: (campaignId: string, chapterId: string) =>
        [...chapterKeys.details(), campaignId, chapterId] as const,
};

/**
 * Fetches the list of chapters for a given campaign.
 *
 * @returns The query result containing the campaign's chapters and React Query metadata (status, error, etc.).
 */
export function useChapters(campaignId: string) {
    return useQuery({
        queryKey: chapterKeys.list(campaignId),
        queryFn: () => chaptersApi.list(campaignId),
        enabled: !!campaignId,
    });
}

/**
 * Fetches a single chapter for a campaign.
 *
 * @param options - Optional settings
 * @param options.enabled - If provided, forces the query enabled state; otherwise the query is enabled when both `campaignId` and `chapterId` are truthy
 * @returns The React Query result for the requested chapter
 */
export function useChapter(
    campaignId: string,
    chapterId: string,
    options?: { enabled?: boolean }
) {
    return useQuery({
        queryKey: chapterKeys.detail(campaignId, chapterId),
        queryFn: () => chaptersApi.get(campaignId, chapterId),
        enabled: options?.enabled ?? (!!campaignId && !!chapterId),
    });
}

/**
 * Provides a React Query mutation to create a new chapter.
 *
 * The mutation expects an object with `campaignId` and `input` (CreateChapterInput). On successful creation it invalidates the chapter list cache for the affected campaign.
 *
 * @returns A mutation result whose mutate function accepts `{ campaignId: string; input: CreateChapterInput }`.
 */
export function useCreateChapter() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            input,
        }: {
            campaignId: string;
            input: CreateChapterInput;
        }) => chaptersApi.create(campaignId, input),
        onSuccess: (data: Chapter) => {
            // Invalidate chapter lists for this campaign
            queryClient.invalidateQueries({
                queryKey: chapterKeys.list(data.campaignId),
            });
        },
    });
}

/**
 * Provides a React Query mutation for updating a chapter.
 *
 * The mutation calls the API to update a chapter and, on success, updates the cached chapter detail for the affected campaign and invalidates the campaign's chapter list so it will be refreshed.
 *
 * @returns A React Query mutation result for performing chapter updates. The mutation accepts an object with `campaignId`, `chapterId`, and `input` and resolves with the updated `Chapter`.
 */
export function useUpdateChapter() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            chapterId,
            input,
        }: {
            campaignId: string;
            chapterId: string;
            input: UpdateChapterInput;
        }) => chaptersApi.update(campaignId, chapterId, input),
        onSuccess: (data: Chapter) => {
            // Update the specific chapter in cache
            queryClient.setQueryData(
                chapterKeys.detail(data.campaignId, data.id),
                data
            );
            // Invalidate chapter lists for this campaign
            queryClient.invalidateQueries({
                queryKey: chapterKeys.list(data.campaignId),
            });
        },
    });
}

/**
 * Provides a React Query mutation for deleting a chapter.
 *
 * On success the specific chapter detail is removed from the cache and the campaign's chapter list is invalidated.
 *
 * @returns A mutation object whose `mutate`/`mutateAsync` function accepts `{ campaignId: string; chapterId: string }` to delete the specified chapter.
 */
export function useDeleteChapter() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            chapterId,
        }: {
            campaignId: string;
            chapterId: string;
        }) => chaptersApi.delete(campaignId, chapterId),
        onSuccess: (
            _data: void,
            variables: { campaignId: string; chapterId: string }
        ) => {
            // Remove from cache
            queryClient.removeQueries({
                queryKey: chapterKeys.detail(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
            // Invalidate chapter lists for this campaign
            queryClient.invalidateQueries({
                queryKey: chapterKeys.list(variables.campaignId),
            });
        },
    });
}