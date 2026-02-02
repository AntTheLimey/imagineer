/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package evernotelocal provides an importer for the local Evernote macOS
// application using AppleScript.
package evernotelocal

import (
	"time"
)

// Notebook represents an Evernote notebook.
type Notebook struct {
	Name      string `json:"name"`
	NoteCount int    `json:"noteCount"`
}

// NoteSummary represents a summary of a note without its full content.
type NoteSummary struct {
	NoteLink string    `json:"noteLink"`
	Title    string    `json:"title"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
	TagNames []string  `json:"tagNames,omitempty"`
}

// NoteContent represents the full content of a single note.
type NoteContent struct {
	Title       string    `json:"title"`
	HTMLContent string    `json:"htmlContent"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Tags        []string  `json:"tags,omitempty"`
}

// EvernoteVersion indicates the version of Evernote installed.
type EvernoteVersion string

const (
	// EvernoteVersionUnknown indicates version could not be determined.
	EvernoteVersionUnknown EvernoteVersion = "unknown"

	// EvernoteVersionLegacy indicates Evernote Legacy (version 7.x) which
	// has full AppleScript support.
	EvernoteVersionLegacy EvernoteVersion = "legacy"

	// EvernoteVersion10 indicates Evernote 10.x (Electron-based) which
	// has very limited AppleScript support.
	EvernoteVersion10 EvernoteVersion = "10.x"
)

// EvernoteStatus represents the status of the Evernote application.
type EvernoteStatus struct {
	Available bool            `json:"available"`
	Running   bool            `json:"running"`
	Version   EvernoteVersion `json:"version,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// ImportRequest represents a request to import notes from a notebook.
type ImportRequest struct {
	CampaignID         string `json:"campaignId"`
	NotebookName       string `json:"notebookName"`
	AutoDetectEntities bool   `json:"autoDetectEntities"`
}
