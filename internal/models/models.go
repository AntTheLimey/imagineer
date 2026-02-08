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
type Relationship struct {
	ID               int64             `json:"id"`
	CampaignID       int64             `json:"campaignId"`
	SourceEntityID   int64             `json:"sourceEntityId"`
	TargetEntityID   int64             `json:"targetEntityId"`
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
	SourceEntityID   int64             `json:"sourceEntityId"`
	TargetEntityID   int64             `json:"targetEntityId"`
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

// RelationshipType defines a relationship type with its inverse mapping.
type RelationshipType struct {
	ID                  int64     `json:"id"`
	CampaignID          *int64    `json:"campaignId,omitempty"` // nil = system default
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
	SessionStagePrep   SessionStage = "prep"
	SessionStagePlay   SessionStage = "play"
	SessionStageWrapUp SessionStage = "wrap_up"
)

// Session represents a game session within a campaign.
type Session struct {
	ID              int64           `json:"id"`
	CampaignID      int64           `json:"campaignId"`
	ChapterID       *int64          `json:"chapterId,omitempty"`
	Title           *string         `json:"title,omitempty"`
	SessionNumber   *int            `json:"sessionNumber,omitempty"`
	PlannedDate     *time.Time      `json:"plannedDate,omitempty"`
	ActualDate      *time.Time      `json:"actualDate,omitempty"`
	Status          SessionStatus   `json:"status"`
	Stage           SessionStage    `json:"stage"`
	PrepNotes       *string         `json:"prepNotes,omitempty"`
	PlannedScenes   json.RawMessage `json:"plannedScenes,omitempty"`
	ActualNotes     *string         `json:"actualNotes,omitempty"`
	Discoveries     json.RawMessage `json:"discoveries,omitempty"`
	PlayerDecisions json.RawMessage `json:"playerDecisions,omitempty"`
	Consequences    json.RawMessage `json:"consequences,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
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
	ChapterID     *int64          `json:"chapterId,omitempty"`
	Title         *string         `json:"title,omitempty"`
	SessionNumber *int            `json:"sessionNumber,omitempty"`
	PlannedDate   *time.Time      `json:"plannedDate,omitempty"`
	Stage         *SessionStage   `json:"stage,omitempty"`
	PrepNotes     *string         `json:"prepNotes,omitempty"`
	PlannedScenes json.RawMessage `json:"plannedScenes,omitempty"`
}

// UpdateSessionRequest represents the request body for updating a session.
type UpdateSessionRequest struct {
	ChapterID       *int64          `json:"chapterId,omitempty"`
	Title           *string         `json:"title,omitempty"`
	SessionNumber   *int            `json:"sessionNumber,omitempty"`
	PlannedDate     *time.Time      `json:"plannedDate,omitempty"`
	ActualDate      *time.Time      `json:"actualDate,omitempty"`
	Status          *SessionStatus  `json:"status,omitempty"`
	Stage           *SessionStage   `json:"stage,omitempty"`
	PrepNotes       *string         `json:"prepNotes,omitempty"`
	PlannedScenes   json.RawMessage `json:"plannedScenes,omitempty"`
	ActualNotes     *string         `json:"actualNotes,omitempty"`
	Discoveries     json.RawMessage `json:"discoveries,omitempty"`
	PlayerDecisions json.RawMessage `json:"playerDecisions,omitempty"`
	Consequences    json.RawMessage `json:"consequences,omitempty"`
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

// SearchResult represents a content chunk returned by hybrid search.
type SearchResult struct {
	SourceTable   string  `json:"sourceTable"`
	SourceID      int64   `json:"sourceId"`
	SourceName    string  `json:"sourceName"`
	ChunkContent  string  `json:"chunkContent"`
	VectorScore   float64 `json:"vectorScore"`
	CombinedScore float64 `json:"combinedScore"`
}
