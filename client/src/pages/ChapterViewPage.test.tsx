// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Tests for the ChapterViewPage component.
 *
 * Verifies rendering of chapter data in read-only mode, including the
 * overview with markdown, entity groupings, sessions, relationships,
 * metadata, and interactive elements (edit/delete buttons). Also tests
 * inline edit mode with PhaseStrip, form fields, and cancel behaviour.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ChapterViewPage from './ChapterViewPage';

// Mock hooks
vi.mock('../hooks', () => ({
    useChapter: vi.fn(),
    useDeleteChapter: vi.fn(),
    useUpdateChapter: vi.fn(),
    useUserSettings: vi.fn(),
    useEntities: vi.fn(),
    useSessionsByChapter: vi.fn(),
    useChapterEntities: vi.fn(),
    useChapterRelationships: vi.fn(),
    useCreateChapterEntity: vi.fn(),
    useDeleteChapterEntity: vi.fn(),
}));

vi.mock('../components/MarkdownRenderer', () => ({
    MarkdownRenderer: ({ content }: { content: string }) => (
        <div data-testid="markdown">{content}</div>
    ),
}));

vi.mock('../components/Sessions/SessionStageIndicator', () => ({
    default: ({ stage }: { stage: string }) => (
        <span data-testid="stage-indicator">{stage}</span>
    ),
}));

vi.mock('../components/PhaseStrip', () => ({
    PhaseStrip: ({ onSave }: { onSave: (p: Record<string, boolean>) => void }) => (
        <button
            data-testid="phase-strip"
            onClick={() => onSave({ identify: false, revise: false, enrich: false })}
        >
            Save &amp; Go
        </button>
    ),
}));

vi.mock('../components/MarkdownEditor', () => ({
    MarkdownEditor: ({ value, onChange }: { value: string; onChange: (v: string) => void }) => (
        <textarea
            data-testid="markdown-editor"
            value={value}
            onChange={(e) => onChange(e.target.value)}
        />
    ),
}));

vi.mock('../components/EntityAutocomplete', () => ({
    default: () => <div data-testid="entity-autocomplete" />,
}));

import {
    useChapter,
    useDeleteChapter,
    useUpdateChapter,
    useUserSettings,
    useEntities,
    useSessionsByChapter,
    useChapterEntities,
    useChapterRelationships,
    useCreateChapterEntity,
    useDeleteChapterEntity,
} from '../hooks';
import type { Mock } from 'vitest';

/**
 * Creates a fresh QueryClient configured for testing.
 */
function createTestQueryClient() {
    return new QueryClient({
        defaultOptions: {
            queries: { retry: false },
        },
    });
}

/**
 * Renders ChapterViewPage inside the necessary providers and router
 * context for testing.
 */
function renderPage() {
    const queryClient = createTestQueryClient();
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/campaigns/1/chapters/10']}>
                <Routes>
                    <Route
                        path="/campaigns/:campaignId/chapters/:chapterId"
                        element={<ChapterViewPage />}
                    />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );
}

/**
 * Sets default return values for all mocked hooks.
 */
function setupDefaultMocks() {
    (useChapter as Mock).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
    });
    (useDeleteChapter as Mock).mockReturnValue({
        mutateAsync: vi.fn(),
    });
    (useUpdateChapter as Mock).mockReturnValue({
        mutateAsync: vi.fn().mockResolvedValue({}),
        isPending: false,
    });
    (useUserSettings as Mock).mockReturnValue({ data: null });
    (useEntities as Mock).mockReturnValue({ data: undefined });
    (useSessionsByChapter as Mock).mockReturnValue({ data: undefined });
    (useChapterEntities as Mock).mockReturnValue({ data: undefined });
    (useChapterRelationships as Mock).mockReturnValue({ data: undefined });
    (useCreateChapterEntity as Mock).mockReturnValue({ mutate: vi.fn() });
    (useDeleteChapterEntity as Mock).mockReturnValue({ mutate: vi.fn() });
}

/**
 * Chapter data fixture used across multiple tests.
 */
const CHAPTER_FIXTURE = {
    id: 10,
    campaignId: 1,
    title: 'The Dark Beginning',
    overview: 'A storm gathers over Arkham.',
    sortOrder: 1,
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-02-01T12:00:00Z',
};

describe('ChapterViewPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        setupDefaultMocks();
    });

    it('shows loading spinner when data is loading', () => {
        (useChapter as Mock).mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        });

        renderPage();

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('shows error alert when chapter fails to load', () => {
        (useChapter as Mock).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('Network error'),
        });

        renderPage();

        expect(
            screen.getByText('Failed to load chapter. Please try again later.')
        ).toBeInTheDocument();
    });

    it('shows not found alert when chapter data is absent', () => {
        (useChapter as Mock).mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        });

        renderPage();

        expect(screen.getByText('Chapter not found.')).toBeInTheDocument();
    });

    it('renders chapter title and overview', () => {
        (useChapter as Mock).mockReturnValue({
            data: CHAPTER_FIXTURE,
            isLoading: false,
            error: null,
        });

        renderPage();

        expect(screen.getByText('The Dark Beginning')).toBeInTheDocument();
        expect(screen.getByTestId('markdown')).toHaveTextContent(
            'A storm gathers over Arkham.'
        );
    });

    it('shows "No overview yet." when overview is empty', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Empty Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        expect(screen.getByText('No overview yet.')).toBeInTheDocument();
    });

    it('renders entities grouped by mention type', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Entity Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        (useChapterEntities as Mock).mockReturnValue({
            data: [
                {
                    id: 1,
                    chapterId: 10,
                    entityId: 100,
                    mentionType: 'featured',
                    createdAt: '2026-01-15T10:00:00Z',
                    entity: {
                        id: 100,
                        name: 'Professor Armitage',
                        entityType: 'npc',
                    },
                },
                {
                    id: 2,
                    chapterId: 10,
                    entityId: 101,
                    mentionType: 'linked',
                    createdAt: '2026-01-15T10:00:00Z',
                    entity: {
                        id: 101,
                        name: 'Miskatonic University',
                        entityType: 'location',
                    },
                },
                {
                    id: 3,
                    chapterId: 10,
                    entityId: 102,
                    mentionType: 'mentioned',
                    createdAt: '2026-01-15T10:00:00Z',
                    entity: {
                        id: 102,
                        name: 'Necronomicon',
                        entityType: 'item',
                    },
                },
            ],
        });

        renderPage();

        expect(screen.getByText('Entities')).toBeInTheDocument();
        // "Featured" appears as both a group heading and a mention chip
        expect(screen.getAllByText('Featured')).toHaveLength(2);
        expect(screen.getByText('Professor Armitage')).toBeInTheDocument();
        expect(screen.getByText('NPC')).toBeInTheDocument();
        expect(screen.getByText('Miskatonic University')).toBeInTheDocument();
        expect(screen.getByText('Necronomicon')).toBeInTheDocument();
    });

    it('renders sessions with stage indicators', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Session Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        (useSessionsByChapter as Mock).mockReturnValue({
            data: [
                {
                    id: 50,
                    campaignId: 1,
                    chapterId: 10,
                    title: 'Into the Library',
                    sessionNumber: 3,
                    plannedDate: '2026-02-20T18:00:00Z',
                    status: 'PLANNED',
                    stage: 'prep',
                    createdAt: '2026-01-15T10:00:00Z',
                    updatedAt: '2026-02-01T12:00:00Z',
                },
                {
                    id: 51,
                    campaignId: 1,
                    chapterId: 10,
                    title: 'The Ritual',
                    sessionNumber: 4,
                    status: 'COMPLETED',
                    stage: 'completed',
                    createdAt: '2026-01-15T10:00:00Z',
                    updatedAt: '2026-02-01T12:00:00Z',
                },
            ],
        });

        renderPage();

        expect(screen.getByText('Sessions (2)')).toBeInTheDocument();
        expect(screen.getByText('Into the Library')).toBeInTheDocument();
        expect(screen.getByText('The Ritual')).toBeInTheDocument();
        const indicators = screen.getAllByTestId('stage-indicator');
        expect(indicators).toHaveLength(2);
        expect(indicators[0]).toHaveTextContent('prep');
        expect(indicators[1]).toHaveTextContent('completed');
    });

    it('renders relationships section', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Relationship Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        (useChapterEntities as Mock).mockReturnValue({
            data: [
                {
                    id: 1,
                    chapterId: 10,
                    entityId: 100,
                    mentionType: 'featured',
                    createdAt: '2026-01-15T10:00:00Z',
                    entity: {
                        id: 100,
                        name: 'Dr Ward',
                        entityType: 'npc',
                    },
                },
            ],
        });

        (useChapterRelationships as Mock).mockReturnValue({
            data: [
                {
                    id: 200,
                    campaignId: 1,
                    sourceEntityId: 100,
                    targetEntityId: 103,
                    relationshipType: 'ally_of',
                    displayLabel: 'Ally of',
                    sourceEntityName: 'Dr Ward',
                    targetEntityName: 'Inspector Legrasse',
                    createdAt: '2026-01-15T10:00:00Z',
                    updatedAt: '2026-01-15T10:00:00Z',
                },
            ],
        });

        renderPage();

        expect(screen.getByText('Relationships')).toBeInTheDocument();
        // "Dr Ward" appears in both the entities section and relationships
        expect(screen.getAllByText('Dr Ward').length).toBeGreaterThanOrEqual(1);
        expect(screen.getByText('Ally of')).toBeInTheDocument();
        expect(screen.getByText('Inspector Legrasse')).toBeInTheDocument();
    });

    it('shows metadata (sort order, dates)', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Metadata Chapter',
                sortOrder: 5,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-28T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        expect(screen.getByText('Sort Order')).toBeInTheDocument();
        expect(screen.getByText('5')).toBeInTheDocument();
        expect(screen.getByText('Created')).toBeInTheDocument();
        expect(screen.getByText('Updated')).toBeInTheDocument();
    });

    it('shows Edit and Delete buttons', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Button Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        expect(
            screen.getByRole('button', { name: /edit/i })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: /delete/i })
        ).toBeInTheDocument();
    });

    it('clicking Delete opens the confirmation dialog', () => {
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Doomed Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        fireEvent.click(screen.getByRole('button', { name: /delete/i }));

        expect(
            screen.getByText(/are you sure you want to delete/i)
        ).toBeInTheDocument();
    });

    it('clicking Cancel closes the dialog without deleting', async () => {
        const mockMutateAsync = vi.fn();
        (useDeleteChapter as Mock).mockReturnValue({
            mutateAsync: mockMutateAsync,
        });
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Safe Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        // Open the delete dialog
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));
        expect(
            screen.getByText(/are you sure you want to delete/i)
        ).toBeInTheDocument();

        // Click Cancel in the dialog
        const dialog = screen.getByRole('dialog');
        fireEvent.click(within(dialog).getByRole('button', { name: /cancel/i }));

        // Dialog should close (wait for MUI transition to complete)
        await waitFor(() => {
            expect(
                screen.queryByText(/are you sure you want to delete/i)
            ).not.toBeInTheDocument();
        });

        // Delete mutation should not have been called
        expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it('clicking Delete in the dialog calls the delete mutation', async () => {
        const mockMutateAsync = vi.fn().mockResolvedValue(undefined);
        (useDeleteChapter as Mock).mockReturnValue({
            mutateAsync: mockMutateAsync,
        });
        (useChapter as Mock).mockReturnValue({
            data: {
                id: 10,
                campaignId: 1,
                title: 'Deleted Chapter',
                sortOrder: 1,
                createdAt: '2026-01-15T10:00:00Z',
                updatedAt: '2026-02-01T12:00:00Z',
            },
            isLoading: false,
            error: null,
        });

        renderPage();

        // Open the delete dialog
        fireEvent.click(screen.getByRole('button', { name: /delete/i }));

        // Find the dialog and click the Delete button within it
        const dialog = screen.getByRole('dialog');
        const dialogDeleteButton = within(dialog).getByRole('button', {
            name: /delete/i,
        });
        fireEvent.click(dialogDeleteButton);

        await waitFor(() => {
            expect(mockMutateAsync).toHaveBeenCalledWith({
                campaignId: 1,
                chapterId: 10,
            });
        });
    });

    describe('edit mode', () => {
        it('clicking Edit switches to edit mode', () => {
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // Click Edit
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // Title becomes a text input pre-filled with chapter title
            const titleInput = screen.getByLabelText(/chapter title/i);
            expect(titleInput).toBeInTheDocument();
            expect(titleInput).toHaveValue('The Dark Beginning');

            // PhaseStrip appears
            expect(screen.getByTestId('phase-strip')).toBeInTheDocument();

            // Edit and Delete buttons disappear
            expect(
                screen.queryByRole('button', { name: /^edit$/i })
            ).not.toBeInTheDocument();
            // The header Delete button is gone (only delete icons on
            // entity rows may remain)
        });

        it('Cancel exits edit mode without saving', async () => {
            const mockMutateAsync = vi.fn().mockResolvedValue({});
            (useUpdateChapter as Mock).mockReturnValue({
                mutateAsync: mockMutateAsync,
                isPending: false,
            });
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));
            expect(screen.getByTestId('phase-strip')).toBeInTheDocument();

            // Click Cancel
            fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

            // Should be back in read mode
            await waitFor(() => {
                expect(screen.queryByTestId('phase-strip')).not.toBeInTheDocument();
            });
            expect(screen.getByText('The Dark Beginning')).toBeInTheDocument();

            // Update mutation should not have been called
            expect(mockMutateAsync).not.toHaveBeenCalled();
        });

        it('Save without phases exits edit mode', async () => {
            const mockMutateAsync = vi.fn().mockResolvedValue({});
            (useUpdateChapter as Mock).mockReturnValue({
                mutateAsync: mockMutateAsync,
                isPending: false,
            });
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // Click Save & Go (mocked PhaseStrip fires with all phases false)
            fireEvent.click(screen.getByTestId('phase-strip'));

            await waitFor(() => {
                expect(mockMutateAsync).toHaveBeenCalledWith(
                    expect.objectContaining({
                        campaignId: 1,
                        chapterId: 10,
                        input: expect.objectContaining({
                            title: 'The Dark Beginning',
                        }),
                    })
                );
            });

            // Should exit edit mode
            await waitFor(() => {
                expect(screen.queryByTestId('phase-strip')).not.toBeInTheDocument();
            });
        });

        it('PhaseStrip renders in edit mode', () => {
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // PhaseStrip should not be visible initially
            expect(screen.queryByTestId('phase-strip')).not.toBeInTheDocument();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // PhaseStrip should now be visible
            expect(screen.getByTestId('phase-strip')).toBeInTheDocument();
        });

        it('shows MarkdownEditor in edit mode', () => {
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // MarkdownEditor should be visible
            expect(screen.getByTestId('markdown-editor')).toBeInTheDocument();

            // MarkdownRenderer should not be visible
            expect(screen.queryByTestId('markdown')).not.toBeInTheDocument();
        });

        it('shows EntityAutocomplete in edit mode', () => {
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // EntityAutocomplete should not be visible initially
            expect(screen.queryByTestId('entity-autocomplete')).not.toBeInTheDocument();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // EntityAutocomplete should now be visible
            expect(screen.getByTestId('entity-autocomplete')).toBeInTheDocument();
        });

        it('shows edit mode info banner with Cancel button', () => {
            (useChapter as Mock).mockReturnValue({
                data: CHAPTER_FIXTURE,
                isLoading: false,
                error: null,
            });

            renderPage();

            // Enter edit mode
            fireEvent.click(screen.getByRole('button', { name: /edit/i }));

            // Info banner should be visible
            expect(
                screen.getByText(/you are in edit mode/i)
            ).toBeInTheDocument();
        });
    });
});
