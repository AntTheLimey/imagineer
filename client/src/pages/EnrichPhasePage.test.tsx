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
 * Tests for the EnrichPhasePage component.
 *
 * Mock strategy: mock the useWizardContext hook via the context module
 * and mock useResolveItem, useTriggerEnrichment, useEnrichmentStream,
 * and useCancelEnrichment so no real API calls are made.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { AnalysisWizardState } from '../hooks/useAnalysisWizard';
import type { ContentAnalysisItem, ContentAnalysisJob } from '../api/contentAnalysis';

// ---------------------------------------------------------------------------
// Module mocks
// ---------------------------------------------------------------------------

vi.mock('../contexts/AnalysisWizardContext', () => ({
    useWizardContext: vi.fn(),
}));

vi.mock('../hooks/useContentAnalysis', () => ({
    useResolveItem: vi.fn(),
    useTriggerEnrichment: vi.fn(),
    useEnrichmentStream: vi.fn(),
    useCancelEnrichment: vi.fn(),
}));

import { useWizardContext } from '../contexts/AnalysisWizardContext';
import {
    useResolveItem,
    useTriggerEnrichment,
    useEnrichmentStream,
    useCancelEnrichment,
} from '../hooks/useContentAnalysis';
import EnrichPhasePage from './EnrichPhasePage';

const mockUseWizardContext = vi.mocked(useWizardContext);
const mockUseResolveItem = vi.mocked(useResolveItem);
const mockUseTriggerEnrichment = vi.mocked(useTriggerEnrichment);
const mockUseEnrichmentStream = vi.mocked(useEnrichmentStream);
const mockUseCancelEnrichment = vi.mocked(useCancelEnrichment);

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const mockEnrichmentItems: ContentAnalysisItem[] = [
    {
        id: 201,
        jobId: 1,
        detectionType: 'description_update',
        matchedText: 'Professor Armitage',
        entityId: 10,
        entityName: 'Prof. Armitage',
        entityType: 'npc',
        contextSnippet:
            'Professor Armitage greeted them at the library entrance.',
        suggestedContent: {
            currentDescription: 'A professor at Miskatonic University.',
            suggestedDescription:
                'A professor of medieval metaphysics at Miskatonic University, known for his role in the Dunwich Horror.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 202,
        jobId: 1,
        detectionType: 'description_update',
        matchedText: 'Arkham',
        entityId: 20,
        entityName: 'Arkham',
        entityType: 'location',
        contextSnippet: 'The investigators returned to Arkham.',
        suggestedContent: {
            currentDescription: 'A city in Massachusetts.',
            suggestedDescription:
                'A city in Massachusetts, home to Miskatonic University and numerous eldritch occurrences.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 203,
        jobId: 1,
        detectionType: 'log_entry',
        matchedText: 'Session 3 Log',
        contextSnippet: 'The party explored the ruins.',
        suggestedContent: {
            content: 'The investigators explored the ruins beneath Arkham.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 204,
        jobId: 1,
        detectionType: 'relationship_suggestion',
        matchedText: 'Armitage knows Wilbur',
        entityId: 10,
        entityName: 'Prof. Armitage',
        entityType: 'npc',
        contextSnippet: 'Armitage recalled his encounter with Wilbur.',
        suggestedContent: {
            relationshipType: 'knows',
            targetEntity: 'Wilbur Whateley',
            description: 'Armitage encountered Wilbur at the university.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 205,
        jobId: 1,
        detectionType: 'graph_warning',
        matchedText: 'Orphaned faction node',
        suggestedContent: {
            description: 'The Esoteric Order of Dagon has no connections.',
            recommendation: 'Add relationships to connect this faction.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
    {
        id: 206,
        jobId: 1,
        detectionType: 'redundant_edge',
        matchedText: 'Duplicate ally relationship',
        suggestedContent: {
            description:
                'Armitage-Wilbur has both "knows" and "ally" edges.',
            recommendation: 'Remove one of the redundant relationships.',
        },
        resolution: 'pending',
        phase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
    },
];

function makeJob(
    overrides: Partial<ContentAnalysisJob> = {},
): ContentAnalysisJob {
    return {
        id: 1,
        campaignId: 1,
        sourceTable: 'chapters',
        sourceId: 1,
        sourceField: 'content',
        status: 'identified',
        totalItems: 5,
        resolvedItems: 0,
        enrichmentTotal: 0,
        enrichmentResolved: 0,
        phases: ['identification', 'analysis', 'enrichment'],
        currentPhase: 'enrichment',
        createdAt: '2025-06-01T00:00:00Z',
        updatedAt: '2025-06-01T00:00:00Z',
        ...overrides,
    };
}

function makeWizardState(
    overrides: Partial<AnalysisWizardState> = {},
): AnalysisWizardState {
    return {
        job: makeJob(),
        items: [],
        isLoading: false,
        error: null,
        phases: ['identification', 'analysis', 'enrichment'],
        currentPhase: 'enrichment',
        currentPhaseIndex: 2,
        phaseItems: [],
        pendingCount: 0,
        canAdvance: false,
        canGoBack: true,
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
    route = '/campaigns/1/analysis/1/enrich',
) => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={[route]}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId/enrich"
                        element={<EnrichPhasePage />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('EnrichPhasePage', () => {
    beforeEach(() => {
        queryClient.clear();
        mockUseResolveItem.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useResolveItem>);
        mockUseTriggerEnrichment.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
            data: undefined,
        } as unknown as ReturnType<typeof useTriggerEnrichment>);
        mockUseEnrichmentStream.mockReturnValue(
            undefined as unknown as ReturnType<typeof useEnrichmentStream>,
        );
        mockUseCancelEnrichment.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useCancelEnrichment>);
    });

    // -- Grouping ----------------------------------------------------------

    it('renders enrichment items grouped by detection type', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockEnrichmentItems,
                pendingCount: 6,
            }),
        );

        renderPage();

        // Group headers should be visible
        expect(
            screen.getByText('Description Updates'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Log Entries'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Relationship Suggestions'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Graph Warnings'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Redundant Edges'),
        ).toBeInTheDocument();

        // Matched text values should be present. Non-graph items
        // appear once; graph items appear twice (in the grouped list
        // and in the Graph Health summary section).
        expect(
            screen.getByText('Professor Armitage'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Session 3 Log'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Armitage knows Wilbur'),
        ).toBeInTheDocument();
        expect(
            screen.getAllByText('Orphaned faction node'),
        ).toHaveLength(2);
        expect(
            screen.getAllByText('Duplicate ally relationship'),
        ).toHaveLength(2);
    });

    // -- Placeholder when nothing selected ---------------------------------

    it('shows "Select an item" placeholder when nothing selected', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockEnrichmentItems,
                pendingCount: 6,
            }),
        );

        renderPage();

        expect(
            screen.getByText('Select an item to view details'),
        ).toBeInTheDocument();
    });

    // -- Start Enrichment button -------------------------------------------

    it('shows "Start Enrichment" button', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: [],
                pendingCount: 0,
            }),
        );

        renderPage();

        expect(
            screen.getByRole('button', {
                name: /start enrichment/i,
            }),
        ).toBeInTheDocument();
    });

    // -- Progress indicator during enrichment ------------------------------

    it('shows progress indicator when enrichment is in progress', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                job: makeJob({ status: 'enriching' }),
                phaseItems: [],
                pendingCount: 0,
            }),
        );

        renderPage();

        expect(
            screen.getByText('Enrichment in progress...'),
        ).toBeInTheDocument();
        // Cancel button should be visible during enrichment
        expect(
            screen.getByRole('button', {
                name: /cancel enrichment/i,
            }),
        ).toBeInTheDocument();
        // Start button should NOT be visible during enrichment
        expect(
            screen.queryByRole('button', {
                name: /start enrichment/i,
            }),
        ).not.toBeInTheDocument();
    });

    // -- Item counts per group ---------------------------------------------

    it('shows item counts per group', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockEnrichmentItems,
                pendingCount: 6,
            }),
        );

        renderPage();

        // Description Updates has 2 items; Graph Health summary also
        // has 2 graph items. Log Entries, Relationship Suggestions,
        // Graph Warnings, and Redundant Edges each have 1.
        const twoChips = screen.getAllByText('2');
        expect(twoChips.length).toBeGreaterThanOrEqual(2);
        const oneChips = screen.getAllByText('1');
        expect(oneChips.length).toBeGreaterThanOrEqual(4);
    });

    // -- Graph Health section ----------------------------------------------

    it('renders the Graph Health section when graph items exist', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockEnrichmentItems,
                pendingCount: 6,
            }),
        );

        renderPage();

        // The "Graph Health" header should appear.
        expect(
            screen.getByText('Graph Health'),
        ).toBeInTheDocument();

        // Graph items should show their detection type labels as
        // secondary text in the Graph Health section.
        expect(
            screen.getByText('Graph Warning'),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Redundant Edge'),
        ).toBeInTheDocument();
    });

    it('does not render Graph Health section when no graph items exist', () => {
        const nonGraphItems = mockEnrichmentItems.filter(
            (item) =>
                !['graph_warning', 'redundant_edge', 'invalid_type_pair', 'orphan_warning'].includes(
                    item.detectionType,
                ),
        );
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: nonGraphItems,
                pendingCount: nonGraphItems.length,
            }),
        );

        renderPage();

        expect(
            screen.queryByText('Graph Health'),
        ).not.toBeInTheDocument();
    });

    it('shows Acknowledge and Dismiss buttons on graph health items', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockEnrichmentItems,
                pendingCount: 6,
            }),
        );

        renderPage();

        // There should be Acknowledge tooltips for graph items.
        // Each graph item in the health section has its own pair.
        const acknowledgeButtons = screen.getAllByRole('button', {
            name: /acknowledge/i,
        });
        expect(acknowledgeButtons.length).toBe(2);

        // Dismiss buttons exist in both the main grouped list and
        // the graph health section. We just verify some exist.
        const dismissButtons = screen.getAllByRole('button', {
            name: /dismiss/i,
        });
        expect(dismissButtons.length).toBeGreaterThanOrEqual(2);
    });
});
