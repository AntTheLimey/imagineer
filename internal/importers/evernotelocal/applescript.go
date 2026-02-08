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
	// Using ASCII Unit Separator (0x1F) to avoid collision with user content.
	fieldSeparator = "\x1F"

	// Record separator used in AppleScript output.
	// Using ASCII Record Separator (0x1E) to avoid collision with user content.
	recordSeparator = "\x1E"

	// Content separator used in AppleScript output.
	// Using ASCII Group Separator (0x1D) to avoid collision with user content.
	contentSeparator = "\x1D"
)

// escapeForAppleScript escapes a string for safe use in AppleScript string literals.
// Backslashes must be escaped first, then double quotes.
func escapeForAppleScript(s string) string {
	result := strings.ReplaceAll(s, "\\", "\\\\")
	result = strings.ReplaceAll(result, "\"", "\\\"")
	return result
}

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

// CheckEvernoteStatus checks if Evernote is installed and running.
// It also detects whether the installed version is Evernote Legacy (7.x)
// or Evernote 10.x, as only Legacy supports the required AppleScript commands.
func (e *AppleScriptExecutor) CheckEvernoteStatus(ctx context.Context) EvernoteStatus {
	// Check if Evernote is installed and get its status
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
			Version:   EvernoteVersionUnknown,
			Error:     fmt.Sprintf("Failed to check Evernote status: %v", err),
		}
	}

	switch output {
	case "running":
		// Evernote is running, now check the version
		version := e.detectEvernoteVersion(ctx)
		status := EvernoteStatus{
			Available: true,
			Running:   true,
			Version:   version,
		}
		if version == EvernoteVersion10 {
			status.Error = Evernote10Error
		}
		return status

	case "installed":
		// Evernote is installed but not running
		// We can still try to detect version from the app bundle
		version := e.detectEvernoteVersionFromBundle(ctx)
		status := EvernoteStatus{
			Available: true,
			Running:   false,
			Version:   version,
		}
		if version == EvernoteVersion10 {
			status.Error = Evernote10Error
		} else {
			status.Error = "Evernote is installed but not running. Please start Evernote first."
		}
		return status

	default:
		return EvernoteStatus{
			Available: false,
			Running:   false,
			Version:   EvernoteVersionUnknown,
			Error:     "Evernote is not installed on this system.",
		}
	}
}

// detectEvernoteVersion detects whether the running Evernote is Legacy or 10.x
// by attempting to access the notebooks property, which only exists in Legacy.
func (e *AppleScriptExecutor) detectEvernoteVersion(ctx context.Context) EvernoteVersion {
	// Try to access the notebooks property which exists in Legacy but not in 10.x
	// If it fails with "notebooks is not defined", we're on 10.x
	versionScript := `
tell application "Evernote"
	try
		set nbCount to count of notebooks
		return "legacy"
	on error errMsg
		if errMsg contains "is not defined" or errMsg contains "doesn't understand" then
			return "10.x"
		else
			return "unknown"
		end if
	end try
end tell
`
	output, err := e.executeScript(ctx, versionScript, 10*time.Second)
	if err != nil {
		// If the script fails entirely, try to detect from the error
		errStr := err.Error()
		if strings.Contains(errStr, "is not defined") || strings.Contains(errStr, "doesn't understand") {
			return EvernoteVersion10
		}
		return EvernoteVersionUnknown
	}

	switch output {
	case "legacy":
		return EvernoteVersionLegacy
	case "10.x":
		return EvernoteVersion10
	default:
		return EvernoteVersionUnknown
	}
}

// detectEvernoteVersionFromBundle attempts to detect Evernote version from the
// application bundle's Info.plist when Evernote is not running.
func (e *AppleScriptExecutor) detectEvernoteVersionFromBundle(ctx context.Context) EvernoteVersion {
	// Get version from the app bundle's Info.plist
	versionScript := `
try
	set appPath to (POSIX path of (path to application "Evernote"))
	set plistPath to appPath & "Contents/Info.plist"
	set versionStr to do shell script "defaults read " & quoted form of plistPath & " CFBundleShortVersionString"
	return versionStr
on error
	return "unknown"
end try
`
	output, err := e.executeScript(ctx, versionScript, 10*time.Second)
	if err != nil || output == "unknown" {
		return EvernoteVersionUnknown
	}

	// Parse version string - Evernote Legacy is 7.x, Evernote 10 is 10.x
	output = strings.TrimSpace(output)
	if strings.HasPrefix(output, "7.") || strings.HasPrefix(output, "6.") {
		return EvernoteVersionLegacy
	}
	if strings.HasPrefix(output, "10.") {
		return EvernoteVersion10
	}

	return EvernoteVersionUnknown
}

// ErrEvernote10Unsupported is returned when attempting to use features that
// require Evernote Legacy on Evernote 10.x.
var ErrEvernote10Unsupported = fmt.Errorf("evernote 10.x does not support this operation")

// ListNotebooks retrieves all notebooks from Evernote.
// This function requires Evernote Legacy (7.x) and will return an error
// with helpful guidance if Evernote 10.x is detected.
func (e *AppleScriptExecutor) ListNotebooks(ctx context.Context) ([]Notebook, error) {
	script := `
tell application "Evernote"
	set nbList to {}
	repeat with nb in notebooks
		set nbName to name of nb
		set nbCount to count of notes of nb
		set end of nbList to (nbName & (ASCII character 31) & nbCount)
	end repeat
	set AppleScript's text item delimiters to (ASCII character 30)
	return nbList as string
end tell
`
	output, err := e.executeScript(ctx, script, ListNotebooksTimeout)
	if err != nil {
		// Check if this is an Evernote 10.x compatibility error
		errStr := err.Error()
		if strings.Contains(errStr, "notebooks is not defined") ||
			strings.Contains(errStr, "variable notebooks is not defined") ||
			strings.Contains(errStr, "doesn't understand") {
			return nil, fmt.Errorf("%w: %s", ErrEvernote10Unsupported, Evernote10Error)
		}
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
	// Escape for safe AppleScript string literal
	escapedName := escapeForAppleScript(notebookName)

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
		set noteInfo to noteLink & (ASCII character 31) & noteTitle & (ASCII character 31) & creationDate & (ASCII character 31) & modDate & (ASCII character 31) & tagStr
		set end of noteList to noteInfo
	end repeat
	set AppleScript's text item delimiters to (ASCII character 30)
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
	// Escape for safe AppleScript string literal
	escapedLink := escapeForAppleScript(noteLink)

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
	return noteTitle & (ASCII character 29) & creationDate & (ASCII character 29) & modDate & (ASCII character 29) & tagStr & (ASCII character 29) & htmlContent
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
