/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package models defines the data structures for the Imagineer API.
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GameSystem represents a TTRPG system definition (e.g., Call of Cthulhu 7e).
type GameSystem struct {
	ID                     uuid.UUID       `json:"id"`
	Name                   string          `json:"name"`
	Code                   string          `json:"code"`
	AttributeSchema        json.RawMessage `json:"attributeSchema,omitempty"`
	SkillSchema            json.RawMessage `json:"skillSchema,omitempty"`
	CharacterSheetTemplate json.RawMessage `json:"characterSheetTemplate,omitempty"`
	DiceConventions        json.RawMessage `json:"diceConventions,omitempty"`
	CreatedAt              time.Time       `json:"createdAt"`
}

// Campaign represents an individual TTRPG campaign.
type Campaign struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	SystemID    *uuid.UUID      `json:"systemId,omitempty"`
	Description *string         `json:"description,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`

	// Joined fields (not in database)
	System *GameSystem `json:"system,omitempty"`
}

// CreateCampaignRequest represents the request body for creating a campaign.
type CreateCampaignRequest struct {
	Name        string          `json:"name"`
	SystemID    *uuid.UUID      `json:"systemId,omitempty"`
	Description *string         `json:"description,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

// UpdateCampaignRequest represents the request body for updating a campaign.
type UpdateCampaignRequest struct {
	Name        *string         `json:"name,omitempty"`
	SystemID    *uuid.UUID      `json:"systemId,omitempty"`
	Description *string         `json:"description,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

// EntityType represents the type of a campaign entity.
type EntityType string

const (
	EntityTypeNPC          EntityType = "npc"
	EntityTypeLocation     EntityType = "location"
	EntityTypeItem         EntityType = "item"
	EntityTypeFaction      EntityType = "faction"
	EntityTypeClue         EntityType = "clue"
	EntityTypeCreature     EntityType = "creature"
	EntityTypeOrganization EntityType = "organization"
	EntityTypeEvent        EntityType = "event"
	EntityTypeDocument     EntityType = "document"
	EntityTypeOther        EntityType = "other"
)

// SourceConfidence represents the canon status of an entity.
type SourceConfidence string

const (
	SourceConfidenceDraft         SourceConfidence = "DRAFT"
	SourceConfidenceAuthoritative SourceConfidence = "AUTHORITATIVE"
	SourceConfidenceSuperseded    SourceConfidence = "SUPERSEDED"
)

// Entity represents a polymorphic campaign entity (NPC, location, item, etc.).
type Entity struct {
	ID                uuid.UUID        `json:"id"`
	CampaignID        uuid.UUID        `json:"campaignId"`
	EntityType        EntityType       `json:"entityType"`
	Name              string           `json:"name"`
	Description       *string          `json:"description,omitempty"`
	Attributes        json.RawMessage  `json:"attributes,omitempty"`
	Tags              []string         `json:"tags,omitempty"`
	KeeperNotes       *string          `json:"keeperNotes,omitempty"`
	DiscoveredSession *uuid.UUID       `json:"discoveredSession,omitempty"`
	SourceDocument    *string          `json:"sourceDocument,omitempty"`
	SourceConfidence  SourceConfidence `json:"sourceConfidence"`
	Version           int              `json:"version"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

// CreateEntityRequest represents the request body for creating an entity.
type CreateEntityRequest struct {
	EntityType        EntityType        `json:"entityType"`
	Name              string            `json:"name"`
	Description       *string           `json:"description,omitempty"`
	Attributes        json.RawMessage   `json:"attributes,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	KeeperNotes       *string           `json:"keeperNotes,omitempty"`
	DiscoveredSession *uuid.UUID        `json:"discoveredSession,omitempty"`
	SourceDocument    *string           `json:"sourceDocument,omitempty"`
	SourceConfidence  *SourceConfidence `json:"sourceConfidence,omitempty"`
}

// UpdateEntityRequest represents the request body for updating an entity.
type UpdateEntityRequest struct {
	EntityType        *EntityType       `json:"entityType,omitempty"`
	Name              *string           `json:"name,omitempty"`
	Description       *string           `json:"description,omitempty"`
	Attributes        json.RawMessage   `json:"attributes,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	KeeperNotes       *string           `json:"keeperNotes,omitempty"`
	DiscoveredSession *uuid.UUID        `json:"discoveredSession,omitempty"`
	SourceDocument    *string           `json:"sourceDocument,omitempty"`
	SourceConfidence  *SourceConfidence `json:"sourceConfidence,omitempty"`
}

// RelationshipTone represents the emotional quality of a relationship.
type RelationshipTone string

const (
	RelationshipToneFriendly     RelationshipTone = "friendly"
	RelationshipToneHostile      RelationshipTone = "hostile"
	RelationshipToneNeutral      RelationshipTone = "neutral"
	RelationshipToneRomantic     RelationshipTone = "romantic"
	RelationshipToneProfessional RelationshipTone = "professional"
	RelationshipToneFearful      RelationshipTone = "fearful"
	RelationshipToneRespectful   RelationshipTone = "respectful"
	RelationshipToneUnknown      RelationshipTone = "unknown"
)

// Relationship represents a connection between two entities.
type Relationship struct {
	ID               uuid.UUID         `json:"id"`
	CampaignID       uuid.UUID         `json:"campaignId"`
	SourceEntityID   uuid.UUID         `json:"sourceEntityId"`
	TargetEntityID   uuid.UUID         `json:"targetEntityId"`
	RelationshipType string            `json:"relationshipType"`
	Tone             *RelationshipTone `json:"tone,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Bidirectional    bool              `json:"bidirectional"`
	Strength         *int              `json:"strength,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`

	// Joined fields (not in database)
	SourceEntity *Entity `json:"sourceEntity,omitempty"`
	TargetEntity *Entity `json:"targetEntity,omitempty"`
}

// CreateRelationshipRequest represents the request body for creating a relationship.
type CreateRelationshipRequest struct {
	SourceEntityID   uuid.UUID         `json:"sourceEntityId"`
	TargetEntityID   uuid.UUID         `json:"targetEntityId"`
	RelationshipType string            `json:"relationshipType"`
	Tone             *RelationshipTone `json:"tone,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Bidirectional    bool              `json:"bidirectional"`
	Strength         *int              `json:"strength,omitempty"`
}

// UpdateRelationshipRequest represents the request body for updating a relationship.
type UpdateRelationshipRequest struct {
	RelationshipType *string           `json:"relationshipType,omitempty"`
	Tone             *RelationshipTone `json:"tone,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Bidirectional    *bool             `json:"bidirectional,omitempty"`
	Strength         *int              `json:"strength,omitempty"`
}

// SessionStatus represents the status of a game session.
type SessionStatus string

const (
	SessionStatusPlanned   SessionStatus = "PLANNED"
	SessionStatusCompleted SessionStatus = "COMPLETED"
	SessionStatusSkipped   SessionStatus = "SKIPPED"
)

// Session represents a game session within a campaign.
type Session struct {
	ID              uuid.UUID       `json:"id"`
	CampaignID      uuid.UUID       `json:"campaignId"`
	SessionNumber   *int            `json:"sessionNumber,omitempty"`
	PlannedDate     *time.Time      `json:"plannedDate,omitempty"`
	ActualDate      *time.Time      `json:"actualDate,omitempty"`
	Status          SessionStatus   `json:"status"`
	PrepNotes       *string         `json:"prepNotes,omitempty"`
	PlannedScenes   json.RawMessage `json:"plannedScenes,omitempty"`
	ActualNotes     *string         `json:"actualNotes,omitempty"`
	Discoveries     json.RawMessage `json:"discoveries,omitempty"`
	PlayerDecisions json.RawMessage `json:"playerDecisions,omitempty"`
	Consequences    json.RawMessage `json:"consequences,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// DatePrecision represents the precision level of a date.
type DatePrecision string

const (
	DatePrecisionExact       DatePrecision = "exact"
	DatePrecisionApproximate DatePrecision = "approximate"
	DatePrecisionMonth       DatePrecision = "month"
	DatePrecisionYear        DatePrecision = "year"
	DatePrecisionUnknown     DatePrecision = "unknown"
)

// TimelineEvent represents an in-game chronological event.
type TimelineEvent struct {
	ID             uuid.UUID     `json:"id"`
	CampaignID     uuid.UUID     `json:"campaignId"`
	EventDate      *time.Time    `json:"eventDate,omitempty"`
	EventTime      *string       `json:"eventTime,omitempty"`
	DatePrecision  DatePrecision `json:"datePrecision"`
	Description    string        `json:"description"`
	EntityIDs      []uuid.UUID   `json:"entityIds,omitempty"`
	SessionID      *uuid.UUID    `json:"sessionId,omitempty"`
	IsPlayerKnown  bool          `json:"isPlayerKnown"`
	SourceDocument *string       `json:"sourceDocument,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// CreateTimelineEventRequest represents the request body for creating a timeline event.
type CreateTimelineEventRequest struct {
	EventDate      *time.Time    `json:"eventDate,omitempty"`
	EventTime      *string       `json:"eventTime,omitempty"`
	DatePrecision  DatePrecision `json:"datePrecision"`
	Description    string        `json:"description"`
	EntityIDs      []uuid.UUID   `json:"entityIds,omitempty"`
	SessionID      *uuid.UUID    `json:"sessionId,omitempty"`
	IsPlayerKnown  bool          `json:"isPlayerKnown"`
	SourceDocument *string       `json:"sourceDocument,omitempty"`
}

// UpdateTimelineEventRequest represents the request body for updating a timeline event.
type UpdateTimelineEventRequest struct {
	EventDate      *time.Time     `json:"eventDate,omitempty"`
	EventTime      *string        `json:"eventTime,omitempty"`
	DatePrecision  *DatePrecision `json:"datePrecision,omitempty"`
	Description    *string        `json:"description,omitempty"`
	EntityIDs      []uuid.UUID    `json:"entityIds,omitempty"`
	SessionID      *uuid.UUID     `json:"sessionId,omitempty"`
	IsPlayerKnown  *bool          `json:"isPlayerKnown,omitempty"`
	SourceDocument *string        `json:"sourceDocument,omitempty"`
}

// ConflictStatus represents the status of a canon conflict.
type ConflictStatus string

const (
	ConflictStatusDetected     ConflictStatus = "DETECTED"
	ConflictStatusAcknowledged ConflictStatus = "ACKNOWLEDGED"
	ConflictStatusResolved     ConflictStatus = "RESOLVED"
)

// CanonConflict represents a contradiction between different sources.
type CanonConflict struct {
	ID                uuid.UUID       `json:"id"`
	CampaignID        uuid.UUID       `json:"campaignId"`
	EntityID          *uuid.UUID      `json:"entityId,omitempty"`
	FieldName         *string         `json:"fieldName,omitempty"`
	ConflictingValues json.RawMessage `json:"conflictingValues"`
	Status            ConflictStatus  `json:"status"`
	Resolution        *string         `json:"resolution,omitempty"`
	ResolvedAt        *time.Time      `json:"resolvedAt,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
}

// DashboardStats represents statistics for the dashboard.
type DashboardStats struct {
	TotalCampaigns     int            `json:"totalCampaigns"`
	TotalEntities      int            `json:"totalEntities"`
	TotalRelationships int            `json:"totalRelationships"`
	TotalSessions      int            `json:"totalSessions"`
	EntitiesByType     map[string]int `json:"entitiesByType"`
	RecentCampaigns    []Campaign     `json:"recentCampaigns"`
}

// FrontendDashboardStats represents statistics in the format expected by the frontend.
type FrontendDashboardStats struct {
	CampaignCount      int `json:"campaignCount"`
	NPCCount           int `json:"npcCount"`
	LocationCount      int `json:"locationCount"`
	TimelineEventCount int `json:"timelineEventCount"`
	ItemCount          int `json:"itemCount"`
	FactionCount       int `json:"factionCount"`
	TotalEntityCount   int `json:"totalEntityCount"`
}

// CampaignStats represents statistics for a specific campaign.
type CampaignStats struct {
	EntityCounts       map[string]int `json:"entityCounts"`
	RelationshipCount  int            `json:"relationshipCount"`
	TimelineEventCount int            `json:"timelineEventCount"`
	SessionCount       int            `json:"sessionCount"`
	ConflictCount      int            `json:"conflictCount"`
}

// APIError represents an error response from the API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ImportRequest represents a request to import content.
type ImportRequest struct {
	CampaignID           uuid.UUID `json:"campaignId"`
	GameSystemCode       string    `json:"gameSystemCode,omitempty"`
	SourceDocument       string    `json:"sourceDocument,omitempty"`
	AutoDetectEntities   bool      `json:"autoDetectEntities"`
	ExtractRelationships bool      `json:"extractRelationships"`
	ExtractEvents        bool      `json:"extractEvents"`
}
