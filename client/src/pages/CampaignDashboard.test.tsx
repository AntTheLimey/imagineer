// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Tests for CampaignDashboard page component.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CampaignDashboard from './CampaignDashboard';
import { CampaignProvider } from '../contexts/CampaignContext';
import { DraftProvider } from '../contexts/DraftContext';

// Mock the hooks
vi.mock('../hooks', async () => {
    const actual = await vi.importActual('../hooks');
    return {
        ...actual,
        useCampaign: vi.fn(),
        useUpdateCampaign: vi.fn(() => ({
            mutateAsync: vi.fn(),
            isPending: false,
            error: null,
        })),
        useGameSystems: vi.fn(() => ({
            data: [
                { id: 1, name: 'Call of Cthulhu 7th Edition' },
                { id: 2, name: 'GURPS 4th Edition' },
            ],
            isLoading: false,
            error: null,
        })),
    };
});

// Import the mocked hook
import { useCampaign } from '../hooks';

describe('CampaignDashboard', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        vi.clearAllMocks();
        queryClient = new QueryClient({
            defaultOptions: {
                queries: {
                    retry: false,
                },
            },
        });
    });

    afterEach(() => {
        cleanup();
        queryClient.clear();
    });

    /**
     * Create a wrapper with all required providers.
     * Uses the shared queryClient from beforeEach so it can be
     * properly cleaned up in afterEach.
     */
    function createWrapper() {
        return function Wrapper({ children }: { children: React.ReactNode }) {
            return (
                <QueryClientProvider client={queryClient}>
                    <MemoryRouter initialEntries={['/campaigns/1/dashboard']}>
                        <CampaignProvider>
                            <DraftProvider>
                                <Routes>
                                    <Route
                                        path="/campaigns/:id/dashboard"
                                        element={children}
                                    />
                                </Routes>
                            </DraftProvider>
                        </CampaignProvider>
                    </MemoryRouter>
                </QueryClientProvider>
            );
        };
    }

    it('renders loading state when campaign is loading', () => {
        vi.mocked(useCampaign).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as ReturnType<typeof useCampaign>);

        render(<CampaignDashboard />, { wrapper: createWrapper() });

        expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('renders error state when campaign fails to load', () => {
        vi.mocked(useCampaign).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('Failed to load'),
        } as ReturnType<typeof useCampaign>);

        render(<CampaignDashboard />, { wrapper: createWrapper() });

        expect(screen.getByText(/Failed to load campaign/)).toBeInTheDocument();
    });

    it('renders campaign settings when campaign is loaded', () => {
        vi.mocked(useCampaign).mockReturnValue({
            data: {
                id: 1,
                name: 'Test Campaign',
                systemId: 1,
                description: 'A test campaign',
                settings: {},
                createdAt: '2025-01-01T00:00:00Z',
                updatedAt: '2025-01-01T00:00:00Z',
            },
            isLoading: false,
            error: null,
        } as ReturnType<typeof useCampaign>);

        render(<CampaignDashboard />, { wrapper: createWrapper() });

        // Campaign name appears in title and breadcrumbs
        expect(screen.getAllByText('Test Campaign').length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText('Campaign Settings')).toBeInTheDocument();
    });

    it('renders navigation items', () => {
        vi.mocked(useCampaign).mockReturnValue({
            data: {
                id: 1,
                name: 'Test Campaign',
                systemId: 1,
                settings: {},
                createdAt: '2025-01-01T00:00:00Z',
                updatedAt: '2025-01-01T00:00:00Z',
            },
            isLoading: false,
            error: null,
        } as ReturnType<typeof useCampaign>);

        render(<CampaignDashboard />, { wrapper: createWrapper() });

        // Check that navigation items are rendered (on desktop view)
        // Note: Navigation is hidden on mobile, so we check for at least one instance
        expect(screen.getAllByText('Overview').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Manage Entities').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Manage Sessions').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Player Characters').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Import Campaign Notes').length).toBeGreaterThanOrEqual(1);
        expect(screen.getAllByText('Import Knowledge').length).toBeGreaterThanOrEqual(1);
    });

    it('renders breadcrumbs with campaign name', () => {
        vi.mocked(useCampaign).mockReturnValue({
            data: {
                id: 1,
                name: 'My Awesome Campaign',
                systemId: 1,
                settings: {},
                createdAt: '2025-01-01T00:00:00Z',
                updatedAt: '2025-01-01T00:00:00Z',
            },
            isLoading: false,
            error: null,
        } as ReturnType<typeof useCampaign>);

        render(<CampaignDashboard />, { wrapper: createWrapper() });

        // Check breadcrumbs
        expect(screen.getByText('Home')).toBeInTheDocument();
        // The campaign name appears both in breadcrumbs and title
        expect(screen.getAllByText('My Awesome Campaign').length).toBeGreaterThanOrEqual(1);
    });
});
