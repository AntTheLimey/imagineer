// Game System types
export interface GameSystem {
    id: string;
    name: string;
    code: string;
    attributeSchema: Record<string, unknown>;
    skillSchema: Record<string, unknown>;
    diceConventions: Record<string, unknown>;
    createdAt: string;
}

// Campaign types
export interface Campaign {
    id: string;
    name: string;
    systemId: string;
    ownerId?: string;
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
    id: string;
    campaignId: string;
    entityType: EntityType;
    name: string;
    description?: string;
    attributes: Record<string, unknown>;
    tags: string[];
    gmNotes?: string;
    discoveredSession?: string;
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
    id: string;
    campaignId: string;
    sourceEntityId: string;
    targetEntityId: string;
    relationshipType: string;
    tone?: RelationshipTone;
    description?: string;
    bidirectional: boolean;
    strength?: number;
    createdAt: string;
    updatedAt: string;
}

// Chapter types
export interface Chapter {
    id: string;
    campaignId: string;
    title: string;
    overview?: string;
    sortOrder: number;
    createdAt: string;
    updatedAt: string;
}

// Session types
export type SessionStage = 'prep' | 'play' | 'wrap_up';

export type SessionStatus = 'PLANNED' | 'COMPLETED' | 'SKIPPED';

export interface Session {
    id: string;
    campaignId: string;
    chapterId?: string;
    title?: string;
    sessionNumber?: number;
    plannedDate?: string;
    actualDate?: string;
    status: SessionStatus;
    stage: SessionStage;
    prepNotes?: string;
    plannedScenes?: unknown[];
    actualNotes?: string;
    discoveries?: Array<{ entityId: string; howDiscovered: string }>;
    playerDecisions?: unknown[];
    consequences?: unknown[];
    createdAt: string;
    updatedAt: string;
}

// Timeline types
export type DatePrecision = 'exact' | 'approximate' | 'month' | 'year' | 'unknown';

export interface TimelineEvent {
    id: string;
    campaignId: string;
    eventDate?: string;
    eventTime?: string;
    datePrecision: DatePrecision;
    description: string;
    entityIds: string[];
    sessionId?: string;
    isPlayerKnown: boolean;
    sourceDocument?: string;
    createdAt: string;
    updatedAt: string;
}

// Canon conflict types
export type ConflictStatus = 'DETECTED' | 'ACKNOWLEDGED' | 'RESOLVED';

export interface CanonConflict {
    id: string;
    campaignId: string;
    entityId?: string;
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
    id: string;
    campaignId?: string | null; // null = system default
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
