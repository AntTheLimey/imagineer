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
