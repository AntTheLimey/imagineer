// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterEditor - Create or edit a chapter.
 *
 * Features:
 * - Title field (required)
 * - Overview rich text field
 * - Sort order field
 * - Save/Cancel buttons
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
    TextField,
} from '@mui/material';
import {
    useChapter,
    useCreateChapter,
    useUpdateChapter,
} from '../../hooks/useChapters';
import type { Chapter } from '../../types';

/**
 * Props for the ChapterEditor component.
 */
export interface ChapterEditorProps {
    /** The campaign ID. */
    campaignId: number;
    /** The chapter ID to edit (if undefined, creates a new chapter). */
    chapterId?: number;
    /** Callback fired when the chapter is successfully saved. */
    onSave?: (chapter: Chapter) => void;
    /** Callback fired when the user cancels editing. */
    onCancel?: () => void;
    /** Whether the dialog is open. */
    open: boolean;
    /** Callback fired when the dialog should close. */
    onClose: () => void;
}

/**
 * Form data for creating or editing a chapter.
 */
interface ChapterFormData {
    title: string;
    overview: string;
    sortOrder: number;
}

/**
 * Initial form data for a new chapter.
 */
const initialFormData: ChapterFormData = {
    title: '',
    overview: '',
    sortOrder: 0,
};

/**
 * Dialog component for creating or editing a chapter.
 *
 * Provides form fields for title, overview, and sort order with
 * validation and error handling.
 *
 * @param props - The component props.
 * @returns A React element containing the chapter editor dialog.
 *
 * @example
 * ```tsx
 * // Create mode:
 * <ChapterEditor
 *     campaignId={campaignId}
 *     open={isOpen}
 *     onClose={() => setIsOpen(false)}
 *     onSave={handleSave}
 * />
 *
 * // Edit mode:
 * <ChapterEditor
 *     campaignId={campaignId}
 *     chapterId={selectedChapterId}
 *     open={isOpen}
 *     onClose={() => setIsOpen(false)}
 *     onSave={handleSave}
 * />
 * ```
 */
export default function ChapterEditor({
    campaignId,
    chapterId,
    onSave,
    onCancel,
    open,
    onClose,
}: ChapterEditorProps) {
    const isEditMode = !!chapterId;

    const [formData, setFormData] = useState<ChapterFormData>(initialFormData);
    const [error, setError] = useState<string | null>(null);
    const [validationError, setValidationError] = useState<string | null>(null);

    // Fetch existing chapter data for edit mode
    const {
        data: existingChapter,
        isLoading: isLoadingChapter,
        error: fetchError,
    } = useChapter(campaignId, chapterId ?? 0, {
        enabled: isEditMode && open,
    });

    const createChapterMutation = useCreateChapter();
    const updateChapterMutation = useUpdateChapter();

    const isSaving =
        createChapterMutation.isPending || updateChapterMutation.isPending;

    /**
     * Resets the form to its initial state.
     */
    const resetForm = useCallback(() => {
        setFormData(initialFormData);
        setError(null);
        setValidationError(null);
    }, []);

    /**
     * Populates the form with existing chapter data.
     */
    useEffect(() => {
        if (isEditMode && existingChapter) {
            setFormData({
                title: existingChapter.title,
                overview: existingChapter.overview ?? '',
                sortOrder: existingChapter.sortOrder,
            });
        } else if (!isEditMode && open) {
            resetForm();
        }
    }, [isEditMode, existingChapter, open, resetForm]);

    /**
     * Handles input changes.
     */
    const handleInputChange = (
        event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
    ) => {
        const { name, value } = event.target;

        setFormData((prev) => ({
            ...prev,
            [name]: name === 'sortOrder' ? parseInt(value, 10) || 0 : value,
        }));

        // Clear validation error when user types
        if (name === 'title' && validationError) {
            setValidationError(null);
        }
    };

    /**
     * Validates the form data.
     */
    const validateForm = (): boolean => {
        if (!formData.title.trim()) {
            setValidationError('Title is required');
            return false;
        }

        return true;
    };

    /**
     * Handles form submission.
     */
    const handleSubmit = async () => {
        if (!validateForm()) return;

        setError(null);

        try {
            let savedChapter: Chapter;

            if (isEditMode && chapterId) {
                savedChapter = await updateChapterMutation.mutateAsync({
                    campaignId,
                    chapterId,
                    input: {
                        title: formData.title.trim(),
                        overview: formData.overview.trim() || undefined,
                        sortOrder: formData.sortOrder,
                    },
                });
            } else {
                savedChapter = await createChapterMutation.mutateAsync({
                    campaignId,
                    input: {
                        title: formData.title.trim(),
                        overview: formData.overview.trim() || undefined,
                        sortOrder: formData.sortOrder,
                    },
                });
            }

            onSave?.(savedChapter);
            handleClose();
        } catch (err) {
            const message =
                err instanceof Error
                    ? err.message
                    : 'Failed to save chapter';
            setError(message);
        }
    };

    /**
     * Handles dialog close.
     */
    const handleClose = () => {
        if (!isSaving) {
            resetForm();
            onCancel?.();
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

    const dialogTitle = isEditMode ? 'Edit Chapter' : 'Create Chapter';

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            maxWidth="sm"
            fullWidth
            aria-labelledby="chapter-editor-dialog-title"
        >
            <DialogTitle id="chapter-editor-dialog-title">
                {dialogTitle}
            </DialogTitle>
            <DialogContent>
                {/* Loading state for edit mode */}
                {isEditMode && isLoadingChapter && (
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
                        Failed to load chapter: {fetchError.message}
                    </Alert>
                )}

                {/* Save error */}
                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}

                {/* Form fields */}
                {(!isEditMode || !isLoadingChapter) && !fetchError && (
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
                            required
                            fullWidth
                            autoFocus
                            error={!!validationError}
                            helperText={validationError}
                            disabled={isSaving}
                            placeholder="Enter chapter title"
                        />

                        <TextField
                            name="overview"
                            label="Overview"
                            value={formData.overview}
                            onChange={handleInputChange}
                            multiline
                            rows={4}
                            fullWidth
                            disabled={isSaving}
                            placeholder="Enter a brief overview of this chapter (optional)"
                            helperText="A summary of the story arc or events in this chapter"
                        />

                        <TextField
                            name="sortOrder"
                            label="Sort Order"
                            type="number"
                            value={formData.sortOrder}
                            onChange={handleInputChange}
                            fullWidth
                            disabled={isSaving}
                            helperText="Lower numbers appear first in the chapter list"
                            inputProps={{ min: 0 }}
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
                    disabled={isSaving || (isEditMode && isLoadingChapter)}
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
