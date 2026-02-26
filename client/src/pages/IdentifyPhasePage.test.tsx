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
    useBatchResolve: vi.fn(),
}));

import { useWizardContext } from '../contexts/AnalysisWizardContext';
import { useResolveItem, useBatchResolve } from '../hooks/useContentAnalysis';
import IdentifyPhasePage from './IdentifyPhasePage';

const mockUseWizardContext = vi.mocked(useWizardContext);
const mockUseResolveItem = vi.mocked(useResolveItem);
const mockUseBatchResolve = vi.mocked(useBatchResolve);

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
        mockUseBatchResolve.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useBatchResolve>);
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

    // -- Accept All button visibility --------------------------------------

    it('shows Accept All for groups where all pending items have entities', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                job: { id: 1 } as AnalysisWizardState['job'],
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        // There should be 4 Accept All buttons (resolved wiki links,
        // untagged mentions, potential aliases, misspellings -- all
        // have entityId set). The unresolved wiki-link group should
        // NOT have Accept All since its item has no entityId.
        const acceptAllButtons = screen.getAllByRole('button', {
            name: /accept all/i,
        });
        expect(acceptAllButtons).toHaveLength(4);
    });

    it('hides Accept All for groups with items missing an entity', () => {
        // Only include the unresolved wiki-link item (no entityId)
        const unresolvedOnly: ContentAnalysisItem[] = [
            mockItems[1], // wiki_link_unresolved, no entityId
        ];
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                job: { id: 1 } as AnalysisWizardState['job'],
                phaseItems: unresolvedOnly,
                pendingCount: 1,
            }),
        );

        renderPage();

        expect(
            screen.queryByRole('button', { name: /accept all/i }),
        ).not.toBeInTheDocument();
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

    // -- Done section (resolved items) ------------------------------------

    it('shows resolved items in Done section, not in pending groups', () => {
        const mixedItems: ContentAnalysisItem[] = [
            { ...mockItems[0], resolution: 'pending' },
            {
                ...mockItems[2],
                id: 3,
                resolution: 'accepted',
            },
        ];

        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mixedItems,
                pendingCount: 1,
            }),
        );

        renderPage();

        // The Done header should be present with count
        expect(screen.getByText('Done')).toBeInTheDocument();

        // The resolved item (Arkham) should appear in the Done
        // section, not under the Untagged Mentions group header
        expect(
            screen.queryByText(/Untagged Mentions/),
        ).not.toBeInTheDocument();

        // Pending item still appears in its group
        expect(
            screen.getByText(/Wiki Links \(Resolved\)/),
        ).toBeInTheDocument();
        expect(
            screen.getByText('Professor Armitage'),
        ).toBeInTheDocument();
    });

    it('hides Done section when no items are resolved', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mockItems,
                pendingCount: 5,
            }),
        );

        renderPage();

        expect(screen.queryByText('Done')).not.toBeInTheDocument();
    });

    it('allows clicking a resolved item in Done to view details', () => {
        const mixedItems: ContentAnalysisItem[] = [
            { ...mockItems[0], resolution: 'pending' },
            {
                ...mockItems[2],
                id: 3,
                resolution: 'accepted',
            },
        ];

        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: mixedItems,
                pendingCount: 1,
            }),
        );

        renderPage();

        // Click the resolved item in the Done section
        fireEvent.click(screen.getAllByText('Arkham')[0]);

        // Detail panel should show the resolution status alert
        expect(
            screen.getByText(/This item has been resolved/),
        ).toBeInTheDocument();
    });

    it('toggles the Done section expand/collapse icon on click', () => {
        const resolvedOnly: ContentAnalysisItem[] = [
            {
                ...mockItems[0],
                resolution: 'accepted',
            },
        ];

        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: resolvedOnly,
                pendingCount: 0,
            }),
        );

        renderPage();

        // Done header is present and item is visible
        expect(screen.getByText('Done')).toBeInTheDocument();
        expect(
            screen.getByText('Professor Armitage'),
        ).toBeInTheDocument();

        // Initially expanded: ExpandLess icon is shown
        const doneHeader = screen.getByText('Done').closest('div')!;
        expect(
            doneHeader.querySelector('[data-testid="ExpandLessIcon"]'),
        ).toBeInTheDocument();

        // Click to collapse
        fireEvent.click(screen.getByText('Done'));

        // Now ExpandMore icon is shown instead
        expect(
            doneHeader.querySelector('[data-testid="ExpandMoreIcon"]'),
        ).toBeInTheDocument();
        expect(
            doneHeader.querySelector('[data-testid="ExpandLessIcon"]'),
        ).not.toBeInTheDocument();

        // Click again to re-expand
        fireEvent.click(screen.getByText('Done'));
        expect(
            doneHeader.querySelector('[data-testid="ExpandLessIcon"]'),
        ).toBeInTheDocument();
    });

    it('shows empty state when no pending or resolved items exist', () => {
        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: [],
                pendingCount: 0,
            }),
        );

        renderPage();

        expect(
            screen.getByText('No items in this phase.'),
        ).toBeInTheDocument();
    });

    it('does not show action buttons on resolved items in Done', () => {
        const resolvedOnly: ContentAnalysisItem[] = [
            {
                ...mockItems[0],
                resolution: 'accepted',
            },
        ];

        mockUseWizardContext.mockReturnValue(
            makeWizardState({
                phaseItems: resolvedOnly,
                pendingCount: 0,
            }),
        );

        renderPage();

        // The Accept/Dismiss icon buttons should not be present
        // for resolved items
        expect(
            screen.queryByRole('button', { name: /accept$/i }),
        ).not.toBeInTheDocument();
        expect(
            screen.queryByRole('button', { name: /dismiss/i }),
        ).not.toBeInTheDocument();
    });
});
