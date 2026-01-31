// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * API module exports - centralizes all API services for easy importing.
 */

export { apiClient, ApiError } from './client';
export type { PaginatedResponse } from './client';

export { campaignsApi } from './campaigns';
export type {
    ListCampaignsParams,
    CreateCampaignInput,
    UpdateCampaignInput,
} from './campaigns';

export { entitiesApi } from './entities';
export type {
    ListEntitiesParams,
    CreateEntityInput,
    UpdateEntityInput,
} from './entities';

export { relationshipsApi } from './relationships';
export type {
    ListRelationshipsParams,
    CreateRelationshipInput,
    UpdateRelationshipInput,
} from './relationships';

export { timelineApi } from './timeline';
export type {
    ListTimelineEventsParams,
    CreateTimelineEventInput,
    UpdateTimelineEventInput,
} from './timeline';

export { gameSystemsApi } from './gameSystems';

export { statsApi } from './stats';
export type { DashboardStats, CampaignStats } from './stats';

export { importApi } from './import';
export type { GoogleDocsImportInput, FileUploadImportInput } from './import';
