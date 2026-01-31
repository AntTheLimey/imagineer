// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Layout from './Layout';

// Mock MUI icons to avoid rendering SVGs in tests
vi.mock('@mui/icons-material', () => ({
    Menu: () => <span data-testid="menu-icon">Menu</span>,
    Dashboard: () => <span data-testid="dashboard-icon">Dashboard</span>,
    Folder: () => <span data-testid="folder-icon">Folder</span>,
    Upload: () => <span data-testid="upload-icon">Upload</span>,
}));

const renderWithRouter = (ui: React.ReactElement, { route = '/' } = {}) => {
    return render(
        <MemoryRouter initialEntries={[route]}>
            {ui}
        </MemoryRouter>
    );
};

describe('Layout', () => {
    it('should render the app bar with title', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>
        );

        expect(screen.getByText('TTRPG Campaign Manager')).toBeInTheDocument();
    });

    it('should render the Imagineer branding in drawer', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>
        );

        // There are two drawers (mobile and desktop), so we expect 2 instances
        const brandingElements = screen.getAllByText('Imagineer');
        expect(brandingElements.length).toBeGreaterThan(0);
    });

    it('should render children content', () => {
        renderWithRouter(
            <Layout>
                <div data-testid="child-content">Test Content</div>
            </Layout>
        );

        expect(screen.getByTestId('child-content')).toBeInTheDocument();
        expect(screen.getByText('Test Content')).toBeInTheDocument();
    });

    it('should render all navigation items', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>
        );

        // Navigation items appear in both mobile and desktop drawers
        const dashboardItems = screen.getAllByText('Dashboard');
        const campaignsItems = screen.getAllByText('Campaigns');
        const importItems = screen.getAllByText('Import');

        expect(dashboardItems.length).toBeGreaterThan(0);
        expect(campaignsItems.length).toBeGreaterThan(0);
        expect(importItems.length).toBeGreaterThan(0);
    });

    it('should highlight the current navigation item based on route', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>,
            { route: '/campaigns' }
        );

        // The Campaigns item should be selected when on /campaigns route
        const campaignsButtons = screen.getAllByRole('button', { name: /campaigns/i });
        // At least one should have the selected class
        const hasSelected = campaignsButtons.some(
            (button) => button.classList.contains('Mui-selected')
        );
        expect(hasSelected).toBe(true);
    });

    it('should highlight Dashboard when on root route', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>,
            { route: '/' }
        );

        const dashboardButtons = screen.getAllByRole('button', { name: /dashboard/i });
        const hasSelected = dashboardButtons.some(
            (button) => button.classList.contains('Mui-selected')
        );
        expect(hasSelected).toBe(true);
    });

    it('should render navigation icons', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>
        );

        // Check that our mocked icons are rendered
        expect(screen.getAllByTestId('dashboard-icon').length).toBeGreaterThan(0);
        expect(screen.getAllByTestId('folder-icon').length).toBeGreaterThan(0);
        expect(screen.getAllByTestId('upload-icon').length).toBeGreaterThan(0);
    });

    it('should render the mobile menu toggle button', () => {
        renderWithRouter(
            <Layout>
                <div>Test Content</div>
            </Layout>
        );

        // The menu icon should be rendered (it's in the IconButton for mobile toggle)
        expect(screen.getByTestId('menu-icon')).toBeInTheDocument();
    });
});
