// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import {
    Box,
    Chip,
    IconButton,
    Paper,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Delete as DeleteIcon,
    Edit as EditIcon,
    SwapHoriz as BidirectionalIcon,
    ArrowForward as DirectionalIcon,
} from '@mui/icons-material';
import type { Relationship } from '../../types';
import type { PendingRelationship } from './RelationshipEditor';

/**
 * Props for the RelationshipRow component.
 */
export interface RelationshipRowProps {
    /** The relationship to display (existing or pending). */
    relationship: Relationship | PendingRelationship;
    /** If true, the relationship is pending (not yet saved). */
    isPending?: boolean;
    /** Callback fired when the edit button is clicked. */
    onEdit?: () => void;
    /** Callback fired when the delete button is clicked. */
    onDelete?: () => void;
    /** If true, the row is in a read-only state. */
    readOnly?: boolean;
}

/**
 * Determines if the relationship is a PendingRelationship.
 */
function isPendingRelationship(
    rel: Relationship | PendingRelationship
): rel is PendingRelationship {
    return 'targetEntityName' in rel;
}

/**
 * Formats a relationship type for display.
 */
function formatRelationshipType(type: string): string {
    return type
        .split('_')
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

/**
 * A single row displaying a relationship.
 *
 * Shows the target entity name (with chip styling), relationship type,
 * bidirectional indicator, and optional edit/delete action buttons.
 * Pending relationships are visually distinguished with a dashed border.
 *
 * @param props - The component props.
 * @returns A React element containing the relationship row.
 *
 * @example
 * ```tsx
 * <RelationshipRow
 *     relationship={relationship}
 *     isPending={false}
 *     onEdit={() => handleEdit(relationship)}
 *     onDelete={() => handleDelete(relationship)}
 * />
 * ```
 */
export default function RelationshipRow({
    relationship,
    isPending = false,
    onEdit,
    onDelete,
    readOnly = false,
}: RelationshipRowProps) {
    const targetName = isPendingRelationship(relationship)
        ? relationship.targetEntityName
        : relationship.targetEntityId; // For existing relationships, we'd need to fetch the name

    const bidirectional = relationship.bidirectional;
    const relationshipType = relationship.relationshipType;
    const description = relationship.description;

    return (
        <Paper
            variant="outlined"
            sx={{
                p: 1.5,
                display: 'flex',
                alignItems: 'center',
                gap: 1.5,
                borderStyle: isPending ? 'dashed' : 'solid',
                borderColor: isPending ? 'warning.main' : 'divider',
                backgroundColor: isPending
                    ? 'action.hover'
                    : 'background.paper',
            }}
        >
            {/* Direction indicator */}
            <Tooltip
                title={
                    bidirectional
                        ? 'Bidirectional relationship'
                        : 'One-way relationship'
                }
            >
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        color: bidirectional ? 'primary.main' : 'text.secondary',
                    }}
                >
                    {bidirectional ? (
                        <BidirectionalIcon fontSize="small" />
                    ) : (
                        <DirectionalIcon fontSize="small" />
                    )}
                </Box>
            </Tooltip>

            {/* Target entity */}
            <Chip
                label={targetName}
                size="small"
                color="primary"
                variant="outlined"
                sx={{ minWidth: 80 }}
            />

            {/* Relationship type */}
            <Typography
                variant="body2"
                color="text.secondary"
                sx={{ minWidth: 80 }}
            >
                {formatRelationshipType(relationshipType)}
            </Typography>

            {/* Description (if present) */}
            {description && (
                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{
                        flexGrow: 1,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                    }}
                >
                    {description}
                </Typography>
            )}

            {/* Spacer */}
            {!description && <Box sx={{ flexGrow: 1 }} />}

            {/* Pending indicator */}
            {isPending && (
                <Chip
                    label="Pending"
                    size="small"
                    color="warning"
                    variant="filled"
                />
            )}

            {/* Action buttons */}
            {!readOnly && (
                <Box sx={{ display: 'flex', gap: 0.5 }}>
                    {onEdit && (
                        <Tooltip title="Edit relationship">
                            <IconButton
                                size="small"
                                onClick={onEdit}
                                aria-label="Edit relationship"
                            >
                                <EditIcon fontSize="small" />
                            </IconButton>
                        </Tooltip>
                    )}
                    {onDelete && (
                        <Tooltip title="Delete relationship">
                            <IconButton
                                size="small"
                                onClick={onDelete}
                                color="error"
                                aria-label="Delete relationship"
                            >
                                <DeleteIcon fontSize="small" />
                            </IconButton>
                        </Tooltip>
                    )}
                </Box>
            )}
        </Paper>
    );
}
