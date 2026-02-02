// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ChapterCard - Compact chapter display with session count.
 *
 * A card component for displaying chapter information in the dashboard,
 * with hover effects and action buttons.
 */

import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    IconButton,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Delete as DeleteIcon,
    Edit as EditIcon,
} from '@mui/icons-material';
import type { Chapter } from '../../types';

/**
 * Props for the ChapterCard component.
 */
export interface ChapterCardProps {
    /** The chapter to display. */
    chapter: Chapter;
    /** The number of sessions in this chapter. */
    sessionCount: number;
    /** Whether this card is currently selected. */
    isSelected: boolean;
    /** Callback fired when the card is clicked. */
    onClick: () => void;
    /** Callback fired when the edit button is clicked. */
    onEdit: () => void;
    /** Callback fired when the delete button is clicked. */
    onDelete: () => void;
}

/**
 * Compact card component for displaying chapter summary.
 *
 * Displays the chapter title, session count, and overview excerpt
 * with hover effects and action buttons for editing and deleting.
 *
 * @param props - The component props.
 * @returns A React element containing the chapter card.
 *
 * @example
 * ```tsx
 * <ChapterCard
 *     chapter={chapter}
 *     sessionCount={3}
 *     isSelected={selectedChapterId === chapter.id}
 *     onClick={() => handleSelectChapter(chapter.id)}
 *     onEdit={() => handleEditChapter(chapter)}
 *     onDelete={() => handleDeleteChapter(chapter)}
 * />
 * ```
 */
export default function ChapterCard({
    chapter,
    sessionCount,
    isSelected,
    onClick,
    onEdit,
    onDelete,
}: ChapterCardProps) {
    /**
     * Handles edit button click without triggering card click.
     */
    const handleEditClick = (event: React.MouseEvent) => {
        event.stopPropagation();
        onEdit();
    };

    /**
     * Handles delete button click without triggering card click.
     */
    const handleDeleteClick = (event: React.MouseEvent) => {
        event.stopPropagation();
        onDelete();
    };

    return (
        <Card
            elevation={isSelected ? 4 : 1}
            sx={{
                position: 'relative',
                transition: 'all 0.2s ease-in-out',
                border: isSelected ? 2 : 1,
                borderColor: isSelected ? 'primary.main' : 'divider',
                '&:hover': {
                    elevation: 3,
                    transform: 'translateY(-2px)',
                    '& .chapter-card-actions': {
                        opacity: 1,
                    },
                },
            }}
        >
            {/* Action buttons - shown on hover */}
            <Box
                className="chapter-card-actions"
                sx={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    display: 'flex',
                    gap: 0.5,
                    opacity: 0,
                    transition: 'opacity 0.2s ease-in-out',
                    zIndex: 1,
                    backgroundColor: 'background.paper',
                    borderRadius: 1,
                    boxShadow: 1,
                }}
            >
                <Tooltip title="Edit chapter">
                    <IconButton
                        size="small"
                        onClick={handleEditClick}
                        aria-label="edit chapter"
                    >
                        <EditIcon fontSize="small" />
                    </IconButton>
                </Tooltip>
                <Tooltip title="Delete chapter">
                    <IconButton
                        size="small"
                        onClick={handleDeleteClick}
                        aria-label="delete chapter"
                    >
                        <DeleteIcon fontSize="small" />
                    </IconButton>
                </Tooltip>
            </Box>

            <CardActionArea onClick={onClick}>
                <CardContent>
                    {/* Chapter title */}
                    <Typography
                        variant="subtitle1"
                        component="h3"
                        fontWeight={isSelected ? 600 : 500}
                        noWrap
                        sx={{ pr: 8 }}
                    >
                        {chapter.title}
                    </Typography>

                    {/* Session count */}
                    <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mt: 0.5 }}
                    >
                        {sessionCount} session{sessionCount !== 1 ? 's' : ''}
                    </Typography>

                    {/* Overview excerpt */}
                    {chapter.overview && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{
                                mt: 1,
                                display: '-webkit-box',
                                WebkitLineClamp: 2,
                                WebkitBoxOrient: 'vertical',
                                overflow: 'hidden',
                                textOverflow: 'ellipsis',
                            }}
                        >
                            {chapter.overview}
                        </Typography>
                    )}
                </CardContent>
            </CardActionArea>
        </Card>
    );
}
