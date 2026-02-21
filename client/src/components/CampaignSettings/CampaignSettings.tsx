// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * CampaignSettings - Settings form for campaign configuration.
 *
 * Provides editable fields for campaign name, description, game system,
 * genre, and image style prompt. Uses controlled form state with dirty
 * tracking for save functionality.
 */

import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    FormControl,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    TextField,
    Typography,
} from '@mui/material';
import { MarkdownEditor } from '../MarkdownEditor';
import { useDeleteCampaign, useGameSystems } from '../../hooks';
import { useCampaignContext } from '../../contexts/CampaignContext';
import type { Campaign, GameSystem } from '../../types';

/**
 * Available genre options for campaigns.
 * Re-exported from index.ts for fast refresh compatibility.
 */
const GENRE_OPTIONS = [
    'Comedy',
    'Cyberpunk',
    'Dark Fantasy',
    'Fantasy',
    'Historical',
    'Horror',
    'Military',
    'Modern Day',
    'Mystery',
    'Mythological',
    'Noir',
    'Post-Apocalyptic',
    'Pulp Adventure',
    'Sci-Fi',
    'Space Opera',
    'Steampunk',
    'Superhero',
    'Urban Fantasy',
    'Western',
    'Wuxia',
] as const;

/**
 * Form data structure for campaign settings.
 */
export interface CampaignSettingsData {
    name: string;
    description: string;
    gameSystemId: string;
    genre: string;
    imageStylePrompt: string;
}

/**
 * Props for the CampaignSettings component.
 */
interface CampaignSettingsProps {
    /** The campaign being edited */
    campaign: Campaign | null;
    /** Whether campaign data is loading */
    isLoading: boolean;
    /** Error from loading campaign data */
    error: Error | null;
    /** Callback when form data changes */
    onChange: (data: CampaignSettingsData, isDirty: boolean) => void;
    /** Optional initial form data (for external state management) */
    formData?: CampaignSettingsData;
}

/**
 * Default form values for a new campaign.
 */
const DEFAULT_FORM_DATA: CampaignSettingsData = {
    name: '',
    description: '',
    gameSystemId: '',
    genre: '',
    imageStylePrompt: '',
};

/**
 * Convert a campaign object to form data structure.
 */
function campaignToFormData(campaign: Campaign): CampaignSettingsData {
    return {
        name: campaign.name,
        description: campaign.description ?? '',
        gameSystemId: String(campaign.systemId),
        genre: (campaign.settings?.genre as string) ?? '',
        imageStylePrompt: (campaign.settings?.imageStylePrompt as string) ?? '',
    };
}

/**
 * Campaign settings form with editable fields for name, description,
 * game system, genre, and image style prompt.
 *
 * Tracks form changes and notifies parent component of dirty state
 * for save functionality.
 *
 * @param campaign - The campaign being edited
 * @param isLoading - Whether campaign data is still loading
 * @param error - Error from loading campaign data
 * @param onChange - Callback invoked when form data changes
 * @param formData - Optional external form data state
 * @returns A form component for editing campaign settings
 */
export default function CampaignSettings({
    campaign,
    isLoading,
    error,
    onChange,
    formData: externalFormData,
}: CampaignSettingsProps) {
    // Form state - use external state if provided, otherwise manage internally
    const [internalFormData, setInternalFormData] =
        useState<CampaignSettingsData>(DEFAULT_FORM_DATA);
    const [initialFormData, setInitialFormData] =
        useState<CampaignSettingsData>(DEFAULT_FORM_DATA);

    const formData = externalFormData ?? internalFormData;
    const isExternallyManaged = !!externalFormData;

    // Delete campaign state
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deleteConfirmName, setDeleteConfirmName] = useState('');
    const [deleteError, setDeleteError] = useState<string | null>(null);

    const navigate = useNavigate();
    const { setCurrentCampaignId } = useCampaignContext();
    const deleteCampaign = useDeleteCampaign();

    // Fetch game systems
    const {
        data: gameSystems,
        isLoading: gameSystemsLoading,
        error: gameSystemsError,
    } = useGameSystems();

    // Initialize form data when campaign loads
    useEffect(() => {
        if (campaign && !externalFormData) {
            const data = campaignToFormData(campaign);
            setInternalFormData(data);
            setInitialFormData(data);
        }
    }, [campaign, externalFormData]);

    // Track initial form data for external form data
    useEffect(() => {
        if (campaign && externalFormData) {
            setInitialFormData(campaignToFormData(campaign));
        }
    }, [campaign, externalFormData]);

    // Store latest onChange in ref to avoid effect re-runs
    const onChangeRef = useRef(onChange);
    useEffect(() => {
        onChangeRef.current = onChange;
    }, [onChange]);

    // Notify parent of changes when using internal state
    useEffect(() => {
        if (!externalFormData) {
            const isDirty =
                JSON.stringify(internalFormData) !== JSON.stringify(initialFormData);
            onChangeRef.current(internalFormData, isDirty);
        }
    }, [internalFormData, initialFormData, externalFormData]);

    /**
     * Update a form field.
     */
    const updateField = useCallback(
        <K extends keyof CampaignSettingsData>(
            field: K,
            value: CampaignSettingsData[K]
        ) => {
            const newData = { ...formData, [field]: value };
            if (isExternallyManaged) {
                const isDirty =
                    JSON.stringify(newData) !== JSON.stringify(initialFormData);
                onChange(newData, isDirty);
            } else {
                setInternalFormData(newData);
            }
        },
        [formData, isExternallyManaged, initialFormData, onChange]
    );

    /**
     * Get game system name by ID.
     */
    const getGameSystemName = (id: string): string => {
        const numericId = Number(id);
        const system = gameSystems?.find((gs: GameSystem) => gs.id === numericId);
        return system?.name ?? 'Unknown System';
    };

    /**
     * Open the delete confirmation dialog.
     */
    const handleOpenDeleteDialog = () => {
        setDeleteConfirmName('');
        setDeleteError(null);
        setDeleteDialogOpen(true);
    };

    /**
     * Close the delete confirmation dialog.
     */
    const handleCloseDeleteDialog = () => {
        setDeleteDialogOpen(false);
        setDeleteConfirmName('');
        setDeleteError(null);
    };

    /**
     * Delete the campaign after confirmation.
     */
    const handleDeleteCampaign = () => {
        if (!campaign) return;

        deleteCampaign.mutate(campaign.id, {
            onSuccess: () => {
                setCurrentCampaignId(null);
                navigate('/campaigns');
            },
            onError: (err: Error) => {
                setDeleteError(
                    err.message || 'Failed to delete campaign. Please try again.'
                );
            },
        });
    };

    // Loading state
    if (isLoading) {
        return (
            <Paper sx={{ p: 3, maxWidth: 800, mx: 'auto' }}>
                <Skeleton variant="rectangular" height={56} sx={{ mb: 3 }} />
                <Skeleton variant="rectangular" height={200} sx={{ mb: 3 }} />
                <Skeleton variant="rectangular" height={56} sx={{ mb: 3 }} />
                <Skeleton variant="rectangular" height={56} sx={{ mb: 3 }} />
                <Skeleton variant="rectangular" height={120} />
            </Paper>
        );
    }

    // Error state
    if (error) {
        return (
            <Alert severity="error" sx={{ maxWidth: 800, mx: 'auto' }}>
                Failed to load campaign settings. Please try again later.
            </Alert>
        );
    }

    // No campaign state
    if (!campaign) {
        return (
            <Alert severity="info" sx={{ maxWidth: 800, mx: 'auto' }}>
                No campaign selected.
            </Alert>
        );
    }

    return (
        <Paper sx={{ p: 3, maxWidth: 800, mx: 'auto' }}>
            <Typography
                variant="h5"
                component="h2"
                gutterBottom
                sx={{ fontFamily: 'Cinzel', mb: 3 }}
            >
                Campaign Settings
            </Typography>

            {/* Campaign Name */}
            <TextField
                label="Campaign Name"
                fullWidth
                required
                value={formData.name}
                onChange={(e) => updateField('name', e.target.value)}
                sx={{ mb: 3 }}
            />

            {/* Description */}
            <Box sx={{ mb: 3 }}>
                <MarkdownEditor
                    label="Description"
                    value={formData.description}
                    onChange={(md) => updateField('description', md)}
                    placeholder="Describe your campaign setting, themes, and background..."
                    minHeight={150}
                    maxHeight={300}
                    campaignId={campaign?.id}
                />
            </Box>

            {/* RPG System */}
            <FormControl fullWidth sx={{ mb: 3 }}>
                <InputLabel required>RPG System</InputLabel>
                <Select
                    value={formData.gameSystemId}
                    label="RPG System"
                    onChange={(e) => updateField('gameSystemId', e.target.value)}
                    disabled={gameSystemsLoading}
                >
                    {gameSystemsLoading ? (
                        <MenuItem disabled>Loading systems...</MenuItem>
                    ) : gameSystemsError ? (
                        <MenuItem disabled>Error loading systems</MenuItem>
                    ) : (
                        gameSystems?.map((system: GameSystem) => (
                            <MenuItem key={system.id} value={String(system.id)}>
                                {system.name}
                            </MenuItem>
                        ))
                    )}
                </Select>
                {formData.gameSystemId && (
                    <Typography
                        variant="caption"
                        color="text.secondary"
                        sx={{ mt: 0.5 }}
                    >
                        Currently: {getGameSystemName(formData.gameSystemId)}
                    </Typography>
                )}
            </FormControl>

            {/* Genre */}
            <FormControl fullWidth sx={{ mb: 3 }}>
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
                <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5 }}>
                    The overall genre or theme of your campaign
                </Typography>
            </FormControl>

            {/* Image Style Prompt */}
            <TextField
                label="Image Style Prompt"
                fullWidth
                multiline
                rows={4}
                value={formData.imageStylePrompt}
                onChange={(e) => updateField('imageStylePrompt', e.target.value)}
                placeholder="e.g., 'dark gothic horror, oil painting style, muted colors, 1920s aesthetic'"
                helperText="Default style prompt used when generating AI images for this campaign"
            />

            {/* Danger Zone */}
            <Divider sx={{ my: 4 }} />
            <Box
                sx={{
                    border: 1,
                    borderColor: 'error.main',
                    borderRadius: 1,
                    p: 3,
                    bgcolor: (theme) =>
                        theme.palette.mode === 'dark'
                            ? 'rgba(211, 47, 47, 0.08)'
                            : 'rgba(211, 47, 47, 0.04)',
                }}
            >
                <Typography
                    variant="h6"
                    color="error"
                    gutterBottom
                    sx={{ fontFamily: 'Cinzel' }}
                >
                    Danger Zone
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Permanently delete this campaign and all of its data,
                    including chapters, sessions, entities, and relationships.
                    This action cannot be undone.
                </Typography>
                <Button
                    variant="outlined"
                    color="error"
                    onClick={handleOpenDeleteDialog}
                >
                    Delete Campaign
                </Button>
            </Box>

            {/* Delete Confirmation Dialog */}
            <Dialog
                open={deleteDialogOpen}
                onClose={handleCloseDeleteDialog}
                maxWidth="sm"
                fullWidth
            >
                <DialogTitle>Delete Campaign</DialogTitle>
                <DialogContent>
                    <DialogContentText sx={{ mb: 2 }}>
                        This action is permanent and cannot be undone. All
                        campaign data will be lost, including chapters,
                        sessions, entities, relationships, and timeline events.
                    </DialogContentText>
                    <DialogContentText sx={{ mb: 2 }}>
                        To confirm, type the campaign name below:
                        {' '}<strong>{campaign.name}</strong>
                    </DialogContentText>
                    {deleteError && (
                        <Alert severity="error" sx={{ mb: 2 }}>
                            {deleteError}
                        </Alert>
                    )}
                    <TextField
                        autoFocus
                        fullWidth
                        label="Campaign name"
                        value={deleteConfirmName}
                        onChange={(e) => setDeleteConfirmName(e.target.value)}
                        placeholder={campaign.name}
                    />
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={handleCloseDeleteDialog}
                        disabled={deleteCampaign.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleDeleteCampaign}
                        color="error"
                        variant="contained"
                        disabled={
                            deleteConfirmName !== campaign.name ||
                            deleteCampaign.isPending
                        }
                    >
                        {deleteCampaign.isPending ? 'Deleting...' : 'Delete'}
                    </Button>
                </DialogActions>
            </Dialog>
        </Paper>
    );
}
