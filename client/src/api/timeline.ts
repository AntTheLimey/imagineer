// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Timeline API service - operations for timeline events.
 */

import { apiClient } from './client';
import type { TimelineEvent, DatePrecision } from '../types';

/**
 * Parameters for listing timeline events.
 */
export interface ListTimelineEventsParams {
    campaignId: string;
    page?: number;
    pageSize?: number;
    isPlayerKnown?: boolean;
    startDate?: string;
    endDate?: string;
    entityId?: string;
}

/**
 * Input for creating a new timeline event.
 */
export interface CreateTimelineEventInput {
    campaignId: string;
    eventDate?: string;
    eventTime?: string;
    datePrecision: DatePrecision;
    description: string;
    entityIds?: string[];
    sessionId?: string;
    isPlayerKnown?: boolean;
    sourceDocument?: string;
}

/**
 * Input for updating an existing timeline event.
 */
export interface UpdateTimelineEventInput {
    eventDate?: string;
    eventTime?: string;
    datePrecision?: DatePrecision;
    description?: string;
    entityIds?: string[];
    sessionId?: string;
    isPlayerKnown?: boolean;
    sourceDocument?: string;
}

/**
 * Timeline API service.
 */
export const timelineApi = {
    /**
     * List timeline events for a campaign with optional filtering.
     */
    list(params: ListTimelineEventsParams): Promise<TimelineEvent[]> {
        const { campaignId, ...rest } = params;
        return apiClient.get<TimelineEvent[]>(
            `/campaigns/${campaignId}/timeline`,
            rest
        );
    },

    /**
     * Get a single timeline event by ID.
     */
    get(campaignId: string, eventId: string): Promise<TimelineEvent> {
        return apiClient.get<TimelineEvent>(
            `/campaigns/${campaignId}/timeline/${eventId}`
        );
    },

    /**
     * Create a new timeline event.
     */
    create(input: CreateTimelineEventInput): Promise<TimelineEvent> {
        const { campaignId, ...body } = input;
        return apiClient.post<TimelineEvent>(
            `/campaigns/${campaignId}/timeline`,
            body
        );
    },

    /**
     * Update an existing timeline event.
     */
    update(
        campaignId: string,
        eventId: string,
        input: UpdateTimelineEventInput
    ): Promise<TimelineEvent> {
        return apiClient.put<TimelineEvent>(
            `/campaigns/${campaignId}/timeline/${eventId}`,
            input
        );
    },

    /**
     * Delete a timeline event.
     */
    delete(campaignId: string, eventId: string): Promise<void> {
        return apiClient.delete<void>(`/campaigns/${campaignId}/timeline/${eventId}`);
    },

    /**
     * Get timeline events involving a specific entity.
     */
    getForEntity(campaignId: string, entityId: string): Promise<TimelineEvent[]> {
        return apiClient.get<TimelineEvent[]>(
            `/campaigns/${campaignId}/entities/${entityId}/timeline`
        );
    },
};

export default timelineApi;
