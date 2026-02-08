// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * SessionEditor - Create or edit a session.
 *
 * Fields:
 * - Title (optional)
 * - Date picker for planned date
 * - Chapter selector (dropdown)
 */

import { useState, useEffect, useCallback } from 'react';
import {
    Alert,
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    InputLabel,
    MenuItem,
    Select,
    TextField,
    type SelectChangeEvent,
} from '@mui/material';
import {
    useSession,
    useCreateSession,
    useUpdateSession,
} from '../../hooks/useSessions';
import { useChapters } from '../../hooks/useChapters';
import type { Session } from '../../types';

/**
 * Props for the SessionEditor component.
 */
export interface SessionEditorProps {
    /** The campaign ID. */
    campaignId: number;
    /** Pre-selected chapter ID (optional). */
    chapterId?: number;
    /** The session ID to edit (if undefined, creates a new session). */
    sessionId?: number;
    /** Whether the dialog is open. */
    open: boolean;
    /** Callback fired when the dialog should close. */
    onClose: () => void;
    /** Callback fired when the session is successfully saved. */
    onSave?: (session: Session) => void;
}

/**
 * Form data for creating or editing a session.
 */
interface SessionFormData {
    title: string;
    chapterId: string;
    plannedDate: string;
}

/**
 * Initial form data for a new session.
 */
const initialFormData: SessionFormData = {
    title: '',
    chapterId: '',
    plannedDate: '',
};

/**
 * Convert an ISO or full-date string to YYYY-MM-DD suitable for a date input.
 *
 * @param dateString - An ISO datetime or date string (e.g., "2023-05-01T12:34:56Z" or "2023-05-01"); may be undefined
 * @returns The date portion in `YYYY-MM-DD` format, or an empty string if `dateString` is undefined or falsy
 */
function formatDateForInput(dateString?: string): string {
    if (!dateString) return '';
    // Handle both ISO format and YYYY-MM-DD format
    return dateString.split('T')[0];
}

/**
 * Dialog component for creating or editing a session.
 *
 * Provides form fields for title, planned date, and chapter selection
 * with validation and error handling.
 *
 * @param props - The component props.
 * @returns A React element containing the session editor dialog.
 *
 * @example
 * ```tsx
 * // Create mode:
 * <SessionEditor
 *     campaignId={campaignId}
 *     chapterId={selectedChapterId}
 *     open={isOpen}
 *     onClose={() => setIsOpen(false)}
 *     onSave={handleSave}
 * />
 *
 * // Edit mode:
 * <SessionEditor
 *     campaignId={campaignId}
 *     sessionId={selectedSessionId}
 *     open={isOpen}
 *     onClose={() => setIsOpen(false)}
 *     onSave={handleSave}
 * />
 * ```
 */
export default function SessionEditor({
    campaignId,
    chapterId: preselectedChapterId,
    sessionId,
    open,
    onClose,
    onSave,
}: SessionEditorProps) {
    const isEditMode = !!sessionId;

    const [formData, setFormData] = useState<SessionFormData>(initialFormData);
    const [error, setError] = useState<string | null>(null);

    // Fetch existing session data for edit mode
    const {
        data: existingSession,
        isLoading: isLoadingSession,
        error: fetchError,
    } = useSession(campaignId, sessionId ?? 0, {
        enabled: isEditMode && open,
    });

    // Fetch chapters for the dropdown
    const { data: chapters, isLoading: isLoadingChapters } =
        useChapters(campaignId);

    const createSessionMutation = useCreateSession();
    const updateSessionMutation = useUpdateSession();

    const isSaving =
        createSessionMutation.isPending || updateSessionMutation.isPending;

    /**
     * Resets the form to its initial state.
     */
    const resetForm = useCallback(() => {
        setFormData({
            ...initialFormData,
            chapterId: preselectedChapterId ? String(preselectedChapterId) : '',
        });
        setError(null);
    }, [preselectedChapterId]);

    /**
     * Populates the form with existing session data.
     */
    useEffect(() => {
        if (isEditMode && existingSession) {
            setFormData({
                title: existingSession.title ?? '',
                chapterId: existingSession.chapterId ? String(existingSession.chapterId) : '',
                plannedDate: formatDateForInput(existingSession.plannedDate),
            });
        } else if (!isEditMode && open) {
            resetForm();
        }
    }, [isEditMode, existingSession, open, resetForm]);

    /**
     * Handles text input changes.
     */
    const handleInputChange = (
        event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
    ) => {
        const { name, value } = event.target;
        setFormData((prev) => ({
            ...prev,
            [name]: value,
        }));
    };

    /**
     * Handles select input changes.
     */
    const handleSelectChange = (event: SelectChangeEvent<string>) => {
        const { name, value } = event.target;
        setFormData((prev) => ({
            ...prev,
            [name]: value,
        }));
    };

    /**
     * Handles form submission.
     */
    const handleSubmit = async () => {
        setError(null);

        try {
            const input = {
                title: formData.title.trim() || undefined,
                chapterId: formData.chapterId ? Number(formData.chapterId) : undefined,
                plannedDate: formData.plannedDate || undefined,
            };

            let savedSession: Session;

            if (isEditMode && sessionId) {
                savedSession = await updateSessionMutation.mutateAsync({
                    campaignId,
                    sessionId,
                    input,
                });
            } else {
                savedSession = await createSessionMutation.mutateAsync({
                    campaignId,
                    input,
                });
            }

            onSave?.(savedSession);
            handleClose();
        } catch (err) {
            const message =
                err instanceof Error
                    ? err.message
                    : 'Failed to save session';
            setError(message);
        }
    };

    /**
     * Handles dialog close.
     */
    const handleClose = () => {
        if (!isSaving) {
            resetForm();
            onClose();
        }
    };

    /**
     * Handles key press events for form submission.
     */
    const handleKeyPress = (event: React.KeyboardEvent) => {
        if (event.key === 'Enter' && event.ctrlKey) {
            handleSubmit();
        }
    };

    const dialogTitle = isEditMode ? 'Edit Session' : 'Create Session';
    const sortedChapters = chapters
        ? [...chapters].sort((a, b) => a.sortOrder - b.sortOrder)
        : [];

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            maxWidth="sm"
            fullWidth
            aria-labelledby="session-editor-dialog-title"
        >
            <DialogTitle id="session-editor-dialog-title">
                {dialogTitle}
            </DialogTitle>
            <DialogContent>
                {/* Loading state for edit mode */}
                {isEditMode && isLoadingSession && (
                    <Box
                        sx={{
                            display: 'flex',
                            justifyContent: 'center',
                            py: 4,
                        }}
                    >
                        <CircularProgress />
                    </Box>
                )}

                {/* Fetch error */}
                {fetchError && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        Failed to load session: {fetchError.message}
                    </Alert>
                )}

                {/* Save error */}
                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}

                {/* Form fields */}
                {(!isEditMode || !isLoadingSession) && !fetchError && (
                    <Box
                        component="form"
                        onKeyDown={handleKeyPress}
                        sx={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 2,
                            mt: 1,
                        }}
                    >
                        <TextField
                            name="title"
                            label="Title"
                            value={formData.title}
                            onChange={handleInputChange}
                            fullWidth
                            autoFocus
                            disabled={isSaving}
                            placeholder="Enter session title (optional)"
                            helperText="Leave blank to use session number"
                        />

                        <FormControl fullWidth disabled={isSaving || isLoadingChapters}>
                            <InputLabel id="chapter-select-label">
                                Chapter
                            </InputLabel>
                            <Select
                                labelId="chapter-select-label"
                                name="chapterId"
                                value={formData.chapterId}
                                onChange={handleSelectChange}
                                label="Chapter"
                            >
                                <MenuItem value="">
                                    <em>No chapter</em>
                                </MenuItem>
                                {sortedChapters.map((chapter) => (
                                    <MenuItem
                                        key={chapter.id}
                                        value={String(chapter.id)}
                                    >
                                        {chapter.title}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <TextField
                            name="plannedDate"
                            label="Planned Date"
                            type="date"
                            value={formData.plannedDate}
                            onChange={handleInputChange}
                            fullWidth
                            disabled={isSaving}
                            helperText="When do you plan to run this session?"
                            InputLabelProps={{
                                shrink: true,
                            }}
                        />
                    </Box>
                )}
            </DialogContent>
            <DialogActions>
                <Button
                    onClick={handleClose}
                    disabled={isSaving}
                    color="inherit"
                >
                    Cancel
                </Button>
                <Button
                    onClick={handleSubmit}
                    disabled={isSaving || (isEditMode && isLoadingSession)}
                    variant="contained"
                    startIcon={
                        isSaving ? <CircularProgress size={16} /> : null
                    }
                >
                    {isSaving ? 'Saving...' : 'Save'}
                </Button>
            </DialogActions>
        </Dialog>
    );
}