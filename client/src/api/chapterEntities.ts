// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Chapter Entity API service - CRUD operations for chapter-entity links.
 */

import { apiClient } from './client';
import type { ChapterEntity, ChapterEntityMentionType, Relationship } from '../types';

/**
 * Input for creating a chapter-entity link.
 */
export interface CreateChapterEntityInput {
    entityId: number;
    mentionType?: ChapterEntityMentionType;
}

/**
 * Input for updating a chapter-entity link.
 */
export interface UpdateChapterEntityInput {
    mentionType?: ChapterEntityMentionType;
}

/**
 * Chapter Entity API service.
 */
export const chapterEntitiesApi = {
    /**
     * List all entity links for a chapter.
     */
    list(campaignId: number, chapterId: number): Promise<ChapterEntity[]> {
        return apiClient.get<ChapterEntity[]>(
            `/campaigns/${campaignId}/chapters/${chapterId}/entities`
        );
    },

    /**
     * Create a new chapter-entity link.
     */
    create(
        campaignId: number,
        chapterId: number,
        input: CreateChapterEntityInput
    ): Promise<ChapterEntity> {
        return apiClient.post<ChapterEntity>(
            `/campaigns/${campaignId}/chapters/${chapterId}/entities`,
            input
        );
    },

    /**
     * Update an existing chapter-entity link.
     */
    update(
        campaignId: number,
        chapterId: number,
        linkId: number,
        input: UpdateChapterEntityInput
    ): Promise<ChapterEntity> {
        return apiClient.put<ChapterEntity>(
            `/campaigns/${campaignId}/chapters/${chapterId}/entities/${linkId}`,
            input
        );
    },

    /**
     * Delete a chapter-entity link.
     */
    delete(
        campaignId: number,
        chapterId: number,
        linkId: number
    ): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/chapters/${chapterId}/entities/${linkId}`
        );
    },

    /**
     * List relationships involving entities linked to a chapter.
     */
    listRelationships(
        campaignId: number,
        chapterId: number,
    ): Promise<Relationship[]> {
        return apiClient.get<Relationship[]>(
            `/campaigns/${campaignId}/chapters/${chapterId}/relationships`,
        );
    },
};

export default chapterEntitiesApi;
