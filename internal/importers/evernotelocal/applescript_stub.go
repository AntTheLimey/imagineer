//go:build !darwin

/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package evernotelocal

import (
	"context"
	"errors"
	"time"
)

const (
	// DefaultTimeout is the default timeout for AppleScript execution.
	DefaultTimeout = 60 * time.Second

	// ListNotebooksTimeout is the timeout for listing notebooks.
	ListNotebooksTimeout = 30 * time.Second

	// ListNotesTimeout is the timeout for listing notes in a notebook.
	ListNotesTimeout = 120 * time.Second

	// GetNoteContentTimeout is the timeout for getting note content.
	GetNoteContentTimeout = 30 * time.Second
)

// ErrNotSupported is returned when AppleScript operations are attempted on
// non-macOS platforms.
var ErrNotSupported = errors.New("Evernote local import is only supported on macOS")

// ErrEvernote10Unsupported is returned when attempting to use features that
// require Evernote Legacy on Evernote 10.x.
var ErrEvernote10Unsupported = errors.New("evernote 10.x does not support this operation")

// Evernote10Error is the error message shown when Evernote 10.x is detected.
const Evernote10Error = `Evernote 10.x detected. This version has limited AppleScript support and cannot be used for direct import.

To import your Evernote notes, please use one of these alternatives:

1. Export as ENEX files (recommended):
   - In Evernote 10.x, right-click a notebook and select "Export notebook..."
   - Choose ENEX format and save the file
   - Use Imagineer's ENEX file import feature instead

2. Install Evernote Legacy:
   - Download Evernote Legacy from https://help.evernote.com/hc/en-us/articles/360052560314
   - Evernote Legacy (version 7.x) has full AppleScript support
   - Note: Evernote Legacy may no longer sync with Evernote servers`

// AppleScriptExecutor handles AppleScript execution for Evernote.
// On non-darwin platforms, all operations return ErrNotSupported.
type AppleScriptExecutor struct {
	timeout time.Duration
}

// NewAppleScriptExecutor creates a new AppleScriptExecutor.
func NewAppleScriptExecutor() *AppleScriptExecutor {
	return &AppleScriptExecutor{
		timeout: DefaultTimeout,
	}
}

// NewAppleScriptExecutorWithTimeout creates a new AppleScriptExecutor with a
// custom timeout.
func NewAppleScriptExecutorWithTimeout(timeout time.Duration) *AppleScriptExecutor {
	return &AppleScriptExecutor{
		timeout: timeout,
	}
}

// CheckEvernoteStatus returns not available on non-macOS platforms.
func (e *AppleScriptExecutor) CheckEvernoteStatus(ctx context.Context) EvernoteStatus {
	return EvernoteStatus{
		Available: false,
		Running:   false,
		Version:   EvernoteVersionUnknown,
		Error:     "Evernote local import is only supported on macOS",
	}
}

// ListNotebooks returns ErrNotSupported on non-macOS platforms.
func (e *AppleScriptExecutor) ListNotebooks(ctx context.Context) ([]Notebook, error) {
	return nil, ErrNotSupported
}

// ListNotesInNotebook returns ErrNotSupported on non-macOS platforms.
func (e *AppleScriptExecutor) ListNotesInNotebook(ctx context.Context, notebookName string) ([]NoteSummary, error) {
	return nil, ErrNotSupported
}

// GetNoteContent returns ErrNotSupported on non-macOS platforms.
func (e *AppleScriptExecutor) GetNoteContent(ctx context.Context, noteLink string) (*NoteContent, error) {
	return nil, ErrNotSupported
}
