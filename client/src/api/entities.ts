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
import type { Entity, EntityType, SourceConfidence } from '../types';

/**
 * Parameters for listing entities.
 */
export interface ListEntitiesParams {
    campaignId: string;
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
    campaignId: string;
    entityType: EntityType;
    name: string;
    description?: string;
    attributes?: Record<string, unknown>;
    tags?: string[];
    gmNotes?: string;
    discoveredSession?: string;
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
    discoveredSession?: string;
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
        const { campaignId, tags, ...rest } = params;
        const queryParams: Record<string, string | number | boolean | undefined> = {
            ...rest,
        };
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
    get(_campaignId: string, entityId: string): Promise<Entity> {
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
    update(_campaignId: string, entityId: string, input: UpdateEntityInput): Promise<Entity> {
        return apiClient.put<Entity>(`/entities/${entityId}`, input);
    },

    /**
     * Delete an entity.
     * Note: campaignId is kept for API consistency but not used in the URL.
     * The server verifies campaign ownership via the entity's campaign_id.
     */
    delete(_campaignId: string, entityId: string): Promise<void> {
        return apiClient.delete<void>(`/entities/${entityId}`);
    },

    /**
     * Search for similar entities by name (for duplicate detection).
     */
    searchSimilar(campaignId: string, name: string): Promise<Entity[]> {
        return apiClient.get<Entity[]>(`/campaigns/${campaignId}/entities/search`, {
            name,
        });
    },
};

export default entitiesApi;
