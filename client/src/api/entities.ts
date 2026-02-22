// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Entity API service - CRUD operations for entities.
 */

import { apiClient } from './client';
import type { AnalysisOptions } from './types';
import type { Entity, EntityType, SourceConfidence } from '../types';

/**
 * Parameters for listing entities.
 */
export interface ListEntitiesParams {
    campaignId: number;
    page?: number;
    pageSize?: number;
    entityType?: EntityType;
    tags?: string[];
    searchTerm?: string;
}

/**
 * Input for creating a new entity.
 */
export interface CreateEntityInput {
    campaignId: number;
    entityType: EntityType;
    name: string;
    description?: string;
    attributes?: Record<string, unknown>;
    tags?: string[];
    gmNotes?: string;
    discoveredSession?: number;
    sourceDocument?: string;
    sourceConfidence?: SourceConfidence;
}

/**
 * Input for updating an existing entity.
 */
export interface UpdateEntityInput {
    name?: string;
    description?: string;
    attributes?: Record<string, unknown>;
    tags?: string[];
    gmNotes?: string;
    discoveredSession?: number;
    sourceDocument?: string;
    sourceConfidence?: SourceConfidence;
}

/**
 * Entity API service.
 */
export const entitiesApi = {
    /**
     * List entities for a campaign with optional filtering.
     */
    list(params: ListEntitiesParams): Promise<Entity[]> {
        const { campaignId, entityType, tags, ...rest } = params;
        const queryParams: Record<string, string | number | boolean | undefined> = {
            ...rest,
        };
        // Backend expects 'type' query parameter for entity type filter
        if (entityType) {
            queryParams.type = entityType;
        }
        // Convert tags array to comma-separated string for query param
        if (tags && tags.length > 0) {
            queryParams.tags = tags.join(',');
        }
        return apiClient.get<Entity[]>(
            `/campaigns/${campaignId}/entities`,
            queryParams
        );
    },

    /**
     * Get a single entity by ID.
     * Note: campaignId is kept for API consistency but not used in the URL.
     * The server verifies campaign ownership via the entity's campaign_id.
     */
    get(_campaignId: number, entityId: number): Promise<Entity> {
        return apiClient.get<Entity>(`/entities/${entityId}`);
    },

    /**
     * Create a new entity.
     */
    create(input: CreateEntityInput): Promise<Entity> {
        const { campaignId, ...body } = input;
        return apiClient.post<Entity>(`/campaigns/${campaignId}/entities`, body);
    },

    /**
     * Update an existing entity.
     * Note: campaignId is kept for API consistency but not used in the URL.
     * The server verifies campaign ownership via the entity's campaign_id.
     */
    update(
        _campaignId: number,
        entityId: number,
        input: UpdateEntityInput,
        options?: AnalysisOptions,
    ): Promise<Entity> {
        const params: Record<string, string> = {};
        if (options?.analyze) params.analyze = 'true';
        if (options?.enrich) params.enrich = 'true';
        if (options?.phases?.length)
            params.phases = options.phases.join(',');
        return apiClient.put<Entity>(
            `/entities/${entityId}`, input, params,
        );
    },

    /**
     * Delete an entity.
     * Note: campaignId is kept for API consistency but not used in the URL.
     * The server verifies campaign ownership via the entity's campaign_id.
     */
    delete(_campaignId: number, entityId: number): Promise<void> {
        return apiClient.delete<void>(`/entities/${entityId}`);
    },

    /**
     * Search for similar entities by name (for duplicate detection).
     */
    searchSimilar(campaignId: number, name: string): Promise<Entity[]> {
        return apiClient.get<Entity[]>(`/campaigns/${campaignId}/entities/search`, {
            name,
        });
    },
};

export default entitiesApi;
