//go:build darwin

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
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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

	// Field separator used in AppleScript output.
	fieldSeparator = "||"

	// Record separator used in AppleScript output.
	recordSeparator = "|||RECORD|||"

	// Content separator used in AppleScript output.
	contentSeparator = "|||SEPARATOR|||"
)

// AppleScriptExecutor handles AppleScript execution for Evernote.
type AppleScriptExecutor struct {
	timeout time.Duration
}

// NewAppleScriptExecutor creates a new AppleScriptExecutor with the default
// timeout.
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

// executeScript runs an AppleScript and returns the output.
func (e *AppleScriptExecutor) executeScript(ctx context.Context, script string, timeout time.Duration) (string, error) {
	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "osascript", "-e", script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("AppleScript execution timed out after %v", timeout)
		}
		// Include stderr in error message if available
		errMsg := stderr.String()
		if errMsg != "" {
			return "", fmt.Errorf("AppleScript error: %s", strings.TrimSpace(errMsg))
		}
		return "", fmt.Errorf("AppleScript execution failed: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CheckEvernoteStatus checks if Evernote is installed and running.
func (e *AppleScriptExecutor) CheckEvernoteStatus(ctx context.Context) EvernoteStatus {
	// Check if Evernote is installed
	checkInstalledScript := `
tell application "System Events"
	set appExists to exists (application process "Evernote")
	if not appExists then
		try
			set appPath to (POSIX path of (path to application "Evernote"))
			return "installed"
		on error
			return "not_installed"
		end try
	else
		return "running"
	end if
end tell
`
	output, err := e.executeScript(ctx, checkInstalledScript, 10*time.Second)
	if err != nil {
		return EvernoteStatus{
			Available: false,
			Running:   false,
			Error:     fmt.Sprintf("Failed to check Evernote status: %v", err),
		}
	}

	switch output {
	case "running":
		return EvernoteStatus{
			Available: true,
			Running:   true,
		}
	case "installed":
		return EvernoteStatus{
			Available: true,
			Running:   false,
			Error:     "Evernote is installed but not running. Please start Evernote first.",
		}
	default:
		return EvernoteStatus{
			Available: false,
			Running:   false,
			Error:     "Evernote is not installed on this system.",
		}
	}
}

// ListNotebooks retrieves all notebooks from Evernote.
func (e *AppleScriptExecutor) ListNotebooks(ctx context.Context) ([]Notebook, error) {
	script := `
tell application "Evernote"
	set nbList to {}
	repeat with nb in notebooks
		set nbName to name of nb
		set nbCount to count of notes of nb
		set end of nbList to (nbName & "||" & nbCount)
	end repeat
	set AppleScript's text item delimiters to "|||RECORD|||"
	return nbList as string
end tell
`
	output, err := e.executeScript(ctx, script, ListNotebooksTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list notebooks: %w", err)
	}

	if output == "" {
		return []Notebook{}, nil
	}

	notebooks := []Notebook{}
	records := strings.Split(output, recordSeparator)

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		fields := strings.Split(record, fieldSeparator)
		if len(fields) != 2 {
			continue
		}

		count, err := strconv.Atoi(strings.TrimSpace(fields[1]))
		if err != nil {
			count = 0
		}

		notebooks = append(notebooks, Notebook{
			Name:      strings.TrimSpace(fields[0]),
			NoteCount: count,
		})
	}

	return notebooks, nil
}

// ListNotesInNotebook retrieves all notes from a specific notebook.
func (e *AppleScriptExecutor) ListNotesInNotebook(ctx context.Context, notebookName string) ([]NoteSummary, error) {
	// Escape quotes in notebook name for AppleScript
	escapedName := strings.ReplaceAll(notebookName, "\"", "\\\"")

	script := fmt.Sprintf(`
tell application "Evernote"
	set noteList to {}
	set nb to notebook "%s"
	repeat with n in notes of nb
		set noteLink to note link of n
		set noteTitle to title of n
		set creationDate to creation date of n as string
		set modDate to modification date of n as string
		set tagList to {}
		repeat with t in tags of n
			set end of tagList to name of t
		end repeat
		set AppleScript's text item delimiters to ","
		set tagStr to tagList as string
		set noteInfo to noteLink & "||" & noteTitle & "||" & creationDate & "||" & modDate & "||" & tagStr
		set end of noteList to noteInfo
	end repeat
	set AppleScript's text item delimiters to "|||RECORD|||"
	return noteList as string
end tell
`, escapedName)

	output, err := e.executeScript(ctx, script, ListNotesTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes in notebook %q: %w", notebookName, err)
	}

	if output == "" {
		return []NoteSummary{}, nil
	}

	notes := []NoteSummary{}
	records := strings.Split(output, recordSeparator)

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		fields := strings.Split(record, fieldSeparator)
		if len(fields) < 4 {
			continue
		}

		summary := NoteSummary{
			NoteLink: strings.TrimSpace(fields[0]),
			Title:    strings.TrimSpace(fields[1]),
			Created:  parseAppleScriptDate(strings.TrimSpace(fields[2])),
			Modified: parseAppleScriptDate(strings.TrimSpace(fields[3])),
		}

		// Parse tags if present
		if len(fields) >= 5 && fields[4] != "" {
			tagStr := strings.TrimSpace(fields[4])
			if tagStr != "" {
				tags := strings.Split(tagStr, ",")
				for _, tag := range tags {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						summary.TagNames = append(summary.TagNames, tag)
					}
				}
			}
		}

		notes = append(notes, summary)
	}

	return notes, nil
}

// GetNoteContent retrieves the full content of a note by its note link.
func (e *AppleScriptExecutor) GetNoteContent(ctx context.Context, noteLink string) (*NoteContent, error) {
	// Escape quotes in note link for AppleScript
	escapedLink := strings.ReplaceAll(noteLink, "\"", "\\\"")

	script := fmt.Sprintf(`
tell application "Evernote"
	set n to find note "%s"
	set noteTitle to title of n
	set htmlContent to HTML content of n
	set creationDate to creation date of n as string
	set modDate to modification date of n as string
	set tagList to {}
	repeat with t in tags of n
		set end of tagList to name of t
	end repeat
	set AppleScript's text item delimiters to ","
	set tagStr to tagList as string
	return noteTitle & "|||SEPARATOR|||" & creationDate & "|||SEPARATOR|||" & modDate & "|||SEPARATOR|||" & tagStr & "|||SEPARATOR|||" & htmlContent
end tell
`, escapedLink)

	output, err := e.executeScript(ctx, script, GetNoteContentTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to get note content: %w", err)
	}

	parts := strings.SplitN(output, contentSeparator, 5)
	if len(parts) < 5 {
		return nil, fmt.Errorf("unexpected response format from Evernote")
	}

	content := &NoteContent{
		Title:       strings.TrimSpace(parts[0]),
		Created:     parseAppleScriptDate(strings.TrimSpace(parts[1])),
		Modified:    parseAppleScriptDate(strings.TrimSpace(parts[2])),
		HTMLContent: strings.TrimSpace(parts[4]),
	}

	// Parse tags
	tagStr := strings.TrimSpace(parts[3])
	if tagStr != "" {
		tags := strings.Split(tagStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				content.Tags = append(content.Tags, tag)
			}
		}
	}

	return content, nil
}

// parseAppleScriptDate parses a date string from AppleScript output.
// AppleScript dates are typically in the format "day, month date, year at time"
// or similar locale-dependent formats.
func parseAppleScriptDate(dateStr string) time.Time {
	// Try multiple date formats that AppleScript might return
	formats := []string{
		"Monday, January 2, 2006 at 3:04:05 PM",
		"January 2, 2006 at 3:04:05 PM",
		"2006-01-02 15:04:05",
		"1/2/2006, 3:04:05 PM",
		"2/1/2006 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	// Try parsing with time.ParseInLocation using local timezone
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, time.Local); err == nil {
			return t
		}
	}

	// Return zero time if parsing fails
	return time.Time{}
}
