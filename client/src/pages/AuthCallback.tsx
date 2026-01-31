// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, CircularProgress, Typography, Alert } from '@mui/material';
import { useAuth, User } from '../contexts/AuthContext';

/**
 * AuthCallback handles the OAuth callback from the server.
 * It extracts the token and user info from URL parameters,
 * stores them in the auth context, and redirects to the dashboard.
 */
export default function AuthCallback() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const { login } = useAuth();
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const token = searchParams.get('token');
        const userParam = searchParams.get('user');
        const errorParam = searchParams.get('error');

        // Handle error from OAuth flow
        if (errorParam) {
            setError(decodeURIComponent(errorParam));
            return;
        }

        // Validate required parameters
        if (!token) {
            setError('Missing authentication token');
            return;
        }

        if (!userParam) {
            setError('Missing user information');
            return;
        }

        // Parse user data
        let user: User;
        try {
            user = JSON.parse(decodeURIComponent(userParam));
        } catch {
            setError('Invalid user data');
            return;
        }

        // Validate user object has required fields
        if (!user.id || !user.email || !user.name) {
            setError('Incomplete user information');
            return;
        }

        // Store auth data and redirect to dashboard
        login(token, user);
        navigate('/', { replace: true });
    }, [searchParams, login, navigate]);

    if (error) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '100vh',
                    bgcolor: 'background.default',
                    p: 2,
                }}
            >
                <Alert severity="error" sx={{ mb: 2, maxWidth: 400 }}>
                    {error}
                </Alert>
                <Typography variant="body2" color="text.secondary">
                    <a href="/login" style={{ color: 'inherit' }}>
                        Return to login
                    </a>
                </Typography>
            </Box>
        );
    }

    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '100vh',
                bgcolor: 'background.default',
            }}
        >
            <CircularProgress sx={{ mb: 2 }} />
            <Typography variant="body1" color="text.secondary">
                Completing sign in...
            </Typography>
        </Box>
    );
}
