// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Relationship Types API service - operations for relationship type definitions.
 */

import { apiClient } from './client';
import type { RelationshipType, CreateRelationshipTypeInput } from '../types';

/**
 * Relationship Types API service.
 */
export const relationshipTypesApi = {
    /**
     * List relationship types for a campaign.
     * Returns both system defaults and campaign-specific custom types.
     */
    list(campaignId: string): Promise<RelationshipType[]> {
        return apiClient.get<RelationshipType[]>(
            `/campaigns/${campaignId}/relationship-types`
        );
    },

    /**
     * Create a custom relationship type for a campaign.
     */
    create(
        campaignId: string,
        input: CreateRelationshipTypeInput
    ): Promise<RelationshipType> {
        return apiClient.post<RelationshipType>(
            `/campaigns/${campaignId}/relationship-types`,
            input
        );
    },

    /**
     * Delete a custom relationship type.
     * Only campaign-specific types can be deleted (not system defaults).
     */
    delete(campaignId: string, typeId: string): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/relationship-types/${typeId}`
        );
    },
};

export default relationshipTypesApi;
