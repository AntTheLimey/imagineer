// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Relationship API service - operations for entity relationships.
 */

import { apiClient } from './client';
import type { Relationship, RelationshipTone } from '../types';

/**
 * Parameters for listing relationships.
 */
export interface ListRelationshipsParams {
    campaignId: string;
    page?: number;
    pageSize?: number;
    entityId?: string; // Filter by source or target entity
    relationshipType?: string;
}

/**
 * Input for creating a new relationship.
 */
export interface CreateRelationshipInput {
    campaignId: string;
    sourceEntityId: string;
    targetEntityId: string;
    relationshipType: string;
    tone?: RelationshipTone;
    description?: string;
    bidirectional?: boolean;
    strength?: number;
}

/**
 * Input for updating an existing relationship.
 */
export interface UpdateRelationshipInput {
    relationshipType?: string;
    tone?: RelationshipTone;
    description?: string;
    bidirectional?: boolean;
    strength?: number;
}

/**
 * Relationship API service.
 */
export const relationshipsApi = {
    /**
     * List relationships for a campaign with optional filtering.
     */
    list(params: ListRelationshipsParams): Promise<Relationship[]> {
        const { campaignId, ...rest } = params;
        return apiClient.get<Relationship[]>(
            `/campaigns/${campaignId}/relationships`,
            rest
        );
    },

    /**
     * Get a single relationship by ID.
     */
    get(campaignId: string, relationshipId: string): Promise<Relationship> {
        return apiClient.get<Relationship>(
            `/campaigns/${campaignId}/relationships/${relationshipId}`
        );
    },

    /**
     * Create a new relationship between entities.
     */
    create(input: CreateRelationshipInput): Promise<Relationship> {
        const { campaignId, ...body } = input;
        return apiClient.post<Relationship>(
            `/campaigns/${campaignId}/relationships`,
            body
        );
    },

    /**
     * Update an existing relationship.
     */
    update(
        campaignId: string,
        relationshipId: string,
        input: UpdateRelationshipInput
    ): Promise<Relationship> {
        return apiClient.put<Relationship>(
            `/campaigns/${campaignId}/relationships/${relationshipId}`,
            input
        );
    },

    /**
     * Delete a relationship.
     */
    delete(campaignId: string, relationshipId: string): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/relationships/${relationshipId}`
        );
    },

    /**
     * Get all relationships for a specific entity.
     */
    getForEntity(campaignId: string, entityId: string): Promise<Relationship[]> {
        return apiClient.get<Relationship[]>(
            `/campaigns/${campaignId}/entities/${entityId}/relationships`
        );
    },
};

export default relationshipsApi;
