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

export { relationshipTypesApi } from './relationshipTypes';

export { userSettingsApi } from './userSettings';
export type {
    ContentGenService,
    EmbeddingService,
    ImageGenService,
    UserSettingsResponse,
    UserSettingsUpdateRequest,
} from './userSettings';

export { evernoteImportApi } from './evernoteImport';
export type {
    EvernoteStatus,
    Notebook,
    NoteSummary,
    EvernoteLocalImportRequest,
} from './evernoteImport';

export { chaptersApi } from './chapters';
export type { CreateChapterInput, UpdateChapterInput } from './chapters';

export { sessionsApi } from './sessions';
export type {
    ListSessionsParams,
    CreateSessionInput,
    UpdateSessionInput,
} from './sessions';

export { entityDetectionApi } from './entityDetection';
export type {
    DetectEntitiesRequest,
    EntitySuggestion,
    DetectEntitiesResponse,
} from './entityDetection';

export { entityResolveApi } from './entityResolve';
export type { EntityResolveResult } from './entityResolve';

export { contentAnalysisApi } from './contentAnalysis';
export type {
    ContentAnalysisJob,
    ContentAnalysisItem,
    AnalysisSummary,
    ResolveAnalysisItemRequest,
    TriggerAnalysisRequest,
    TriggerAnalysisResponse,
    PendingCountResponse,
} from './contentAnalysis';

export { entityLogApi } from './entityLog';
export type {
    EntityLog,
    CreateEntityLogInput,
    UpdateEntityLogInput,
} from './entityLog';

export { draftsApi } from './drafts';
export type {
    Draft,
    DraftIndicator,
    SaveDraftInput,
} from './drafts';
