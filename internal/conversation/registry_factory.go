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
	"github.com/antonypegg/imagineer/internal/database"
	"github.com/antonypegg/imagineer/internal/llm"
)

// BuildToolRegistry creates a ToolRegistry populated
// with all conversation tools: 8 procedural, 2 document,
// and 3 agent tools (13 total).
//
// Pass nil for db and provider when only tool definitions
// are needed (e.g. in tests). The database and provider
// are captured by closures and only used at execution
// time.
func BuildToolRegistry(
	db *database.DB,
	campaignID int64,
	provider llm.StreamingProvider,
	schemasDir string,
) *ToolRegistry {
	registry := NewToolRegistry()

	// Register procedural tools (8).
	for _, tool := range BuildProceduralTools(
		db, campaignID, schemasDir) {
		registry.Register(tool)
	}

	// Register document tools (2).
	for _, tool := range BuildDocumentTools(
		db, campaignID) {
		registry.Register(tool)
	}

	// Register agent tools (3).
	// Agent tools need access to sub-sets of the
	// procedural tools. BuildAgentTools handles
	// the sub-tool filtering internally.
	for _, tool := range BuildAgentTools(
		registry, provider) {
		registry.Register(tool)
	}

	return registry
}
