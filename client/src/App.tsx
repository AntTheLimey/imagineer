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

// Pages
import HomePage from './pages/HomePage';
import CampaignOverview from './pages/CampaignOverview';
import CreateCampaign from './pages/CreateCampaign';
import Entities from './pages/Entities';
import EntityEditor from './pages/EntityEditor';
import Sessions from './pages/Sessions';
import CampaignImport from './pages/CampaignImport';
import Timeline from './pages/Timeline';
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
 * @returns The protected nested routes via `<Outlet />` when authenticated,
 *          a centered loading indicator while authentication is in progress,
 *          or a redirect to `/login` when not authenticated.
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
 * Layout wrapper for routes that use the AppShell with persistent navigation.
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
 * Provides Campaign and Draft contexts to full-screen editor routes
 * without applying the AppShell.
 *
 * Renders an Outlet so nested routes receive the necessary providers
 * while remaining outside the AppShell layout.
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
 * - Protected routes with AppShell:
 *   - / - Home (redirects to campaign overview or shows campaign selection)
 *   - /campaigns - Redirects to home
 *   - /campaigns/:id/overview - Campaign overview
 *   - /campaigns/:id/entities - Entities list
 *   - /campaigns/:id/sessions - Sessions (placeholder)
 *   - /campaigns/:id/import - Import content
 *   - /campaigns/:id/timeline - Timeline
 * - Full-screen routes:
 *   - /campaigns/new - Create new campaign
 *   - /campaigns/:campaignId/entities/new - New entity editor
 *   - /campaigns/:campaignId/entities/:entityId/edit - Edit entity
 *   - /settings - Account settings
 */
function AppRoutes() {
    return (
        <Routes>
            {/* Public routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/auth/callback" element={<AuthCallback />} />

            {/* Protected routes with AppShell layout */}
            <Route element={<ProtectedOutlet />}>
                <Route element={<AppShellWrapper />}>
                    {/* Home - handles smart redirect to current/latest campaign */}
                    <Route path="/" element={<HomePage />} />

                    {/* Campaigns list - redirect to home */}
                    <Route path="/campaigns" element={<Navigate to="/" replace />} />

                    {/* Campaign-specific routes */}
                    <Route path="/campaigns/:id/overview" element={<CampaignOverview />} />
                    <Route path="/campaigns/:id/entities" element={<Entities />} />
                    <Route path="/campaigns/:id/sessions" element={<Sessions />} />
                    <Route path="/campaigns/:id/import" element={<CampaignImport />} />
                    <Route path="/campaigns/:id/timeline" element={<Timeline />} />

                    {/* Legacy route redirect - dashboard to overview */}
                    <Route
                        path="/campaigns/:id/dashboard"
                        element={<Navigate to="../overview" replace />}
                    />
                </Route>

                {/* Full-screen editor routes (outside AppShell) */}
                <Route element={<FullScreenWrapper />}>
                    {/* Create new campaign */}
                    <Route
                        path="/campaigns/new"
                        element={<CreateCampaign />}
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

            {/* Account Settings - Full-screen layout without sidebar */}
            <Route
                path="/settings"
                element={
                    <ProtectedRoute>
                        <AccountSettings />
                    </ProtectedRoute>
                }
            />

            {/* Fallback - redirect to home */}
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
