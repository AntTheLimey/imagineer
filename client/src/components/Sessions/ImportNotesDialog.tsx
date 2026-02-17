// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ImportNotesDialog - Import notes into a session via paste or file upload.
 *
 * Provides a dialog for pasting text or uploading .txt/.md files,
 * with a toggle for append vs. replace mode.
 */

import { useState, useRef, type ChangeEvent } from 'react';
import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControlLabel,
    Switch,
    TextField,
} from '@mui/material';
import { UploadFile as UploadFileIcon } from '@mui/icons-material';

/**
 * Props for the ImportNotesDialog component.
 */
export interface ImportNotesDialogProps {
    /** Whether the dialog is open. */
    open: boolean;
    /** Callback fired when the dialog should close. */
    onClose: () => void;
    /** Callback fired when the user confirms the import. */
    onImport: (content: string, mode: 'append' | 'replace') => void;
}

/**
 * Dialog component for importing notes into a session.
 *
 * Users can paste text directly into a multiline text field or upload
 * a `.txt` or `.md` file. A switch controls whether the imported
 * content appends to or replaces the existing notes.
 *
 * @param props - The component props.
 * @returns A React element containing the import notes dialog.
 *
 * @example
 * ```tsx
 * <ImportNotesDialog
 *     open={isImportOpen}
 *     onClose={() => setIsImportOpen(false)}
 *     onImport={(content, mode) => handleImport(content, mode)}
 * />
 * ```
 */
export function ImportNotesDialog({
    open,
    onClose,
    onImport,
}: ImportNotesDialogProps) {
    const [content, setContent] = useState('');
    const [replaceMode, setReplaceMode] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    const mode: 'append' | 'replace' = replaceMode ? 'replace' : 'append';

    /**
     * Resets the dialog state and closes it.
     */
    const handleClose = () => {
        setContent('');
        setReplaceMode(false);
        if (fileInputRef.current) {
            fileInputRef.current.value = '';
        }
        onClose();
    };

    /**
     * Handles the import action, passing content and mode to the caller.
     */
    const handleImport = () => {
        onImport(content, mode);
        handleClose();
    };

    /**
     * Opens the native file picker by clicking the hidden file input.
     */
    const handleUploadClick = () => {
        fileInputRef.current?.click();
    };

    /**
     * Reads the selected file and populates the text field with its
     * contents.
     */
    const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (e) => {
            const text = e.target?.result;
            if (typeof text === 'string') {
                setContent(text);
            }
        };
        reader.onerror = () => {
            console.error('Failed to read file:', reader.error);
        };
        reader.readAsText(file);
    };

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            maxWidth="sm"
            fullWidth
            aria-labelledby="import-notes-dialog-title"
        >
            <DialogTitle id="import-notes-dialog-title">
                Import Notes
            </DialogTitle>
            <DialogContent>
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 2,
                        mt: 1,
                    }}
                >
                    <TextField
                        multiline
                        rows={10}
                        fullWidth
                        placeholder="Paste your notes here..."
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        aria-label="Notes content"
                    />

                    <Box>
                        <Button
                            variant="outlined"
                            startIcon={<UploadFileIcon />}
                            onClick={handleUploadClick}
                        >
                            Upload File
                        </Button>
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept=".txt,.md"
                            onChange={handleFileChange}
                            hidden
                        />
                    </Box>

                    <FormControlLabel
                        control={
                            <Switch
                                checked={replaceMode}
                                onChange={(e) =>
                                    setReplaceMode(e.target.checked)
                                }
                            />
                        }
                        label={
                            replaceMode
                                ? 'Replace existing notes'
                                : 'Append to existing notes'
                        }
                    />
                </Box>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose} color="inherit">
                    Cancel
                </Button>
                <Button
                    onClick={handleImport}
                    variant="contained"
                    disabled={!content.trim()}
                >
                    Import
                </Button>
            </DialogActions>
        </Dialog>
    );
}

export default ImportNotesDialog;
