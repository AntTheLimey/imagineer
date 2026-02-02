// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Session API service - CRUD operations for sessions.
 */

import { apiClient } from './client';
import type { Session, SessionStage, SessionStatus } from '../types';

/**
 * Parameters for listing sessions.
 */
export interface ListSessionsParams {
    chapterId?: string;
    status?: SessionStatus;
}

/**
 * Input for creating a new session.
 */
export interface CreateSessionInput {
    chapterId?: string;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    stage?: SessionStage;
    prepNotes?: string;
    plannedScenes?: unknown[];
}

/**
 * Input for updating an existing session.
 */
export interface UpdateSessionInput {
    chapterId?: string;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    actualDate?: string;
    status?: SessionStatus;
    stage?: SessionStage;
    prepNotes?: string;
    plannedScenes?: unknown[];
    actualNotes?: string;
    discoveries?: Array<{ entityId: string; howDiscovered: string }>;
    playerDecisions?: unknown[];
    consequences?: unknown[];
}

/**
 * Session API service.
 */
export const sessionsApi = {
    /**
     * List all sessions for a campaign with optional filtering.
     */
    list(campaignId: string, params?: ListSessionsParams): Promise<Session[]> {
        return apiClient.get<Session[]>(
            `/campaigns/${campaignId}/sessions`,
            params ? { ...params } : undefined
        );
    },

    /**
     * List all sessions for a specific chapter.
     */
    listByChapter(campaignId: string, chapterId: string): Promise<Session[]> {
        return apiClient.get<Session[]>(
            `/campaigns/${campaignId}/chapters/${chapterId}/sessions`
        );
    },

    /**
     * Get a single session by ID.
     */
    get(campaignId: string, sessionId: string): Promise<Session> {
        return apiClient.get<Session>(
            `/campaigns/${campaignId}/sessions/${sessionId}`
        );
    },

    /**
     * Create a new session.
     */
    create(campaignId: string, input: CreateSessionInput): Promise<Session> {
        return apiClient.post<Session>(
            `/campaigns/${campaignId}/sessions`,
            input
        );
    },

    /**
     * Update an existing session.
     */
    update(
        campaignId: string,
        sessionId: string,
        input: UpdateSessionInput
    ): Promise<Session> {
        return apiClient.put<Session>(
            `/campaigns/${campaignId}/sessions/${sessionId}`,
            input
        );
    },

    /**
     * Delete a session.
     */
    delete(campaignId: string, sessionId: string): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/sessions/${sessionId}`
        );
    },
};

export default sessionsApi;
