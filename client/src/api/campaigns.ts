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
import type { Campaign } from '../types';

/**
 * Parameters for listing campaigns.
 */
export interface ListCampaignsParams {
    page?: number;
    pageSize?: number;
    systemId?: string;
}

/**
 * Input for creating a new campaign.
 */
export interface CreateCampaignInput {
    name: string;
    systemId: string;
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
    get(id: string): Promise<Campaign> {
        return apiClient.get<Campaign>(`/campaigns/${id}`);
    },

    /**
     * Create a new campaign.
     */
    create(input: CreateCampaignInput): Promise<Campaign> {
        return apiClient.post<Campaign>('/campaigns', input);
    },

    /**
     * Update an existing campaign.
     */
    update(id: string, input: UpdateCampaignInput): Promise<Campaign> {
        return apiClient.put<Campaign>(`/campaigns/${id}`, input);
    },

    /**
     * Delete a campaign.
     */
    delete(id: string): Promise<void> {
        return apiClient.delete<void>(`/campaigns/${id}`);
    },
};

export default campaignsApi;
