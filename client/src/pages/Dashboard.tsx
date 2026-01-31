// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { Box, Card, CardContent, Grid, Typography, Skeleton, Alert } from '@mui/material';
import {
    Folder as CampaignIcon,
    People as NPCIcon,
    Place as LocationIcon,
    Timeline as TimelineIcon,
} from '@mui/icons-material';
import { useDashboardStats } from '../hooks';

interface StatCardProps {
    label: string;
    value: number | undefined;
    icon: React.ReactNode;
    isLoading: boolean;
}

function StatCard({ label, value, icon, isLoading }: StatCardProps) {
    return (
        <Card>
            <CardContent sx={{ textAlign: 'center' }}>
                <Box sx={{ color: 'primary.main', mb: 1 }}>
                    {icon}
                </Box>
                {isLoading ? (
                    <Skeleton variant="text" width={60} sx={{ mx: 'auto' }}>
                        <Typography variant="h4">0</Typography>
                    </Skeleton>
                ) : (
                    <Typography variant="h4">{value ?? 0}</Typography>
                )}
                <Typography color="text.secondary">{label}</Typography>
            </CardContent>
        </Card>
    );
}

export default function Dashboard() {
    const { data: stats, isLoading, error } = useDashboardStats();

    const statCards = [
        {
            label: 'Campaigns',
            value: stats?.campaignCount,
            icon: <CampaignIcon fontSize="large" />,
        },
        {
            label: 'NPCs',
            value: stats?.npcCount,
            icon: <NPCIcon fontSize="large" />,
        },
        {
            label: 'Locations',
            value: stats?.locationCount,
            icon: <LocationIcon fontSize="large" />,
        },
        {
            label: 'Timeline Events',
            value: stats?.timelineEventCount,
            icon: <TimelineIcon fontSize="large" />,
        },
    ];

    return (
        <Box>
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Dashboard
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
                Welcome to Imagineer. Manage your TTRPG campaigns, track NPCs,
                locations, and timeline events.
            </Typography>

            {error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to load dashboard statistics. Please try again later.
                </Alert>
            )}

            <Grid container spacing={3} sx={{ mt: 2 }}>
                {statCards.map((stat) => (
                    <Grid item xs={12} sm={6} md={3} key={stat.label}>
                        <StatCard
                            label={stat.label}
                            value={stat.value}
                            icon={stat.icon}
                            isLoading={isLoading}
                        />
                    </Grid>
                ))}
            </Grid>

            <Box sx={{ mt: 4 }}>
                <Typography variant="h5" gutterBottom>
                    Getting Started
                </Typography>
                <Typography variant="body1" paragraph>
                    1. Create a new campaign or import existing content
                </Typography>
                <Typography variant="body1" paragraph>
                    2. Add NPCs, locations, items, and other entities
                </Typography>
                <Typography variant="body1" paragraph>
                    3. Define relationships between entities
                </Typography>
                <Typography variant="body1" paragraph>
                    4. Track your campaign timeline and session notes
                </Typography>
            </Box>
        </Box>
    );
}
