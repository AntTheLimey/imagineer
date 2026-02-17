// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Scene API service - CRUD operations for session scenes.
 */

import { apiClient } from './client';

/**
 * A scene within a game session.
 */
export interface Scene {
    id: number;
    sessionId: number;
    campaignId: number;
    title: string;
    description?: string;
    sceneType: string;
    status: string;
    sortOrder: number;
    objective?: string;
    gmNotes?: string;
    entityIds: number[];
    systemData?: Record<string, unknown>;
    source: string;
    sourceConfidence: string;
    connections?: unknown[];
    createdAt: string;
    updatedAt: string;
}

/**
 * Input for creating a new scene.
 */
export interface CreateSceneInput {
    title: string;
    description?: string;
    sceneType?: string;
    sortOrder?: number;
    objective?: string;
    gmNotes?: string;
    entityIds?: number[];
    systemData?: Record<string, unknown>;
    source?: string;
    sourceConfidence?: string;
    connections?: unknown[];
}

/**
 * Input for updating an existing scene.
 */
export interface UpdateSceneInput {
    title?: string;
    description?: string;
    sceneType?: string;
    status?: string;
    sortOrder?: number;
    objective?: string;
    gmNotes?: string;
    entityIds?: number[];
    systemData?: Record<string, unknown>;
    source?: string;
    sourceConfidence?: string;
    connections?: unknown[];
}

/**
 * Scene API service.
 */
export const scenesApi = {
    /**
     * List all scenes for a session.
     */
    list(campaignId: number, sessionId: number): Promise<Scene[]> {
        return apiClient.get<Scene[]>(
            `/campaigns/${campaignId}/sessions/${sessionId}/scenes`
        );
    },

    /**
     * Get a single scene by ID.
     */
    get(
        campaignId: number,
        sessionId: number,
        sceneId: number
    ): Promise<Scene> {
        return apiClient.get<Scene>(
            `/campaigns/${campaignId}/sessions/${sessionId}/scenes/${sceneId}`
        );
    },

    /**
     * Create a new scene.
     */
    create(
        campaignId: number,
        sessionId: number,
        input: CreateSceneInput
    ): Promise<Scene> {
        return apiClient.post<Scene>(
            `/campaigns/${campaignId}/sessions/${sessionId}/scenes`,
            input
        );
    },

    /**
     * Update an existing scene.
     */
    update(
        campaignId: number,
        sessionId: number,
        sceneId: number,
        input: UpdateSceneInput
    ): Promise<Scene> {
        return apiClient.put<Scene>(
            `/campaigns/${campaignId}/sessions/${sessionId}/scenes/${sceneId}`,
            input
        );
    },

    /**
     * Delete a scene.
     */
    delete(
        campaignId: number,
        sessionId: number,
        sceneId: number
    ): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/sessions/${sessionId}/scenes/${sceneId}`
        );
    },
};

export default scenesApi;
