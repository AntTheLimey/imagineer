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
import { useParams, useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CardContent,
    Chip,
    FormControl,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Snackbar,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Edit as EditIcon,
    Close as CancelIcon,
    Check as CheckIcon,
} from '@mui/icons-material';
import { AnalysisBadge } from '../components/AnalysisBadge';
import { PhaseStrip } from '../components/PhaseStrip';
import type { PhaseSelection } from '../components/PhaseStrip';
import { MarkdownEditor } from '../components/MarkdownEditor';
import { GENRE_OPTIONS } from '../components/CampaignSettings';
import {
    useCampaign,
    useUpdateCampaign,
    useGameSystems,
    useCampaignStats,
    useEntities,
} from '../hooks';
import { useUserSettings } from '../hooks/useUserSettings';
import { useCampaignContext } from '../contexts/CampaignContext';
import { MarkdownRenderer } from '../components/MarkdownRenderer';
import type { WikiLinkEntity } from '../components/MarkdownRenderer';
import type { GameSystem } from '../types';

/**
 * Small helper that redirects to the campaigns list when a campaign can't
 * be loaded (e.g. after deletion). Extracted as a component so the
 * useEffect/useNavigate hooks run at the right lifecycle point.
 */
function CampaignErrorRedirect() {
    const navigate = useNavigate();
    useEffect(() => {
        navigate('/campaigns', { replace: true });
    }, [navigate]);
    return null;
}

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
    const { id } = useParams<{ id: string }>();
    const campaignId = id ? Number(id) : undefined;
    const navigate = useNavigate();
    const { setCurrentCampaignId } = useCampaignContext();

    /**
     * Navigate to entities page filtered by the clicked wiki link name.
     */
    const handleEntityClick = useCallback((name: string) => {
        if (campaignId) {
            navigate(`/campaigns/${campaignId}/entities?search=${encodeURIComponent(name)}`);
        }
    }, [campaignId, navigate]);

    /**
     * Navigate to the entity view page for a specific entity.
     */
    const handleEntityNavigate = useCallback((entityId: number) => {
        if (campaignId) {
            navigate(`/campaigns/${campaignId}/entities/${entityId}`);
        }
    }, [campaignId, navigate]);

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
    } = useCampaign(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Fetch entities for wiki link resolution
    const { data: entitiesData } = useEntities({
        campaignId: campaignId ?? 0,
    });

    // Map entities to WikiLinkEntity shape for MarkdownRenderer
    const wikiLinkEntities: WikiLinkEntity[] | undefined = entitiesData?.map(
        (e) => ({
            id: e.id,
            name: e.name,
            entityType: e.entityType,
            description: e.description ?? null,
        })
    );

    // Fetch campaign stats
    const { data: stats } = useCampaignStats(campaignId ?? 0, {
        enabled: !!campaignId,
    });

    // Fetch game systems
    const { data: gameSystems } = useGameSystems();

    // Update campaign mutation
    const updateCampaign = useUpdateCampaign();

    // LLM availability check — used to disable AI-dependent phases
    const { data: userSettings } = useUserSettings();
    const hasLLM = !!userSettings?.contentGenService
                && !!userSettings?.contentGenApiKey;

    const disabledPhases = !hasLLM
        ? { revise: 'Configure an LLM in Account Settings',
            enrich: 'Configure an LLM in Account Settings' }
        : undefined;

    // Analysis snackbar state
    const [analysisSnackbar, setAnalysisSnackbar] = useState<{
        open: boolean;
        jobId: number;
        count: number;
        message?: string;
    }>({ open: false, jobId: 0, count: 0 });

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
    const getGameSystemName = (systemId: number): string => {
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
            const result = await updateCampaign.mutateAsync({
                id: campaignId!,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    settings: {
                        genre: formData.genre || undefined,
                        imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                    },
                },
            });

            // Check for analysis results
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const analysisResult = (result as any)?._analysis;
            if (analysisResult?.pendingCount > 0) {
                setEditingField(null);
                navigate(`/campaigns/${campaignId}/analysis/${analysisResult.jobId}`);
                return;
            }

            setEditingField(null);
        } catch (error) {
            console.error('Failed to save field:', error);
        }
    }, [campaignId, editingField, formData, updateCampaign, navigate]);

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
     * Save all changes with phase selection that controls which pipeline
     * stages run after saving.
     *
     * @param phases - Which workflow phases the GM has selected.
     */
    const handleSaveAll = useCallback(async (phases: PhaseSelection) => {
        if (!campaignId) return;

        if (!formData.name.trim()) {
            // TODO: Show validation error
            return;
        }

        const hasAnyPhase = phases.identify || phases.revise || phases.enrich;

        // Build phases array from selection
        const phaseKeys: string[] = [];
        if (phases.identify) phaseKeys.push('identify');
        if (phases.revise) phaseKeys.push('revise');
        if (phases.enrich) phaseKeys.push('enrich');

        try {
            const result = await updateCampaign.mutateAsync({
                id: campaignId!,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    settings: {
                        genre: formData.genre || undefined,
                        imageStylePrompt: formData.imageStylePrompt.trim() || undefined,
                    },
                },
                options: {
                    analyze: phases.identify || phases.revise,
                    enrich: phases.enrich,
                    phases: phaseKeys.length > 0 ? phaseKeys : undefined,
                },
            });

            // Check for analysis results
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const analysisResult = (result as any)?._analysis;
            if (hasAnyPhase && analysisResult?.jobId) {
                navigate(`/campaigns/${campaignId}/analysis/${analysisResult.jobId}`);
            } else if (hasAnyPhase && analysisResult?.pendingCount > 0) {
                setAnalysisSnackbar({
                    open: true,
                    jobId: analysisResult.jobId,
                    count: analysisResult.pendingCount,
                });
            } else if (!hasAnyPhase) {
                // Plain save - no navigation
                setIsEditing(false);
            } else {
                setAnalysisSnackbar({
                    open: true,
                    jobId: analysisResult?.jobId ?? 0,
                    count: 0,
                    message: 'Analysis complete: no issues found.',
                });
            }

            setIsEditing(false);
        } catch (error) {
            console.error('Failed to save campaign:', error);
        }
    }, [campaignId, formData, updateCampaign, navigate]);

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

    // Error state - redirect to campaigns list (e.g. after deletion)
    if (campaignError) {
        return <CampaignErrorRedirect />;
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
                {isEditing ? (
                    <PhaseStrip
                        onSave={handleSaveAll}
                        isDirty={isEditing}
                        isSaving={updateCampaign.isPending}
                        disabledPhases={disabledPhases}
                    />
                ) : (
                    <Button
                        variant="outlined"
                        startIcon={<EditIcon />}
                        onClick={handleToggleEdit}
                    >
                        Edit Campaign
                    </Button>
                )}
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
                    You are in edit mode. Make your changes and use Save &amp; Go to save.
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={handleToggleEdit}
                        sx={{ ml: 2 }}
                    >
                        Cancel
                    </Button>
                </Alert>
            )}

            {/* Pending analysis banner */}
            {campaignId && (
                <Box sx={{ mb: 2 }}>
                    <AnalysisBadge
                        campaignId={campaignId}
                        variant="banner"
                    />
                </Box>
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
                                campaignId={campaignId}
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
                        >
                            <MarkdownRenderer
                                content={campaign.description}
                                onEntityClick={handleEntityClick}
                                entities={wikiLinkEntities}
                                onEntityNavigate={handleEntityNavigate}
                            />
                        </Box>
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

            {/* Analysis results snackbar */}
            <Snackbar
                open={analysisSnackbar.open}
                autoHideDuration={analysisSnackbar.count === 0 ? 4000 : 10000}
                onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
            >
                {analysisSnackbar.count === 0 ? (
                    <Alert
                        severity="success"
                        onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
                    >
                        {analysisSnackbar.message ?? 'Analysis complete: no issues found.'}
                    </Alert>
                ) : (
                    <Alert
                        severity="warning"
                        onClose={() => setAnalysisSnackbar(prev => ({ ...prev, open: false }))}
                        action={
                            <Button
                                color="inherit"
                                size="small"
                                onClick={() => {
                                    setAnalysisSnackbar(prev => ({ ...prev, open: false }));
                                    navigate(`/campaigns/${campaignId}/analysis/${analysisSnackbar.jobId}`);
                                }}
                            >
                                Review Now
                            </Button>
                        }
                    >
                        {`Analysis found ${analysisSnackbar.count} item${analysisSnackbar.count === 1 ? '' : 's'} to review`}
                    </Alert>
                )}
            </Snackbar>

        </Box>
    );
}