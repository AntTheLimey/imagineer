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
 * Hook to fetch list of chapters for a campaign.
 */
export function useChapters(campaignId: string) {
    return useQuery({
        queryKey: chapterKeys.list(campaignId),
        queryFn: () => chaptersApi.list(campaignId),
        enabled: !!campaignId,
    });
}

/**
 * Hook to fetch a single chapter by ID.
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
 * Hook to create a new chapter.
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
 * Hook to update an existing chapter.
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
 * Hook to delete a chapter.
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
