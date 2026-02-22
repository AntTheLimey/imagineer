// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Chapter API service - CRUD operations for chapters.
 */

import { apiClient } from './client';
import type { Chapter } from '../types';

/**
 * Input for creating a new chapter.
 */
export interface CreateChapterInput {
    title: string;
    overview?: string;
    sortOrder?: number;
}

/**
 * Input for updating an existing chapter.
 */
export interface UpdateChapterInput {
    title?: string;
    overview?: string;
    sortOrder?: number;
}

/**
 * Chapter API service.
 */
export const chaptersApi = {
    /**
     * List all chapters for a campaign.
     */
    list(campaignId: number): Promise<Chapter[]> {
        return apiClient.get<Chapter[]>(`/campaigns/${campaignId}/chapters`);
    },

    /**
     * Get a single chapter by ID.
     */
    get(campaignId: number, chapterId: number): Promise<Chapter> {
        return apiClient.get<Chapter>(
            `/campaigns/${campaignId}/chapters/${chapterId}`
        );
    },

    /**
     * Create a new chapter.
     */
    create(campaignId: number, input: CreateChapterInput): Promise<Chapter> {
        return apiClient.post<Chapter>(
            `/campaigns/${campaignId}/chapters`,
            input
        );
    },

    /**
     * Update an existing chapter.
     */
    update(
        campaignId: number,
        chapterId: number,
        input: UpdateChapterInput,
        options?: { analyze?: boolean; enrich?: boolean; phases?: string[] },
    ): Promise<Chapter> {
        const params: Record<string, string> = {};
        if (options?.analyze) params.analyze = 'true';
        if (options?.enrich) params.enrich = 'true';
        if (options?.phases?.length)
            params.phases = options.phases.join(',');
        return apiClient.put<Chapter>(
            `/campaigns/${campaignId}/chapters/${chapterId}`,
            input,
            params,
        );
    },

    /**
     * Delete a chapter.
     */
    delete(campaignId: number, chapterId: number): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/chapters/${chapterId}`
        );
    },
};

export default chaptersApi;
