// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Entity Detection API service - detects existing entities in text segments.
 */

import { apiClient } from './client';
import type { Entity } from '../types';

/**
 * Request payload for detecting entities in text.
 */
export interface DetectEntitiesRequest {
    textSegments: string[];
    excludeEntityIds?: number[];
}

/**
 * A single entity suggestion with match information.
 */
export interface EntitySuggestion {
    text: string;
    entity: Entity;
    similarity: number;
}

/**
 * Response from the entity detection endpoint.
 */
export interface DetectEntitiesResponse {
    suggestions: EntitySuggestion[];
    configured: boolean;
}

/**
 * Entity Detection API service.
 */
export const entityDetectionApi = {
    /**
     * Detect entities in text segments for a campaign.
     * Returns suggestions for existing entities that match the provided text.
     */
    detect(
        campaignId: number,
        request: DetectEntitiesRequest
    ): Promise<DetectEntitiesResponse> {
        return apiClient.post<DetectEntitiesResponse>(
            `/campaigns/${campaignId}/chapters/detect-entities`,
            request
        );
    },
};

export default entityDetectionApi;
