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
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AnalysisTriagePage from './AnalysisTriagePage';

vi.mock('../hooks/useContentAnalysis', () => ({
    useAnalysisJob: vi.fn(),
    useAnalysisItems: vi.fn(),
    useResolveItem: vi.fn(),
}));

import {
    useAnalysisJob,
    useAnalysisItems,
    useResolveItem,
} from '../hooks/useContentAnalysis';

const mockUseAnalysisJob = vi.mocked(useAnalysisJob);
const mockUseAnalysisItems = vi.mocked(useAnalysisItems);
const mockUseResolveItem = vi.mocked(useResolveItem);

const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
});

const renderPage = () => {
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/campaigns/1/analysis/1']}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId"
                        element={<AnalysisTriagePage />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const mockJob = {
    id: 1,
    campaignId: 1,
    sourceTable: 'chapters',
    sourceId: 5,
    sourceField: 'overview',
    status: 'completed',
    totalItems: 4,
    resolvedItems: 0,
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
};

const mockItems = [
    {
        id: 1,
        jobId: 1,
        detectionType: 'wiki_link_resolved' as const,
        matchedText: 'Professor Armitage',
        entityId: 10,
        entityName: 'Prof. Armitage',
        entityType: 'npc' as const,
        contextSnippet: 'They went to see Professor Armitage at the university.',
        resolution: 'pending' as const,
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 2,
        jobId: 1,
        detectionType: 'wiki_link_unresolved' as const,
        matchedText: 'The Silver Key',
        contextSnippet: 'He carried The Silver Key in his pack.',
        resolution: 'pending' as const,
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 3,
        jobId: 1,
        detectionType: 'untagged_mention' as const,
        matchedText: 'Arkham',
        entityId: 20,
        entityName: 'Arkham',
        entityType: 'location' as const,
        similarity: 1.0,
        contextSnippet: 'The investigators returned to Arkham that evening.',
        resolution: 'pending' as const,
        createdAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 4,
        jobId: 1,
        detectionType: 'misspelling' as const,
        matchedText: 'Cthuhlu',
        entityId: 30,
        entityName: 'Cthulhu',
        entityType: 'creature' as const,
        similarity: 0.85,
        contextSnippet: 'The cultists chanted the name of Cthuhlu.',
        resolution: 'pending' as const,
        createdAt: '2025-01-01T00:00:00Z',
    },
];

describe('AnalysisTriagePage', () => {
    beforeEach(() => {
        mockUseResolveItem.mockReturnValue({
            mutate: vi.fn(),
            isPending: false,
        } as unknown as ReturnType<typeof useResolveItem>);
    });

    it('renders loading state', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: undefined,
            isLoading: true,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: undefined,
            isLoading: true,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('renders with mock job data', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: mockItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(screen.getByText('Content Analysis')).toBeInTheDocument();
    });

    it('groups items by detection type', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: mockItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(
            screen.getByText(/Wiki Links \(Resolved\)/)
        ).toBeInTheDocument();
        expect(
            screen.getByText(/Wiki Links \(Unresolved\)/)
        ).toBeInTheDocument();
        expect(screen.getByText(/Untagged Mentions/)).toBeInTheDocument();
        expect(screen.getByText(/Misspellings/)).toBeInTheDocument();
    });

    it('renders matched text for items', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: mockItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(screen.getByText('Professor Armitage')).toBeInTheDocument();
        expect(screen.getByText('The Silver Key')).toBeInTheDocument();
        expect(screen.getAllByText('Arkham').length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText('Cthuhlu')).toBeInTheDocument();
    });

    it('"Skip for now" button is present', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: mockItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(
            screen.getByRole('button', { name: /skip for now/i })
        ).toBeInTheDocument();
    });

    it('"Done" button is present with pending items', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);
        mockUseAnalysisItems.mockReturnValue({
            data: mockItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(
            screen.getByRole('button', { name: /^Done$/i })
        ).toBeInTheDocument();
    });

    it('"Done" button shows "All resolved" when every item is resolved', () => {
        mockUseAnalysisJob.mockReturnValue({
            data: mockJob,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisJob>);

        const resolvedItems = mockItems.map((item) => ({
            ...item,
            resolution: 'accepted' as const,
        }));
        mockUseAnalysisItems.mockReturnValue({
            data: resolvedItems,
            isLoading: false,
        } as unknown as ReturnType<typeof useAnalysisItems>);

        renderPage();

        expect(
            screen.getByRole('button', { name: /all resolved/i })
        ).toBeInTheDocument();
    });
});
