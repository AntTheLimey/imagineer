// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterList - Displays chapters for a campaign with CRUD operations.
 *
 * Features:
 * - List chapters in sort order
 * - Click row to select chapter (shows sessions)
 * - Click title to navigate to chapter view page
 * - Show sessions count per chapter
 * - Add new chapter button
 * - Delete chapter action
 */

import type { MouseEvent } from 'react';
import { useState, useCallback } from 'react';
import {
    Alert,
    Box,
    Button,
    CircularProgress,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemText,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
} from '@mui/icons-material';
import { useChapters, useDeleteChapter } from '../../hooks/useChapters';
import { useSessionsByChapter } from '../../hooks/useSessions';
import { useDraftIndicators } from '../../hooks/useDraftIndicators';
import { DraftIndicator } from '../DraftIndicator';
import type { Chapter } from '../../types';

/**
 * Props for the ChapterList component.
 */
export interface ChapterListProps {
    /** The campaign ID to list chapters for. */
    campaignId: number;
    /** The currently selected chapter ID. */
    selectedChapterId?: number;
    /** Callback fired when a chapter is selected. */
    onSelectChapter: (chapterId: number) => void;
    /** Optional callback to open the create chapter dialog. */
    onCreateChapter?: () => void;
    /** Optional callback to open the edit chapter dialog. */
    onEditChapter?: (chapter: Chapter) => void;
}

/**
 * Props for the ChapterListItem component.
 */
interface ChapterListItemProps {
    /** The chapter to display. */
    chapter: Chapter;
    /** The campaign ID. */
    campaignId: number;
    /** Whether this chapter is selected. */
    isSelected: boolean;
    /** Callback fired when the chapter is clicked. */
    onSelect: () => void;
    /** Callback fired when the title is clicked (navigates to view page). */
    onEdit: () => void;
    /** Callback fired when the delete button is clicked. */
    onDelete: () => void;
    /** Whether a delete operation is in progress. */
    isDeleting: boolean;
    /** Whether a draft exists for this chapter. */
    hasDraft: boolean;
}

/**
 * Render a single chapter row with selection, title navigation, and delete control.
 *
 * Clicking the row selects the chapter (showing its sessions). Clicking the
 * chapter title navigates to the chapter view page.
 *
 * @param chapter - The chapter to display.
 * @param campaignId - Campaign ID used to fetch this chapter's sessions for the displayed session count.
 * @param isSelected - Whether the chapter is currently selected.
 * @param onSelect - Callback invoked when the main list item is clicked.
 * @param onEdit - Callback invoked when the chapter title is clicked (navigates to view page).
 * @param onDelete - Callback invoked when the delete action is triggered.
 * @param isDeleting - True while a delete operation for this chapter is in progress (shows a spinner).
 * @param hasDraft - Whether a draft exists for this chapter.
 */
function ChapterListItem({
    chapter,
    campaignId,
    isSelected,
    onSelect,
    onEdit,
    onDelete,
    isDeleting,
    hasDraft,
}: ChapterListItemProps) {
    const { data: sessions } = useSessionsByChapter(campaignId, chapter.id);
    const sessionCount = sessions?.length ?? 0;

    const handleDeleteClick = (event: MouseEvent) => {
        event.stopPropagation();
        onDelete();
    };

    return (
        <ListItem
            disablePadding
            secondaryAction={
                <Tooltip title="Delete chapter">
                    <IconButton
                        edge="end"
                        size="small"
                        onClick={handleDeleteClick}
                        disabled={isDeleting}
                        aria-label="delete chapter"
                    >
                        {isDeleting ? (
                            <CircularProgress size={18} />
                        ) : (
                            <DeleteIcon fontSize="small" />
                        )}
                    </IconButton>
                </Tooltip>
            }
        >
            <ListItemButton
                selected={isSelected}
                onClick={onSelect}
                sx={{ pr: 7 }}
            >
                <ListItemText
                    primary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            <Typography
                                component="span"
                                noWrap
                                fontWeight={isSelected ? 600 : 400}
                                sx={{
                                    flexGrow: 1,
                                    minWidth: 0,
                                    cursor: 'pointer',
                                    '&:hover': { textDecoration: 'underline' },
                                }}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onEdit();
                                }}
                            >
                                {chapter.title}
                            </Typography>
                            <DraftIndicator hasDraft={hasDraft} />
                        </Box>
                    }
                    secondary={`${sessionCount} session${sessionCount !== 1 ? 's' : ''}`}
                    primaryTypographyProps={{
                        component: 'div',
                    }}
                />
            </ListItemButton>
        </ListItem>
    );
}

/**
 * Component for displaying and managing a list of chapters.
 *
 * Provides chapter selection, clickable titles for navigation to the
 * chapter view page, and actions for creating and deleting chapters.
 *
 * @param props - The component props.
 * @returns A React element containing the chapter list.
 *
 * @example
 * ```tsx
 * <ChapterList
 *     campaignId={campaignId}
 *     selectedChapterId={selectedChapterId}
 *     onSelectChapter={handleSelectChapter}
 *     onCreateChapter={() => setEditorOpen(true)}
 *     onEditChapter={handleEditChapter}
 * />
 * ```
 */
export default function ChapterList({
    campaignId,
    selectedChapterId,
    onSelectChapter,
    onCreateChapter,
    onEditChapter,
}: ChapterListProps) {
    const [deleteError, setDeleteError] = useState<string | null>(null);
    const { hasDraft } = useDraftIndicators(campaignId, 'chapters');

    const {
        data: chapters,
        isLoading,
        error: fetchError,
    } = useChapters(campaignId);

    const deleteChapterMutation = useDeleteChapter();

    /**
     * Handles deleting a chapter with confirmation.
     */
    const handleDeleteChapter = useCallback(
        async (chapter: Chapter) => {
            const confirmed = window.confirm(
                `Are you sure you want to delete the chapter "${chapter.title}"? ` +
                    'This action cannot be undone.'
            );

            if (!confirmed) return;

            setDeleteError(null);

            try {
                await deleteChapterMutation.mutateAsync({
                    campaignId,
                    chapterId: chapter.id,
                });
            } catch (err) {
                const message =
                    err instanceof Error
                        ? err.message
                        : 'Failed to delete chapter';
                setDeleteError(message);
            }
        },
        [campaignId, deleteChapterMutation]
    );

    // Sort chapters by sortOrder
    const sortedChapters = chapters
        ? [...chapters].sort((a, b) => a.sortOrder - b.sortOrder)
        : [];

    // Loading state
    if (isLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    py: 4,
                }}
            >
                <CircularProgress size={24} />
            </Box>
        );
    }

    // Error state
    if (fetchError) {
        return (
            <Alert severity="error" sx={{ m: 2 }}>
                Failed to load chapters: {fetchError.message}
            </Alert>
        );
    }

    return (
        <Box>
            {/* Header with add button */}
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    px: 2,
                    py: 1,
                    borderBottom: 1,
                    borderColor: 'divider',
                }}
            >
                <Typography variant="subtitle1" fontWeight={500}>
                    Chapters
                </Typography>
                {onCreateChapter && (
                    <Button
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={onCreateChapter}
                    >
                        Add
                    </Button>
                )}
            </Box>

            {/* Delete error alert */}
            {deleteError && (
                <Alert
                    severity="error"
                    sx={{ m: 1 }}
                    onClose={() => setDeleteError(null)}
                >
                    {deleteError}
                </Alert>
            )}

            {/* Chapter list */}
            {sortedChapters.length > 0 ? (
                <List disablePadding>
                    {sortedChapters.map((chapter) => (
                        <ChapterListItem
                            key={chapter.id}
                            chapter={chapter}
                            campaignId={campaignId}
                            isSelected={selectedChapterId === chapter.id}
                            hasDraft={hasDraft(chapter.id)}
                            onSelect={() => onSelectChapter(chapter.id)}
                            onEdit={() => onEditChapter?.(chapter)}
                            onDelete={() => handleDeleteChapter(chapter)}
                            isDeleting={
                                deleteChapterMutation.isPending &&
                                deleteChapterMutation.variables?.chapterId ===
                                    chapter.id
                            }
                        />
                    ))}
                </List>
            ) : (
                <Box sx={{ p: 2, textAlign: 'center' }}>
                    <Typography variant="body2" color="text.secondary">
                        No chapters yet. Click &quot;Add&quot; to create your
                        first chapter.
                    </Typography>
                </Box>
            )}
        </Box>
    );
}