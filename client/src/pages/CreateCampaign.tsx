// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * CreateCampaign - Full-page form for creating a new campaign.
 *
 * Uses the FullScreenLayout with breadcrumbs and provides all campaign
 * fields including name, description, game system, genre, and image style
 * prompt.
 */

import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    CircularProgress,
    FormControl,
    FormHelperText,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    TextField,
    Typography,
} from '@mui/material';
import { FullScreenLayout } from '../layouts';
import { RichTextEditor } from '../components/RichTextEditor';
import { useCreateCampaign, useGameSystems } from '../hooks';
import { useCampaignContext } from '../contexts/CampaignContext';

/**
 * Available genre options for campaigns.
 */
const GENRE_OPTIONS = [
    'Fantasy',
    'Sci-Fi',
    'Horror',
    'Mystery',
    'Cyberpunk',
    'Steampunk',
    'Post-Apocalyptic',
    'Superhero',
    'Historical',
    'Urban Fantasy',
    'Space Opera',
    'Dark Fantasy',
    'Pulp Adventure',
    'Western',
    'Mythological',
    'Modern Day',
    'Noir',
    'Wuxia',
    'Military',
    'Comedy',
] as const;

/**
 * Form state for creating a campaign.
 */
interface CampaignFormData {
    name: string;
    description: string;
    systemId: string;
    genre: string;
    imageStylePrompt: string;
}

/**
 * CreateCampaign page component.
 *
 * Provides a full-page form for creating a new campaign with all available
 * fields. After successful creation, sets the new campaign as current and
 * navigates to its overview page.
 *
 * @returns The CreateCampaign page component
 */
export default function CreateCampaign() {
    const navigate = useNavigate();
    const { setCurrentCampaignId } = useCampaignContext();

    // Form state
    const [formData, setFormData] = useState<CampaignFormData>({
        name: '',
        description: '',
        systemId: '',
        genre: '',
        imageStylePrompt: '',
    });

    // Validation state
    const [errors, setErrors] = useState<Partial<Record<keyof CampaignFormData, string>>>({});

    // Fetch game systems
    const { data: gameSystems, isLoading: gameSystemsLoading } = useGameSystems();

    // Create campaign mutation
    const createCampaign = useCreateCampaign();

    /**
     * Check if the form has been modified.
     */
    const isDirty = Boolean(
        formData.name ||
        formData.description ||
        formData.systemId ||
        formData.genre ||
        formData.imageStylePrompt
    );

    /**
     * Update a form field.
     */
    const updateField = <K extends keyof CampaignFormData>(
        field: K,
        value: CampaignFormData[K]
    ) => {
        setFormData((prev) => ({ ...prev, [field]: value }));
        // Clear error when field is modified
        if (errors[field]) {
            setErrors((prev) => {
                const newErrors = { ...prev };
                delete newErrors[field];
                return newErrors;
            });
        }
    };

    /**
     * Validate the form and return whether it's valid.
     */
    const validateForm = useCallback((): boolean => {
        const newErrors: Partial<Record<keyof CampaignFormData, string>> = {};

        if (!formData.name.trim()) {
            newErrors.name = 'Campaign name is required';
        }

        if (!formData.systemId) {
            newErrors.systemId = 'Game system is required';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    }, [formData.name, formData.systemId]);

    /**
     * Handle form submission.
     */
    const handleSave = useCallback(async () => {
        if (!validateForm()) {
            return;
        }

        try {
            const created = await createCampaign.mutateAsync({
                name: formData.name.trim(),
                systemId: formData.systemId,
                description: formData.description.trim() || undefined,
                settings: {
                    genre: formData.genre || undefined,
                    imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                },
            });

            // Set as current campaign and navigate to overview
            setCurrentCampaignId(created.id);
            navigate(`/campaigns/${created.id}/overview`);
        } catch (error) {
            console.error('Failed to create campaign:', error);
        }
    }, [formData, validateForm, createCampaign, setCurrentCampaignId, navigate]);

    /**
     * Handle cancel - navigate back.
     */
    const handleCancel = useCallback(() => {
        // Navigate to home if there's no history, otherwise go back
        if (window.history.length <= 2) {
            navigate('/');
        } else {
            navigate(-1);
        }
    }, [navigate]);

    // Breadcrumbs for navigation
    const breadcrumbs = [
        { label: 'Home', path: '/' },
        { label: 'Create Campaign' },
    ];

    return (
        <FullScreenLayout
            title="Create Campaign"
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={createCampaign.isPending}
            onSave={handleSave}
            onSaveAndClose={handleSave}
            onBack={handleCancel}
            backPath="/"
        >
            <Box sx={{ maxWidth: 800, mx: 'auto' }}>
                {/* Error alert */}
                {createCampaign.error && (
                    <Alert severity="error" sx={{ mb: 3 }}>
                        Failed to create campaign. Please try again.
                    </Alert>
                )}

                <Paper sx={{ p: 3 }}>
                    <Typography variant="h6" gutterBottom sx={{ mb: 3 }}>
                        Campaign Details
                    </Typography>

                    {/* Campaign Name */}
                    <TextField
                        label="Campaign Name"
                        fullWidth
                        required
                        value={formData.name}
                        onChange={(e) => updateField('name', e.target.value)}
                        error={Boolean(errors.name)}
                        helperText={errors.name}
                        sx={{ mb: 3 }}
                        autoFocus
                    />

                    {/* Game System */}
                    <FormControl
                        fullWidth
                        required
                        error={Boolean(errors.systemId)}
                        sx={{ mb: 3 }}
                    >
                        <InputLabel>Game System</InputLabel>
                        <Select
                            value={formData.systemId}
                            label="Game System"
                            onChange={(e) => updateField('systemId', e.target.value)}
                            disabled={gameSystemsLoading}
                        >
                            {gameSystemsLoading ? (
                                <MenuItem disabled>
                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                        <CircularProgress size={16} />
                                        Loading...
                                    </Box>
                                </MenuItem>
                            ) : (
                                gameSystems?.map((system) => (
                                    <MenuItem key={system.id} value={system.id}>
                                        {system.name}
                                    </MenuItem>
                                ))
                            )}
                        </Select>
                        {errors.systemId && (
                            <FormHelperText>{errors.systemId}</FormHelperText>
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
                        <FormHelperText>
                            Optional. Helps set the tone for AI-generated content.
                        </FormHelperText>
                    </FormControl>

                    {/* Description */}
                    <Box sx={{ mb: 3 }}>
                        <RichTextEditor
                            label="Description"
                            value={formData.description}
                            onChange={(html) => updateField('description', html)}
                            placeholder="Describe your campaign setting, themes, and background..."
                            minHeight={150}
                            maxHeight={300}
                        />
                    </Box>

                    {/* Image Style Prompt */}
                    <TextField
                        label="Image Style Prompt"
                        fullWidth
                        multiline
                        rows={3}
                        value={formData.imageStylePrompt}
                        onChange={(e) => updateField('imageStylePrompt', e.target.value)}
                        placeholder="e.g., 'dark gothic horror, oil painting style, muted colors, 1920s aesthetic'"
                        helperText="Default style prompt used when generating AI images for this campaign"
                    />
                </Paper>
            </Box>
        </FullScreenLayout>
    );
}
