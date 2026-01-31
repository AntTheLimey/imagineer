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

import (
	"testing"
	"time"
)

// TestDefaultPort verifies the server uses port 3001 by default.
// Port 3001 is used to avoid conflict with the MCP server which runs on
// port 8080. Using 8080 would prevent both services from running
// simultaneously during development.
func TestDefaultPort(t *testing.T) {
	expected := "3001"
	if DefaultPort != expected {
		t.Errorf("DefaultPort = %q, want %q (port 8080 conflicts with MCP server)", DefaultPort, expected)
	}
}

// TestDefaultConfigPath verifies the default database configuration path.
func TestDefaultConfigPath(t *testing.T) {
	expected := "config/db/db.json"
	if DefaultConfigPath != expected {
		t.Errorf("DefaultConfigPath = %q, want %q", DefaultConfigPath, expected)
	}
}

// TestShutdownTimeout verifies the graceful shutdown timeout value.
func TestShutdownTimeout(t *testing.T) {
	expected := 30 * time.Second
	if ShutdownTimeout != expected {
		t.Errorf("ShutdownTimeout = %v, want %v", ShutdownTimeout, expected)
	}
}
