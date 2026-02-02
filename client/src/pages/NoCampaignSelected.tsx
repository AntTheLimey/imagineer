// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * NoCampaignSelected - Displayed when no campaign is selected.
 *
 * Shows different content based on whether the user has campaigns:
 * - If no campaigns exist: Welcome/onboarding screen explaining what campaigns are
 * - If campaigns exist: Campaign selection cards to choose from
 */

import { useNavigate } from 'react-router-dom';
import {
    Box,
    Button,
    Card,
    CardActions,
    CardContent,
    Grid,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Paper,
    Skeleton,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    AutoStories as CampaignIcon,
    People as NPCIcon,
    Timeline as TimelineIcon,
    FileUpload as ImportIcon,
} from '@mui/icons-material';
import { useCampaigns, useGameSystems } from '../hooks';
import { useCampaignContext } from '../contexts/CampaignContext';
import type { Campaign } from '../types';

/**
 * NoCampaignSelected page shown when no campaign is currently selected.
 *
 * Displays existing campaigns as cards for quick selection and provides
 * the option to create a new campaign.
 *
 * @returns The NoCampaignSelected page component
 */
export default function NoCampaignSelected() {
    const navigate = useNavigate();
    const { setCurrentCampaignId } = useCampaignContext();

    // Fetch campaigns and game systems
    const {
        data: campaigns,
        isLoading: campaignsLoading,
        error: campaignsError,
    } = useCampaigns();

    const { data: gameSystems } = useGameSystems();

    /**
     * Handle campaign card click - select and navigate to overview.
     */
    const handleCampaignClick = (campaign: Campaign) => {
        setCurrentCampaignId(campaign.id);
        navigate(`/campaigns/${campaign.id}/overview`);
    };

    /**
     * Navigate to create campaign page.
     */
    const handleCreateCampaign = () => {
        navigate('/campaigns/new');
    };

    // Loading state
    if (campaignsLoading) {
        return (
            <Box>
                <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                    Welcome to Imagineer
                </Typography>
                <Grid container spacing={3} sx={{ mt: 2 }}>
                    {[1, 2, 3].map((i) => (
                        <Grid item xs={12} sm={6} md={4} key={i}>
                            <Skeleton variant="rectangular" height={180} />
                        </Grid>
                    ))}
                </Grid>
            </Box>
        );
    }

    // Error state
    if (campaignsError) {
        return (
            <Box>
                <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                    Welcome to Imagineer
                </Typography>
                <Paper sx={{ p: 3, textAlign: 'center' }}>
                    <Typography color="error">
                        Failed to load campaigns. Please try again later.
                    </Typography>
                </Paper>
            </Box>
        );
    }

    const hasCampaigns = campaigns && campaigns.length > 0;

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Welcome to Imagineer
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
                Select a campaign to get started, or create a new one.
            </Typography>

            {/* Create Campaign Button */}
            <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={handleCreateCampaign}
                sx={{ mb: 3 }}
            >
                Create New Campaign
            </Button>

            {/* Campaign Cards or Welcome Screen */}
            {hasCampaigns ? (
                <Grid container spacing={3}>
                    {campaigns.map((campaign) => {
                        const gameSystem = gameSystems?.find(
                            (gs) => gs.id === campaign.systemId
                        );
                        return (
                            <Grid item xs={12} sm={6} md={4} key={campaign.id}>
                                <Card
                                    sx={{
                                        cursor: 'pointer',
                                        height: '100%',
                                        display: 'flex',
                                        flexDirection: 'column',
                                        '&:hover': {
                                            boxShadow: 6,
                                        },
                                    }}
                                    onClick={() => handleCampaignClick(campaign)}
                                >
                                    <CardContent sx={{ flexGrow: 1 }}>
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
                                        <Button size="small">Open Campaign</Button>
                                    </CardActions>
                                </Card>
                            </Grid>
                        );
                    })}
                </Grid>
            ) : (
                /* Welcome/Onboarding screen when no campaigns exist */
                <Paper
                    sx={{
                        p: { xs: 3, sm: 4, md: 6 },
                        textAlign: 'center',
                        maxWidth: 800,
                        mx: 'auto',
                    }}
                >
                    {/* Campaign icon */}
                    <Box
                        sx={{
                            display: 'flex',
                            justifyContent: 'center',
                            mb: 3,
                        }}
                    >
                        <CampaignIcon
                            sx={{
                                fontSize: 80,
                                color: 'primary.main',
                                opacity: 0.8,
                            }}
                        />
                    </Box>

                    <Typography
                        variant="h5"
                        gutterBottom
                        sx={{ fontWeight: 500 }}
                    >
                        Start Your Adventure
                    </Typography>

                    <Typography
                        variant="body1"
                        color="text.secondary"
                        sx={{ mb: 3, maxWidth: 600, mx: 'auto' }}
                    >
                        A campaign is your central hub for managing a tabletop RPG
                        adventure. Everything you need to run an unforgettable game,
                        organized in one place.
                    </Typography>

                    {/* Feature list */}
                    <Box sx={{ textAlign: 'left', maxWidth: 500, mx: 'auto', mb: 4 }}>
                        <List dense>
                            <ListItem>
                                <ListItemIcon>
                                    <NPCIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Track NPCs, locations, items, and other entities"
                                    secondary="Build a rich world with interconnected characters and places"
                                />
                            </ListItem>
                            <ListItem>
                                <ListItemIcon>
                                    <TimelineIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Build timelines and manage game sessions"
                                    secondary="Keep track of what happened and when"
                                />
                            </ListItem>
                            <ListItem>
                                <ListItemIcon>
                                    <ImportIcon color="primary" />
                                </ListItemIcon>
                                <ListItemText
                                    primary="Import your existing notes"
                                    secondary="Bring in content from Evernote or other sources"
                                />
                            </ListItem>
                        </List>
                    </Box>

                    {/* Create button */}
                    <Button
                        variant="contained"
                        size="large"
                        startIcon={<AddIcon />}
                        onClick={handleCreateCampaign}
                        sx={{ px: 4, py: 1.5 }}
                    >
                        Create Your First Campaign
                    </Button>
                </Paper>
            )}
        </Box>
    );
}
