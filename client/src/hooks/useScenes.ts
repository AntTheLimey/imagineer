// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for scene management — listing, creating,
 * updating, and deleting scenes within game sessions.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { scenesApi } from '../api/scenes';
import type {
    Scene,
    CreateSceneInput,
    UpdateSceneInput,
} from '../api/scenes';

/**
 * Query keys for scene queries.
 */
export const sceneKeys = {
    all: ['scenes'] as const,
    lists: () => [...sceneKeys.all, 'list'] as const,
    list: (campaignId: number, sessionId: number) =>
        [...sceneKeys.lists(), campaignId, sessionId] as const,
    details: () => [...sceneKeys.all, 'detail'] as const,
    detail: (campaignId: number, sessionId: number, sceneId: number) =>
        [...sceneKeys.details(), campaignId, sessionId, sceneId] as const,
};

/**
 * Fetch all scenes for a specific session within a campaign.
 */
export function useScenes(campaignId: number, sessionId: number) {
    return useQuery<Scene[]>({
        queryKey: sceneKeys.list(campaignId, sessionId),
        queryFn: () => scenesApi.list(campaignId, sessionId),
        enabled: !!campaignId && !!sessionId,
    });
}

/**
 * Fetch a single scene by ID.
 */
export function useScene(
    campaignId: number,
    sessionId: number,
    sceneId: number
) {
    return useQuery<Scene>({
        queryKey: sceneKeys.detail(campaignId, sessionId, sceneId),
        queryFn: () => scenesApi.get(campaignId, sessionId, sceneId),
        enabled: !!campaignId && !!sessionId && !!sceneId,
    });
}

/**
 * Mutation to create a new scene. Invalidates the scene list on success.
 */
export function useCreateScene(campaignId: number, sessionId: number) {
    const queryClient = useQueryClient();
    return useMutation<Scene, Error, CreateSceneInput>({
        mutationFn: (input) =>
            scenesApi.create(campaignId, sessionId, input),
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: sceneKeys.list(campaignId, sessionId),
            });
        },
    });
}

/**
 * Mutation to update an existing scene. Updates cache and invalidates
 * the scene list on success.
 */
export function useUpdateScene(campaignId: number, sessionId: number) {
    const queryClient = useQueryClient();
    return useMutation<
        Scene,
        Error,
        { sceneId: number; input: UpdateSceneInput }
    >({
        mutationFn: ({ sceneId, input }) =>
            scenesApi.update(campaignId, sessionId, sceneId, input),
        onSuccess: (data) => {
            queryClient.setQueryData(
                sceneKeys.detail(campaignId, sessionId, data.id),
                data
            );
            queryClient.invalidateQueries({
                queryKey: sceneKeys.list(campaignId, sessionId),
            });
        },
    });
}

/**
 * Mutation to delete a scene. Removes from cache and invalidates
 * the scene list on success.
 */
export function useDeleteScene(campaignId: number, sessionId: number) {
    const queryClient = useQueryClient();
    return useMutation<void, Error, number>({
        mutationFn: (sceneId) =>
            scenesApi.delete(campaignId, sessionId, sceneId),
        onSuccess: (_data, sceneId) => {
            queryClient.removeQueries({
                queryKey: sceneKeys.detail(
                    campaignId,
                    sessionId,
                    sceneId
                ),
            });
            queryClient.invalidateQueries({
                queryKey: sceneKeys.list(campaignId, sessionId),
            });
        },
    });
}
