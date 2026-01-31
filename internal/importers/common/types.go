// Package common provides shared types and interfaces for content importers.
package common

import (
	"context"
	"io"
)

// EntityType represents the type of entity being imported.
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

// ExtractedEntity represents an entity extracted from imported content.
type ExtractedEntity struct {
	Name        string                 `json:"name"`
	Type        EntityType             `json:"type"`
	Description string                 `json:"description,omitempty"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	SourceDoc   string                 `json:"sourceDoc,omitempty"`
}

// ExtractedRelationship represents a relationship extracted from imported content.
type ExtractedRelationship struct {
	SourceName       string `json:"sourceName"`
	TargetName       string `json:"targetName"`
	RelationshipType string `json:"relationshipType"`
	Description      string `json:"description,omitempty"`
	Bidirectional    bool   `json:"bidirectional"`
}

// ExtractedEvent represents a timeline event extracted from imported content.
type ExtractedEvent struct {
	Description   string   `json:"description"`
	Date          string   `json:"date,omitempty"`
	DatePrecision string   `json:"datePrecision,omitempty"`
	EntityNames   []string `json:"entityNames,omitempty"`
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
	Events        []ExtractedEvent        `json:"events"`
	Warnings      []string                `json:"warnings"`
	Errors        []string                `json:"errors"`
}

// Importer defines the interface for content importers.
type Importer interface {
	// Import reads content from the provided source and extracts entities.
	Import(ctx context.Context, source io.Reader, options ImportOptions) (*ImportResult, error)

	// Name returns the name of the importer.
	Name() string

	// SupportedFormats returns the file formats this importer supports.
	SupportedFormats() []string
}

// ImportOptions provides configuration for import operations.
type ImportOptions struct {
	// CampaignID is the target campaign for imported entities.
	CampaignID string `json:"campaignId"`

	// GameSystemCode identifies the game system (e.g., "coc-7e").
	GameSystemCode string `json:"gameSystemCode"`

	// SourceDocument is a label for tracking where content came from.
	SourceDocument string `json:"sourceDocument"`

	// AutoDetectEntities enables automatic entity type detection.
	AutoDetectEntities bool `json:"autoDetectEntities"`

	// ExtractRelationships enables relationship extraction from text.
	ExtractRelationships bool `json:"extractRelationships"`

	// ExtractEvents enables timeline event extraction from text.
	ExtractEvents bool `json:"extractEvents"`
}
