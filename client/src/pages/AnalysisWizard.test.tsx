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
 * Smoke tests for the AnalysisWizard shell component.
 *
 * Mock strategy: mock the useAnalysisWizard hook so the component
 * never makes real API calls. Each test sets up the wizard state it
 * needs and verifies the correct structural elements render.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { AnalysisWizardState } from '../hooks/useAnalysisWizard';

// ---------------------------------------------------------------------------
// Module mocks
// ---------------------------------------------------------------------------

vi.mock('../hooks/useAnalysisWizard', () => ({
    useAnalysisWizard: vi.fn(),
    PHASE_ROUTES: {
        identification: 'identify',
        analysis: 'revise',
        enrichment: 'enrich',
    },
}));

import { useAnalysisWizard } from '../hooks/useAnalysisWizard';
import AnalysisWizard from './AnalysisWizard';

const mockUseAnalysisWizard = vi.mocked(useAnalysisWizard);

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

const renderWizard = (route = '/campaigns/1/analysis/1/identify') => {
    const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
    });

    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={[route]}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/analysis/:jobId/*"
                        element={<AnalysisWizard />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('AnalysisWizard', () => {
    beforeEach(() => {
        mockUseAnalysisWizard.mockReturnValue(makeWizardState());
    });

    // -- Loading state -----------------------------------------------------

    it('renders loading spinner when isLoading is true', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({ isLoading: true }),
        );

        renderWizard();

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    // -- NaN guard ---------------------------------------------------------

    it('renders error alert when campaignId is not a number', () => {
        renderWizard('/campaigns/abc/analysis/1/identify');

        expect(screen.getByRole('alert')).toBeInTheDocument();
        expect(
            screen.getByText('Invalid campaign or job ID.'),
        ).toBeInTheDocument();
    });

    it('renders error alert when jobId is not a number', () => {
        renderWizard('/campaigns/1/analysis/xyz/identify');

        expect(screen.getByRole('alert')).toBeInTheDocument();
        expect(
            screen.getByText('Invalid campaign or job ID.'),
        ).toBeInTheDocument();
    });

    // -- Stepper -----------------------------------------------------------

    it('renders a stepper with three phase labels', () => {
        renderWizard();

        expect(screen.getByText('Identify')).toBeInTheDocument();
        expect(screen.getByText('Revise')).toBeInTheDocument();
        expect(screen.getByText('Enrich')).toBeInTheDocument();
    });

    // -- Header with source info ------------------------------------------

    it('renders header with source info when job is loaded', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                job: {
                    id: 1,
                    campaignId: 1,
                    sourceTable: 'chapters',
                    sourceId: 5,
                    sourceField: 'overview',
                    status: 'completed',
                    totalItems: 4,
                    resolvedItems: 0,
                    enrichmentTotal: 0,
                    enrichmentResolved: 0,
                    phases: [
                        'identification',
                        'analysis',
                        'enrichment',
                    ],
                    currentPhase: 'identification',
                    createdAt: '2025-01-01T00:00:00Z',
                    updatedAt: '2025-01-01T00:00:00Z',
                },
            }),
        );

        renderWizard();

        expect(screen.getByText('Content Analysis')).toBeInTheDocument();
        expect(
            screen.getByText(/chapters.*overview/i),
        ).toBeInTheDocument();
    });

    // -- Error banner: quota ----------------------------------------------

    it('renders quota error message', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                job: {
                    id: 1,
                    campaignId: 1,
                    sourceTable: 'chapters',
                    sourceId: 5,
                    sourceField: 'overview',
                    status: 'failed',
                    totalItems: 0,
                    resolvedItems: 0,
                    enrichmentTotal: 0,
                    enrichmentResolved: 0,
                    phases: ['identification'],
                    currentPhase: 'identification',
                    failureReason: 'API quota exceeded',
                    createdAt: '2025-01-01T00:00:00Z',
                    updatedAt: '2025-01-01T00:00:00Z',
                },
            }),
        );

        renderWizard();

        expect(
            screen.getByText(
                'Your API token limit has been reached. Please check your account and try again later.',
            ),
        ).toBeInTheDocument();
    });

    // -- Error banner: rate limit -----------------------------------------

    it('renders rate limit error message', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                job: {
                    id: 1,
                    campaignId: 1,
                    sourceTable: 'chapters',
                    sourceId: 5,
                    sourceField: 'overview',
                    status: 'failed',
                    totalItems: 0,
                    resolvedItems: 0,
                    enrichmentTotal: 0,
                    enrichmentResolved: 0,
                    phases: ['identification'],
                    currentPhase: 'identification',
                    failureReason: 'Rate limited by provider',
                    createdAt: '2025-01-01T00:00:00Z',
                    updatedAt: '2025-01-01T00:00:00Z',
                },
            }),
        );

        renderWizard();

        expect(
            screen.getByText(
                'The AI service is temporarily unavailable. Please try again in a few minutes.',
            ),
        ).toBeInTheDocument();
    });

    // -- Error banner: generic failure ------------------------------------

    it('renders generic error message for other failures', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                job: {
                    id: 1,
                    campaignId: 1,
                    sourceTable: 'chapters',
                    sourceId: 5,
                    sourceField: 'overview',
                    status: 'failed',
                    totalItems: 0,
                    resolvedItems: 0,
                    enrichmentTotal: 0,
                    enrichmentResolved: 0,
                    phases: ['identification'],
                    currentPhase: 'identification',
                    failureReason: 'Unexpected server error',
                    createdAt: '2025-01-01T00:00:00Z',
                    updatedAt: '2025-01-01T00:00:00Z',
                },
            }),
        );

        renderWizard();

        expect(
            screen.getByText(
                'Enrichment encountered an error: Unexpected server error',
            ),
        ).toBeInTheDocument();
    });

    // -- Navigation buttons -----------------------------------------------

    it('renders "Back to Overview" button', () => {
        renderWizard();

        expect(
            screen.getByRole('button', { name: /back to overview/i }),
        ).toBeInTheDocument();
    });

    it('renders "Continue" button when canAdvance is true', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                canAdvance: true,
                currentPhaseIndex: 0,
            }),
        );

        renderWizard();

        expect(
            screen.getByRole('button', { name: /continue to revise/i }),
        ).toBeInTheDocument();
    });

    it('calls goToNextPhase when Continue button is clicked', () => {
        const goToNextPhase = vi.fn();
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                canAdvance: true,
                currentPhaseIndex: 0,
                goToNextPhase,
            }),
        );

        renderWizard();

        fireEvent.click(
            screen.getByRole('button', { name: /continue to revise/i }),
        );
        expect(goToNextPhase).toHaveBeenCalledTimes(1);
    });

    it('does not render "Continue" button when canAdvance is false', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                canAdvance: false,
                currentPhaseIndex: 2,
            }),
        );

        renderWizard();

        expect(
            screen.queryByRole('button', { name: /continue/i }),
        ).not.toBeInTheDocument();
    });

    // -- No error banner when status is not failed ------------------------

    it('does not render error banner when job status is not failed', () => {
        mockUseAnalysisWizard.mockReturnValue(
            makeWizardState({
                job: {
                    id: 1,
                    campaignId: 1,
                    sourceTable: 'chapters',
                    sourceId: 5,
                    sourceField: 'overview',
                    status: 'completed',
                    totalItems: 4,
                    resolvedItems: 0,
                    enrichmentTotal: 0,
                    enrichmentResolved: 0,
                    phases: [
                        'identification',
                        'analysis',
                        'enrichment',
                    ],
                    currentPhase: 'identification',
                    createdAt: '2025-01-01T00:00:00Z',
                    updatedAt: '2025-01-01T00:00:00Z',
                },
            }),
        );

        renderWizard();

        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
});
