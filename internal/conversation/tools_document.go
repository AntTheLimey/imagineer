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
	"strings"

	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
)

// BuildDocumentTools builds the read_document and
// edit_document tools and returns them as a slice.
// Pass nil for db when only tool definitions are
// needed (e.g. in tests).
func BuildDocumentTools(
	db *database.DB,
	campaignID int64,
) []Tool {
	return []Tool{
		buildReadDocumentTool(db, campaignID),
		buildEditDocumentTool(db, campaignID),
	}
}

// documentResult is the JSON structure returned by
// both read_document and edit_document tools.
type documentResult struct {
	ScopeType string `json:"scope_type"`
	ScopeID   int64  `json:"scope_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
}

// validScopeTypes lists the scope types that have a
// readable document. Campaign is excluded because it
// does not have a single document body.
var validDocumentScopeTypes = map[models.ScopeType]bool{
	models.ScopeTypeEntity:  true,
	models.ScopeTypeChapter: true,
	models.ScopeTypeSession: true,
	models.ScopeTypeScene:   true,
}

// derefStr safely dereferences a *string, returning
// an empty string if the pointer is nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// readDocument fetches the title and content for a
// given scope type and ID from the database. It also
// verifies the record belongs to the expected campaign.
func readDocument(
	ctx context.Context,
	db *database.DB,
	scopeType models.ScopeType,
	scopeID int64,
	campaignID int64,
) (title, content string, err error) {
	switch scopeType {
	case models.ScopeTypeChapter:
		ch, e := db.GetChapter(ctx, scopeID)
		if e != nil {
			return "", "", fmt.Errorf(
				"failed to get chapter: %w", e)
		}
		if ch.CampaignID != campaignID {
			return "", "", fmt.Errorf(
				"document does not belong to this campaign")
		}
		return ch.Title, derefStr(ch.Overview), nil

	case models.ScopeTypeSession:
		s, e := db.GetSession(ctx, scopeID)
		if e != nil {
			return "", "", fmt.Errorf(
				"failed to get session: %w", e)
		}
		if s.CampaignID != campaignID {
			return "", "", fmt.Errorf(
				"document does not belong to this campaign")
		}
		return derefStr(s.Title), derefStr(s.PrepNotes), nil

	case models.ScopeTypeScene:
		sc, e := db.GetScene(ctx, scopeID)
		if e != nil {
			return "", "", fmt.Errorf(
				"failed to get scene: %w", e)
		}
		if sc.CampaignID != campaignID {
			return "", "", fmt.Errorf(
				"document does not belong to this campaign")
		}
		return sc.Title, derefStr(sc.Description), nil

	case models.ScopeTypeEntity:
		ent, e := db.GetEntity(ctx, scopeID)
		if e != nil {
			return "", "", fmt.Errorf(
				"failed to get entity: %w", e)
		}
		if ent.CampaignID != campaignID {
			return "", "", fmt.Errorf(
				"document does not belong to this campaign")
		}
		return ent.Name, derefStr(ent.Description), nil

	default:
		return "", "", fmt.Errorf(
			"unsupported scope_type: %s", scopeType)
	}
}

// buildReadDocumentTool creates a tool that reads the
// document content for a chapter, session, scene, or
// entity. It deliberately excludes GM notes for
// security purposes.
func buildReadDocumentTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name: "read_document",
			Description: "Read the text content of a chapter, session, scene, or entity. " +
				"Returns the title and main body text. " +
				"Does not include GM notes.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"scope_type": {
						"type": "string",
						"description": "The type of document to read",
						"enum": ["entity", "chapter", "session", "scene"]
					},
					"scope_id": {
						"type": "integer",
						"description": "The ID of the document to read"
					}
				},
				"required": ["scope_type", "scope_id"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				ScopeType string `json:"scope_type"`
				ScopeID   int64  `json:"scope_id"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.ScopeType == "" {
				return nil, fmt.Errorf(
					"scope_type is required")
			}
			if params.ScopeID == 0 {
				return nil, fmt.Errorf(
					"scope_id is required")
			}

			st := models.ScopeType(params.ScopeType)

			if st == models.ScopeTypeCampaign {
				return nil, fmt.Errorf(
					"campaign scope does not have a document")
			}
			if !validDocumentScopeTypes[st] {
				return nil, fmt.Errorf(
					"invalid scope_type: %s", params.ScopeType)
			}

			title, content, err := readDocument(
				ctx, db, st, params.ScopeID,
				campaignID)
			if err != nil {
				return nil, err
			}

			return json.Marshal(documentResult{
				ScopeType: params.ScopeType,
				ScopeID:   params.ScopeID,
				Title:     title,
				Content:   content,
			})
		},
	}
}

// writeContent saves the updated content back to the
// database for the given scope type and ID.
func writeContent(
	ctx context.Context,
	db *database.DB,
	scopeType models.ScopeType,
	scopeID int64,
	content string,
) error {
	switch scopeType {
	case models.ScopeTypeChapter:
		_, err := db.UpdateChapter(ctx, scopeID,
			models.UpdateChapterRequest{
				Overview: &content,
			})
		if err != nil {
			return fmt.Errorf(
				"failed to update chapter: %w", err)
		}
		return nil

	case models.ScopeTypeSession:
		_, err := db.UpdateSession(ctx, scopeID,
			models.UpdateSessionRequest{
				PrepNotes: &content,
			})
		if err != nil {
			return fmt.Errorf(
				"failed to update session: %w", err)
		}
		return nil

	case models.ScopeTypeScene:
		_, err := db.UpdateScene(ctx, scopeID,
			models.UpdateSceneRequest{
				Description: &content,
			})
		if err != nil {
			return fmt.Errorf(
				"failed to update scene: %w", err)
		}
		return nil

	case models.ScopeTypeEntity:
		_, err := db.UpdateEntity(ctx, scopeID,
			models.UpdateEntityRequest{
				Description: &content,
			})
		if err != nil {
			return fmt.Errorf(
				"failed to update entity: %w", err)
		}
		return nil

	default:
		return fmt.Errorf(
			"unsupported scope_type for write: %s",
			scopeType)
	}
}

// buildEditDocumentTool creates a tool that edits the
// document content for a chapter, session, scene, or
// entity. Supports replace, append, and insert
// (prepend) operations.
func buildEditDocumentTool(
	db *database.DB,
	campaignID int64,
) Tool {
	return Tool{
		Definition: llm.ToolDefinition{
			Name: "edit_document",
			Description: "Edit the content of a chapter, session, scene, or entity. " +
				"Supports replace (first occurrence), insert (prepend), " +
				"and append operations.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"scope_type": {
						"type": "string",
						"description": "The type of document to edit",
						"enum": ["entity", "chapter", "session", "scene"]
					},
					"scope_id": {
						"type": "integer",
						"description": "The ID of the document to edit"
					},
					"operation": {
						"type": "string",
						"description": "The edit operation to perform",
						"enum": ["replace", "append", "insert"]
					},
					"old_text": {
						"type": "string",
						"description": "The text to find and replace (required for replace operation)"
					},
					"new_text": {
						"type": "string",
						"description": "The new text to insert, append, or use as replacement"
					}
				},
				"required": ["scope_type", "scope_id", "operation", "new_text"]
			}`),
		},
		Execute: func(
			ctx context.Context,
			input json.RawMessage,
		) (json.RawMessage, error) {
			var params struct {
				ScopeType string `json:"scope_type"`
				ScopeID   int64  `json:"scope_id"`
				Operation string `json:"operation"`
				OldText   string `json:"old_text"`
				NewText   string `json:"new_text"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf(
					"invalid input: %w", err)
			}
			if params.ScopeType == "" {
				return nil, fmt.Errorf(
					"scope_type is required")
			}
			if params.ScopeID == 0 {
				return nil, fmt.Errorf(
					"scope_id is required")
			}
			if params.Operation == "" {
				return nil, fmt.Errorf(
					"operation is required")
			}
			if params.NewText == "" {
				return nil, fmt.Errorf(
					"new_text is required")
			}

			// Validate operation.
			switch params.Operation {
			case "replace", "append", "insert":
				// valid
			default:
				return nil, fmt.Errorf(
					"invalid operation: %s (must be replace, append, or insert)",
					params.Operation)
			}

			// Replace requires old_text.
			if params.Operation == "replace" &&
				params.OldText == "" {
				return nil, fmt.Errorf(
					"old_text is required for replace operation")
			}

			st := models.ScopeType(params.ScopeType)

			if st == models.ScopeTypeCampaign {
				return nil, fmt.Errorf(
					"campaign scope does not have a document")
			}
			if !validDocumentScopeTypes[st] {
				return nil, fmt.Errorf(
					"invalid scope_type: %s", params.ScopeType)
			}

			// Note: read-modify-write is not transactional.
			// Acceptable because conversations serialise
			// requests per session.
			title, content, err := readDocument(
				ctx, db, st, params.ScopeID,
				campaignID)
			if err != nil {
				return nil, err
			}

			// Apply the operation.
			switch params.Operation {
			case "replace":
				if !strings.Contains(content, params.OldText) {
					return nil, fmt.Errorf(
						"old_text not found in document content")
				}
				content = strings.Replace(
					content, params.OldText,
					params.NewText, 1)

			case "append":
				content = content + params.NewText

			case "insert":
				content = params.NewText + content
			}

			// Write back.
			if err := writeContent(
				ctx, db, st, params.ScopeID, content); err != nil {
				return nil, fmt.Errorf(
					"failed to save document: %w", err)
			}

			return json.Marshal(documentResult{
				ScopeType: params.ScopeType,
				ScopeID:   params.ScopeID,
				Title:     title,
				Content:   content,
			})
		},
	}
}
