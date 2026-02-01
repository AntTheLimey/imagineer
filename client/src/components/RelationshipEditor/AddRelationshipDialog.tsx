// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useEffect } from 'react';
import {
    Autocomplete,
    Box,
    Button,
    Checkbox,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControlLabel,
    TextField,
} from '@mui/material';
import { EntitySelector } from '../EntitySelector';
import type { Entity } from '../../types';
import type { PendingRelationship } from './RelationshipEditor';

/**
 * Common relationship type suggestions.
 */
const RELATIONSHIP_TYPE_SUGGESTIONS = [
    'knows',
    'works_for',
    'related_to',
    'friend_of',
    'enemy_of',
    'located_at',
    'member_of',
    'owns',
    'created',
    'reports_to',
    'allied_with',
    'rivals_with',
];

/**
 * Props for the AddRelationshipDialog component.
 */
export interface AddRelationshipDialogProps {
    /** If true, the dialog is open. */
    open: boolean;
    /** Callback fired when the dialog is closed. */
    onClose: () => void;
    /** Callback fired when a relationship is saved. */
    onSave: (relationship: PendingRelationship) => void;
    /** The campaign ID for entity search. */
    campaignId: string;
    /** Entity IDs to exclude from selection. */
    excludeEntityIds: string[];
    /** Existing relationship data for editing (optional). */
    existingRelationship?: PendingRelationship;
}

/**
 * Dialog for adding or editing relationships.
 *
 * Provides form fields for selecting a target entity, relationship type,
 * optional description, and bidirectional flag. Validates that required
 * fields are filled before allowing save.
 *
 * @param props - The component props.
 * @returns A React element containing the add/edit relationship dialog.
 *
 * @example
 * ```tsx
 * <AddRelationshipDialog
 *     open={dialogOpen}
 *     onClose={() => setDialogOpen(false)}
 *     onSave={handleSaveRelationship}
 *     campaignId={campaignId}
 *     excludeEntityIds={[currentEntityId]}
 * />
 * ```
 */
export default function AddRelationshipDialog({
    open,
    onClose,
    onSave,
    campaignId,
    excludeEntityIds,
    existingRelationship,
}: AddRelationshipDialogProps) {
    const [targetEntity, setTargetEntity] = useState<Entity | null>(null);
    const [relationshipType, setRelationshipType] = useState('');
    const [description, setDescription] = useState('');
    const [bidirectional, setBidirectional] = useState(false);
    const [errors, setErrors] = useState<{
        targetEntity?: string;
        relationshipType?: string;
    }>({});

    // Reset form when dialog opens/closes or existing relationship changes
    useEffect(() => {
        if (open) {
            if (existingRelationship) {
                // Editing: populate form with existing data
                // Note: We don't have the full Entity object, just the ID and name
                setTargetEntity({
                    id: existingRelationship.targetEntityId,
                    name: existingRelationship.targetEntityName,
                    campaignId,
                    entityType: 'other', // Placeholder; actual type unknown
                    attributes: {},
                    tags: [],
                    sourceConfidence: 'DRAFT',
                    version: 1,
                    createdAt: '',
                    updatedAt: '',
                } as Entity);
                setRelationshipType(existingRelationship.relationshipType);
                setDescription(existingRelationship.description ?? '');
                setBidirectional(existingRelationship.bidirectional);
            } else {
                // Adding: reset to defaults
                setTargetEntity(null);
                setRelationshipType('');
                setDescription('');
                setBidirectional(false);
            }
            setErrors({});
        }
    }, [open, existingRelationship, campaignId]);

    /**
     * Validates the form and returns true if valid.
     */
    const validateForm = (): boolean => {
        const newErrors: { targetEntity?: string; relationshipType?: string } =
            {};

        if (!targetEntity) {
            newErrors.targetEntity = 'Please select a target entity';
        }
        if (!relationshipType.trim()) {
            newErrors.relationshipType = 'Please enter a relationship type';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    /**
     * Handles form submission.
     */
    const handleSave = () => {
        if (!validateForm() || !targetEntity) return;

        const pendingRelationship: PendingRelationship = {
            targetEntityId: targetEntity.id,
            targetEntityName: targetEntity.name,
            relationshipType: relationshipType.trim().toLowerCase().replace(/\s+/g, '_'),
            description: description.trim() || undefined,
            bidirectional,
        };

        onSave(pendingRelationship);
        onClose();
    };

    /**
     * Handles dialog close without saving.
     */
    const handleClose = () => {
        onClose();
    };

    const isEditing = !!existingRelationship;

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            maxWidth="sm"
            fullWidth
            aria-labelledby="add-relationship-dialog-title"
        >
            <DialogTitle id="add-relationship-dialog-title">
                {isEditing ? 'Edit Relationship' : 'Add Relationship'}
            </DialogTitle>
            <DialogContent>
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 2.5,
                        pt: 1,
                    }}
                >
                    {/* Target Entity */}
                    <EntitySelector
                        campaignId={campaignId}
                        value={targetEntity}
                        onChange={setTargetEntity}
                        excludeIds={excludeEntityIds}
                        label="Target Entity"
                        placeholder="Search for an entity..."
                        error={!!errors.targetEntity}
                        helperText={errors.targetEntity}
                    />

                    {/* Relationship Type */}
                    <Autocomplete
                        freeSolo
                        options={RELATIONSHIP_TYPE_SUGGESTIONS}
                        value={relationshipType}
                        onChange={(_event, newValue) => {
                            setRelationshipType(newValue ?? '');
                        }}
                        onInputChange={(_event, newInputValue) => {
                            setRelationshipType(newInputValue);
                        }}
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Relationship Type"
                                placeholder="e.g., knows, works_for, enemy_of"
                                error={!!errors.relationshipType}
                                helperText={
                                    errors.relationshipType ??
                                    'Type or select a relationship type'
                                }
                            />
                        )}
                    />

                    {/* Description */}
                    <TextField
                        label="Description"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="Optional details about this relationship..."
                        multiline
                        rows={2}
                        helperText="Optional: Add context about this relationship"
                    />

                    {/* Bidirectional */}
                    <FormControlLabel
                        control={
                            <Checkbox
                                checked={bidirectional}
                                onChange={(e) =>
                                    setBidirectional(e.target.checked)
                                }
                            />
                        }
                        label="Bidirectional relationship (applies in both directions)"
                    />
                </Box>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose}>Cancel</Button>
                <Button onClick={handleSave} variant="contained">
                    {isEditing ? 'Save Changes' : 'Add Relationship'}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
