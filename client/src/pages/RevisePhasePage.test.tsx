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
 * Tests for the RevisePhasePage component.
 *
 * Mock strategy: mock the useWizardContext hook via the context module
 * and mock useResolveItem, useGenerateItemRevision, and useApplyItemRevision
 * so no real API calls are made.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
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
    useGenerateItemRevision: vi.fn(() => ({
        mutate: vi.fn(),
        isPending: false,
        data: null,
        error: null,
    })),
    useApplyItemRevision: vi.fn(() => ({
        mutate: vi.fn(),
        isPending: false,
        data: null,
        error: null,
    })),
}));

import { useWizardContext } from '../contexts/AnalysisWizardContext';
import {
    useResolveItem,
    useGenerateItemRevision,
    useApplyItemRevision,
} from '../hooks/useContentAnalysis';
import RevisePhasePage from './RevisePhasePage';

const mockUseWizardContext = vi.mocked(useWizardContext);
const mockUseResolveItem = vi.mocked(useResolveItem);
const mockUseGenerateItemRevision = vi.mocked(useGenerateItemRevision);
const mockUseApplyItemRevision = vi.mocked(useApplyItemRevision);

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const mockAnalysisItems: ContentAnalysisItem[] = [
    {
        id: 101,
        jobId: 1,
        detectionType: 'canon_contradiction',
        matchedText: 'Timeline conflict in Chapter 3',
        contextSnippet:
            'The events in Chapter 3 contradict the established timeline.',
        resolution: 'pending',
        phase: 'analysis',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 102,
        jobId: 1,
        detectionType: 'content_suggestion',
        matchedText: 'Add foreshadowing for the ritual',
        contextSnippet:
            'The ritual scene would benefit from earlier hints.',
        resolution: 'acknowledged',
        phase: 'analysis',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 103,
        jobId: 1,
        detectionType: 'mechanics_warning',
        matchedText: 'Sanity check DC too high',
        contextSnippet:
            'The DC 25 Sanity check is above normal thresholds.',
        resolution: 'pending',
        phase: 'analysis',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 104,
        jobId: 1,
        detectionType: 'pacing_note',
        matchedText: 'Scene drags in the middle',
        contextSnippet:
            'The investigation scene in the library runs too long.',
        resolution: 'pending',
        phase: 'analysis',
        createdAt: '2025-06-01T00:00:00Z',
    },
];

function makeWizardState(
    overrides: Partial<AnalysisWizardState> = {},
): AnalysisWizardState {
    return {
        job: { id: 1 } as AnalysisWizardState['job'],
        items: [],
        isLoading: false,
        error: null,
        phases: ['identification', 'analysis', 'enrichment'],
        currentPhase: 'analysis',
        currentPhaseIndex: 1,
        phaseItems: [],
        pendingCount: 0,
        canAdvance: true,
        canGoBack: true,
        nextPhaseLabel: 'enrich',
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
    route = '/campaigns/1/analysis/1/revise',
) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={[route]}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId/revise"
                        element={<RevisePhasePage />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('RevisePhasePage', () => {
    beforeEach(() => {
        queryClient.clear();
        mockUseResolveItem.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useResolveItem>);
        mockUseGenerateItemRevision.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: null,
            error: null,
        } as unknown as ReturnType<typeof useGenerateItemRevision>);
        mockUseApplyItemRevision.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: null,
            error: null,
        } as unknown as ReturnType<typeof useApplyItemRevision>);
    });

    // -- Grouping ----------------------------------------------------------

    it('renders analysis findings grouped by detection type', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockAnalysisItems,
                items: mockAnalysisItems,
                pendingCount: 3,
            }),
        );

        renderPage();

        // Group headers should be visible
        expect(
            screen.getByText('Canon Contradictions'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Content Suggestions'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Mechanics Warnings'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Pacing Notes'),
        ).toBeInTheDocument();

        // Matched text values should be present
        expect(
            screen.getByText('Timeline conflict in Chapter 3'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Add foreshadowing for the ritual'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Sanity check DC too high'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Scene drags in the middle'),
        ).toBeInTheDocument();
    });

    // -- Placeholder when nothing selected ---------------------------------

    it('shows "Select an item" placeholder when nothing selected', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockAnalysisItems,
                items: mockAnalysisItems,
                pendingCount: 3,
            }),
        );

        renderPage();

        expect(
            screen.getByText('Select an item to view details'),
        ).toBeInTheDocument();
    });

    // -- Revision workflow section -----------------------------------------

    it('shows revision workflow section with Revise Findings header', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockAnalysisItems,
                items: mockAnalysisItems,
                pendingCount: 3,
            }),
        );

        renderPage();

        expect(
            screen.getByText('Revise Findings'),
        ).toBeInTheDocument();
    });

    // -- Generate button disabled without acknowledged items ---------------

    it('does not show Generate Suggestion button when no acknowledged items exist', () => {
        const allPending = mockAnalysisItems.map((item) => ({
            ...item,
            resolution: 'pending' as const,
        }));
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: allPending,
                items: allPending,
                pendingCount: 4,
            }),
        );

        renderPage();

        // Per-item Generate Suggestion buttons only appear for acknowledged items
        expect(
            screen.queryByRole('button', {
                name: /generate suggestion/i,
            }),
        ).not.toBeInTheDocument();
    });

    it('shows Generate Suggestion button for acknowledged items', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockAnalysisItems,
                items: mockAnalysisItems,
                pendingCount: 3,
            }),
        );

        renderPage();

        // mockAnalysisItems[1] has resolution 'acknowledged', so one button should appear
        const generateBtns = screen.getAllByRole('button', {
            name: /generate suggestion/i,
        });
        expect(generateBtns.length).toBeGreaterThanOrEqual(1);
    });

    // -- Severity chips ----------------------------------------------------

    it('shows severity chips on groups', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockAnalysisItems,
                items: mockAnalysisItems,
                pendingCount: 3,
            }),
        );

        renderPage();

        // Each severity label should appear as a chip.
        // Some severities appear on multiple groups so use
        // getAllByText where needed.
        expect(screen.getByText('high')).toBeInTheDocument();
        expect(screen.getByText('medium')).toBeInTheDocument();
        expect(screen.getAllByText('low').length).toBeGreaterThanOrEqual(1);
    });
});
