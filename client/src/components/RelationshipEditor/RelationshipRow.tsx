// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useMemo } from 'react';
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
import type { Relationship, RelationshipType } from '../../types';
import type { PendingRelationship } from './RelationshipEditor';

/**
 * Props for the RelationshipRow component.
 */
export interface RelationshipRowProps {
    /** The relationship to display (existing or pending). */
    relationship: Relationship | PendingRelationship;
    /** The current entity name for display in natural sentences. */
    currentEntityName?: string;
    /** Map of relationship type IDs/names to their definitions. */
    relationshipTypeMap?: Map<string, RelationshipType>;
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
 * Formats a relationship type for display (fallback when no type definition found).
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
 * Shows the relationship as a natural sentence using the relationship type's
 * display labels. For example: "Billy Bob owns Thornwood Farm" or
 * "Thornwood Farm is owned by Billy Bob".
 *
 * @param props - The component props.
 * @returns A React element containing the relationship row.
 *
 * @example
 * ```tsx
 * <RelationshipRow
 *     relationship={relationship}
 *     currentEntityName="Billy Bob"
 *     relationshipTypeMap={typeMap}
 *     isPending={false}
 *     onEdit={() => handleEdit(relationship)}
 *     onDelete={() => handleDelete(relationship)}
 * />
 * ```
 */
export default function RelationshipRow({
    relationship,
    currentEntityName = 'This entity',
    relationshipTypeMap,
    isPending = false,
    onEdit,
    onDelete,
    readOnly = false,
}: RelationshipRowProps) {
    const targetName = isPendingRelationship(relationship)
        ? relationship.targetEntityName
        : relationship.targetEntityId; // For existing relationships, this should be resolved by the parent

    const bidirectional = relationship.bidirectional;
    const relationshipType = relationship.relationshipType;
    const description = relationship.description;
    const isReversed = isPendingRelationship(relationship)
        ? relationship.isReversed ?? false
        : false;

    /**
     * Get the display label for the relationship.
     * Uses the relationship type definition if available, otherwise formats
     * the raw type name.
     */
    const displayLabel = useMemo(() => {
        if (!relationshipTypeMap) {
            return formatRelationshipType(relationshipType);
        }

        // Try to find the type definition
        const typeDef = isPendingRelationship(relationship) && relationship.relationshipTypeId
            ? relationshipTypeMap.get(relationship.relationshipTypeId)
            : relationshipTypeMap.get(relationshipType);

        if (!typeDef) {
            return formatRelationshipType(relationshipType);
        }

        // If the relationship is reversed (target -> source), use inverse label
        // Otherwise use the normal display label
        if (isReversed) {
            return typeDef.inverseDisplayLabel;
        }
        return typeDef.displayLabel;
    }, [relationshipTypeMap, relationshipType, relationship, isReversed]);

    /**
     * Build the natural sentence describing the relationship.
     */
    const relationshipSentence = useMemo(() => {
        if (bidirectional) {
            // For bidirectional relationships, show both entities connected
            return (
                <>
                    <Typography
                        component="span"
                        variant="body2"
                        fontWeight={500}
                    >
                        {currentEntityName}
                    </Typography>
                    <Typography
                        component="span"
                        variant="body2"
                        sx={{ mx: 0.5 }}
                    >
                        {displayLabel}
                    </Typography>
                    <Typography
                        component="span"
                        variant="body2"
                        fontWeight={500}
                    >
                        {targetName}
                    </Typography>
                </>
            );
        }

        // For directional relationships, the sentence depends on direction
        if (isReversed) {
            // Target -> Current entity (e.g., "Thornwood Farm is owned by Billy Bob")
            return (
                <>
                    <Typography
                        component="span"
                        variant="body2"
                        fontWeight={500}
                    >
                        {targetName}
                    </Typography>
                    <Typography
                        component="span"
                        variant="body2"
                        sx={{ mx: 0.5 }}
                    >
                        {displayLabel}
                    </Typography>
                    <Typography
                        component="span"
                        variant="body2"
                        fontWeight={500}
                    >
                        {currentEntityName}
                    </Typography>
                </>
            );
        }

        // Current entity -> Target (e.g., "Billy Bob owns Thornwood Farm")
        return (
            <>
                <Typography
                    component="span"
                    variant="body2"
                    fontWeight={500}
                >
                    {currentEntityName}
                </Typography>
                <Typography
                    component="span"
                    variant="body2"
                    sx={{ mx: 0.5 }}
                >
                    {displayLabel}
                </Typography>
                <Typography
                    component="span"
                    variant="body2"
                    fontWeight={500}
                >
                    {targetName}
                </Typography>
            </>
        );
    }, [bidirectional, isReversed, currentEntityName, targetName, displayLabel]);

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

            {/* Relationship sentence */}
            <Box sx={{ flexGrow: 1, display: 'flex', alignItems: 'center', flexWrap: 'wrap' }}>
                {relationshipSentence}
            </Box>

            {/* Description (if present) */}
            {description && (
                <Tooltip title={description}>
                    <Chip
                        label="..."
                        size="small"
                        variant="outlined"
                        sx={{ cursor: 'help' }}
                    />
                </Tooltip>
            )}

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
