// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useCallback, useEffect } from 'react';
import {
    Alert,
    Box,
    Button,
    CircularProgress,
    Divider,
    Typography,
} from '@mui/material';
import { Add as AddIcon } from '@mui/icons-material';
import {
    useEntityRelationships,
    useDeleteRelationship,
} from '../../hooks/useRelationships';
import { useEntity } from '../../hooks/useEntities';
import RelationshipRow from './RelationshipRow';
import AddRelationshipDialog from './AddRelationshipDialog';
import type { Relationship } from '../../types';

/**
 * Represents a relationship that has not yet been saved to the database.
 */
export interface PendingRelationship {
    /** The ID of the target entity. */
    targetEntityId: string;
    /** The display name of the target entity. */
    targetEntityName: string;
    /** The type of relationship (e.g., "knows", "works_for"). */
    relationshipType: string;
    /** Optional description of the relationship. */
    description?: string;
    /** If true, the relationship applies in both directions. */
    bidirectional: boolean;
}

/**
 * Props for the RelationshipEditor component.
 */
export interface RelationshipEditorProps {
    /** The campaign ID. */
    campaignId: string;
    /** The entity ID (undefined for new entities). */
    entityId?: string;
    /** Callback fired when pending relationships change (for new entities). */
    onPendingRelationshipsChange?: (pending: PendingRelationship[]) => void;
    /** If true, the editor is in read-only mode. */
    readOnly?: boolean;
}

/**
 * Component for managing entity relationships.
 *
 * For existing entities, fetches and displays relationships from the server
 * with options to edit or delete. For new entities, manages pending
 * relationships in local state and reports changes via callback.
 *
 * @param props - The component props.
 * @returns A React element containing the relationship editor.
 *
 * @example
 * ```tsx
 * // For an existing entity:
 * <RelationshipEditor
 *     campaignId={campaignId}
 *     entityId={entityId}
 * />
 *
 * // For a new entity:
 * <RelationshipEditor
 *     campaignId={campaignId}
 *     onPendingRelationshipsChange={setPendingRelationships}
 * />
 * ```
 */
export default function RelationshipEditor({
    campaignId,
    entityId,
    onPendingRelationshipsChange,
    readOnly = false,
}: RelationshipEditorProps) {
    const [pendingRelationships, setPendingRelationships] = useState<
        PendingRelationship[]
    >([]);
    const [dialogOpen, setDialogOpen] = useState(false);
    const [editingIndex, setEditingIndex] = useState<number | null>(null);

    const isNewEntity = !entityId;

    // Fetch existing relationships for existing entities
    const {
        data: existingRelationships,
        isLoading,
        error,
    } = useEntityRelationships(campaignId, entityId ?? '', {
        enabled: !isNewEntity && !!entityId,
    });

    const deleteRelationshipMutation = useDeleteRelationship();

    // Notify parent of pending relationship changes
    useEffect(() => {
        if (onPendingRelationshipsChange) {
            onPendingRelationshipsChange(pendingRelationships);
        }
    }, [pendingRelationships, onPendingRelationshipsChange]);

    /**
     * Gets the list of entity IDs to exclude from selection.
     */
    const getExcludeIds = useCallback((): string[] => {
        const ids: string[] = [];

        // Exclude the current entity
        if (entityId) {
            ids.push(entityId);
        }

        // Exclude entities that already have relationships
        if (existingRelationships) {
            existingRelationships.forEach((rel) => {
                ids.push(rel.targetEntityId);
            });
        }

        // Exclude pending relationship targets
        pendingRelationships.forEach((rel) => {
            ids.push(rel.targetEntityId);
        });

        return ids;
    }, [entityId, existingRelationships, pendingRelationships]);

    /**
     * Opens the add relationship dialog.
     */
    const handleAddClick = () => {
        setEditingIndex(null);
        setDialogOpen(true);
    };

    /**
     * Opens the dialog for editing a pending relationship.
     */
    const handleEditPending = (index: number) => {
        setEditingIndex(index);
        setDialogOpen(true);
    };

    /**
     * Handles saving a relationship from the dialog.
     */
    const handleSaveRelationship = (relationship: PendingRelationship) => {
        if (editingIndex !== null) {
            // Update existing pending relationship
            setPendingRelationships((prev) =>
                prev.map((rel, idx) =>
                    idx === editingIndex ? relationship : rel
                )
            );
        } else {
            // Add new pending relationship
            setPendingRelationships((prev) => [...prev, relationship]);
        }
    };

    /**
     * Handles deleting a pending relationship.
     */
    const handleDeletePending = (index: number) => {
        setPendingRelationships((prev) =>
            prev.filter((_, idx) => idx !== index)
        );
    };

    /**
     * Handles deleting an existing relationship.
     */
    const handleDeleteExisting = async (relationshipId: string) => {
        if (!campaignId) return;

        try {
            await deleteRelationshipMutation.mutateAsync({
                campaignId,
                relationshipId,
            });
        } catch (err) {
            console.error('Failed to delete relationship:', err);
        }
    };

    /**
     * Closes the dialog.
     */
    const handleCloseDialog = () => {
        setDialogOpen(false);
        setEditingIndex(null);
    };

    // Loading state
    if (!isNewEntity && isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    // Error state
    if (error) {
        return (
            <Alert severity="error" sx={{ mb: 2 }}>
                Failed to load relationships: {error.message}
            </Alert>
        );
    }

    const hasExistingRelationships =
        existingRelationships && existingRelationships.length > 0;
    const hasPendingRelationships = pendingRelationships.length > 0;
    const hasAnyRelationships =
        hasExistingRelationships || hasPendingRelationships;

    return (
        <Box>
            {/* Header */}
            <Box
                sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    mb: 2,
                }}
            >
                <Typography variant="subtitle1" fontWeight={500}>
                    Relationships
                </Typography>
                {!readOnly && (
                    <Button
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={handleAddClick}
                    >
                        Add Relationship
                    </Button>
                )}
            </Box>

            {/* Relationship list */}
            {hasAnyRelationships ? (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                    {/* Existing relationships */}
                    {existingRelationships?.map((rel) => (
                        <RelationshipRowWithName
                            key={rel.id}
                            relationship={rel}
                            campaignId={campaignId}
                            readOnly={readOnly}
                            onDelete={() => handleDeleteExisting(rel.id)}
                        />
                    ))}

                    {/* Divider between existing and pending */}
                    {hasExistingRelationships && hasPendingRelationships && (
                        <Divider sx={{ my: 1 }}>
                            <Typography
                                variant="caption"
                                color="text.secondary"
                            >
                                Pending (will be saved with entity)
                            </Typography>
                        </Divider>
                    )}

                    {/* Pending relationships */}
                    {pendingRelationships.map((rel, index) => (
                        <RelationshipRow
                            key={`pending-${index}`}
                            relationship={rel}
                            isPending
                            readOnly={readOnly}
                            onEdit={() => handleEditPending(index)}
                            onDelete={() => handleDeletePending(index)}
                        />
                    ))}
                </Box>
            ) : (
                <Typography variant="body2" color="text.secondary">
                    No relationships defined. Click &quot;Add Relationship&quot; to
                    connect this entity to others.
                </Typography>
            )}

            {/* Add/Edit Dialog */}
            <AddRelationshipDialog
                open={dialogOpen}
                onClose={handleCloseDialog}
                onSave={handleSaveRelationship}
                campaignId={campaignId}
                excludeEntityIds={getExcludeIds()}
                existingRelationship={
                    editingIndex !== null
                        ? pendingRelationships[editingIndex]
                        : undefined
                }
            />
        </Box>
    );
}

/**
 * Props for RelationshipRowWithName component.
 */
interface RelationshipRowWithNameProps {
    relationship: Relationship;
    campaignId: string;
    readOnly?: boolean;
    onDelete?: () => void;
}

/**
 * Wrapper component that fetches the target entity name for display.
 */
function RelationshipRowWithName({
    relationship,
    campaignId,
    readOnly = false,
    onDelete,
}: RelationshipRowWithNameProps) {
    const { data: targetEntity } = useEntity(
        campaignId,
        relationship.targetEntityId
    );

    const resolvedRelationship: PendingRelationship = {
        targetEntityId: relationship.targetEntityId,
        targetEntityName: targetEntity?.name ?? 'Loading...',
        relationshipType: relationship.relationshipType,
        description: relationship.description,
        bidirectional: relationship.bidirectional,
    };

    return (
        <RelationshipRow
            relationship={resolvedRelationship}
            isPending={false}
            readOnly={readOnly}
            onDelete={onDelete}
        />
    );
}
