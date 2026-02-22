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
import EntityView from './pages/EntityView';
import ChapterEditorPage from './pages/ChapterEditorPage';
import SessionsManagement from './pages/SessionsManagement';
import CampaignImport from './pages/CampaignImport';
import Timeline from './pages/Timeline';
import Login from './pages/Login';
import AuthCallback from './pages/AuthCallback';
import AccountSettings from './pages/AccountSettings';
import AnalysisTriagePage from './pages/AnalysisTriagePage';
import SessionEditorPage from './pages/SessionEditorPage';
import CampaignSettingsPage from './pages/CampaignSettingsPage';

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
 * Render nested routes only for authenticated users, showing a loading state or redirecting otherwise.
 *
 * @returns The nested route UI: an `<Outlet />` when authenticated, a `<Navigate to="/login" />` redirect when not authenticated, or a centered `CircularProgress` while authentication is loading.
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
 * Wraps protected routes with the AppShell layout and provides campaign and draft contexts.
 *
 * Renders the AppShell containing an Outlet, with CampaignProvider and DraftProvider applied
 * so nested routes receive campaign and draft context values.
 *
 * @returns The React element tree composing CampaignProvider, DraftProvider, AppShell, and an Outlet for nested routes.
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
 * Wraps nested routes with Campaign and Draft providers for full-screen routes outside the AppShell.
 *
 * @returns The element that provides campaign and draft contexts and mounts an Outlet for nested routes.
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
 * Defines the application's route tree, including public, authenticated, AppShell-wrapped, and full-screen routes.
 *
 * Routes:
 * - Public: /login, /auth/callback
 * - Protected (requires authentication, rendered inside the AppShell): /, /campaigns (redirects to /), /campaigns/:id/overview, /campaigns/:id/entities, /campaigns/:id/sessions, /campaigns/:id/import, /campaigns/:id/timeline, /campaigns/:id/settings, and legacy /campaigns/:id/dashboard -> ../overview
 * - Full-screen (requires authentication, rendered outside the AppShell): /campaigns/new, /campaigns/:campaignId/entities/new, /campaigns/:campaignId/entities/:entityId/edit, /campaigns/:campaignId/chapters/new, /campaigns/:campaignId/chapters/:chapterId/edit
 * - Account settings (full-screen, requires authentication): /settings
 * - Fallback: any other path redirects to /
 *
 * @returns The React Routes element composing public routes, authentication guards, layout wrappers, page routes, and redirects
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
                    <Route path="/campaigns/:campaignId/entities/:entityId" element={<EntityView />} />
                    <Route path="/campaigns/:id/sessions" element={<SessionsManagement />} />
                    <Route path="/campaigns/:id/import" element={<CampaignImport />} />
                    <Route path="/campaigns/:id/timeline" element={<Timeline />} />
                    <Route path="/campaigns/:id/settings" element={<CampaignSettingsPage />} />

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
                    {/* New chapter editor */}
                    <Route
                        path="/campaigns/:campaignId/chapters/new"
                        element={<ChapterEditorPage />}
                    />
                    {/* Edit existing chapter */}
                    <Route
                        path="/campaigns/:campaignId/chapters/:chapterId/edit"
                        element={<ChapterEditorPage />}
                    />
                    {/* New session editor */}
                    <Route
                        path="/campaigns/:campaignId/sessions/new"
                        element={<SessionEditorPage />}
                    />
                    {/* Edit existing session */}
                    <Route
                        path="/campaigns/:campaignId/sessions/:sessionId/edit"
                        element={<SessionEditorPage />}
                    />
                    {/* Analysis triage */}
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId"
                        element={<AnalysisTriagePage />}
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