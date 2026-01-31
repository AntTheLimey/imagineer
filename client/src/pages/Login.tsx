// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import {
    Box,
    Button,
    Card,
    CardContent,
    CircularProgress,
    Typography,
} from '@mui/material';
import { Google as GoogleIcon } from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

/**
 * Render the login page and initiate Google OAuth sign-in.
 *
 * While authentication state is resolving, displays a full-screen loading indicator. If the user is already authenticated, redirects to the root path ("/"). Otherwise renders a branded sign-in card with a button that navigates the browser to the backend OAuth endpoint to start Google sign-in.
 *
 * @returns The React element for the login page.
 */
export default function Login() {
    const { isAuthenticated, isLoading: authLoading } = useAuth();
    const [isRedirecting, setIsRedirecting] = useState(false);

    // If already authenticated, redirect to dashboard
    if (authLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '100vh',
                    bgcolor: 'background.default',
                }}
            >
                <CircularProgress />
            </Box>
        );
    }

    if (isAuthenticated) {
        return <Navigate to="/" replace />;
    }

    const handleGoogleLogin = () => {
        setIsRedirecting(true);
        // Redirect to the backend OAuth endpoint
        window.location.href = '/api/auth/google';
    };

    return (
        <Box
            sx={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '100vh',
                bgcolor: 'background.default',
                p: 2,
            }}
        >
            <Card
                sx={{
                    maxWidth: 400,
                    width: '100%',
                    textAlign: 'center',
                }}
            >
                <CardContent sx={{ p: 4 }}>
                    <Typography
                        variant="h3"
                        component="h1"
                        sx={{
                            fontFamily: 'Cinzel',
                            mb: 1,
                            color: 'primary.main',
                        }}
                    >
                        Imagineer
                    </Typography>
                    <Typography
                        variant="subtitle1"
                        color="text.secondary"
                        sx={{ mb: 4 }}
                    >
                        TTRPG Campaign Intelligence Platform
                    </Typography>

                    <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mb: 3 }}
                    >
                        Sign in to manage your campaigns, track entities,
                        and bring your tabletop adventures to life.
                    </Typography>

                    <Button
                        variant="contained"
                        size="large"
                        startIcon={
                            isRedirecting ? (
                                <CircularProgress size={20} color="inherit" />
                            ) : (
                                <GoogleIcon />
                            )
                        }
                        onClick={handleGoogleLogin}
                        disabled={isRedirecting}
                        sx={{
                            py: 1.5,
                            px: 4,
                            textTransform: 'none',
                            fontSize: '1rem',
                        }}
                    >
                        {isRedirecting ? 'Redirecting...' : 'Sign in with Google'}
                    </Button>
                </CardContent>
            </Card>
        </Box>
    );
}