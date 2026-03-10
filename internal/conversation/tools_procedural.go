/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// BuildProceduralTools builds all 8 procedural tools
// and returns them as a slice. Pass nil for db when
// only tool definitions are needed (e.g. in tests).
func BuildProceduralTools(
	db *database.DB,
	campaignID int64,
	schemasDir string,
) []Tool {
	return []Tool{
		buildSearchEntitiesTool(db, campaignID),
		buildGetEntityTool(db),
		buildCreateEntityTool(db, campaignID),
		buildUpdateEntityTool(db),
		buildCreateRelationshipTool(db, campaignID),
		buildGetRelatedEntitiesTool(db),
		buildSearchContentTool(db, campaignID),
		buildReadGameSchemaTool(schemasDir),
	}
}

// buildSearchEntitiesTool creates a tool that searches
// entities by name within a campaign using trigram
// similarity matching.
func buildSearchEntitiesTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_entities",
			Description: "Search for entities by name within the campaign. Uses fuzzy matching to find similar names. Optionally filter by entity type.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The name or partial name to search for"
					},
					"entity_type": {
						"type": "string",
						"description": "Optional entity type filter (e.g. npc, location, item, faction)"
					}
				},
				"required": ["query"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				Query      string `json:"query"`
				EntityType string `json:"entity_type"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.Query == "" {
				return nil, fmt.Errorf(
					"query is required")
			}

			if params.EntityType != "" {
				entities, err := db.ListEntitiesByType(
					ctx, campaignID,
					models.EntityType(params.EntityType))
				if err != nil {
					return nil, fmt.Errorf(
						"failed to list entities by type: %w", err)
				}
				// Filter by name similarity client-side
				// since ListEntitiesByType does not
				// support fuzzy search.
				queryLower := strings.ToLower(params.Query)
				var filtered []models.Entity
				for _, e := range entities {
					if strings.Contains(
						strings.ToLower(e.Name),
						queryLower,
					) {
						filtered = append(filtered, e)
					}
				}
				return json.Marshal(filtered)
			}

			entities, err := db.SearchEntitiesByName(
				ctx, campaignID, params.Query, 20)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to search entities: %w", err)
			}
			return json.Marshal(entities)
		},
	}
}

// buildGetEntityTool creates a tool that retrieves a
// single entity by its ID.
func buildGetEntityTool(db *database.DB) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "get_entity",
			Description: "Get a single entity by its ID. Returns the full entity including attributes, tags, and metadata.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {
						"type": "integer",
						"description": "The entity ID"
					}
				},
				"required": ["id"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.ID == 0 {
				return nil, fmt.Errorf("id is required")
			}

			entity, err := db.GetEntity(ctx, params.ID)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to get entity: %w", err)
			}
			return json.Marshal(entity)
		},
	}
}

// buildCreateEntityTool creates a tool that adds a new
// entity to the campaign.
func buildCreateEntityTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "create_entity",
			Description: "Create a new entity in the campaign. Entities represent NPCs, locations, items, factions, and other world elements.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The entity name"
					},
					"entity_type": {
						"type": "string",
						"description": "The type of entity",
						"enum": ["npc", "location", "item", "faction", "clue", "creature", "event", "document", "other"]
					},
					"description": {
						"type": "string",
						"description": "A description of the entity"
					}
				},
				"required": ["name", "entity_type"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				Name        string `json:"name"`
				EntityType  string `json:"entity_type"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.Name == "" {
				return nil, fmt.Errorf(
					"name is required")
			}
			if params.EntityType == "" {
				return nil, fmt.Errorf(
					"entity_type is required")
			}

			req := models.CreateEntityRequest{
				EntityType: models.EntityType(
					params.EntityType),
				Name: params.Name,
			}
			if params.Description != "" {
				req.Description = &params.Description
			}

			entity, err := db.CreateEntity(
				ctx, campaignID, req)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to create entity: %w", err)
			}
			return json.Marshal(entity)
		},
	}
}

// buildUpdateEntityTool creates a tool that updates an
// existing entity by ID.
func buildUpdateEntityTool(db *database.DB) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "update_entity",
			Description: "Update an existing entity. Only the fields provided will be changed; omitted fields remain unchanged.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"id": {
						"type": "integer",
						"description": "The entity ID to update"
					},
					"name": {
						"type": "string",
						"description": "New name for the entity"
					},
					"description": {
						"type": "string",
						"description": "New description for the entity"
					}
				},
				"required": ["id"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				ID          int64   `json:"id"`
				Name        *string `json:"name"`
				Description *string `json:"description"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.ID == 0 {
				return nil, fmt.Errorf("id is required")
			}

			req := models.UpdateEntityRequest{
				Name:        params.Name,
				Description: params.Description,
			}

			entity, err := db.UpdateEntity(
				ctx, params.ID, req)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to update entity: %w", err)
			}
			return json.Marshal(entity)
		},
	}
}

// buildCreateRelationshipTool creates a tool that
// establishes a relationship between two entities.
func buildCreateRelationshipTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "create_relationship",
			Description: "Create a relationship between two entities. Requires the relationship_type_id which identifies the kind of relationship (e.g. ally_of, located_in). Use get_related_entities first to see existing relationships.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"source_entity_id": {
						"type": "integer",
						"description": "The ID of the source entity"
					},
					"target_entity_id": {
						"type": "integer",
						"description": "The ID of the target entity"
					},
					"relationship_type_id": {
						"type": "integer",
						"description": "The ID of the relationship type"
					},
					"description": {
						"type": "string",
						"description": "A description of the relationship"
					}
				},
				"required": ["source_entity_id", "target_entity_id", "relationship_type_id"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				SourceEntityID     int64   `json:"source_entity_id"`
				TargetEntityID     int64   `json:"target_entity_id"`
				RelationshipTypeID int64   `json:"relationship_type_id"`
				Description        *string `json:"description"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.SourceEntityID == 0 {
				return nil, fmt.Errorf(
					"source_entity_id is required")
			}
			if params.TargetEntityID == 0 {
				return nil, fmt.Errorf(
					"target_entity_id is required")
			}
			if params.RelationshipTypeID == 0 {
				return nil, fmt.Errorf(
					"relationship_type_id is required")
			}

			req := models.CreateRelationshipRequest{
				SourceEntityID:     params.SourceEntityID,
				TargetEntityID:     params.TargetEntityID,
				RelationshipTypeID: params.RelationshipTypeID,
				Description:        params.Description,
			}

			rel, err := db.CreateRelationship(
				ctx, campaignID, req)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to create relationship: %w",
					err)
			}
			return json.Marshal(rel)
		},
	}
}

// buildGetRelatedEntitiesTool creates a tool that
// retrieves all relationships for a given entity.
func buildGetRelatedEntitiesTool(
	db *database.DB,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "get_related_entities",
			Description: "Get all relationships for a given entity, including both forward and inverse relationships. Shows relationship types, connected entities, and descriptions.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"entity_id": {
						"type": "integer",
						"description": "The entity ID to get relationships for"
					}
				},
				"required": ["entity_id"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				EntityID int64 `json:"entity_id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.EntityID == 0 {
				return nil, fmt.Errorf(
					"entity_id is required")
			}

			rels, err := db.GetEntityRelationships(
				ctx, params.EntityID)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to get relationships: %w",
					err)
			}
			return json.Marshal(rels)
		},
	}
}

// buildSearchContentTool creates a tool that performs
// hybrid vector and BM25 search across campaign content.
func buildSearchContentTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "search_content",
			Description: "Search across all campaign content (chapters, sessions, memories) using hybrid vector and text search. Returns relevant content chunks ranked by relevance.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The search query"
					},
					"limit": {
						"type": "integer",
						"description": "Maximum number of results to return (default 10, max 100)"
					}
				},
				"required": ["query"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.Query == "" {
				return nil, fmt.Errorf(
					"query is required")
			}
			if params.Limit <= 0 {
				params.Limit = 10
			}
			if params.Limit > 100 {
				params.Limit = 100
			}

			results, err := db.SearchCampaignContent(
				ctx, campaignID,
				params.Query, params.Limit)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to search content: %w", err)
			}
			return json.Marshal(results)
		},
	}
}

// buildReadGameSchemaTool creates a tool that reads a
// YAML game system schema file from disk. This tool
// does not require a database connection.
func buildReadGameSchemaTool(
	schemasDir string,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name:        "read_game_schema",
			Description: "Read a game system schema definition (YAML). Contains character attributes, skills, dice conventions, and entity attribute requirements for a specific TTRPG system.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"system_code": {
						"type": "string",
						"description": "The game system code (e.g. coc7e, gurps4e, fitd, dnd5e2024)"
					}
				},
				"required": ["system_code"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				SystemCode string `json:"system_code"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.SystemCode == "" {
				return nil, fmt.Errorf(
					"system_code is required")
			}

			// Sanitise the system code to prevent
			// path traversal.
			cleaned := filepath.Base(params.SystemCode)
			if cleaned != params.SystemCode ||
				strings.Contains(cleaned, "..") {
				return nil, fmt.Errorf(
					"invalid system_code: %q",
					params.SystemCode)
			}

			filename := cleaned + ".yaml"
			path := filepath.Join(schemasDir, filename)

			data, err := os.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, fmt.Errorf(
						"schema not found for system: %s",
						params.SystemCode)
				}
				return nil, fmt.Errorf(
					"failed to read schema: %w", err)
			}

			result := map[string]string{
				"system_code": params.SystemCode,
				"content":     string(data),
			}
			return json.Marshal(result)
		},
	}
}
