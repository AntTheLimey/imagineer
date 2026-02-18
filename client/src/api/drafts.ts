// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Drafts API service - server-side draft persistence for editor pages.
 */

import { apiClient } from './client';
import { getStoredToken } from '../contexts/authUtils';

/**
 * A server-persisted draft of an editor form.
 */
export interface Draft {
    id: number;
    campaignId: number;
    userId: number;
    sourceTable: string;
    sourceId: number;
    isNew: boolean;
    draftData: Record<string, unknown>;
    serverVersion?: number;
    createdAt: string;
    updatedAt: string;
}

/**
 * Lightweight indicator showing that a draft exists for a given source.
 */
export interface DraftIndicator {
    sourceTable: string;
    sourceId: number;
    isNew: boolean;
    updatedAt: string;
}

/**
 * Input for saving a draft to the server.
 */
export interface SaveDraftInput {
    sourceTable: string;
    sourceId: number;
    isNew: boolean;
    draftData: Record<string, unknown>;
    serverVersion?: number;
}

/**
 * Drafts API service.
 */
export const draftsApi = {
    /**
     * Save or update a draft for the given campaign.
     */
    saveDraft(campaignId: number, input: SaveDraftInput): Promise<Draft> {
        return apiClient.put<Draft>(
            `/campaigns/${campaignId}/drafts`,
            input
        );
    },

    /**
     * Get a single draft by source table and source ID.
     */
    getDraft(
        campaignId: number,
        sourceTable: string,
        sourceId: number
    ): Promise<Draft> {
        return apiClient.get<Draft>(
            `/campaigns/${campaignId}/drafts/${sourceTable}/${sourceId}`
        );
    },

    /**
     * Delete a draft by source table and source ID.
     */
    deleteDraft(
        campaignId: number,
        sourceTable: string,
        sourceId: number
    ): Promise<void> {
        return apiClient.delete<void>(
            `/campaigns/${campaignId}/drafts/${sourceTable}/${sourceId}`
        );
    },

    /**
     * List draft indicators for a campaign, optionally filtered by source table.
     */
    listIndicators(
        campaignId: number,
        sourceTable?: string
    ): Promise<DraftIndicator[]> {
        return apiClient.get<DraftIndicator[]>(
            `/campaigns/${campaignId}/drafts`,
            sourceTable ? { source_table: sourceTable } : undefined
        );
    },

    /**
     * Save a draft using fetch with keepalive for reliable delivery during
     * page unload. Since navigator.sendBeacon cannot set Authorization
     * headers, this uses fetch with keepalive: true instead.
     */
    beaconSave(campaignId: number, input: SaveDraftInput): void {
        const token = getStoredToken();
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const body = JSON.stringify(input);

        try {
            fetch(`/api/campaigns/${campaignId}/drafts/beacon`, {
                method: 'POST',
                headers,
                body,
                keepalive: true,
            });
        } catch {
            // Best-effort: beacon saves are fire-and-forget
            console.warn('Failed to send beacon draft save');
        }
    },
};

export default draftsApi;
