// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Sessions - Placeholder page for session management.
 *
 * This page will eventually allow users to manage game sessions,
 * including session notes, attendance, and session summaries.
 */

import {
    Alert,
    Box,
    Paper,
    Typography,
} from '@mui/material';
import { EventNote as SessionsIcon } from '@mui/icons-material';

/**
 * Sessions management page placeholder.
 *
 * This feature is under development. The page will eventually provide:
 * - Session list and management
 * - Session notes and summaries
 * - Player attendance tracking
 * - Session scheduling
 *
 * @returns The Sessions page component
 */
export default function Sessions() {
    return (
        <Box>
            {/* Header */}
            <Typography variant="h4" gutterBottom sx={{ fontFamily: 'Cinzel' }}>
                Sessions
            </Typography>

            {/* Coming Soon Notice */}
            <Paper
                sx={{
                    p: 4,
                    textAlign: 'center',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: 2,
                }}
            >
                <SessionsIcon sx={{ fontSize: 64, color: 'text.secondary' }} />

                <Typography variant="h5" color="text.secondary">
                    Sessions - Coming Soon
                </Typography>

                <Typography variant="body1" color="text.secondary" sx={{ maxWidth: 600 }}>
                    The Sessions feature is currently under development. When complete,
                    you will be able to manage your game sessions, track attendance,
                    record session notes, and organize your campaign timeline by session.
                </Typography>

                <Alert severity="info" sx={{ mt: 2, maxWidth: 600 }}>
                    <Typography variant="body2" component="div">
                        <strong>Planned Features:</strong>
                        <ul style={{ margin: '8px 0 0 0', paddingLeft: 20 }}>
                            <li>Create and schedule game sessions</li>
                            <li>Record session notes and summaries</li>
                            <li>Track player attendance</li>
                            <li>Link timeline events to sessions</li>
                            <li>Generate session recaps using AI</li>
                        </ul>
                    </Typography>
                </Alert>
            </Paper>
        </Box>
    );
}
