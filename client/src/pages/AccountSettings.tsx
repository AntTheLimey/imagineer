// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * AccountSettings - Full-screen page for managing user account settings.
 *
 * Provides configuration for LLM API keys used for content generation,
 * embeddings, and image generation.
 */

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Divider,
    FormControl,
    FormHelperText,
    IconButton,
    InputAdornment,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Snackbar,
    TextField,
    Typography,
} from '@mui/material';
import {
    Visibility as VisibilityIcon,
    VisibilityOff as VisibilityOffIcon,
} from '@mui/icons-material';
import { FullScreenLayout } from '../layouts';
import { useUnsavedChanges } from '../hooks';
import { useUserSettings, useUpdateUserSettings } from '../hooks/useUserSettings';
import type {
    ContentGenService,
    EmbeddingService,
    ImageGenService,
    UserSettingsUpdateRequest,
} from '../api/userSettings';

/**
 * Form data structure for account settings.
 */
interface SettingsFormData {
    contentGenService: ContentGenService | '';
    contentGenApiKey: string;
    embeddingService: EmbeddingService | '';
    embeddingApiKey: string;
    imageGenService: ImageGenService | '';
    imageGenApiKey: string;
}

/**
 * Track which API key fields have been modified from their masked values.
 */
interface ApiKeyModified {
    contentGenApiKey: boolean;
    embeddingApiKey: boolean;
    imageGenApiKey: boolean;
}

/**
 * Default form values.
 */
const DEFAULT_FORM_DATA: SettingsFormData = {
    contentGenService: '',
    contentGenApiKey: '',
    embeddingService: '',
    embeddingApiKey: '',
    imageGenService: '',
    imageGenApiKey: '',
};

/**
 * Default API key modified state.
 */
const DEFAULT_API_KEY_MODIFIED: ApiKeyModified = {
    contentGenApiKey: false,
    embeddingApiKey: false,
    imageGenApiKey: false,
};

/**
 * Service options for content generation.
 */
const CONTENT_GEN_SERVICES: { value: ContentGenService; label: string }[] = [
    { value: 'anthropic', label: 'Anthropic (Claude)' },
    { value: 'openai', label: 'OpenAI (GPT)' },
    { value: 'gemini', label: 'Google (Gemini)' },
];

/**
 * Service options for embedding generation.
 */
const EMBEDDING_SERVICES: { value: EmbeddingService; label: string }[] = [
    { value: 'voyage', label: 'Voyage AI' },
    { value: 'openai', label: 'OpenAI' },
    { value: 'gemini', label: 'Google (Gemini)' },
];

/**
 * Service options for image generation.
 */
const IMAGE_GEN_SERVICES: { value: ImageGenService; label: string }[] = [
    { value: 'openai', label: 'OpenAI (DALL-E)' },
    { value: 'stability', label: 'Stability AI' },
];

/**
 * Check if a value is a masked API key (contains asterisks).
 */
function isMaskedValue(value: string): boolean {
    return value.includes('****');
}

/**
 * Props for the API key input component.
 */
interface ApiKeyInputProps {
    label: string;
    value: string;
    onChange: (value: string) => void;
    error?: string;
    helperText?: string;
}

/**
 * Password-style input with show/hide toggle for API keys.
 */
function ApiKeyInput({ label, value, onChange, error, helperText }: ApiKeyInputProps) {
    const [showKey, setShowKey] = useState(false);

    const handleToggleVisibility = () => {
        setShowKey((prev) => !prev);
    };

    return (
        <TextField
            fullWidth
            label={label}
            type={showKey ? 'text' : 'password'}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            error={!!error}
            helperText={error || helperText}
            InputProps={{
                endAdornment: (
                    <InputAdornment position="end">
                        <IconButton
                            onClick={handleToggleVisibility}
                            edge="end"
                            aria-label={showKey ? 'Hide API key' : 'Show API key'}
                        >
                            {showKey ? <VisibilityOffIcon /> : <VisibilityIcon />}
                        </IconButton>
                    </InputAdornment>
                ),
            }}
        />
    );
}

/**
 * Full-screen page for managing user account settings including LLM API keys.
 *
 * @returns The React element for the Account Settings page.
 */
export default function AccountSettings() {
    const navigate = useNavigate();

    // Fetch current settings
    const { data: settings, isLoading, error: fetchError } = useUserSettings();
    const updateSettings = useUpdateUserSettings();

    // Form state
    const [formData, setFormData] = useState<SettingsFormData>(DEFAULT_FORM_DATA);
    const [apiKeyModified, setApiKeyModified] = useState<ApiKeyModified>(
        DEFAULT_API_KEY_MODIFIED
    );
    const [formErrors, setFormErrors] = useState<Partial<Record<keyof SettingsFormData, string>>>(
        {}
    );

    // Snackbar state
    const [snackbar, setSnackbar] = useState<{
        open: boolean;
        message: string;
        severity: 'success' | 'error';
    }>({
        open: false,
        message: '',
        severity: 'success',
    });

    // Unsaved changes protection
    const { isDirty, setIsDirty, clearDirty, checkUnsavedChanges, ConfirmDialog } =
        useUnsavedChanges({
            message:
                'You have unsaved changes to your settings. Are you sure you want to leave?',
        });

    // Initialize form data from fetched settings
    useEffect(() => {
        if (settings) {
            setFormData({
                contentGenService: settings.content_gen_service ?? '',
                contentGenApiKey: settings.content_gen_api_key ?? '',
                embeddingService: settings.embedding_service ?? '',
                embeddingApiKey: settings.embedding_api_key ?? '',
                imageGenService: settings.image_gen_service ?? '',
                imageGenApiKey: settings.image_gen_api_key ?? '',
            });
            // Reset modified state since we just loaded fresh data
            setApiKeyModified(DEFAULT_API_KEY_MODIFIED);
        }
    }, [settings]);

    /**
     * Update a form field and mark as dirty.
     */
    const updateField = useCallback(
        <K extends keyof SettingsFormData>(field: K, value: SettingsFormData[K]) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
            setIsDirty(true);
            // Clear error for this field
            setFormErrors((prev) => ({ ...prev, [field]: undefined }));

            // Track if an API key field has been modified
            if (field === 'contentGenApiKey' || field === 'embeddingApiKey' || field === 'imageGenApiKey') {
                setApiKeyModified((prev) => ({ ...prev, [field]: true }));
            }
        },
        [setIsDirty]
    );

    /**
     * Validate the form.
     */
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof SettingsFormData, string>> = {};

        // If a service is selected, the API key is required
        if (formData.contentGenService && !formData.contentGenApiKey) {
            errors.contentGenApiKey = 'API key is required when a service is selected';
        }
        if (formData.embeddingService && !formData.embeddingApiKey) {
            errors.embeddingApiKey = 'API key is required when a service is selected';
        }
        if (formData.imageGenService && !formData.imageGenApiKey) {
            errors.imageGenApiKey = 'API key is required when a service is selected';
        }

        // If a service is selected and API key was modified, it cannot be empty
        // (allows keeping masked value if not modified)
        if (
            formData.contentGenService &&
            apiKeyModified.contentGenApiKey &&
            !formData.contentGenApiKey.trim()
        ) {
            errors.contentGenApiKey = 'API key cannot be empty';
        }
        if (
            formData.embeddingService &&
            apiKeyModified.embeddingApiKey &&
            !formData.embeddingApiKey.trim()
        ) {
            errors.embeddingApiKey = 'API key cannot be empty';
        }
        if (
            formData.imageGenService &&
            apiKeyModified.imageGenApiKey &&
            !formData.imageGenApiKey.trim()
        ) {
            errors.imageGenApiKey = 'API key cannot be empty';
        }

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, [formData, apiKeyModified]);

    /**
     * Build the update request, only including API keys if they were modified.
     */
    const buildUpdateRequest = useCallback((): UserSettingsUpdateRequest => {
        const request: UserSettingsUpdateRequest = {};

        // Always include service selections (even if clearing)
        request.content_gen_service = formData.contentGenService || null;
        request.embedding_service = formData.embeddingService || null;
        request.image_gen_service = formData.imageGenService || null;

        // Only include API keys if they were actually modified
        // (not if they still contain masked values)
        if (apiKeyModified.contentGenApiKey && !isMaskedValue(formData.contentGenApiKey)) {
            request.content_gen_api_key = formData.contentGenApiKey;
        }
        if (apiKeyModified.embeddingApiKey && !isMaskedValue(formData.embeddingApiKey)) {
            request.embedding_api_key = formData.embeddingApiKey;
        }
        if (apiKeyModified.imageGenApiKey && !isMaskedValue(formData.imageGenApiKey)) {
            request.image_gen_api_key = formData.imageGenApiKey;
        }

        return request;
    }, [formData, apiKeyModified]);

    /**
     * Save the settings.
     */
    const handleSave = useCallback(async (): Promise<boolean> => {
        if (!validateForm()) {
            return false;
        }

        try {
            const request = buildUpdateRequest();
            await updateSettings.mutateAsync(request);

            clearDirty();
            setApiKeyModified(DEFAULT_API_KEY_MODIFIED);
            setSnackbar({
                open: true,
                message: 'Settings saved successfully',
                severity: 'success',
            });
            return true;
        } catch (error) {
            console.error('Failed to save settings:', error);
            setSnackbar({
                open: true,
                message: 'Failed to save settings. Please try again.',
                severity: 'error',
            });
            return false;
        }
    }, [validateForm, buildUpdateRequest, updateSettings, clearDirty]);

    /**
     * Save and close.
     */
    const handleSaveAndClose = useCallback(async () => {
        const saved = await handleSave();
        if (saved) {
            navigate('/');
        }
    }, [handleSave, navigate]);

    /**
     * Handle back navigation with unsaved changes check.
     */
    const handleBack = useCallback(() => {
        const goBack = () => navigate('/');
        if (!checkUnsavedChanges(goBack)) {
            goBack();
        }
    }, [navigate, checkUnsavedChanges]);

    /**
     * Close the snackbar.
     */
    const handleCloseSnackbar = useCallback(() => {
        setSnackbar((prev) => ({ ...prev, open: false }));
    }, []);

    // Build breadcrumbs
    const breadcrumbs = useMemo(
        () => [
            { label: 'Home', path: '/' },
            { label: 'Account Settings' },
        ],
        []
    );

    // Loading state
    if (isLoading) {
        return (
            <FullScreenLayout
                title="Account Settings"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                backPath="/"
            >
                <Box sx={{ maxWidth: 600, mx: 'auto' }}>
                    <Skeleton variant="rectangular" height={200} sx={{ mb: 3 }} />
                    <Skeleton variant="rectangular" height={200} sx={{ mb: 3 }} />
                    <Skeleton variant="rectangular" height={200} />
                </Box>
            </FullScreenLayout>
        );
    }

    // Fetch error state
    if (fetchError) {
        return (
            <FullScreenLayout
                title="Account Settings"
                breadcrumbs={breadcrumbs}
                showSaveButtons={false}
                backPath="/"
            >
                <Alert severity="error">
                    Failed to load settings. Please try again later.
                </Alert>
            </FullScreenLayout>
        );
    }

    return (
        <FullScreenLayout
            title="Account Settings"
            breadcrumbs={breadcrumbs}
            isDirty={isDirty}
            isSaving={updateSettings.isPending}
            onSave={handleSave}
            onSaveAndClose={handleSaveAndClose}
            onBack={handleBack}
        >
            <Box sx={{ maxWidth: 600, mx: 'auto' }}>
                {/* Mutation error alert */}
                {updateSettings.error && (
                    <Alert severity="error" sx={{ mb: 3 }}>
                        Failed to save settings. Please try again.
                    </Alert>
                )}

                {/* Content Generation Section */}
                <Paper sx={{ p: 3, mb: 3 }}>
                    <Typography variant="h6" gutterBottom>
                        Content Generation
                    </Typography>
                    <Typography variant="body2" color="text.secondary" paragraph>
                        Configure the AI service used for generating campaign content,
                        descriptions, and suggestions.
                    </Typography>
                    <Divider sx={{ mb: 3 }} />

                    <FormControl fullWidth sx={{ mb: 3 }}>
                        <InputLabel>Service</InputLabel>
                        <Select
                            value={formData.contentGenService}
                            label="Service"
                            onChange={(e) =>
                                updateField(
                                    'contentGenService',
                                    e.target.value as ContentGenService | ''
                                )
                            }
                        >
                            <MenuItem value="">
                                <em>None</em>
                            </MenuItem>
                            {CONTENT_GEN_SERVICES.map((service) => (
                                <MenuItem key={service.value} value={service.value}>
                                    {service.label}
                                </MenuItem>
                            ))}
                        </Select>
                        <FormHelperText>
                            Select the AI provider for content generation
                        </FormHelperText>
                    </FormControl>

                    <ApiKeyInput
                        label="API Key"
                        value={formData.contentGenApiKey}
                        onChange={(value) => updateField('contentGenApiKey', value)}
                        error={formErrors.contentGenApiKey}
                        helperText="Enter your API key for the selected service"
                    />
                </Paper>

                {/* Embedding Generation Section */}
                <Paper sx={{ p: 3, mb: 3 }}>
                    <Typography variant="h6" gutterBottom>
                        Embedding Generation
                    </Typography>
                    <Typography variant="body2" color="text.secondary" paragraph>
                        Configure the AI service used for generating embeddings for
                        semantic search and similarity matching.
                    </Typography>
                    <Divider sx={{ mb: 3 }} />

                    <FormControl fullWidth sx={{ mb: 3 }}>
                        <InputLabel>Service</InputLabel>
                        <Select
                            value={formData.embeddingService}
                            label="Service"
                            onChange={(e) =>
                                updateField(
                                    'embeddingService',
                                    e.target.value as EmbeddingService | ''
                                )
                            }
                        >
                            <MenuItem value="">
                                <em>None</em>
                            </MenuItem>
                            {EMBEDDING_SERVICES.map((service) => (
                                <MenuItem key={service.value} value={service.value}>
                                    {service.label}
                                </MenuItem>
                            ))}
                        </Select>
                        <FormHelperText>
                            Select the AI provider for embedding generation
                        </FormHelperText>
                    </FormControl>

                    <ApiKeyInput
                        label="API Key"
                        value={formData.embeddingApiKey}
                        onChange={(value) => updateField('embeddingApiKey', value)}
                        error={formErrors.embeddingApiKey}
                        helperText="Enter your API key for the selected service"
                    />
                </Paper>

                {/* Image Generation Section */}
                <Paper sx={{ p: 3, mb: 3 }}>
                    <Typography variant="h6" gutterBottom>
                        Image Generation
                    </Typography>
                    <Typography variant="body2" color="text.secondary" paragraph>
                        Configure the AI service used for generating images for
                        characters, locations, and other campaign elements.
                    </Typography>
                    <Divider sx={{ mb: 3 }} />

                    <FormControl fullWidth sx={{ mb: 3 }}>
                        <InputLabel>Service</InputLabel>
                        <Select
                            value={formData.imageGenService}
                            label="Service"
                            onChange={(e) =>
                                updateField(
                                    'imageGenService',
                                    e.target.value as ImageGenService | ''
                                )
                            }
                        >
                            <MenuItem value="">
                                <em>None</em>
                            </MenuItem>
                            {IMAGE_GEN_SERVICES.map((service) => (
                                <MenuItem key={service.value} value={service.value}>
                                    {service.label}
                                </MenuItem>
                            ))}
                        </Select>
                        <FormHelperText>
                            Select the AI provider for image generation
                        </FormHelperText>
                    </FormControl>

                    <ApiKeyInput
                        label="API Key"
                        value={formData.imageGenApiKey}
                        onChange={(value) => updateField('imageGenApiKey', value)}
                        error={formErrors.imageGenApiKey}
                        helperText="Enter your API key for the selected service"
                    />
                </Paper>
            </Box>

            {/* Navigation confirmation dialog */}
            {ConfirmDialog}

            {/* Success/Error snackbar */}
            <Snackbar
                open={snackbar.open}
                autoHideDuration={6000}
                onClose={handleCloseSnackbar}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
            >
                <Alert
                    onClose={handleCloseSnackbar}
                    severity={snackbar.severity}
                    sx={{ width: '100%' }}
                >
                    {snackbar.message}
                </Alert>
            </Snackbar>
        </FullScreenLayout>
    );
}
