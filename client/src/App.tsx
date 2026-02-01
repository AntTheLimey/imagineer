// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { Box, CircularProgress } from '@mui/material';
import { ReactNode } from 'react';

// Contexts
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { CampaignProvider } from './contexts/CampaignContext';
import { DraftProvider } from './contexts/DraftContext';

// Layouts
import { AppShell } from './layouts';
import Layout from './components/Layout';

// Pages
import Dashboard from './pages/Dashboard';
import Campaigns from './pages/Campaigns';
import CampaignDashboard from './pages/CampaignDashboard';
import Entities from './pages/Entities';
import EntityEditor from './pages/EntityEditor';
import Timeline from './pages/Timeline';
import Import from './pages/Import';
import Login from './pages/Login';
import AuthCallback from './pages/AuthCallback';
import AccountSettings from './pages/AccountSettings';

/**
 * Guards access to child routes based on authentication state.
 *
 * Renders a centered loading indicator while authentication is being
 * determined, redirects to `/login` if the user is not authenticated,
 * and renders `children` when authenticated.
 *
 * @param children - Elements to render when the user is authenticated
 * @returns `children` when authenticated, a `<Navigate>` to `/login` when
 *          not authenticated, or a loading indicator while auth is pending
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
 * Guards nested routes using authentication state and renders the appropriate UI.
 *
 * @returns The protected nested routes via `<Outlet />` when authenticated, a centered loading indicator while authentication is in progress, or a redirect to `/login` when not authenticated.
 */
function ProtectedOutlet() {
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

    return <Outlet />;
}

/**
 * Wraps content with the legacy sidebar Layout to preserve backward-compatible pages.
 *
 * @param children - React nodes to render inside the legacy Layout
 */
function LegacyLayoutWrapper({ children }: { children: ReactNode }) {
    return (
        <Box sx={{ display: 'flex', minHeight: '100vh' }}>
            <Layout>{children}</Layout>
        </Box>
    );
}

/**
 * Layout wrapper for routes that use the new AppShell.
 * Wraps content with campaign and draft providers.
 */
function AppShellWrapper() {
    return (
        <CampaignProvider>
            <DraftProvider>
                <AppShell>
                    <Outlet />
                </AppShell>
            </DraftProvider>
        </CampaignProvider>
    );
}

/**
 * Provides Campaign and Draft contexts to full-screen editor routes without applying the AppShell.
 *
 * Renders an Outlet so nested routes receive the necessary providers while remaining outside the AppShell layout.
 */
function FullScreenWrapper() {
    return (
        <CampaignProvider>
            <DraftProvider>
                <Outlet />
            </DraftProvider>
        </CampaignProvider>
    );
}

/**
 * Main application component with routing and authentication.
 *
 * Route structure:
 * - Public routes: /login, /auth/callback
 * - Protected routes with legacy Layout: /, /campaigns, /import
 * - Protected routes with AppShell: /campaigns/:id/entities, /campaigns/:id/timeline
 * - Full-screen routes: /campaigns/:id/dashboard, /campaigns/:id/entities/new,
 *   /campaigns/:id/entities/:entityId/edit
 */
function AppRoutes() {
    return (
        <Routes>
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/auth/callback" element={<AuthCallback />} />

            {/* Protected routes with legacy Layout (backward compatibility) */}
            <Route
                path="/"
                element={
                    <ProtectedRoute>
                        <LegacyLayoutWrapper>
                            <Dashboard />
                        </LegacyLayoutWrapper>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/campaigns"
                element={
                    <ProtectedRoute>
                        <LegacyLayoutWrapper>
                            <Campaigns />
                        </LegacyLayoutWrapper>
                    </ProtectedRoute>
                }
            />
            <Route
                path="/import"
                element={
                    <ProtectedRoute>
                        <LegacyLayoutWrapper>
                            <Import />
                        </LegacyLayoutWrapper>
                    </ProtectedRoute>
                }
            />

            {/* Account Settings - Full-screen layout without sidebar */}
            <Route
                path="/settings"
                element={
                    <ProtectedRoute>
                        <AccountSettings />
                    </ProtectedRoute>
                }
            />

            {/* Protected routes with new AppShell layout */}
            <Route element={<ProtectedOutlet />}>
                <Route element={<AppShellWrapper />}>
                    {/* Entity list view */}
                    <Route
                        path="/campaigns/:id/entities"
                        element={<Entities />}
                    />
                    {/* Timeline view */}
                    <Route
                        path="/campaigns/:id/timeline"
                        element={<Timeline />}
                    />
                </Route>

                {/* Full-screen editor routes (outside AppShell) */}
                <Route element={<FullScreenWrapper />}>
                    {/* Campaign dashboard - main hub for campaign management */}
                    <Route
                        path="/campaigns/:id/dashboard"
                        element={<CampaignDashboard />}
                    />
                    {/* New entity editor */}
                    <Route
                        path="/campaigns/:campaignId/entities/new"
                        element={<EntityEditor />}
                    />
                    {/* Edit existing entity */}
                    <Route
                        path="/campaigns/:campaignId/entities/:entityId/edit"
                        element={<EntityEditor />}
                    />
                </Route>
            </Route>

            {/* Fallback - redirect to dashboard */}
            <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
    );
}

/**
 * Application root that initializes routing and authentication context.
 *
 * @returns The root React element that wraps the app with a BrowserRouter
 *          and AuthProvider, rendering the application routes.
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