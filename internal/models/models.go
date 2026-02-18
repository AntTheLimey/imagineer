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
)

// GameSystem represents a TTRPG system definition (e.g., Call of Cthulhu 7e).
type GameSystem struct {
	ID                     int64           `json:"id"`
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
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	SystemID         *int64          `json:"systemId,omitempty"`
	OwnerID          *int64          `json:"ownerId,omitempty"`
	Description      *string         `json:"description,omitempty"`
	Settings         json.RawMessage `json:"settings,omitempty"`
	Genre            *CampaignGenre  `json:"genre,omitempty"`
	ImageStylePrompt *string         `json:"imageStylePrompt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`

	// Joined fields (not in database)
	System *GameSystem `json:"system,omitempty"`
	Owner  *User       `json:"owner,omitempty"`
}

// CreateCampaignRequest represents the request body for creating a campaign.
type CreateCampaignRequest struct {
	Name             string          `json:"name"`
	SystemID         *int64          `json:"systemId,omitempty"`
	Description      *string         `json:"description,omitempty"`
	Settings         json.RawMessage `json:"settings,omitempty"`
	Genre            *CampaignGenre  `json:"genre,omitempty"`
	ImageStylePrompt *string         `json:"imageStylePrompt,omitempty"`
}

// UpdateCampaignRequest represents the request body for updating a campaign.
type UpdateCampaignRequest struct {
	Name             *string         `json:"name,omitempty"`
	SystemID         *int64          `json:"systemId,omitempty"`
	Description      *string         `json:"description,omitempty"`
	Settings         json.RawMessage `json:"settings,omitempty"`
	Genre            *CampaignGenre  `json:"genre,omitempty"`
	ImageStylePrompt *string         `json:"imageStylePrompt,omitempty"`
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
	ID                int64            `json:"id"`
	CampaignID        int64            `json:"campaignId"`
	EntityType        EntityType       `json:"entityType"`
	Name              string           `json:"name"`
	Description       *string          `json:"description,omitempty"`
	Attributes        json.RawMessage  `json:"attributes,omitempty"`
	Tags              []string         `json:"tags,omitempty"`
	GMNotes           *string          `json:"gmNotes,omitempty"`
	DiscoveredSession *int64           `json:"discoveredSession,omitempty"`
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
	GMNotes           *string           `json:"gmNotes,omitempty"`
	DiscoveredSession *int64            `json:"discoveredSession,omitempty"`
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
	GMNotes           *string           `json:"gmNotes,omitempty"`
	DiscoveredSession *int64            `json:"discoveredSession,omitempty"`
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
// A single row stores the canonical forward direction; the
// entity_relationships_view provides both forward and inverse
// perspectives.
type Relationship struct {
	ID                 int64             `json:"id"`
	CampaignID         int64             `json:"campaignId"`
	SourceEntityID     int64             `json:"sourceEntityId"`
	TargetEntityID     int64             `json:"targetEntityId"`
	RelationshipTypeID int64             `json:"relationshipTypeId"`
	Tone               *RelationshipTone `json:"tone,omitempty"`
	Description        *string           `json:"description,omitempty"`
	Strength           *int              `json:"strength,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`

	// Joined fields from entity_relationships_view
	RelationshipTypeName string `json:"relationshipType,omitempty"`
	DisplayLabel         string `json:"displayLabel,omitempty"`
	Direction            string `json:"direction,omitempty"`

	// Joined entity fields
	SourceEntity     *Entity `json:"sourceEntity,omitempty"`
	TargetEntity     *Entity `json:"targetEntity,omitempty"`
	SourceEntityName string  `json:"sourceEntityName,omitempty"`
	SourceEntityType string  `json:"sourceEntityType,omitempty"`
	TargetEntityName string  `json:"targetEntityName,omitempty"`
	TargetEntityType string  `json:"targetEntityType,omitempty"`
}

// CreateRelationshipRequest represents the request body for
// creating a relationship.
type CreateRelationshipRequest struct {
	SourceEntityID     int64             `json:"sourceEntityId"`
	TargetEntityID     int64             `json:"targetEntityId"`
	RelationshipTypeID int64             `json:"relationshipTypeId"`
	Tone               *RelationshipTone `json:"tone,omitempty"`
	Description        *string           `json:"description,omitempty"`
	Strength           *int              `json:"strength,omitempty"`
}

// UpdateRelationshipRequest represents the request body for
// updating a relationship.
type UpdateRelationshipRequest struct {
	RelationshipTypeID *int64            `json:"relationshipTypeId,omitempty"`
	Tone               *RelationshipTone `json:"tone,omitempty"`
	Description        *string           `json:"description,omitempty"`
	Strength           *int              `json:"strength,omitempty"`
}

// RelationshipType defines a relationship type with its inverse
// mapping. Each campaign has its own set of types, seeded from
// relationship_type_templates on creation.
type RelationshipType struct {
	ID                  int64     `json:"id"`
	CampaignID          int64     `json:"campaignId"`
	Name                string    `json:"name"`
	InverseName         string    `json:"inverseName"`
	IsSymmetric         bool      `json:"isSymmetric"`
	DisplayLabel        string    `json:"displayLabel"`
	InverseDisplayLabel string    `json:"inverseDisplayLabel"`
	Description         *string   `json:"description,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// CreateRelationshipTypeRequest is the request body for creating a relationship type.
type CreateRelationshipTypeRequest struct {
	Name                string  `json:"name"`
	InverseName         string  `json:"inverseName"`
	IsSymmetric         bool    `json:"isSymmetric"`
	DisplayLabel        string  `json:"displayLabel"`
	InverseDisplayLabel string  `json:"inverseDisplayLabel"`
	Description         *string `json:"description,omitempty"`
}

// SessionStatus represents the status of a game session.
type SessionStatus string

const (
	SessionStatusPlanned   SessionStatus = "PLANNED"
	SessionStatusCompleted SessionStatus = "COMPLETED"
	SessionStatusSkipped   SessionStatus = "SKIPPED"
)

// SessionStage represents the workflow stage of a session.
type SessionStage string

const (
	SessionStagePrep      SessionStage = "prep"
	SessionStagePlay      SessionStage = "play"
	SessionStageWrapUp    SessionStage = "wrap_up"
	SessionStageCompleted SessionStage = "completed"
)

// Session represents a game session within a campaign.
type Session struct {
	ID            int64         `json:"id"`
	CampaignID    int64         `json:"campaignId"`
	ChapterID     *int64        `json:"chapterId,omitempty"`
	Title         *string       `json:"title,omitempty"`
	SessionNumber *int          `json:"sessionNumber,omitempty"`
	PlannedDate   *time.Time    `json:"plannedDate,omitempty"`
	ActualDate    *time.Time    `json:"actualDate,omitempty"`
	Status        SessionStatus `json:"status"`
	Stage         SessionStage  `json:"stage"`
	PrepNotes     *string       `json:"prepNotes,omitempty"`
	ActualNotes   *string       `json:"actualNotes,omitempty"`
	PlayNotes     *string       `json:"playNotes,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

// Chapter represents a story arc or grouping of sessions within a campaign.
type Chapter struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaignId"`
	Title      string    `json:"title"`
	Overview   *string   `json:"overview,omitempty"`
	SortOrder  int       `json:"sortOrder"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// CreateChapterRequest represents the request body for creating a chapter.
type CreateChapterRequest struct {
	Title     string  `json:"title"`
	Overview  *string `json:"overview,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
}

// UpdateChapterRequest represents the request body for updating a chapter.
type UpdateChapterRequest struct {
	Title     *string `json:"title,omitempty"`
	Overview  *string `json:"overview,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
}

// CreateSessionRequest represents the request body for creating a session.
type CreateSessionRequest struct {
	ChapterID     *int64        `json:"chapterId,omitempty"`
	Title         *string       `json:"title,omitempty"`
	SessionNumber *int          `json:"sessionNumber,omitempty"`
	PlannedDate   *time.Time    `json:"plannedDate,omitempty"`
	Stage         *SessionStage `json:"stage,omitempty"`
	PrepNotes     *string       `json:"prepNotes,omitempty"`
}

// UpdateSessionRequest represents the request body for updating a session.
type UpdateSessionRequest struct {
	ChapterID     *int64         `json:"chapterId,omitempty"`
	Title         *string        `json:"title,omitempty"`
	SessionNumber *int           `json:"sessionNumber,omitempty"`
	PlannedDate   *time.Time     `json:"plannedDate,omitempty"`
	ActualDate    *time.Time     `json:"actualDate,omitempty"`
	Status        *SessionStatus `json:"status,omitempty"`
	Stage         *SessionStage  `json:"stage,omitempty"`
	PrepNotes     *string        `json:"prepNotes,omitempty"`
	ActualNotes   *string        `json:"actualNotes,omitempty"`
	PlayNotes     *string        `json:"playNotes,omitempty"`
}

// Scene represents a structured scene within a game session.
type Scene struct {
	ID               int64            `json:"id"`
	SessionID        int64            `json:"sessionId"`
	CampaignID       int64            `json:"campaignId"`
	Title            string           `json:"title"`
	Description      *string          `json:"description,omitempty"`
	SceneType        string           `json:"sceneType"`
	Status           string           `json:"status"`
	SortOrder        int              `json:"sortOrder"`
	Objective        *string          `json:"objective,omitempty"`
	GMNotes          *string          `json:"gmNotes,omitempty"`
	EntityIDs        []int64          `json:"entityIds"`
	SystemData       json.RawMessage  `json:"systemData,omitempty"`
	Source           string           `json:"source"`
	SourceConfidence SourceConfidence `json:"sourceConfidence"`
	Connections      json.RawMessage  `json:"connections,omitempty"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// CreateSceneRequest represents the request body for creating a scene.
type CreateSceneRequest struct {
	Title            string            `json:"title"`
	Description      *string           `json:"description,omitempty"`
	SceneType        *string           `json:"sceneType,omitempty"`
	SortOrder        *int              `json:"sortOrder,omitempty"`
	Objective        *string           `json:"objective,omitempty"`
	GMNotes          *string           `json:"gmNotes,omitempty"`
	EntityIDs        []int64           `json:"entityIds,omitempty"`
	SystemData       json.RawMessage   `json:"systemData,omitempty"`
	Source           *string           `json:"source,omitempty"`
	SourceConfidence *SourceConfidence `json:"sourceConfidence,omitempty"`
	Connections      json.RawMessage   `json:"connections,omitempty"`
}

// UpdateSceneRequest represents the request body for updating a scene.
type UpdateSceneRequest struct {
	Title            *string           `json:"title,omitempty"`
	Description      *string           `json:"description,omitempty"`
	SceneType        *string           `json:"sceneType,omitempty"`
	Status           *string           `json:"status,omitempty"`
	SortOrder        *int              `json:"sortOrder,omitempty"`
	Objective        *string           `json:"objective,omitempty"`
	GMNotes          *string           `json:"gmNotes,omitempty"`
	EntityIDs        []int64           `json:"entityIds,omitempty"`
	SystemData       json.RawMessage   `json:"systemData,omitempty"`
	Source           *string           `json:"source,omitempty"`
	SourceConfidence *SourceConfidence `json:"sourceConfidence,omitempty"`
	Connections      json.RawMessage   `json:"connections,omitempty"`
}

// SessionChatMessage represents a chat message within a session workflow.
type SessionChatMessage struct {
	ID         int64     `json:"id"`
	SessionID  int64     `json:"sessionId"`
	CampaignID int64     `json:"campaignId"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	SortOrder  int       `json:"sortOrder"`
	CreatedAt  time.Time `json:"createdAt"`
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
	ID             int64         `json:"id"`
	CampaignID     int64         `json:"campaignId"`
	EventDate      *time.Time    `json:"eventDate,omitempty"`
	EventTime      *string       `json:"eventTime,omitempty"`
	DatePrecision  DatePrecision `json:"datePrecision"`
	Description    string        `json:"description"`
	EntityIDs      []int64       `json:"entityIds,omitempty"`
	SessionID      *int64        `json:"sessionId,omitempty"`
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
	EntityIDs      []int64       `json:"entityIds,omitempty"`
	SessionID      *int64        `json:"sessionId,omitempty"`
	IsPlayerKnown  bool          `json:"isPlayerKnown"`
	SourceDocument *string       `json:"sourceDocument,omitempty"`
}

// UpdateTimelineEventRequest represents the request body for updating a timeline event.
type UpdateTimelineEventRequest struct {
	EventDate      *time.Time     `json:"eventDate,omitempty"`
	EventTime      *string        `json:"eventTime,omitempty"`
	DatePrecision  *DatePrecision `json:"datePrecision,omitempty"`
	Description    *string        `json:"description,omitempty"`
	EntityIDs      []int64        `json:"entityIds,omitempty"`
	SessionID      *int64         `json:"sessionId,omitempty"`
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
	ID                int64           `json:"id"`
	CampaignID        int64           `json:"campaignId"`
	EntityID          *int64          `json:"entityId,omitempty"`
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
	CampaignID           int64  `json:"campaignId"`
	GameSystemCode       string `json:"gameSystemCode,omitempty"`
	SourceDocument       string `json:"sourceDocument,omitempty"`
	AutoDetectEntities   bool   `json:"autoDetectEntities"`
	ExtractRelationships bool   `json:"extractRelationships"`
	ExtractEvents        bool   `json:"extractEvents"`
}

// LLMService represents supported LLM service providers.
type LLMService string

const (
	LLMServiceAnthropic LLMService = "anthropic"
	LLMServiceOpenAI    LLMService = "openai"
	LLMServiceGemini    LLMService = "gemini"
	LLMServiceVoyage    LLMService = "voyage"
	LLMServiceStability LLMService = "stability"
	LLMServiceOllama    LLMService = "ollama"
)

// UserSettings stores a user's API keys and service preferences.
// API key fields are excluded from JSON serialization to prevent secret exposure.
// Use UserSettingsResponse for API responses with masked keys.
type UserSettings struct {
	UserID            int64       `json:"userId"`
	ContentGenService *LLMService `json:"contentGenService,omitempty"`
	ContentGenAPIKey  *string     `json:"-"`
	EmbeddingService  *LLMService `json:"embeddingService,omitempty"`
	EmbeddingAPIKey   *string     `json:"-"`
	ImageGenService   *LLMService `json:"imageGenService,omitempty"`
	ImageGenAPIKey    *string     `json:"-"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         time.Time   `json:"updatedAt"`
}

// UserSettingsResponse is the API response with masked API keys.
type UserSettingsResponse struct {
	UserID            int64       `json:"userId"`
	ContentGenService *LLMService `json:"contentGenService,omitempty"`
	ContentGenAPIKey  *string     `json:"contentGenApiKey,omitempty"`
	EmbeddingService  *LLMService `json:"embeddingService,omitempty"`
	EmbeddingAPIKey   *string     `json:"embeddingApiKey,omitempty"`
	ImageGenService   *LLMService `json:"imageGenService,omitempty"`
	ImageGenAPIKey    *string     `json:"imageGenApiKey,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
	UpdatedAt         time.Time   `json:"updatedAt"`
}

// UpdateUserSettingsRequest is the request body for updating user settings.
type UpdateUserSettingsRequest struct {
	ContentGenService *LLMService `json:"contentGenService,omitempty"`
	ContentGenAPIKey  *string     `json:"contentGenApiKey,omitempty"`
	EmbeddingService  *LLMService `json:"embeddingService,omitempty"`
	EmbeddingAPIKey   *string     `json:"embeddingApiKey,omitempty"`
	ImageGenService   *LLMService `json:"imageGenService,omitempty"`
	ImageGenAPIKey    *string     `json:"imageGenApiKey,omitempty"`
}

// CampaignGenre represents campaign genre types.
type CampaignGenre string

const (
	GenreAnimeManga         CampaignGenre = "anime_manga"
	GenreCyberpunk          CampaignGenre = "cyberpunk"
	GenreEspionage          CampaignGenre = "espionage"
	GenreFantasy            CampaignGenre = "fantasy"
	GenreGothic             CampaignGenre = "gothic"
	GenreHistorical         CampaignGenre = "historical"
	GenreHorror             CampaignGenre = "horror"
	GenreLovecraftian       CampaignGenre = "lovecraftian"
	GenreMilitary           CampaignGenre = "military"
	GenreModernUrbanFantasy CampaignGenre = "modern_urban_fantasy"
	GenreMystery            CampaignGenre = "mystery"
	GenrePostApocalyptic    CampaignGenre = "post_apocalyptic"
	GenrePulpAdventure      CampaignGenre = "pulp_adventure"
	GenreScienceFiction     CampaignGenre = "science_fiction"
	GenreSpaceOpera         CampaignGenre = "space_opera"
	GenreSteampunk          CampaignGenre = "steampunk"
	GenreSuperhero          CampaignGenre = "superhero"
	GenreTimeTravel         CampaignGenre = "time_travel"
	GenreWestern            CampaignGenre = "western"
	GenreOther              CampaignGenre = "other"
)

// ChapterEntityMentionType represents how an entity is associated with a chapter.
type ChapterEntityMentionType string

const (
	ChapterEntityMentionLinked    ChapterEntityMentionType = "linked"    // Explicitly linked by user
	ChapterEntityMentionMentioned ChapterEntityMentionType = "mentioned" // Detected in content by AI
	ChapterEntityMentionFeatured  ChapterEntityMentionType = "featured"  // Primary entity for chapter
)

// ChapterEntity represents a link between a chapter and an entity.
type ChapterEntity struct {
	ID          int64                    `json:"id"`
	ChapterID   int64                    `json:"chapterId"`
	EntityID    int64                    `json:"entityId"`
	MentionType ChapterEntityMentionType `json:"mentionType"`
	CreatedAt   time.Time                `json:"createdAt"`

	// Joined fields (not in database)
	Entity *Entity `json:"entity,omitempty"`
}

// CreateChapterEntityRequest represents the request body for linking an entity to a chapter.
type CreateChapterEntityRequest struct {
	EntityID    int64                     `json:"entityId"`
	MentionType *ChapterEntityMentionType `json:"mentionType,omitempty"`
}

// UpdateChapterEntityRequest represents the request body for updating a chapter-entity link.
type UpdateChapterEntityRequest struct {
	MentionType *ChapterEntityMentionType `json:"mentionType,omitempty"`
}

// PlayerCharacter represents a player character in a campaign.
type PlayerCharacter struct {
	ID            int64     `json:"id"`
	CampaignID    int64     `json:"campaignId"`
	EntityID      *int64    `json:"entityId,omitempty"`
	CharacterName string    `json:"characterName"`
	PlayerName    string    `json:"playerName"`
	Description   *string   `json:"description,omitempty"`
	Background    *string   `json:"background,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// CreatePlayerCharacterRequest is the request body for creating a player character.
type CreatePlayerCharacterRequest struct {
	EntityID      *int64  `json:"entityId,omitempty"`
	CharacterName string  `json:"characterName"`
	PlayerName    string  `json:"playerName"`
	Description   *string `json:"description,omitempty"`
	Background    *string `json:"background,omitempty"`
}

// UpdatePlayerCharacterRequest is the request body for updating a player character.
type UpdatePlayerCharacterRequest struct {
	EntityID      *int64  `json:"entityId,omitempty"`
	CharacterName *string `json:"characterName,omitempty"`
	PlayerName    *string `json:"playerName,omitempty"`
	Description   *string `json:"description,omitempty"`
	Background    *string `json:"background,omitempty"`
}

// EntityResolveResult represents a fuzzy-matched entity returned by the
// entity resolve endpoint for wiki-link autocomplete.
type EntityResolveResult struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	EntityType EntityType `json:"entityType"`
	Similarity float64    `json:"similarity"`
}

// SearchResult represents a content chunk returned by hybrid search.
type SearchResult struct {
	SourceTable   string  `json:"sourceTable"`
	SourceID      int64   `json:"sourceId"`
	SourceName    string  `json:"sourceName"`
	ChunkContent  string  `json:"chunkContent"`
	VectorScore   float64 `json:"vectorScore"`
	CombinedScore float64 `json:"combinedScore"`
}

// ContentAnalysisJob represents an analysis run for a content field.
type ContentAnalysisJob struct {
	ID                 int64     `json:"id"`
	CampaignID         int64     `json:"campaignId"`
	SourceTable        string    `json:"sourceTable"`
	SourceID           int64     `json:"sourceId"`
	SourceField        string    `json:"sourceField"`
	Status             string    `json:"status"`
	TotalItems         int       `json:"totalItems"`
	ResolvedItems      int       `json:"resolvedItems"`
	EnrichmentTotal    int       `json:"enrichmentTotal"`
	EnrichmentResolved int       `json:"enrichmentResolved"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ContentAnalysisItem represents an individual detection within an analysis job.
type ContentAnalysisItem struct {
	ID               int64           `json:"id"`
	JobID            int64           `json:"jobId"`
	DetectionType    string          `json:"detectionType"`
	MatchedText      string          `json:"matchedText"`
	EntityID         *int64          `json:"entityId,omitempty"`
	Similarity       *float64        `json:"similarity,omitempty"`
	ContextSnippet   *string         `json:"contextSnippet,omitempty"`
	PositionStart    *int            `json:"positionStart,omitempty"`
	PositionEnd      *int            `json:"positionEnd,omitempty"`
	Resolution       string          `json:"resolution"`
	ResolvedEntityID *int64          `json:"resolvedEntityId,omitempty"`
	ResolvedAt       *time.Time      `json:"resolvedAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	SuggestedContent json.RawMessage `json:"suggestedContent,omitempty"`
	Phase            string          `json:"phase"`
	AgentName        string          `json:"agentName,omitempty"`
	PipelineRunID    *int64          `json:"pipelineRunId,omitempty"`

	// Joined fields (not in database)
	EntityName *string     `json:"entityName,omitempty"`
	EntityType *EntityType `json:"entityType,omitempty"`
}

// Item type constants for content analysis items.
const (
	ItemTypeNewEntitySuggestion = "new_entity_suggestion"
)

// ResolveAnalysisItemRequest is the request body for resolving an analysis item.
type ResolveAnalysisItemRequest struct {
	Resolution               string                 `json:"resolution"`
	EntityType               *EntityType            `json:"entityType,omitempty"`
	EntityName               *string                `json:"entityName,omitempty"`
	SuggestedContentOverride map[string]interface{} `json:"suggestedContentOverride,omitempty"`
}

// AnalysisSummary provides a brief summary of a content analysis job.
type AnalysisSummary struct {
	JobID        int64 `json:"jobId"`
	PendingCount int   `json:"pendingCount"`
}

// EntityLog represents a chronological event history entry for an entity.
type EntityLog struct {
	ID          int64     `json:"id"`
	EntityID    int64     `json:"entityId"`
	CampaignID  int64     `json:"campaignId"`
	ChapterID   *int64    `json:"chapterId,omitempty"`
	SessionID   *int64    `json:"sessionId,omitempty"`
	SourceTable *string   `json:"sourceTable,omitempty"`
	SourceID    *int64    `json:"sourceId,omitempty"`
	Content     string    `json:"content"`
	OccurredAt  *string   `json:"occurredAt,omitempty"`
	SortOrder   *int      `json:"sortOrder,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CreateEntityLogRequest is the request body for creating an entity log entry.
type CreateEntityLogRequest struct {
	Content    string  `json:"content"`
	ChapterID  *int64  `json:"chapterId,omitempty"`
	SessionID  *int64  `json:"sessionId,omitempty"`
	OccurredAt *string `json:"occurredAt,omitempty"`
	SortOrder  *int    `json:"sortOrder,omitempty"`
}

// UpdateEntityLogRequest is the request body for updating an entity log entry.
type UpdateEntityLogRequest struct {
	Content    *string `json:"content,omitempty"`
	ChapterID  *int64  `json:"chapterId,omitempty"`
	SessionID  *int64  `json:"sessionId,omitempty"`
	OccurredAt *string `json:"occurredAt,omitempty"`
	SortOrder  *int    `json:"sortOrder,omitempty"`
}

// DescriptionUpdateSuggestion is an enrichment suggestion for updating
// an entity's description.
type DescriptionUpdateSuggestion struct {
	CurrentDescription   string `json:"currentDescription"`
	SuggestedDescription string `json:"suggestedDescription"`
	Rationale            string `json:"rationale"`
}

// LogEntrySuggestion is an enrichment suggestion for creating a new
// entity log entry.
type LogEntrySuggestion struct {
	Content    string  `json:"content"`
	OccurredAt *string `json:"occurredAt,omitempty"`
}

// RelationshipSuggestion is an enrichment suggestion for creating a
// new relationship between entities.
type RelationshipSuggestion struct {
	SourceEntityID   int64  `json:"sourceEntityId"`
	SourceEntityName string `json:"sourceEntityName"`
	TargetEntityID   int64  `json:"targetEntityId"`
	TargetEntityName string `json:"targetEntityName"`
	RelationshipType string `json:"relationshipType"`
	Description      string `json:"description"`
}

// Draft represents a server-side auto-saved draft for an editor page.
type Draft struct {
	ID            int64           `json:"id"`
	CampaignID    int64           `json:"campaignId"`
	UserID        int64           `json:"userId"`
	SourceTable   string          `json:"sourceTable"`
	SourceID      int64           `json:"sourceId"`
	IsNew         bool            `json:"isNew"`
	DraftData     json.RawMessage `json:"draftData"`
	ServerVersion *int            `json:"serverVersion,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

// DraftIndicator is a lightweight representation of a draft used in
// list views to show which items have pending drafts.
type DraftIndicator struct {
	SourceTable string    `json:"sourceTable"`
	SourceID    int64     `json:"sourceId"`
	IsNew       bool      `json:"isNew"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SaveDraftRequest is the request body for saving or updating a draft.
type SaveDraftRequest struct {
	SourceTable   string          `json:"sourceTable"`
	SourceID      int64           `json:"sourceId"`
	IsNew         bool            `json:"isNew"`
	DraftData     json.RawMessage `json:"draftData"`
	ServerVersion *int            `json:"serverVersion,omitempty"`
}
