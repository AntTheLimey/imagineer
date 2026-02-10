// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Content Analysis API service - triggers post-save content analysis
 * and manages resolution of detected items (untagged mentions,
 * unresolved wiki links, misspellings).
 */

import { apiClient } from './client';
import type { EntityType } from '../types';

/**
 * A content analysis job tracking the analysis of a single content field.
 */
export interface ContentAnalysisJob {
    id: number;
    campaignId: number;
    sourceTable: string;
    sourceId: number;
    sourceField: string;
    status: string;
    totalItems: number;
    resolvedItems: number;
    enrichmentTotal: number;
    enrichmentResolved: number;
    createdAt: string;
    updatedAt: string;
}

/**
 * A single item detected during content analysis.
 */
export interface ContentAnalysisItem {
    id: number;
    jobId: number;
    detectionType:
        | 'wiki_link_resolved'
        | 'wiki_link_unresolved'
        | 'untagged_mention'
        | 'potential_alias'
        | 'misspelling'
        | 'description_update'
        | 'log_entry'
        | 'relationship_suggestion'
        | 'new_entity_suggestion';
    matchedText: string;
    entityId?: number;
    similarity?: number;
    contextSnippet?: string;
    positionStart?: number;
    positionEnd?: number;
    resolution: 'pending' | 'accepted' | 'new_entity' | 'dismissed';
    resolvedEntityId?: number;
    resolvedAt?: string;
    suggestedContent?: Record<string, unknown>;
    phase: 'identification' | 'enrichment';
    createdAt: string;
    entityName?: string;
    entityType?: EntityType;
}

/**
 * Summary of pending analysis items for a job.
 */
export interface AnalysisSummary {
    jobId: number;
    pendingCount: number;
}

/**
 * Request payload for resolving a single analysis item.
 */
export interface ResolveAnalysisItemRequest {
    resolution: 'accepted' | 'new_entity' | 'dismissed';
    entityType?: EntityType;
    entityName?: string;
    /** Override for the relationship type in a relationship_suggestion. */
    suggestedContentOverride?: Record<string, unknown>;
}

/**
 * Request payload for triggering content analysis on a content field.
 */
export interface TriggerAnalysisRequest {
    sourceTable: string;
    sourceId: number;
    sourceField: string;
}

/**
 * Response from triggering content analysis.
 */
export interface TriggerAnalysisResponse {
    job: ContentAnalysisJob;
    items: ContentAnalysisItem[];
}

/**
 * Response containing a count of pending analysis items.
 */
export interface PendingCountResponse {
    count: number;
}

/**
 * Content Analysis API service.
 */
export const contentAnalysisApi = {
    /**
     * List all analysis jobs for a campaign.
     */
    listJobs(campaignId: number): Promise<ContentAnalysisJob[]> {
        return apiClient.get<ContentAnalysisJob[]>(
            `/campaigns/${campaignId}/analysis/jobs`
        );
    },

    /**
     * Get a single analysis job by ID.
     */
    getJob(campaignId: number, jobId: number): Promise<ContentAnalysisJob> {
        return apiClient.get<ContentAnalysisJob>(
            `/campaigns/${campaignId}/analysis/jobs/${jobId}`
        );
    },

    /**
     * List items for an analysis job, optionally filtered by resolution
     * status.
     */
    listJobItems(campaignId: number, jobId: number, resolution?: string): Promise<ContentAnalysisItem[]> {
        return apiClient.get<ContentAnalysisItem[]>(
            `/campaigns/${campaignId}/analysis/jobs/${jobId}/items`,
            resolution ? { resolution } : undefined
        );
    },

    /**
     * Resolve a single analysis item (accept, create new entity, or
     * dismiss).
     */
    resolveItem(campaignId: number, itemId: number, req: ResolveAnalysisItemRequest): Promise<void> {
        return apiClient.put<void>(
            `/campaigns/${campaignId}/analysis/items/${itemId}`,
            req
        );
    },

    /**
     * Trigger content analysis for a specific content field.
     */
    triggerAnalysis(campaignId: number, req: TriggerAnalysisRequest): Promise<TriggerAnalysisResponse> {
        return apiClient.post<TriggerAnalysisResponse>(
            `/campaigns/${campaignId}/analysis/trigger`,
            req
        );
    },

    /**
     * Batch-resolve all pending items of a given detection type within a
     * job.
     */
    batchResolve(
        campaignId: number,
        jobId: number,
        req: { detectionType: string; resolution: string },
    ): Promise<{ resolved: number }> {
        return apiClient.put<{ resolved: number }>(
            `/campaigns/${campaignId}/analysis/jobs/${jobId}/resolve-all`,
            req,
        );
    },

    /**
     * Revert a previously resolved analysis item back to pending.
     */
    revertItem(
        campaignId: number,
        itemId: number,
    ): Promise<{ status: string }> {
        return apiClient.put<{ status: string }>(
            `/campaigns/${campaignId}/analysis/items/${itemId}/revert`,
        );
    },

    /**
     * Trigger LLM enrichment for a completed analysis job. The backend
     * generates description updates, log entries, and relationship
     * suggestions for entities detected in the analysed content.
     */
    triggerEnrichment(
        campaignId: number,
        jobId: number,
    ): Promise<{ status: string; entityCount?: number; message?: string }> {
        return apiClient.post<{
            status: string;
            entityCount?: number;
            message?: string;
        }>(
            `/campaigns/${campaignId}/analysis/jobs/${jobId}/enrich`
        );
    },

    /**
     * Cancel a running LLM enrichment for a job.
     */
    cancelEnrichment(
        campaignId: number,
        jobId: number,
    ): Promise<{ status: string }> {
        return apiClient.post<{ status: string }>(
            `/campaigns/${campaignId}/analysis/jobs/${jobId}/cancel-enrichment`
        );
    },

    /**
     * Get the count of pending (unresolved) analysis items, optionally
     * scoped to a specific source.
     */
    getPendingCount(campaignId: number, sourceTable?: string, sourceId?: number): Promise<PendingCountResponse> {
        const params: Record<string, string> = {};
        if (sourceTable) params.sourceTable = sourceTable;
        if (sourceId !== undefined) params.sourceId = String(sourceId);
        return apiClient.get<PendingCountResponse>(
            `/campaigns/${campaignId}/analysis/pending-count`,
            params
        );
    },
};

export default contentAnalysisApi;
