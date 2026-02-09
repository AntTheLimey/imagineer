// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Entity Log API service - manages log entries for campaign entities,
 * tracking events and observations tied to chapters, sessions, or
 * other sources.
 */

import { apiClient } from './client';

/**
 * A single log entry for a campaign entity.
 */
export interface EntityLog {
    id: number;
    entityId: number;
    campaignId: number;
    chapterId?: number;
    sessionId?: number;
    sourceTable?: string;
    sourceId?: number;
    content: string;
    occurredAt?: string;
    sortOrder?: number;
    createdAt: string;
}

/**
 * Request payload for creating a new entity log entry.
 */
export interface CreateEntityLogInput {
    content: string;
    chapterId?: number;
    sessionId?: number;
    occurredAt?: string;
    sortOrder?: number;
}

/**
 * Request payload for updating an existing entity log entry.
 */
export interface UpdateEntityLogInput {
    content?: string;
    chapterId?: number;
    sessionId?: number;
    occurredAt?: string;
    sortOrder?: number;
}

/**
 * Entity Log API service.
 */
export const entityLogApi = {
    /**
     * List all log entries for a specific entity within a campaign.
     */
    list(campaignId: number, entityId: number): Promise<EntityLog[]> {
        return apiClient.get<EntityLog[]>(
            `/campaigns/${campaignId}/entities/${entityId}/log`
        );
    },

    /**
     * Create a new log entry for an entity.
     */
    create(
        campaignId: number,
        entityId: number,
        input: CreateEntityLogInput,
    ): Promise<EntityLog> {
        return apiClient.post<EntityLog>(
            `/campaigns/${campaignId}/entities/${entityId}/log`,
            input
        );
    },

    /**
     * Update an existing entity log entry.
     */
    update(
        campaignId: number,
        entityId: number,
        logId: number,
        input: UpdateEntityLogInput,
    ): Promise<EntityLog> {
        return apiClient.put<EntityLog>(
            `/campaigns/${campaignId}/entities/${entityId}/log/${logId}`,
            input
        );
    },

    /**
     * Delete an entity log entry.
     */
    delete(
        campaignId: number,
        entityId: number,
        logId: number,
    ): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/entities/${entityId}/log/${logId}`
        );
    },
};

export default entityLogApi;
