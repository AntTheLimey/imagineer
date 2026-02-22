/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * Tests for the IdentifyPhasePage component.
 *
 * Mock strategy: mock the useWizardContext hook via the context module
 * and mock useResolveItem so no real API calls are made.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { AnalysisWizardState } from '../hooks/useAnalysisWizard';
import type { ContentAnalysisItem } from '../api/contentAnalysis';

// ---------------------------------------------------------------------------
// Module mocks
// ---------------------------------------------------------------------------

vi.mock('../contexts/AnalysisWizardContext', () => ({
    useWizardContext: vi.fn(),
}));

vi.mock('../hooks/useContentAnalysis', () => ({
    useResolveItem: vi.fn(),
}));

import { useWizardContext } from '../contexts/AnalysisWizardContext';
import { useResolveItem } from '../hooks/useContentAnalysis';
import IdentifyPhasePage from './IdentifyPhasePage';

const mockUseWizardContext = vi.mocked(useWizardContext);
const mockUseResolveItem = vi.mocked(useResolveItem);

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const mockItems: ContentAnalysisItem[] = [
    {
        id: 1,
        jobId: 1,
        detectionType: 'wiki_link_resolved',
        matchedText: 'Professor Armitage',
        entityId: 10,
        entityName: 'Prof. Armitage',
        entityType: 'npc',
        contextSnippet:
            'They went to see Professor Armitage at the university.',
        resolution: 'pending',
        phase: 'identification',
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 2,
        jobId: 1,
        detectionType: 'wiki_link_unresolved',
        matchedText: 'The Silver Key',
        contextSnippet: 'He carried The Silver Key in his pack.',
        resolution: 'pending',
        phase: 'identification',
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 3,
        jobId: 1,
        detectionType: 'untagged_mention',
        matchedText: 'Arkham',
        entityId: 20,
        entityName: 'Arkham',
        entityType: 'location',
        similarity: 1.0,
        contextSnippet:
            'The investigators returned to Arkham that evening.',
        resolution: 'pending',
        phase: 'identification',
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 5,
        jobId: 1,
        detectionType: 'potential_alias',
        matchedText: 'Dr. Armitage',
        entityId: 10,
        entityName: 'Prof. Armitage',
        entityType: 'npc',
        similarity: 0.9,
        contextSnippet:
            'Dr. Armitage greeted them at the door.',
        resolution: 'pending',
        phase: 'identification',
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 4,
        jobId: 1,
        detectionType: 'misspelling',
        matchedText: 'Cthuhlu',
        entityId: 30,
        entityName: 'Cthulhu',
        entityType: 'creature',
        similarity: 0.85,
        contextSnippet: 'The cultists chanted the name of Cthuhlu.',
        resolution: 'pending',
        phase: 'identification',
        createdAt: '2025-01-01T00:00:00Z',
    },
];

function makeWizardState(
    overrides: Partial<AnalysisWizardState> = {},
): AnalysisWizardState {
    return {
        job: undefined,
        items: [],
        isLoading: false,
        error: null,
        phases: ['identification', 'analysis', 'enrichment'],
        currentPhase: 'identification',
        currentPhaseIndex: 0,
        phaseItems: [],
        pendingCount: 0,
        canAdvance: false,
        canGoBack: false,
        nextPhaseLabel: null,
        goToPhase: vi.fn(),
        goToNextPhase: vi.fn(),
        ...overrides,
    };
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
});

const renderPage = (
    route = '/campaigns/1/analysis/1/identify',
) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={[route]}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId/identify"
                        element={<IdentifyPhasePage />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('IdentifyPhasePage', () => {
    beforeEach(() => {
        mockUseResolveItem.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useResolveItem>);
    });

    // -- Grouping ----------------------------------------------------------

    it('groups items by detection type', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        // Group headers
        expect(
            screen.getByText(/Wiki Links \(Resolved\)/),
        ).toBeInTheDocument();
        expect(
            screen.getByText(/Wiki Links \(Unresolved\)/),
        ).toBeInTheDocument();
        expect(
            screen.getByText(/Untagged Mentions/),
        ).toBeInTheDocument();
        expect(
            screen.getByText(/Potential Aliases/),
        ).toBeInTheDocument();
        expect(screen.getByText(/Misspellings/)).toBeInTheDocument();

        // Matched text values
        expect(
            screen.getByText('Professor Armitage'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('The Silver Key'),
        ).toBeInTheDocument();
        expect(
            screen.getAllByText('Arkham').length,
        ).toBeGreaterThanOrEqual(1);
        expect(
            screen.getByText('Dr. Armitage'),
        ).toBeInTheDocument();
        expect(screen.getByText('Cthuhlu')).toBeInTheDocument();
    });

    // -- Detail panel on selection -----------------------------------------

    it('shows detail panel when item selected', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        // Click on the first item
        fireEvent.click(screen.getByText('Professor Armitage'));

        // Detail panel shows entity name (also in left-panel chip)
        expect(
            screen.getAllByText('Prof. Armitage').length,
        ).toBeGreaterThanOrEqual(2);
        // Context snippet appears in both the left panel secondary
        // text and the detail panel highlight section
        expect(
            screen.getAllByText(/They went to see/).length,
        ).toBeGreaterThanOrEqual(2);
    });

    // -- Placeholder when nothing selected ---------------------------------

    it('shows placeholder when no item selected', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        expect(
            screen.getByText('Select an item to view details'),
        ).toBeInTheDocument();
    });

    // -- Create entity form for unresolved wiki-links ----------------------

    it('shows create entity form for unresolved wiki-links', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        // Select the unresolved wiki-link item
        fireEvent.click(screen.getByText('The Silver Key'));

        // Click New Entity button
        fireEvent.click(
            screen.getByRole('button', { name: /new entity/i }),
        );

        // Form should be visible with pre-filled name
        expect(
            screen.getByDisplayValue('The Silver Key'),
        ).toBeInTheDocument();

        // Entity type selector should exist
        expect(
            screen.getByRole('button', { name: /create & link/i }),
        ).toBeInTheDocument();
    });
});
