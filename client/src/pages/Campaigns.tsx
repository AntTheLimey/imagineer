// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    Card,
    CardActions,
    CardContent,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    Grid,
    InputLabel,
    MenuItem,
    Select,
    Skeleton,
    TextField,
    Typography,
} from '@mui/material';
import { Add as AddIcon } from '@mui/icons-material';
import { useCampaigns, useCreateCampaign, useGameSystems } from '../hooks';
import type { Campaign } from '../types';

export default function Campaigns() {
    const navigate = useNavigate();
    const [open, setOpen] = useState(false);
    const [newCampaign, setNewCampaign] = useState({
        name: '',
        systemId: '',
        description: '',
    });

    // Fetch campaigns and game systems
    const {
        data: campaignsResponse,
        isLoading: campaignsLoading,
        error: campaignsError,
    } = useCampaigns();

    const {
        data: gameSystems,
        isLoading: gameSystemsLoading,
    } = useGameSystems();

    // Create campaign mutation
    const createCampaign = useCreateCampaign();

    const campaigns = campaignsResponse ?? [];

    const handleCreate = async () => {
        if (!newCampaign.name || !newCampaign.systemId) return;

        try {
            await createCampaign.mutateAsync({
                name: newCampaign.name,
                systemId: newCampaign.systemId,
                description: newCampaign.description || undefined,
            });
            setOpen(false);
            setNewCampaign({ name: '', systemId: '', description: '' });
        } catch (error) {
            // Error is handled by mutation state
            console.error('Failed to create campaign:', error);
        }
    };

    const handleCampaignClick = (campaign: Campaign) => {
        navigate(`/campaigns/${campaign.id}/entities`);
    };

    // Loading state
    if (campaignsLoading) {
        return (
            <Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                    <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                        Campaigns
                    </Typography>
                    <Skeleton variant="rectangular" width={150} height={36} />
                </Box>
                <Grid container spacing={3}>
                    {[1, 2, 3].map((i) => (
                        <Grid item xs={12} sm={6} md={4} key={i}>
                            <Skeleton variant="rectangular" height={180} />
                        </Grid>
                    ))}
                </Grid>
            </Box>
        );
    }

    return (
        <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                    Campaigns
                </Typography>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => setOpen(true)}
                >
                    New Campaign
                </Button>
            </Box>

            {campaignsError && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to load campaigns. Please try again later.
                </Alert>
            )}

            {createCampaign.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to create campaign. Please try again.
                </Alert>
            )}

            <Grid container spacing={3}>
                {campaigns.length === 0 ? (
                    <Grid item xs={12}>
                        <Card sx={{ textAlign: 'center', py: 4 }}>
                            <CardContent>
                                <Typography variant="h6" color="text.secondary">
                                    No campaigns yet
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    Create your first campaign to get started
                                </Typography>
                            </CardContent>
                            <CardActions sx={{ justifyContent: 'center' }}>
                                <Button variant="outlined" onClick={() => setOpen(true)}>
                                    Create Campaign
                                </Button>
                            </CardActions>
                        </Card>
                    </Grid>
                ) : (
                    campaigns.map((campaign) => {
                        const gameSystem = gameSystems?.find(
                            (gs) => gs.id === campaign.systemId
                        );
                        return (
                            <Grid item xs={12} sm={6} md={4} key={campaign.id}>
                                <Card
                                    sx={{
                                        cursor: 'pointer',
                                        '&:hover': {
                                            boxShadow: 6,
                                        },
                                    }}
                                    onClick={() => handleCampaignClick(campaign)}
                                >
                                    <CardContent>
                                        <Typography variant="h6" gutterBottom>
                                            {campaign.name}
                                        </Typography>
                                        <Typography
                                            variant="body2"
                                            color="text.secondary"
                                            sx={{ mb: 1 }}
                                        >
                                            {gameSystem?.name ?? 'Unknown System'}
                                        </Typography>
                                        {campaign.description && (
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                                sx={{
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    display: '-webkit-box',
                                                    WebkitLineClamp: 2,
                                                    WebkitBoxOrient: 'vertical',
                                                }}
                                            >
                                                {campaign.description}
                                            </Typography>
                                        )}
                                    </CardContent>
                                    <CardActions>
                                        <Button
                                            size="small"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                navigate(`/campaigns/${campaign.id}/entities`);
                                            }}
                                        >
                                            View Entities
                                        </Button>
                                        <Button
                                            size="small"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                navigate(`/campaigns/${campaign.id}/timeline`);
                                            }}
                                        >
                                            Timeline
                                        </Button>
                                    </CardActions>
                                </Card>
                            </Grid>
                        );
                    })
                )}
            </Grid>

            <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
                <DialogTitle>Create New Campaign</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="Campaign Name"
                        fullWidth
                        variant="outlined"
                        value={newCampaign.name}
                        onChange={(e) => setNewCampaign({ ...newCampaign, name: e.target.value })}
                        sx={{ mb: 2 }}
                    />
                    <FormControl fullWidth sx={{ mb: 2 }}>
                        <InputLabel>Game System</InputLabel>
                        <Select
                            value={newCampaign.systemId}
                            label="Game System"
                            onChange={(e) =>
                                setNewCampaign({ ...newCampaign, systemId: e.target.value })
                            }
                            disabled={gameSystemsLoading}
                        >
                            {gameSystemsLoading ? (
                                <MenuItem disabled>Loading...</MenuItem>
                            ) : (
                                gameSystems?.map((system) => (
                                    <MenuItem key={system.id} value={system.id}>
                                        {system.name}
                                    </MenuItem>
                                ))
                            )}
                        </Select>
                    </FormControl>
                    <TextField
                        margin="dense"
                        label="Description"
                        fullWidth
                        multiline
                        rows={3}
                        variant="outlined"
                        value={newCampaign.description}
                        onChange={(e) =>
                            setNewCampaign({ ...newCampaign, description: e.target.value })
                        }
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpen(false)} disabled={createCampaign.isPending}>
                        Cancel
                    </Button>
                    <Button
                        onClick={handleCreate}
                        variant="contained"
                        disabled={
                            !newCampaign.name ||
                            !newCampaign.systemId ||
                            createCampaign.isPending
                        }
                    >
                        {createCampaign.isPending ? (
                            <CircularProgress size={24} />
                        ) : (
                            'Create'
                        )}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
}
