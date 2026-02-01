// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hooks module exports - centralizes all custom hooks for easy importing.
 */

// Campaign hooks
export {
    useCampaigns,
    useCampaign,
    useCreateCampaign,
    useUpdateCampaign,
    useDeleteCampaign,
    campaignKeys,
} from './useCampaigns';

// Entity hooks
export {
    useEntities,
    useEntity,
    useSimilarEntities,
    useCreateEntity,
    useUpdateEntity,
    useDeleteEntity,
    entityKeys,
} from './useEntities';

// Relationship hooks
export {
    useRelationships,
    useRelationship,
    useEntityRelationships,
    useCreateRelationship,
    useUpdateRelationship,
    useDeleteRelationship,
    relationshipKeys,
} from './useRelationships';

// Timeline hooks
export {
    useTimelineEvents,
    useTimelineEvent,
    useEntityTimeline,
    useCreateTimelineEvent,
    useUpdateTimelineEvent,
    useDeleteTimelineEvent,
    timelineKeys,
} from './useTimeline';

// Game system hooks
export {
    useGameSystems,
    useGameSystem,
    useGameSystemByCode,
    gameSystemKeys,
} from './useGameSystems';

// Stats hooks
export {
    useDashboardStats,
    useCampaignStats,
    statsKeys,
} from './useStats';

// Import hooks
export {
    useImportEvernote,
    useImportGoogleDocs,
    useImportFile,
    useImportFiles,
} from './useImport';

// Ownership hooks
export {
    useCampaignOwnership,
    isCampaignOwner,
} from './useCampaignOwnership';
export type { CampaignOwnershipResult } from './useCampaignOwnership';

// Draft management hooks
export { useDraft } from './useDraft';
export type { DraftData } from './useDraft';

// Autosave hooks
export { useAutosave } from './useAutosave';

// Unsaved changes hooks
export { useUnsavedChanges } from './useUnsavedChanges';

// Relationship type hooks
export {
    useRelationshipTypes,
    useCreateRelationshipType,
    useDeleteRelationshipType,
    relationshipTypeKeys,
} from './useRelationshipTypes';
