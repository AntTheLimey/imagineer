// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useEffect, useMemo } from 'react';
import {
    Alert,
    Box,
    Button,
    Checkbox,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Divider,
    FormControl,
    FormControlLabel,
    FormHelperText,
    InputLabel,
    ListItemText,
    MenuItem,
    Paper,
    Select,
    TextField,
    ToggleButton,
    ToggleButtonGroup,
    Typography,
} from '@mui/material';
import {
    ArrowForward as ArrowForwardIcon,
    ArrowBack as ArrowBackIcon,
    Add as AddIcon,
} from '@mui/icons-material';
import { EntitySelector } from '../EntitySelector';
import {
    useRelationshipTypes,
    useCreateRelationshipType,
} from '../../hooks/useRelationshipTypes';
import type { Entity, RelationshipType, CreateRelationshipTypeInput } from '../../types';
import type { PendingRelationship } from './RelationshipEditor';

/**
 * Direction of the relationship from the current entity's perspective.
 */
type RelationshipDirection = 'outgoing' | 'incoming';

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
    campaignId: number;
    /** Entity IDs to exclude from selection. */
    excludeEntityIds: number[];
    /** The current entity name for preview. */
    currentEntityName?: string;
    /** Existing relationship data for editing (optional). */
    existingRelationship?: PendingRelationship;
}

/**
 * Special value to indicate "Add Custom Type" option in the select.
 */
const ADD_CUSTOM_TYPE_VALUE = '__add_custom__';

/**
 * Dialog for adding or editing relationships.
 *
 * Provides form fields for selecting a target entity, relationship direction,
 * relationship type from API-provided types, and optional description.
 * Includes a mini-dialog for creating custom types.
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
 *     currentEntityName="Billy Bob"
 * />
 * ```
 */
export default function AddRelationshipDialog({
    open,
    onClose,
    onSave,
    campaignId,
    excludeEntityIds,
    currentEntityName = 'This entity',
    existingRelationship,
}: AddRelationshipDialogProps) {
    // Form state
    const [targetEntity, setTargetEntity] = useState<Entity | null>(null);
    const [direction, setDirection] = useState<RelationshipDirection>('outgoing');
    const [selectedTypeId, setSelectedTypeId] = useState<string>('');
    const [description, setDescription] = useState('');
    const [errors, setErrors] = useState<{
        targetEntity?: string;
        relationshipType?: string;
    }>({});

    // Custom type creation state
    const [customTypeDialogOpen, setCustomTypeDialogOpen] = useState(false);
    const [customTypeName, setCustomTypeName] = useState('');
    const [customTypeInverseName, setCustomTypeInverseName] = useState('');
    const [customTypeIsSymmetric, setCustomTypeIsSymmetric] = useState(false);
    const [customTypeDisplayLabel, setCustomTypeDisplayLabel] = useState('');
    const [customTypeInverseDisplayLabel, setCustomTypeInverseDisplayLabel] = useState('');
    const [customTypeDescription, setCustomTypeDescription] = useState('');

    // Fetch relationship types from API
    const {
        data: relationshipTypes,
        isLoading: typesLoading,
        error: typesError,
    } = useRelationshipTypes(campaignId, { enabled: open });

    const createTypeMutation = useCreateRelationshipType();

    /**
     * Find the selected relationship type object.
     */
    const selectedType = useMemo(() => {
        if (!relationshipTypes || !selectedTypeId) return null;
        const numericId = Number(selectedTypeId);
        return relationshipTypes.find((t) => t.id === numericId) ?? null;
    }, [relationshipTypes, selectedTypeId]);

    /**
     * Get the display label based on direction.
     */
    const getDisplayLabel = (type: RelationshipType): string => {
        return direction === 'outgoing' ? type.displayLabel : type.inverseDisplayLabel;
    };

    /**
     * Get the internal type name based on direction.
     */
    const getTypeName = (type: RelationshipType): string => {
        return direction === 'outgoing' ? type.name : type.inverseName;
    };

    /**
     * Generate preview sentence for the relationship.
     */
    const previewSentence = useMemo(() => {
        if (!targetEntity || !selectedType) return null;

        const label = getDisplayLabel(selectedType);
        if (direction === 'outgoing') {
            return `${currentEntityName} ${label} ${targetEntity.name}`;
        } else {
            return `${targetEntity.name} ${label} ${currentEntityName}`;
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [targetEntity, selectedType, direction, currentEntityName]);

    // Reset form when dialog opens/closes or existing relationship changes
    useEffect(() => {
        if (open) {
            if (existingRelationship) {
                // Editing: populate form with existing data
                setTargetEntity({
                    id: existingRelationship.targetEntityId,
                    name: existingRelationship.targetEntityName,
                    campaignId,
                    entityType: 'other',
                    attributes: {},
                    tags: [],
                    sourceConfidence: 'DRAFT',
                    version: 1,
                    createdAt: '',
                    updatedAt: '',
                } as Entity);
                setDirection('outgoing'); // Default to outgoing when editing
                setDescription(existingRelationship.description ?? '');
                // Try to find matching type by name
                if (relationshipTypes) {
                    const matchingType = relationshipTypes.find(
                        (t) =>
                            t.name === existingRelationship.relationshipType ||
                            t.inverseName === existingRelationship.relationshipType
                    );
                    if (matchingType) {
                        setSelectedTypeId(String(matchingType.id));
                        // Set direction based on which name matched
                        if (matchingType.inverseName === existingRelationship.relationshipType) {
                            setDirection('incoming');
                        }
                    }
                }
            } else {
                // Adding: reset to defaults
                setTargetEntity(null);
                setDirection('outgoing');
                setSelectedTypeId('');
                setDescription('');
            }
            setErrors({});
        }
    }, [open, existingRelationship, campaignId, relationshipTypes]);

    // Auto-fill display labels when custom type name changes
    useEffect(() => {
        if (customTypeName && !customTypeDisplayLabel) {
            setCustomTypeDisplayLabel(customTypeName.replace(/_/g, ' '));
        }
    }, [customTypeName, customTypeDisplayLabel]);

    useEffect(() => {
        if (customTypeInverseName && !customTypeInverseDisplayLabel) {
            setCustomTypeInverseDisplayLabel(customTypeInverseName.replace(/_/g, ' '));
        }
    }, [customTypeInverseName, customTypeInverseDisplayLabel]);

    // For symmetric types, sync inverse name with name
    useEffect(() => {
        if (customTypeIsSymmetric) {
            setCustomTypeInverseName(customTypeName);
            setCustomTypeInverseDisplayLabel(customTypeDisplayLabel);
        }
    }, [customTypeIsSymmetric, customTypeName, customTypeDisplayLabel]);

    /**
     * Validates the form and returns true if valid.
     */
    const validateForm = (): boolean => {
        const newErrors: { targetEntity?: string; relationshipType?: string } = {};

        if (!targetEntity) {
            newErrors.targetEntity = 'Please select a target entity';
        }
        if (!selectedTypeId) {
            newErrors.relationshipType = 'Please select a relationship type';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    /**
     * Handles form submission.
     */
    const handleSave = () => {
        if (!validateForm() || !targetEntity || !selectedType) return;

        const typeName = getTypeName(selectedType);

        const pendingRelationship: PendingRelationship = {
            targetEntityId: targetEntity.id,
            targetEntityName: targetEntity.name,
            relationshipType: typeName,
            relationshipTypeId: selectedType.id,
            description: description.trim() || undefined,
            isReversed: direction === 'incoming',
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

    /**
     * Handles direction toggle change.
     */
    const handleDirectionChange = (
        _event: React.MouseEvent<HTMLElement>,
        newDirection: RelationshipDirection | null
    ) => {
        if (newDirection !== null) {
            setDirection(newDirection);
        }
    };

    /**
     * Handles relationship type selection.
     */
    const handleTypeChange = (event: { target: { value: string } }) => {
        const value = event.target.value;
        if (value === ADD_CUSTOM_TYPE_VALUE) {
            setCustomTypeDialogOpen(true);
        } else {
            setSelectedTypeId(value);
        }
    };

    /**
     * Resets the custom type form.
     */
    const resetCustomTypeForm = () => {
        setCustomTypeName('');
        setCustomTypeInverseName('');
        setCustomTypeIsSymmetric(false);
        setCustomTypeDisplayLabel('');
        setCustomTypeInverseDisplayLabel('');
        setCustomTypeDescription('');
    };

    /**
     * Handles custom type creation.
     */
    const handleCreateCustomType = async () => {
        if (!customTypeName.trim()) return;

        const input: CreateRelationshipTypeInput = {
            name: customTypeName.trim().toLowerCase().replace(/\s+/g, '_'),
            inverseName: customTypeIsSymmetric
                ? customTypeName.trim().toLowerCase().replace(/\s+/g, '_')
                : customTypeInverseName.trim().toLowerCase().replace(/\s+/g, '_'),
            isSymmetric: customTypeIsSymmetric,
            displayLabel: customTypeDisplayLabel.trim() || customTypeName.trim(),
            inverseDisplayLabel: customTypeIsSymmetric
                ? customTypeDisplayLabel.trim() || customTypeName.trim()
                : customTypeInverseDisplayLabel.trim() || customTypeInverseName.trim(),
            description: customTypeDescription.trim() || undefined,
        };

        try {
            const newType = await createTypeMutation.mutateAsync({
                campaignId,
                input,
            });
            setSelectedTypeId(String(newType.id));
            setCustomTypeDialogOpen(false);
            resetCustomTypeForm();
        } catch (err) {
            console.error('Failed to create custom type:', err);
        }
    };

    const isEditing = !!existingRelationship;

    return (
        <>
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

                        {/* Direction Toggle */}
                        <Box>
                            <Typography
                                variant="body2"
                                color="text.secondary"
                                sx={{ mb: 1 }}
                            >
                                Relationship Direction
                            </Typography>
                            <ToggleButtonGroup
                                value={direction}
                                exclusive
                                onChange={handleDirectionChange}
                                aria-label="relationship direction"
                                fullWidth
                            >
                                <ToggleButton
                                    value="outgoing"
                                    aria-label="outgoing relationship"
                                >
                                    <ArrowForwardIcon sx={{ mr: 1 }} />
                                    {currentEntityName} to Target
                                </ToggleButton>
                                <ToggleButton
                                    value="incoming"
                                    aria-label="incoming relationship"
                                >
                                    <ArrowBackIcon sx={{ mr: 1 }} />
                                    Target to {currentEntityName}
                                </ToggleButton>
                            </ToggleButtonGroup>
                        </Box>

                        {/* Relationship Type Selector */}
                        <FormControl
                            fullWidth
                            error={!!errors.relationshipType}
                        >
                            <InputLabel id="relationship-type-label">
                                Relationship Type
                            </InputLabel>
                            {typesLoading ? (
                                <Box sx={{ display: 'flex', alignItems: 'center', py: 2 }}>
                                    <CircularProgress size={20} sx={{ mr: 1 }} />
                                    <Typography variant="body2" color="text.secondary">
                                        Loading relationship types...
                                    </Typography>
                                </Box>
                            ) : typesError ? (
                                <Alert severity="error" sx={{ mt: 1 }}>
                                    Failed to load relationship types
                                </Alert>
                            ) : (
                                <Select
                                    labelId="relationship-type-label"
                                    value={selectedTypeId}
                                    onChange={handleTypeChange}
                                    label="Relationship Type"
                                >
                                    {relationshipTypes?.map((type) => (
                                        <MenuItem key={type.id} value={String(type.id)}>
                                            <ListItemText
                                                primary={getDisplayLabel(type)}
                                                secondary={
                                                    type.isSymmetric
                                                        ? '(symmetric)'
                                                        : `${type.displayLabel} / ${type.inverseDisplayLabel}`
                                                }
                                            />
                                        </MenuItem>
                                    ))}
                                    <Divider />
                                    <MenuItem value={ADD_CUSTOM_TYPE_VALUE}>
                                        <AddIcon sx={{ mr: 1 }} />
                                        Add Custom Type...
                                    </MenuItem>
                                </Select>
                            )}
                            {errors.relationshipType && (
                                <FormHelperText>{errors.relationshipType}</FormHelperText>
                            )}
                        </FormControl>

                        {/* Preview Sentence */}
                        {previewSentence && (
                            <Paper
                                variant="outlined"
                                sx={{
                                    p: 2,
                                    backgroundColor: 'action.hover',
                                }}
                            >
                                <Typography
                                    variant="caption"
                                    color="text.secondary"
                                    display="block"
                                    gutterBottom
                                >
                                    Preview:
                                </Typography>
                                <Typography variant="body1">
                                    {previewSentence}
                                </Typography>
                            </Paper>
                        )}

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

                    </Box>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose}>Cancel</Button>
                    <Button
                        onClick={handleSave}
                        variant="contained"
                        disabled={typesLoading}
                    >
                        {isEditing ? 'Save Changes' : 'Add Relationship'}
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Custom Type Creation Dialog */}
            <Dialog
                open={customTypeDialogOpen}
                onClose={() => {
                    setCustomTypeDialogOpen(false);
                    resetCustomTypeForm();
                }}
                maxWidth="sm"
                fullWidth
                aria-labelledby="create-custom-type-dialog-title"
            >
                <DialogTitle id="create-custom-type-dialog-title">
                    Create Custom Relationship Type
                </DialogTitle>
                <DialogContent>
                    <Box
                        sx={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 2,
                            pt: 1,
                        }}
                    >
                        {/* Symmetric Checkbox */}
                        <FormControlLabel
                            control={
                                <Checkbox
                                    checked={customTypeIsSymmetric}
                                    onChange={(e) =>
                                        setCustomTypeIsSymmetric(e.target.checked)
                                    }
                                />
                            }
                            label="Symmetric relationship (same in both directions, e.g., 'siblings_with')"
                        />

                        {/* Name */}
                        <TextField
                            label="Type Name"
                            value={customTypeName}
                            onChange={(e) => setCustomTypeName(e.target.value)}
                            placeholder="e.g., owns, mentors"
                            helperText="Internal name used in the database (will be converted to snake_case)"
                            required
                        />

                        {/* Display Label */}
                        <TextField
                            label="Display Label"
                            value={customTypeDisplayLabel}
                            onChange={(e) => setCustomTypeDisplayLabel(e.target.value)}
                            placeholder="e.g., owns, mentors"
                            helperText="How the relationship reads (e.g., 'Billy owns the Farm')"
                        />

                        {/* Inverse Name (hidden if symmetric) */}
                        {!customTypeIsSymmetric && (
                            <>
                                <Divider>
                                    <Typography variant="caption" color="text.secondary">
                                        Inverse Relationship
                                    </Typography>
                                </Divider>

                                <TextField
                                    label="Inverse Type Name"
                                    value={customTypeInverseName}
                                    onChange={(e) => setCustomTypeInverseName(e.target.value)}
                                    placeholder="e.g., owned_by, mentored_by"
                                    helperText="The inverse name (how it reads from the target's perspective)"
                                    required
                                />

                                <TextField
                                    label="Inverse Display Label"
                                    value={customTypeInverseDisplayLabel}
                                    onChange={(e) =>
                                        setCustomTypeInverseDisplayLabel(e.target.value)
                                    }
                                    placeholder="e.g., is owned by, is mentored by"
                                    helperText="How the inverse reads (e.g., 'The Farm is owned by Billy')"
                                />
                            </>
                        )}

                        {/* Description */}
                        <TextField
                            label="Description"
                            value={customTypeDescription}
                            onChange={(e) => setCustomTypeDescription(e.target.value)}
                            placeholder="Optional description of this relationship type..."
                            multiline
                            rows={2}
                        />
                    </Box>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setCustomTypeDialogOpen(false);
                            resetCustomTypeForm();
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleCreateCustomType}
                        variant="contained"
                        disabled={
                            !customTypeName.trim() ||
                            (!customTypeIsSymmetric && !customTypeInverseName.trim()) ||
                            createTypeMutation.isPending
                        }
                    >
                        {createTypeMutation.isPending ? (
                            <CircularProgress size={20} />
                        ) : (
                            'Create Type'
                        )}
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
}
