// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

// Game System types
export interface GameSystem {
    id: number;
    name: string;
    code: string;
    attributeSchema: Record<string, unknown>;
    skillSchema: Record<string, unknown>;
    diceConventions: Record<string, unknown>;
    createdAt: string;
}

// Campaign types
export interface Campaign {
    id: number;
    name: string;
    systemId: number;
    ownerId?: number;
    description?: string;
    settings: Record<string, unknown>;
    createdAt: string;
    updatedAt: string;
}

// Entity types
export type EntityType =
    | 'npc'
    | 'location'
    | 'item'
    | 'faction'
    | 'clue'
    | 'creature'
    | 'organization'
    | 'event'
    | 'document'
    | 'other';

export type SourceConfidence = 'DRAFT' | 'AUTHORITATIVE' | 'SUPERSEDED';

export interface Entity {
    id: number;
    campaignId: number;
    entityType: EntityType;
    name: string;
    description?: string;
    attributes: Record<string, unknown>;
    tags: string[];
    gmNotes?: string;
    discoveredSession?: number;
    sourceDocument?: string;
    sourceConfidence: SourceConfidence;
    version: number;
    createdAt: string;
    updatedAt: string;
}

// Relationship types
export type RelationshipTone =
    | 'friendly'
    | 'hostile'
    | 'neutral'
    | 'romantic'
    | 'professional'
    | 'fearful'
    | 'respectful'
    | 'unknown';

export interface Relationship {
    id: number;
    campaignId: number;
    sourceEntityId: number;
    targetEntityId: number;
    relationshipTypeId: number;
    relationshipType: string;
    tone?: RelationshipTone;
    description?: string;
    strength?: number;
    displayLabel?: string;
    direction?: 'forward' | 'inverse';
    sourceEntityName?: string;
    sourceEntityType?: string;
    targetEntityName?: string;
    targetEntityType?: string;
    createdAt: string;
    updatedAt: string;
}

// Chapter types
export interface Chapter {
    id: number;
    campaignId: number;
    title: string;
    overview?: string;
    sortOrder: number;
    createdAt: string;
    updatedAt: string;
}

// Chapter-Entity link types
export type ChapterEntityMentionType = 'linked' | 'mentioned' | 'featured';

export interface ChapterEntity {
    id: number;
    chapterId: number;
    entityId: number;
    mentionType: ChapterEntityMentionType;
    createdAt: string;
    entity?: Entity;
}

// Session types
export type SessionStage = 'prep' | 'play' | 'wrap_up' | 'completed';

export type SessionStatus = 'PLANNED' | 'COMPLETED' | 'SKIPPED';

export interface Session {
    id: number;
    campaignId: number;
    chapterId?: number;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    actualDate?: string;
    status: SessionStatus;
    stage: SessionStage;
    prepNotes?: string;
    actualNotes?: string;
    playNotes?: string;
    createdAt: string;
    updatedAt: string;
}

// Timeline types
export type DatePrecision = 'exact' | 'approximate' | 'month' | 'year' | 'unknown';

export interface TimelineEvent {
    id: number;
    campaignId: number;
    eventDate?: string;
    eventTime?: string;
    datePrecision: DatePrecision;
    description: string;
    entityIds: number[];
    sessionId?: number;
    isPlayerKnown: boolean;
    sourceDocument?: string;
    createdAt: string;
    updatedAt: string;
}

// Canon conflict types
export type ConflictStatus = 'DETECTED' | 'ACKNOWLEDGED' | 'RESOLVED';

export interface CanonConflict {
    id: number;
    campaignId: number;
    entityId?: number;
    fieldName?: string;
    conflictingValues: Array<{ value: string; source: string; date: string }>;
    status: ConflictStatus;
    resolution?: string;
    resolvedAt?: string;
    createdAt: string;
}

// Import types
export interface ImportResult {
    success: boolean;
    entitiesCreated: number;
    relationshipsCreated: number;
    errors: string[];
    warnings: string[];
}

// Relationship type definition types
export interface RelationshipType {
    id: number;
    campaignId: number;
    name: string;
    inverseName: string;
    isSymmetric: boolean;
    displayLabel: string;
    inverseDisplayLabel: string;
    description?: string;
    createdAt: string;
    updatedAt: string;
}

export interface CreateRelationshipTypeInput {
    name: string;
    inverseName: string;
    isSymmetric: boolean;
    displayLabel: string;
    inverseDisplayLabel: string;
    description?: string;
}
