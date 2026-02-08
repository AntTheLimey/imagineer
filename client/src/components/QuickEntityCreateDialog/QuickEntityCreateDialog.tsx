/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useState, useEffect } from 'react';
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
    Typography,
} from '@mui/material';
import { useCreateEntity } from '../../hooks/useEntities';
import type { Entity, EntityType } from '../../types';

/**
 * Props for the QuickEntityCreateDialog component.
 */
export interface QuickEntityCreateDialogProps {
    /** If true, the dialog is open. */
    open: boolean;
    /** The campaign ID for the new entity. */
    campaignId: number;
    /** Optional pre-filled name for the entity. */
    suggestedName?: string;
    /** Optional pre-selected entity type. */
    suggestedType?: EntityType;
    /** Callback fired when the dialog is closed. */
    onClose: () => void;
    /** Callback fired when an entity is successfully created. */
    onCreated: (entity: Entity) => void;
}

/**
 * All available entity types with their display labels.
 */
const ENTITY_TYPE_OPTIONS: { value: EntityType; label: string }[] = [
    { value: 'npc', label: 'NPC' },
    { value: 'location', label: 'Location' },
    { value: 'item', label: 'Item' },
    { value: 'faction', label: 'Faction' },
    { value: 'clue', label: 'Clue' },
    { value: 'creature', label: 'Creature' },
    { value: 'organization', label: 'Organization' },
    { value: 'event', label: 'Event' },
    { value: 'document', label: 'Document' },
    { value: 'other', label: 'Other' },
];

/**
 * A simplified dialog for quickly creating entities inline from the chapter editor.
 *
 * This dialog provides minimal fields (name, type, description) so users can
 * rapidly create entities and continue editing. The created entity can be
 * edited later with full details.
 *
 * @param props - The component props.
 * @returns A React element containing the quick entity creation dialog.
 *
 * @example
 * ```tsx
 * <QuickEntityCreateDialog
 *     open={dialogOpen}
 *     campaignId={campaignId}
 *     suggestedName="Dr. Morgan"
 *     suggestedType="npc"
 *     onClose={() => setDialogOpen(false)}
 *     onCreated={(entity) => handleEntityCreated(entity)}
 * />
 * ```
 */
export default function QuickEntityCreateDialog({
    open,
    campaignId,
    suggestedName,
    suggestedType,
    onClose,
    onCreated,
}: QuickEntityCreateDialogProps) {
    // Form state
    const [name, setName] = useState('');
    const [entityType, setEntityType] = useState<EntityType | ''>('');
    const [description, setDescription] = useState('');
    const [error, setError] = useState<string | null>(null);

    // Create entity mutation
    const createEntityMutation = useCreateEntity();

    // Reset form when dialog opens
    useEffect(() => {
        if (open) {
            setName(suggestedName ?? '');
            setEntityType(suggestedType ?? '');
            setDescription('');
            setError(null);
        }
    }, [open, suggestedName, suggestedType]);

    /**
     * Checks if the form is valid for submission.
     */
    const isFormValid = name.trim().length > 0 && entityType !== '';

    /**
     * Handles form submission.
     */
    const handleCreate = async () => {
        if (!isFormValid) return;

        // TypeScript guard: entityType is guaranteed to be non-empty by isFormValid
        const selectedType = entityType as EntityType;

        setError(null);

        try {
            const entity = await createEntityMutation.mutateAsync({
                campaignId,
                name: name.trim(),
                entityType: selectedType,
                description: description.trim() || undefined,
            });

            onCreated(entity);
            onClose();
        } catch (err) {
            console.error('Failed to create entity:', err);
            setError(
                err instanceof Error
                    ? err.message
                    : 'Failed to create entity. Please try again.'
            );
        }
    };

    /**
     * Handles dialog close without saving.
     */
    const handleClose = () => {
        if (!createEntityMutation.isPending) {
            onClose();
        }
    };

    /**
     * Handles Enter key press to submit the form.
     */
    const handleKeyDown = (event: React.KeyboardEvent) => {
        if (event.key === 'Enter' && !event.shiftKey && isFormValid) {
            event.preventDefault();
            handleCreate();
        }
    };

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            maxWidth="sm"
            fullWidth
            aria-labelledby="quick-entity-create-dialog-title"
        >
            <DialogTitle id="quick-entity-create-dialog-title">
                Quick Create Entity
            </DialogTitle>
            <DialogContent>
                <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ mb: 2 }}
                >
                    Create a minimal entity quickly. You can add more details later.
                </Typography>

                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 2,
                        pt: 1,
                    }}
                >
                    {/* Error Alert */}
                    {error && (
                        <Alert severity="error" onClose={() => setError(null)}>
                            {error}
                        </Alert>
                    )}

                    {/* Name Field */}
                    <TextField
                        label="Name"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Enter entity name..."
                        required
                        fullWidth
                        autoFocus
                        error={name.length > 0 && name.trim().length === 0}
                        helperText={
                            name.length > 0 && name.trim().length === 0
                                ? 'Name cannot be empty'
                                : undefined
                        }
                    />

                    {/* Type Selector */}
                    <FormControl fullWidth required>
                        <InputLabel id="entity-type-label">Type</InputLabel>
                        <Select
                            labelId="entity-type-label"
                            value={entityType}
                            onChange={(e) =>
                                setEntityType(e.target.value as EntityType)
                            }
                            label="Type"
                        >
                            {ENTITY_TYPE_OPTIONS.map((option) => (
                                <MenuItem key={option.value} value={option.value}>
                                    {option.label}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>

                    {/* Description Field */}
                    <TextField
                        label="Description"
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="Optional brief description..."
                        multiline
                        rows={3}
                        fullWidth
                        helperText="Optional: Add a brief description"
                    />
                </Box>
            </DialogContent>
            <DialogActions>
                <Button
                    onClick={handleClose}
                    disabled={createEntityMutation.isPending}
                >
                    Cancel
                </Button>
                <Button
                    onClick={handleCreate}
                    variant="contained"
                    disabled={!isFormValid || createEntityMutation.isPending}
                >
                    {createEntityMutation.isPending ? (
                        <CircularProgress size={20} />
                    ) : (
                        'Create'
                    )}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
