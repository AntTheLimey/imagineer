/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package main

import "time"

// Server configuration defaults.
//
// DefaultPort is set to "3001" to avoid conflict with the MCP server
// which runs on port 8080. This ensures both services can run
// simultaneously during development.
const (
	DefaultPort       = "3001"
	DefaultConfigPath = "config/db/db.json"
	ShutdownTimeout   = 30 * time.Second
)
