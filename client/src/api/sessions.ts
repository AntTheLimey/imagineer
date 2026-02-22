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
import type { AnalysisOptions } from './types';
import type { Session, SessionStage, SessionStatus } from '../types';

/**
 * Parameters for listing sessions.
 */
export interface ListSessionsParams {
    chapterId?: number;
    status?: SessionStatus;
}

/**
 * Input for creating a new session.
 */
export interface CreateSessionInput {
    chapterId?: number;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    stage?: SessionStage;
    prepNotes?: string;
}

/**
 * Input for updating an existing session.
 */
export interface UpdateSessionInput {
    chapterId?: number;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    actualDate?: string;
    status?: SessionStatus;
    stage?: SessionStage;
    prepNotes?: string;
    actualNotes?: string;
    playNotes?: string;
}

/**
 * Session API service.
 */
export const sessionsApi = {
    /**
     * List all sessions for a campaign with optional filtering.
     */
    list(campaignId: number, params?: ListSessionsParams): Promise<Session[]> {
        return apiClient.get<Session[]>(
            `/campaigns/${campaignId}/sessions`,
            params ? { ...params } : undefined
        );
    },

    /**
     * List all sessions for a specific chapter.
     */
    listByChapter(campaignId: number, chapterId: number): Promise<Session[]> {
        return apiClient.get<Session[]>(
            `/campaigns/${campaignId}/chapters/${chapterId}/sessions`
        );
    },

    /**
     * Get a single session by ID.
     */
    get(campaignId: number, sessionId: number): Promise<Session> {
        return apiClient.get<Session>(
            `/campaigns/${campaignId}/sessions/${sessionId}`
        );
    },

    /**
     * Create a new session.
     */
    create(campaignId: number, input: CreateSessionInput): Promise<Session> {
        return apiClient.post<Session>(
            `/campaigns/${campaignId}/sessions`,
            input
        );
    },

    /**
     * Update an existing session.
     */
    update(
        campaignId: number,
        sessionId: number,
        input: UpdateSessionInput,
        options?: AnalysisOptions,
    ): Promise<Session> {
        const params: Record<string, string> = {};
        if (options?.analyze) params.analyze = 'true';
        if (options?.enrich) params.enrich = 'true';
        if (options?.phases?.length)
            params.phases = options.phases.join(',');
        return apiClient.put<Session>(
            `/campaigns/${campaignId}/sessions/${sessionId}`,
            input,
            params,
        );
    },

    /**
     * Delete a session.
     */
    delete(campaignId: number, sessionId: number): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/sessions/${sessionId}`
        );
    },
};

export default sessionsApi;
