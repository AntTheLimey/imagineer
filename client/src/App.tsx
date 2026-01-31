// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Box, CircularProgress } from '@mui/material';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Campaigns from './pages/Campaigns';
import Entities from './pages/Entities';
import Timeline from './pages/Timeline';
import Import from './pages/Import';
import Login from './pages/Login';
import AuthCallback from './pages/AuthCallback';
import { ReactNode } from 'react';

/**
 * Guards access to child routes based on authentication state.
 *
 * Renders a centered loading indicator while authentication is being determined, redirects to `/login` if the user is not authenticated, and renders `children` when authenticated.
 *
 * @param children - Elements to render when the user is authenticated
 * @returns `children` when authenticated, a `<Navigate>` to `/login` when not authenticated, or a loading indicator while authentication is pending
 */
function ProtectedRoute({ children }: { children: ReactNode }) {
    const { isAuthenticated, isLoading } = useAuth();

    if (isLoading) {
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

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }

    return <>{children}</>;
}

/**
 * Main application component with routing and authentication.
 */
function AppRoutes() {
    return (
        <Routes>
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/auth/callback" element={<AuthCallback />} />

            {/* Protected routes */}
            <Route
                path="/"
                element={
                    <ProtectedRoute>
                        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                            <Layout>
                                <Dashboard />
                            </Layout>
                        </Box>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/campaigns"
                element={
                    <ProtectedRoute>
                        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                            <Layout>
                                <Campaigns />
                            </Layout>
                        </Box>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/campaigns/:id/entities"
                element={
                    <ProtectedRoute>
                        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                            <Layout>
                                <Entities />
                            </Layout>
                        </Box>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/campaigns/:id/timeline"
                element={
                    <ProtectedRoute>
                        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                            <Layout>
                                <Timeline />
                            </Layout>
                        </Box>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/import"
                element={
                    <ProtectedRoute>
                        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
                            <Layout>
                                <Import />
                            </Layout>
                        </Box>
                    </ProtectedRoute>
                }
            />

            {/* Fallback - redirect to dashboard */}
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
}

/**
 * Application root that initializes routing and authentication context.
 *
 * @returns The root React element that wraps the app with a BrowserRouter and AuthProvider, rendering the application routes.
 */
function App() {
    return (
        <BrowserRouter>
            <AuthProvider>
                <AppRoutes />
            </AuthProvider>
        </BrowserRouter>
    );
}

export default App;