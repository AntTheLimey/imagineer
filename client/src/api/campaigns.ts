// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Campaign API service - CRUD operations for campaigns.
 */

import { apiClient } from './client';
import type { AnalysisOptions } from './types';
import type { Campaign } from '../types';

/**
 * Parameters for listing campaigns.
 */
export interface ListCampaignsParams {
    page?: number;
    pageSize?: number;
    systemId?: number;
}

/**
 * Input for creating a new campaign.
 */
export interface CreateCampaignInput {
    name: string;
    systemId: number;
    description?: string;
    settings?: Record<string, unknown>;
}

/**
 * Input for updating an existing campaign.
 */
export interface UpdateCampaignInput {
    name?: string;
    description?: string;
    settings?: Record<string, unknown>;
}

/**
 * Campaign API service.
 */
export const campaignsApi = {
    /**
     * List all campaigns with optional filtering.
     */
    list(params?: ListCampaignsParams): Promise<Campaign[]> {
        return apiClient.get<Campaign[]>('/campaigns', params ? { ...params } : undefined);
    },

    /**
     * Get a single campaign by ID.
     */
    get(id: number): Promise<Campaign> {
        return apiClient.get<Campaign>(`/campaigns/${id}`);
    },

    /**
     * Create a new campaign.
     */
    create(input: CreateCampaignInput, options?: AnalysisOptions): Promise<Campaign> {
        const params: Record<string, string> = {};
        if (options?.analyze) params.analyze = 'true';
        if (options?.enrich) params.enrich = 'true';
        if (options?.phases?.length)
            params.phases = options.phases.join(',');
        return apiClient.post<Campaign>('/campaigns', input, params);
    },

    /**
     * Update an existing campaign.
     */
    update(
        id: number,
        input: UpdateCampaignInput,
        options?: AnalysisOptions,
    ): Promise<Campaign> {
        const params: Record<string, string> = {};
        if (options?.analyze) params.analyze = 'true';
        if (options?.enrich) params.enrich = 'true';
        if (options?.phases?.length)
            params.phases = options.phases.join(',');
        return apiClient.put<Campaign>(
            `/campaigns/${id}`, input, params,
        );
    },

    /**
     * Delete a campaign.
     */
    delete(id: number): Promise<void> {
        return apiClient.delete<void>(`/campaigns/${id}`);
    },
};

export default campaignsApi;
