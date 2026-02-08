// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Entity Resolve API service - resolves entity names to matching entities.
 */

import { apiClient } from './client';
import type { EntityType } from '../types';

/**
 * A single entity resolve result with similarity score.
 */
export interface EntityResolveResult {
    id: number;
    name: string;
    entityType: EntityType;
    similarity: number;
}

/**
 * Entity Resolve API service.
 */
export const entityResolveApi = {
    /**
     * Resolve an entity name to matching entities for a campaign.
     * Returns entities that match the provided name, sorted by similarity.
     */
    resolve(campaignId: number, name: string): Promise<EntityResolveResult[]> {
        return apiClient.get<EntityResolveResult[]>(
            `/campaigns/${campaignId}/entities/resolve`,
            { name }
        );
    },
};

export default entityResolveApi;
