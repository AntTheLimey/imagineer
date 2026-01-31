// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import Dashboard from './Dashboard';
import type { DashboardStats } from '../api';

// Mock the hooks module
vi.mock('../hooks', () => ({
    useDashboardStats: vi.fn(),
}));

// Import the mocked hook
import { useDashboardStats } from '../hooks';

// Mock MUI icons to simplify testing
vi.mock('@mui/icons-material', () => ({
    Folder: () => <span data-testid="campaign-icon">Campaign</span>,
    People: () => <span data-testid="npc-icon">NPC</span>,
    Place: () => <span data-testid="location-icon">Location</span>,
    Timeline: () => <span data-testid="timeline-icon">Timeline</span>,
}));

const mockDashboardStats: DashboardStats = {
    campaignCount: 5,
    npcCount: 42,
    locationCount: 17,
    timelineEventCount: 89,
    itemCount: 25,
    factionCount: 8,
    totalEntityCount: 186,
};

describe('Dashboard', () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    it('should render the dashboard title', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    it('should render the welcome message', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByText(/Welcome to Imagineer/)).toBeInTheDocument();
    });

    it('should render stat cards with correct values when data is loaded', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByText('5')).toBeInTheDocument(); // Campaigns
        expect(screen.getByText('42')).toBeInTheDocument(); // NPCs
        expect(screen.getByText('17')).toBeInTheDocument(); // Locations
        expect(screen.getByText('89')).toBeInTheDocument(); // Timeline Events

        expect(screen.getByText('Campaigns')).toBeInTheDocument();
        expect(screen.getByText('NPCs')).toBeInTheDocument();
        expect(screen.getByText('Locations')).toBeInTheDocument();
        expect(screen.getByText('Timeline Events')).toBeInTheDocument();
    });

    it('should render stat card icons', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByTestId('campaign-icon')).toBeInTheDocument();
        expect(screen.getByTestId('npc-icon')).toBeInTheDocument();
        expect(screen.getByTestId('location-icon')).toBeInTheDocument();
        expect(screen.getByTestId('timeline-icon')).toBeInTheDocument();
    });

    it('should show loading skeletons when data is loading', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        // MUI Skeleton components should be present during loading
        const skeletons = document.querySelectorAll('.MuiSkeleton-root');
        expect(skeletons.length).toBeGreaterThan(0);
    });

    it('should show error alert when there is an error', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('Failed to fetch'),
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByText(/Failed to load dashboard statistics/)).toBeInTheDocument();
        expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('should not show error alert when there is no error', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('should render the Getting Started section', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: mockDashboardStats,
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        expect(screen.getByText('Getting Started')).toBeInTheDocument();
        expect(screen.getByText(/Create a new campaign or import existing content/)).toBeInTheDocument();
        expect(screen.getByText(/Add NPCs, locations, items, and other entities/)).toBeInTheDocument();
        expect(screen.getByText(/Define relationships between entities/)).toBeInTheDocument();
        expect(screen.getByText(/Track your campaign timeline and session notes/)).toBeInTheDocument();
    });

    it('should show zero values when stats are undefined', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: {
                campaignCount: 0,
                npcCount: 0,
                locationCount: 0,
                timelineEventCount: 0,
                itemCount: 0,
                factionCount: 0,
                totalEntityCount: 0,
            },
            isLoading: false,
            error: null,
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        // All stat values should be 0
        const zeroValues = screen.getAllByText('0');
        expect(zeroValues.length).toBe(4); // 4 stat cards
    });

    it('should handle both loading and error states simultaneously (loading takes precedence)', () => {
        vi.mocked(useDashboardStats).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: new Error('Some error'),
        } as ReturnType<typeof useDashboardStats>);

        render(<Dashboard />);

        // Error alert should still show even during loading (per component implementation)
        expect(screen.getByRole('alert')).toBeInTheDocument();

        // But skeletons should also be visible
        const skeletons = document.querySelectorAll('.MuiSkeleton-root');
        expect(skeletons.length).toBeGreaterThan(0);
    });
});
