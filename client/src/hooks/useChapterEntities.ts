// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for chapter-entity operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    chapterEntitiesApi,
    CreateChapterEntityInput,
    UpdateChapterEntityInput,
} from '../api/chapterEntities';
import type { ChapterEntity } from '../types';

/**
 * Query keys for chapter entity-related queries.
 */
export const chapterEntityKeys = {
    all: ['chapterEntities'] as const,
    lists: () => [...chapterEntityKeys.all, 'list'] as const,
    list: (campaignId: number, chapterId: number) =>
        [...chapterEntityKeys.lists(), campaignId, chapterId] as const,
    relationships: (campaignId: number, chapterId: number) =>
        [...chapterEntityKeys.all, 'relationships', campaignId, chapterId] as const,
};

/**
 * Fetches the list of entity links for a chapter.
 *
 * @param campaignId - The campaign ID
 * @param chapterId - The chapter ID to list entities for
 * @returns The query result containing the chapter's linked entities
 */
export function useChapterEntities(campaignId: number, chapterId: number) {
    return useQuery({
        queryKey: chapterEntityKeys.list(campaignId, chapterId),
        queryFn: () => chapterEntitiesApi.list(campaignId, chapterId),
        enabled: !!campaignId && !!chapterId,
    });
}

/**
 * Fetches relationships involving entities linked to a chapter.
 *
 * @param campaignId - The campaign ID
 * @param chapterId - The chapter ID
 * @returns The query result containing the chapter's relationships
 */
export function useChapterRelationships(campaignId: number, chapterId: number) {
    return useQuery({
        queryKey: chapterEntityKeys.relationships(campaignId, chapterId),
        queryFn: () => chapterEntitiesApi.listRelationships(campaignId, chapterId),
        enabled: !!campaignId && !!chapterId,
    });
}

/**
 * Provides a React Query mutation to link an entity to a chapter.
 *
 * @returns A mutation result for creating chapter-entity links
 */
export function useCreateChapterEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            chapterId,
            input,
        }: {
            campaignId: number;
            chapterId: number;
            input: CreateChapterEntityInput;
        }) => chapterEntitiesApi.create(campaignId, chapterId, input),
        onSuccess: (_data: ChapterEntity, variables) => {
            // Invalidate chapter entities list
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.list(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
            // Invalidate chapter relationships (entity set changed)
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.relationships(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
        },
    });
}

/**
 * Provides a React Query mutation for updating a chapter-entity link.
 *
 * @returns A mutation result for updating chapter-entity links
 */
export function useUpdateChapterEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            chapterId,
            linkId,
            input,
        }: {
            campaignId: number;
            chapterId: number;
            linkId: number;
            input: UpdateChapterEntityInput;
        }) => chapterEntitiesApi.update(campaignId, chapterId, linkId, input),
        onSuccess: (_data: ChapterEntity, variables) => {
            // Invalidate chapter entities list
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.list(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
            // Invalidate chapter relationships (entity set changed)
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.relationships(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
        },
    });
}

/**
 * Provides a React Query mutation for deleting a chapter-entity link.
 *
 * @returns A mutation result for removing chapter-entity links
 */
export function useDeleteChapterEntity() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({
            campaignId,
            chapterId,
            linkId,
        }: {
            campaignId: number;
            chapterId: number;
            linkId: number;
        }) => chapterEntitiesApi.delete(campaignId, chapterId, linkId),
        onSuccess: (
            _data: void,
            variables: { campaignId: number; chapterId: number; linkId: number }
        ) => {
            // Invalidate chapter entities list
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.list(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
            // Invalidate chapter relationships (entity set changed)
            queryClient.invalidateQueries({
                queryKey: chapterEntityKeys.relationships(
                    variables.campaignId,
                    variables.chapterId
                ),
            });
        },
    });
}
