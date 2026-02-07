// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * CampaignOverview - Read-first campaign overview page.
 *
 * Displays campaign information in a readable format with the ability
 * to edit individual fields or switch to full edit mode.
 */

import { useState, useCallback, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    CircularProgress,
    FormControl,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Edit as EditIcon,
    Save as SaveIcon,
    Close as CancelIcon,
    Check as CheckIcon,
} from '@mui/icons-material';
import { MarkdownEditor } from '../components/MarkdownEditor';
import { GENRE_OPTIONS } from '../components/CampaignSettings';
import {
    useCampaign,
    useUpdateCampaign,
    useGameSystems,
    useCampaignStats,
} from '../hooks';
import { useCampaignContext } from '../contexts/CampaignContext';
import { sanitizeHtml } from '../utils';
import type { GameSystem } from '../types';

/**
 * Fields that can be edited inline.
 */
type EditableField = 'name' | 'description' | 'genre' | 'imageStylePrompt';

/**
 * Render the Campaign Overview page with a read-first layout that supports per-field inline editing and a full-page edit mode.
 *
 * Initializes and synchronizes local edit form state from the loaded campaign, fetches campaign stats and available game systems, and exposes controls to save individual fields or all changes via the updateCampaign mutation.
 *
 * @returns The Campaign Overview page component ready to be mounted in the app UI
 */
export default function CampaignOverview() {
    const { id: campaignId } = useParams<{ id: string }>();
    const { setCurrentCampaignId } = useCampaignContext();

    // Editing state
    const [isEditing, setIsEditing] = useState(false);
    const [editingField, setEditingField] = useState<EditableField | null>(null);

    // Form data for editing
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        genre: '',
        imageStylePrompt: '',
    });

    // Fetch campaign data
    const {
        data: campaign,
        isLoading: campaignLoading,
        error: campaignError,
    } = useCampaign(campaignId ?? '', {
        enabled: !!campaignId,
    });

    // Fetch campaign stats
    const { data: stats } = useCampaignStats(campaignId ?? '', {
        enabled: !!campaignId,
    });

    // Fetch game systems
    const { data: gameSystems } = useGameSystems();

    // Update campaign mutation
    const updateCampaign = useUpdateCampaign();

    // Initialize form data when campaign loads
    useEffect(() => {
        if (campaign) {
            setFormData({
                name: campaign.name,
                description: campaign.description ?? '',
                genre: (campaign.settings?.genre as string) ?? '',
                imageStylePrompt: (campaign.settings?.imageStylePrompt as string) ?? '',
            });
            // Update the campaign context
            setCurrentCampaignId(campaign.id);
        }
    }, [campaign, setCurrentCampaignId]);

    /**
     * Get game system name by ID.
     */
    const getGameSystemName = (systemId: string): string => {
        const system = gameSystems?.find((gs: GameSystem) => gs.id === systemId);
        return system?.name ?? 'Unknown System';
    };

    /**
     * Start editing a specific field.
     */
    const handleFieldEdit = (field: EditableField) => {
        setEditingField(field);
    };

    /**
     * Cancel editing a field.
     */
    const handleFieldCancel = () => {
        if (campaign && editingField) {
            // Only reset the specific field being cancelled
            let originalValue: string;
            switch (editingField) {
                case 'name':
                    originalValue = campaign.name;
                    break;
                case 'description':
                    originalValue = campaign.description ?? '';
                    break;
                case 'genre':
                    originalValue = (campaign.settings?.genre as string) ?? '';
                    break;
                case 'imageStylePrompt':
                    originalValue = (campaign.settings?.imageStylePrompt as string) ?? '';
                    break;
            }
            setFormData(prev => ({
                ...prev,
                [editingField]: originalValue,
            }));
        }
        setEditingField(null);
    };

    /**
     * Save a single field.
     */
    const handleFieldSave = useCallback(async () => {
        if (!campaignId || !editingField) return;

        try {
            await updateCampaign.mutateAsync({
                id: campaignId,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    settings: {
                        genre: formData.genre || undefined,
                        imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                    },
                },
            });
            setEditingField(null);
        } catch (error) {
            console.error('Failed to save field:', error);
        }
    }, [campaignId, editingField, formData, updateCampaign]);

    /**
     * Toggle full edit mode.
     */
    const handleToggleEdit = () => {
        if (isEditing) {
            // Cancel editing
            if (campaign) {
                setFormData({
                    name: campaign.name,
                    description: campaign.description ?? '',
                    genre: (campaign.settings?.genre as string) ?? '',
                    imageStylePrompt: (campaign.settings?.imageStylePrompt as string) ?? '',
                });
            }
        }
        setIsEditing(!isEditing);
        setEditingField(null);
    };

    /**
     * Save all changes.
     */
    const handleSaveAll = useCallback(async () => {
        if (!campaignId) return;

        if (!formData.name.trim()) {
            // TODO: Show validation error
            return;
        }

        try {
            await updateCampaign.mutateAsync({
                id: campaignId,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    settings: {
                        genre: formData.genre || undefined,
                        imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                    },
                },
            });
            setIsEditing(false);
        } catch (error) {
            console.error('Failed to save campaign:', error);
        }
    }, [campaignId, formData, updateCampaign]);

    /**
     * Update form field.
     */
    const updateField = <K extends keyof typeof formData>(
        field: K,
        value: (typeof formData)[K]
    ) => {
        setFormData((prev) => ({ ...prev, [field]: value }));
    };

    // Loading state
    if (campaignLoading) {
        return (
            <Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                    <Skeleton variant="text" width={300} height={48} />
                    <Skeleton variant="rectangular" width={100} height={36} />
                </Box>
                <Paper sx={{ p: 3 }}>
                    <Skeleton variant="text" width={200} height={32} sx={{ mb: 2 }} />
                    <Skeleton variant="rectangular" height={100} sx={{ mb: 3 }} />
                    <Skeleton variant="text" width={150} height={24} sx={{ mb: 2 }} />
                    <Skeleton variant="text" width={250} height={24} />
                </Paper>
            </Box>
        );
    }

    // Error state
    if (campaignError) {
        return (
            <Box>
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to load campaign. Please try again later.
                </Alert>
            </Box>
        );
    }

    // No campaign selected state
    if (!campaign) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    minHeight: 400,
                }}
            >
                <Typography variant="h5" color="text.secondary" gutterBottom>
                    No Campaign Selected
                </Typography>
                <Typography variant="body1" color="text.secondary">
                    Select a campaign from the dropdown above or create a new one to get started.
                </Typography>
            </Box>
        );
    }

    return (
        <Box>
            {/* Header */}
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                    Campaign Overview
                </Typography>
                <Button
                    variant={isEditing ? 'contained' : 'outlined'}
                    startIcon={isEditing ? <SaveIcon /> : <EditIcon />}
                    onClick={isEditing ? handleSaveAll : handleToggleEdit}
                    disabled={updateCampaign.isPending}
                >
                    {updateCampaign.isPending ? (
                        <CircularProgress size={20} />
                    ) : isEditing ? (
                        'Save Changes'
                    ) : (
                        'Edit Campaign'
                    )}
                </Button>
            </Box>

            {/* Error alert */}
            {updateCampaign.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to save changes. Please try again.
                </Alert>
            )}

            {/* Cancel button when in edit mode */}
            {isEditing && (
                <Alert severity="info" sx={{ mb: 2 }}>
                    You are in edit mode. Make your changes and click "Save Changes" to save.
                    <Button
                        size="small"
                        onClick={handleToggleEdit}
                        sx={{ ml: 2 }}
                    >
                        Cancel
                    </Button>
                </Alert>
            )}

            {/* Quick Stats */}
            <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
                <Card sx={{ minWidth: 120 }}>
                    <CardContent sx={{ textAlign: 'center', py: 2, '&:last-child': { pb: 2 } }}>
                        <Typography variant="h4" color="primary">
                            {stats?.entityCounts
                                ? Object.values(stats.entityCounts).reduce((a, b) => a + b, 0)
                                : 0}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            Entities
                        </Typography>
                    </CardContent>
                </Card>
                <Card sx={{ minWidth: 120 }}>
                    <CardContent sx={{ textAlign: 'center', py: 2, '&:last-child': { pb: 2 } }}>
                        <Typography variant="h4" color="primary">
                            {stats?.relationshipCount ?? 0}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            Relationships
                        </Typography>
                    </CardContent>
                </Card>
                <Card sx={{ minWidth: 120 }}>
                    <CardContent sx={{ textAlign: 'center', py: 2, '&:last-child': { pb: 2 } }}>
                        <Typography variant="h4" color="primary">
                            {stats?.timelineEventCount ?? 0}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            Timeline Events
                        </Typography>
                    </CardContent>
                </Card>
            </Box>

            {/* Campaign Details */}
            <Paper sx={{ p: 3 }}>
                {/* Campaign Name */}
                <Box sx={{ mb: 3 }}>
                    {isEditing || editingField === 'name' ? (
                        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
                            <TextField
                                label="Campaign Name"
                                fullWidth
                                required
                                value={formData.name}
                                onChange={(e) => updateField('name', e.target.value)}
                                autoFocus={editingField === 'name'}
                            />
                            {editingField === 'name' && (
                                <>
                                    <IconButton
                                        color="primary"
                                        onClick={handleFieldSave}
                                        disabled={updateCampaign.isPending}
                                    >
                                        <CheckIcon />
                                    </IconButton>
                                    <IconButton onClick={handleFieldCancel}>
                                        <CancelIcon />
                                    </IconButton>
                                </>
                            )}
                        </Box>
                    ) : (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography variant="h5" sx={{ fontFamily: 'Cinzel' }}>
                                {campaign.name}
                            </Typography>
                            {!isEditing && (
                                <Tooltip title="Edit name">
                                    <IconButton
                                        size="small"
                                        onClick={() => handleFieldEdit('name')}
                                    >
                                        <EditIcon fontSize="small" />
                                    </IconButton>
                                </Tooltip>
                            )}
                        </Box>
                    )}
                </Box>

                {/* Game System */}
                <Box sx={{ mb: 3 }}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Game System
                    </Typography>
                    <Chip label={getGameSystemName(campaign.systemId)} />
                </Box>

                {/* Genre */}
                <Box sx={{ mb: 3 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                        <Typography variant="subtitle2" color="text.secondary">
                            Genre
                        </Typography>
                        {!isEditing && editingField !== 'genre' && (
                            <Tooltip title="Edit genre">
                                <IconButton
                                    size="small"
                                    onClick={() => handleFieldEdit('genre')}
                                >
                                    <EditIcon fontSize="small" />
                                </IconButton>
                            </Tooltip>
                        )}
                    </Box>
                    {isEditing || editingField === 'genre' ? (
                        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
                            <FormControl sx={{ minWidth: 200 }}>
                                <InputLabel>Genre</InputLabel>
                                <Select
                                    value={formData.genre}
                                    label="Genre"
                                    onChange={(e) => updateField('genre', e.target.value)}
                                >
                                    <MenuItem value="">
                                        <em>None</em>
                                    </MenuItem>
                                    {GENRE_OPTIONS.map((genre) => (
                                        <MenuItem key={genre} value={genre}>
                                            {genre}
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                            {editingField === 'genre' && (
                                <>
                                    <IconButton
                                        color="primary"
                                        onClick={handleFieldSave}
                                        disabled={updateCampaign.isPending}
                                    >
                                        <CheckIcon />
                                    </IconButton>
                                    <IconButton onClick={handleFieldCancel}>
                                        <CancelIcon />
                                    </IconButton>
                                </>
                            )}
                        </Box>
                    ) : (
                        <Typography variant="body1">
                            {(campaign.settings?.genre as string) || (
                                <em style={{ color: 'text.secondary' }}>Not set</em>
                            )}
                        </Typography>
                    )}
                </Box>

                {/* Description */}
                <Box sx={{ mb: 3 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                        <Typography variant="subtitle2" color="text.secondary">
                            Description
                        </Typography>
                        {!isEditing && editingField !== 'description' && (
                            <Tooltip title="Edit description">
                                <IconButton
                                    size="small"
                                    onClick={() => handleFieldEdit('description')}
                                >
                                    <EditIcon fontSize="small" />
                                </IconButton>
                            </Tooltip>
                        )}
                    </Box>
                    {isEditing || editingField === 'description' ? (
                        <Box>
                            <MarkdownEditor
                                label=""
                                value={formData.description}
                                onChange={(md) => updateField('description', md)}
                                placeholder="Describe your campaign setting, themes, and background..."
                                minHeight={150}
                                maxHeight={300}
                            />
                            {editingField === 'description' && (
                                <Box sx={{ mt: 1, display: 'flex', gap: 1 }}>
                                    <Button
                                        variant="contained"
                                        size="small"
                                        startIcon={<CheckIcon />}
                                        onClick={handleFieldSave}
                                        disabled={updateCampaign.isPending}
                                    >
                                        Save
                                    </Button>
                                    <Button
                                        size="small"
                                        startIcon={<CancelIcon />}
                                        onClick={handleFieldCancel}
                                    >
                                        Cancel
                                    </Button>
                                </Box>
                            )}
                        </Box>
                    ) : campaign.description ? (
                        <Box
                            sx={{
                                '& p': { mt: 0, mb: 1 },
                                '& p:last-child': { mb: 0 },
                            }}
                            dangerouslySetInnerHTML={{
                                __html: sanitizeHtml(campaign.description),
                            }}
                        />
                    ) : (
                        <Typography variant="body1" color="text.secondary" fontStyle="italic">
                            No description provided
                        </Typography>
                    )}
                </Box>

                {/* Image Style Prompt */}
                <Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                        <Typography variant="subtitle2" color="text.secondary">
                            Image Style Prompt
                        </Typography>
                        {!isEditing && editingField !== 'imageStylePrompt' && (
                            <Tooltip title="Edit image style prompt">
                                <IconButton
                                    size="small"
                                    onClick={() => handleFieldEdit('imageStylePrompt')}
                                >
                                    <EditIcon fontSize="small" />
                                </IconButton>
                            </Tooltip>
                        )}
                    </Box>
                    {isEditing || editingField === 'imageStylePrompt' ? (
                        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
                            <TextField
                                fullWidth
                                multiline
                                rows={3}
                                value={formData.imageStylePrompt}
                                onChange={(e) => updateField('imageStylePrompt', e.target.value)}
                                placeholder="e.g., 'dark gothic horror, oil painting style, muted colors, 1920s aesthetic'"
                                helperText="Default style prompt used when generating AI images for this campaign"
                            />
                            {editingField === 'imageStylePrompt' && (
                                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                                    <IconButton
                                        color="primary"
                                        onClick={handleFieldSave}
                                        disabled={updateCampaign.isPending}
                                    >
                                        <CheckIcon />
                                    </IconButton>
                                    <IconButton onClick={handleFieldCancel}>
                                        <CancelIcon />
                                    </IconButton>
                                </Box>
                            )}
                        </Box>
                    ) : (
                        <Typography variant="body1">
                            {(campaign.settings?.imageStylePrompt as string) || (
                                <em style={{ color: 'text.secondary' }}>Not set</em>
                            )}
                        </Typography>
                    )}
                </Box>
            </Paper>

            {/* Metadata */}
            <Box sx={{ mt: 3, display: 'flex', gap: 4 }}>
                <Box>
                    <Typography variant="caption" color="text.secondary">
                        Created
                    </Typography>
                    <Typography variant="body2">
                        {new Date(campaign.createdAt).toLocaleDateString(undefined, {
                            year: 'numeric',
                            month: 'long',
                            day: 'numeric',
                        })}
                    </Typography>
                </Box>
                <Box>
                    <Typography variant="caption" color="text.secondary">
                        Last Updated
                    </Typography>
                    <Typography variant="body2">
                        {new Date(campaign.updatedAt).toLocaleDateString(undefined, {
                            year: 'numeric',
                            month: 'long',
                            day: 'numeric',
                        })}
                    </Typography>
                </Box>
            </Box>
        </Box>
    );
}