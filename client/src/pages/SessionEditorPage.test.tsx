// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Tests for SessionEditorPage component.
 *
 * Mock strategy: synchronous vi.mock factories. Component mocks
 * either render nothing (vi.fn(() => null)) or pass through children.
 * Default hook return values live in the factories; per-test overrides
 * use mockReturnValue in beforeEach or inside individual tests.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DraftProvider } from '../contexts/DraftContext';

// ---------------------------------------------------------------------------
// Module mocks -- all synchronous factories with sensible defaults.
// ---------------------------------------------------------------------------

// Stable mock function references to mimic useCallback behavior in real hooks.
// Creating new vi.fn() instances inside mock factories would cause the
// component's useEffect dependencies to change on every render, triggering
// an infinite re-render loop in edit mode.
const stableGetDraft = vi.fn(() => null);
const stableSaveDraft = vi.fn();
const stableDeleteDraft = vi.fn();
const stableHasDraft = vi.fn(() => false);
const stableSaveNow = vi.fn();
const stableSetIsDirty = vi.fn();
const stableMarkDirty = vi.fn();
const stableClearDirty = vi.fn();
const stableShowConfirmDialog = vi.fn();
const stableHideConfirmDialog = vi.fn();
const stableCheckUnsavedChanges = vi.fn(() => false);
const stableSaveDraftToServer = vi.fn();
const stableDeleteDraftFromServer = vi.fn();

vi.mock('../hooks', () => ({
    useCampaign: vi.fn(() => ({
        data: {
            id: 1, name: 'Test Campaign', systemId: 1, settings: {},
            createdAt: '2025-01-01T00:00:00Z',
            updatedAt: '2025-01-01T00:00:00Z',
        },
        isLoading: false,
        error: null,
    })),
    useEntities: vi.fn(() => ({
        data: [],
        isLoading: false,
        error: null,
    })),
    useChapters: vi.fn(() => ({
        data: [
            {
                id: 1, campaignId: 1, title: 'Chapter 1', sortOrder: 1,
                createdAt: '2025-01-01T00:00:00Z',
                updatedAt: '2025-01-01T00:00:00Z',
            },
            {
                id: 2, campaignId: 1, title: 'Chapter 2', sortOrder: 2,
                createdAt: '2025-01-01T00:00:00Z',
                updatedAt: '2025-01-01T00:00:00Z',
            },
        ],
        isLoading: false,
        error: null,
    })),
    useDraft: vi.fn(() => ({
        getDraft: stableGetDraft,
        saveDraft: stableSaveDraft,
        deleteDraft: stableDeleteDraft,
        hasDraft: stableHasDraft,
    })),
    useAutosave: vi.fn(() => ({
        lastSaved: null,
        saveNow: stableSaveNow,
    })),
    useUnsavedChanges: vi.fn(() => ({
        isDirty: false,
        setIsDirty: stableSetIsDirty,
        markDirty: stableMarkDirty,
        clearDirty: stableClearDirty,
        showConfirmDialog: stableShowConfirmDialog,
        hideConfirmDialog: stableHideConfirmDialog,
        checkUnsavedChanges: stableCheckUnsavedChanges,
        ConfirmDialog: null,
    })),
    useServerDraft: vi.fn(() => ({
        serverDraft: null,
        saveDraftToServer: stableSaveDraftToServer,
        deleteDraftFromServer: stableDeleteDraftFromServer,
        isDirty: false,
        isLoading: false,
        draftUpdatedAt: null,
        hasDraft: false,
        lastSaved: null,
    })),
}));

vi.mock('../hooks/useSessions', () => ({
    sessionKeys: {
        all: ['sessions'],
        lists: () => ['sessions', 'list'],
        list: (cid: number, p?: unknown) => ['sessions', 'list', cid, p],
        byChapter: (cid: number, chid: number) =>
            ['sessions', 'list', 'byChapter', cid, chid],
        details: () => ['sessions', 'detail'],
        detail: (cid: number, sid: number) =>
            ['sessions', 'detail', cid, sid],
    },
    useSession: vi.fn(() => ({
        data: undefined,
        isLoading: false,
        error: null,
    })),
    useSessions: vi.fn(() => ({
        data: [],
        isLoading: false,
        error: null,
    })),
    useSessionsByChapter: vi.fn(() => ({
        data: [],
        isLoading: false,
        error: null,
    })),
    useCreateSession: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
    useUpdateSession: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
    useDeleteSession: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
}));

vi.mock('../hooks/useScenes', () => ({
    sceneKeys: {
        all: ['scenes'],
        lists: () => ['scenes', 'list'],
        list: (cid: number, sid: number) => ['scenes', 'list', cid, sid],
        details: () => ['scenes', 'detail'],
        detail: (cid: number, sid: number, scid: number) =>
            ['scenes', 'detail', cid, sid, scid],
    },
    useScenes: vi.fn(() => ({
        data: undefined,
        isLoading: false,
        error: null,
    })),
    useScene: vi.fn(),
    useCreateScene: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
    useUpdateScene: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
    useDeleteScene: vi.fn(() => ({
        mutateAsync: vi.fn(),
        isPending: false,
    })),
}));

// Stub MarkdownEditor: renders nothing; the "Prep Notes" label comes
// from SessionEditorPage itself, so text assertions still work.
vi.mock('../components/MarkdownEditor', () => ({
    MarkdownEditor: vi.fn(() => null),
    default: vi.fn(() => null),
}));

// Stub SaveSplitButton: renders nothing; tests verify the mock was
// called with the expected props.
vi.mock('../components/SaveSplitButton', () => ({
    SaveSplitButton: vi.fn(() => null),
    default: vi.fn(() => null),
}));

// Stub FullScreenLayout: transparent wrapper that renders only children.
// Title and breadcrumb assertions use mock-call inspection instead.
vi.mock('../layouts', () => ({
    FullScreenLayout: vi.fn(
        ({ children }: { children: unknown }) => children
    ),
    default: vi.fn(() => null),
}));

// Stub Play components: render nothing in tests.
vi.mock('../components/Play', () => ({
    SceneStrip: vi.fn(() => null),
    SceneViewer: vi.fn(() => null),
    PlayEntityDrawer: vi.fn(() => null),
    PlayScratchpad: vi.fn(() => null),
    PlayEntitySidebar: vi.fn(() => null),
}));

// Stub ImportNotesDialog: render nothing in tests.
vi.mock('../components/Sessions', () => ({
    ImportNotesDialog: vi.fn(() => null),
}));

// ---------------------------------------------------------------------------
// Imports after mocks
// ---------------------------------------------------------------------------

import SessionEditorPage from './SessionEditorPage';
import { useSession } from '../hooks/useSessions';
import { useScenes } from '../hooks/useScenes';
import { FullScreenLayout } from '../layouts';
import { PlayScratchpad } from '../components/Play';

const mockUseSession = vi.mocked(useSession);
const mockUseScenes = vi.mocked(useScenes);
const mockFullScreenLayout = vi.mocked(FullScreenLayout);
const mockPlayScratchpad = vi.mocked(PlayScratchpad);

// ---------------------------------------------------------------------------
// Shared test fixtures
// ---------------------------------------------------------------------------

const mockSession = {
    id: 42,
    campaignId: 1,
    chapterId: 1,
    title: 'The Dunwich Horror',
    sessionNumber: 3,
    plannedDate: '2025-03-15',
    status: 'PLANNED' as const,
    stage: 'prep' as const,
    prepNotes: 'Investigate the farmhouse',
    actualNotes: '',
    playNotes: '',
    createdAt: '2025-01-01T00:00:00Z',
    updatedAt: '2025-01-01T00:00:00Z',
};

const mockScenes = [
    {
        id: 1,
        sessionId: 42,
        campaignId: 1,
        title: 'Arrival at Dunwich',
        description:
            'The investigators arrive at the small town of Dunwich.',
        sceneType: 'exploration',
        status: 'planned',
        sortOrder: 1,
        entityIds: [],
        source: 'gm',
        sourceConfidence: 'AUTHORITATIVE',
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 2,
        sessionId: 42,
        campaignId: 1,
        title: 'Farmhouse Investigation',
        description:
            'The group explores the abandoned Whateley farmhouse.',
        sceneType: 'combat',
        status: 'completed',
        sortOrder: 2,
        entityIds: [],
        source: 'gm',
        sourceConfidence: 'AUTHORITATIVE',
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    },
];

// ---------------------------------------------------------------------------
// Test suite
// ---------------------------------------------------------------------------

describe('SessionEditorPage', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        vi.stubGlobal(
            'requestAnimationFrame',
            (cb: FrameRequestCallback) =>
                setTimeout(() => cb(Date.now()), 0) as unknown as number
        );
        vi.stubGlobal('cancelAnimationFrame', (id: number) => {
            clearTimeout(id);
        });
        queryClient = new QueryClient({
            defaultOptions: { queries: { retry: false } },
        });

        // Reset to default hook values (create mode: no session loaded).
        mockUseSession.mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useSession>);

        mockUseScenes.mockReturnValue({
            data: undefined,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useScenes>);
    });

    afterEach(() => {
        cleanup();
        queryClient.clear();
        vi.unstubAllGlobals();
    });

    // Helpers ---------------------------------------------------------------

    function renderCreateMode() {
        return render(
            <QueryClientProvider client={queryClient}>
                <MemoryRouter
                    initialEntries={['/campaigns/1/sessions/new']}
                >
                    <DraftProvider>
                        <Routes>
                            <Route
                                path="/campaigns/:campaignId/sessions/new"
                                element={<SessionEditorPage />}
                            />
                        </Routes>
                    </DraftProvider>
                </MemoryRouter>
            </QueryClientProvider>
        );
    }

    function renderEditMode() {
        return render(
            <QueryClientProvider client={queryClient}>
                <MemoryRouter
                    initialEntries={['/campaigns/1/sessions/42/edit']}
                >
                    <DraftProvider>
                        <Routes>
                            <Route
                                path="/campaigns/:campaignId/sessions/:sessionId/edit"
                                element={<SessionEditorPage />}
                            />
                        </Routes>
                    </DraftProvider>
                </MemoryRouter>
            </QueryClientProvider>
        );
    }

    /** Set up mocks for edit mode with session + optional scenes loaded. */
    function setupEditMocks(scenes: typeof mockScenes | [] = []) {
        mockUseSession.mockReturnValue({
            data: mockSession,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useSession>);

        mockUseScenes.mockReturnValue({
            data: scenes,
            isLoading: false,
            error: null,
        } as unknown as ReturnType<typeof useScenes>);
    }

    // -----------------------------------------------------------------
    // Create mode
    // -----------------------------------------------------------------

    it('renders in create mode with "New Session" title', () => {
        renderCreateMode();

        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({ title: 'New Session' }),
            expect.anything()
        );
    });

    it('renders empty form fields in create mode', () => {
        renderCreateMode();

        const titleInput = screen.getByLabelText('Title');
        expect(titleInput).toBeInTheDocument();
        expect(titleInput).toHaveValue('');
    });

    it('shows "Save the session first" message for new sessions', () => {
        renderCreateMode();

        expect(
            screen.getByText('Save the session first, then add scenes.')
        ).toBeInTheDocument();
    });

    it('disables the Add Scene button in create mode', () => {
        renderCreateMode();

        const addButton = screen.getByRole('button', {
            name: /add scene/i,
        });
        expect(addButton).toBeDisabled();
    });

    // -----------------------------------------------------------------
    // Edit mode
    // -----------------------------------------------------------------

    it('renders loading state while session is loading', () => {
        mockUseSession.mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        } as unknown as ReturnType<typeof useSession>);

        renderEditMode();

        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({ title: 'Loading...' }),
            expect.anything()
        );
    });

    it('renders error state when session fails to load', () => {
        mockUseSession.mockReturnValue({
            data: undefined,
            isLoading: false,
            error: new Error('Not found'),
        } as unknown as ReturnType<typeof useSession>);

        renderEditMode();

        expect(
            screen.getByText(/Failed to load session/)
        ).toBeInTheDocument();
    });

    it('renders "Edit Session" title in edit mode', () => {
        setupEditMocks();
        renderEditMode();

        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({ title: 'Edit Session' }),
            expect.anything()
        );
    });

    it('populates form fields from session data in edit mode', () => {
        setupEditMocks();
        renderEditMode();

        expect(screen.getByLabelText('Title')).toHaveValue(
            'The Dunwich Horror'
        );
        expect(screen.getByLabelText('Session Number')).toHaveValue(3);
    });

    // -----------------------------------------------------------------
    // Stage tabs
    // -----------------------------------------------------------------

    it('renders all four stage tabs', () => {
        renderCreateMode();

        expect(
            screen.getByRole('tab', { name: 'Prep' })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('tab', { name: 'Play' })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('tab', { name: 'Wrap-up' })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('tab', { name: 'Completed' })
        ).toBeInTheDocument();
    });

    it('defaults to Prep tab showing form fields', () => {
        renderCreateMode();

        const prepTab = screen.getByRole('tab', { name: 'Prep' });
        expect(prepTab).toHaveAttribute('aria-selected', 'true');
        expect(screen.getByLabelText('Title')).toBeInTheDocument();
        expect(screen.getByText('Prep Notes')).toBeInTheDocument();
    });

    it('shows Play mode layout when switching to Play tab', async () => {
        renderCreateMode();

        const user = userEvent.setup();
        await user.click(screen.getByRole('tab', { name: 'Play' }));

        // Prep form should not be visible
        expect(screen.queryByLabelText('Title')).not.toBeInTheDocument();
        // Play mode renders (mocked components render null, but placeholder is gone)
        expect(screen.queryByText('Coming in Phase 2')).not.toBeInTheDocument();
        // Positive: verify Play components were rendered
        expect(mockPlayScratchpad).toHaveBeenCalled();
    });

    it('shows placeholder for Wrap-up tab', async () => {
        renderCreateMode();

        const user = userEvent.setup();
        await user.click(screen.getByRole('tab', { name: 'Wrap-up' }));

        expect(screen.getByText('Wrap-up Stage')).toBeInTheDocument();
        expect(screen.getByText('Coming in Phase 3')).toBeInTheDocument();
    });

    it('shows placeholder for Completed tab', async () => {
        renderCreateMode();

        const user = userEvent.setup();
        await user.click(
            screen.getByRole('tab', { name: 'Completed' })
        );

        expect(
            screen.getByText('Completed Stage')
        ).toBeInTheDocument();
        expect(screen.getByText('Coming in Phase 3')).toBeInTheDocument();
    });

    // -----------------------------------------------------------------
    // Prep tab form fields
    // -----------------------------------------------------------------

    it('renders the Chapter selector on the Prep tab', () => {
        renderCreateMode();

        // MUI Select renders "Chapter" in both the InputLabel and the
        // outlined fieldset legend. Query the label element directly.
        const chapterLabel = screen.getAllByText('Chapter')
            .find((el) => el.tagName === 'LABEL');
        expect(chapterLabel).toBeDefined();
    });

    it('renders Session Number field on the Prep tab', () => {
        renderCreateMode();

        expect(
            screen.getByLabelText('Session Number')
        ).toBeInTheDocument();
    });

    it('renders Planned Date field on the Prep tab', () => {
        renderCreateMode();

        expect(
            screen.getByLabelText('Planned Date')
        ).toBeInTheDocument();
    });

    it('renders the Prep Notes markdown editor', () => {
        renderCreateMode();

        expect(screen.getByText('Prep Notes')).toBeInTheDocument();
    });

    // -----------------------------------------------------------------
    // Scene sidebar
    // -----------------------------------------------------------------

    it('renders the Scenes heading in the sidebar', () => {
        setupEditMocks();
        renderEditMode();

        expect(screen.getByText('Scenes')).toBeInTheDocument();
    });

    it('shows empty state when no scenes exist', () => {
        setupEditMocks([]);
        renderEditMode();

        expect(
            screen.getByText(
                'No scenes yet. Add scenes to plan your session structure.'
            )
        ).toBeInTheDocument();
    });

    it('renders scene list with scene titles and types', () => {
        setupEditMocks(mockScenes);
        renderEditMode();

        expect(
            screen.getByText('Arrival at Dunwich')
        ).toBeInTheDocument();
        expect(
            screen.getByText('Farmhouse Investigation')
        ).toBeInTheDocument();
        expect(screen.getByText('exploration')).toBeInTheDocument();
        expect(screen.getByText('combat')).toBeInTheDocument();
    });

    it('shows scene count chip when scenes exist', () => {
        setupEditMocks(mockScenes);
        renderEditMode();

        const scenesHeader = screen
            .getByText('Scenes')
            .closest('div');
        expect(scenesHeader).not.toBeNull();
        expect(
            within(scenesHeader!).getByText('2')
        ).toBeInTheDocument();
    });

    it('renders scene descriptions', () => {
        setupEditMocks(mockScenes);
        renderEditMode();

        expect(
            screen.getByText(
                'The investigators arrive at the small town of Dunwich.'
            )
        ).toBeInTheDocument();
    });

    it('renders scene status chips', () => {
        setupEditMocks(mockScenes);
        renderEditMode();

        expect(screen.getByText('planned')).toBeInTheDocument();
        expect(screen.getByText('completed')).toBeInTheDocument();
    });

    // -----------------------------------------------------------------
    // Save button
    // -----------------------------------------------------------------

    it('passes renderSaveButtons callback to the layout', () => {
        renderCreateMode();

        // The page provides a renderSaveButtons callback to
        // FullScreenLayout.  Since the mock only renders children,
        // we verify the prop was passed.
        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({
                renderSaveButtons: expect.any(Function),
            }),
            expect.anything()
        );
    });

    // -----------------------------------------------------------------
    // Breadcrumbs
    // -----------------------------------------------------------------

    it('passes breadcrumbs with campaign name to the layout', () => {
        renderCreateMode();

        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({
                breadcrumbs: expect.arrayContaining([
                    expect.objectContaining({ label: 'Campaigns' }),
                    expect.objectContaining({ label: 'Test Campaign' }),
                    expect.objectContaining({ label: 'Sessions' }),
                ]),
            }),
            expect.anything()
        );
    });

    it('includes session title in breadcrumbs for edit mode', () => {
        setupEditMocks();
        renderEditMode();

        expect(mockFullScreenLayout).toHaveBeenCalledWith(
            expect.objectContaining({
                breadcrumbs: expect.arrayContaining([
                    expect.objectContaining({
                        label: 'The Dunwich Horror',
                    }),
                ]),
            }),
            expect.anything()
        );
    });

    // -----------------------------------------------------------------
    // User interaction
    // -----------------------------------------------------------------

    it('allows typing in the title field', async () => {
        renderCreateMode();

        const user = userEvent.setup();
        const titleInput = screen.getByLabelText('Title');
        await user.type(titleInput, 'Session Alpha');

        expect(titleInput).toHaveValue('Session Alpha');
    });

    it('allows typing in the session number field', async () => {
        renderCreateMode();

        const user = userEvent.setup();
        const sessionNumberInput =
            screen.getByLabelText('Session Number');
        await user.type(sessionNumberInput, '5');

        expect(sessionNumberInput).toHaveValue(5);
    });
});
