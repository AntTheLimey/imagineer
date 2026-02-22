/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import EntityAutocomplete from './EntityAutocomplete';
import type { Entity } from '../types';

/**
 * Mock entities returned by the useEntities hook.
 */
const mockEntities: Entity[] = [
    {
        id: 1,
        campaignId: 1,
        entityType: 'npc',
        name: 'Sherlock Holmes',
        description: 'A detective',
        attributes: {},
        tags: [],
        sourceConfidence: 'AUTHORITATIVE',
        version: 1,
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 2,
        campaignId: 1,
        entityType: 'location',
        name: 'Baker Street',
        description: 'A famous address',
        attributes: {},
        tags: [],
        sourceConfidence: 'AUTHORITATIVE',
        version: 1,
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    },
    {
        id: 3,
        campaignId: 1,
        entityType: 'item',
        name: 'Magnifying Glass',
        description: 'An investigative tool',
        attributes: {},
        tags: [],
        sourceConfidence: 'DRAFT',
        version: 1,
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    },
];

/**
 * Mock the useEntities hook so tests do not require a live API.
 */
const mockUseEntities = vi.fn();

vi.mock('../hooks/useEntities', () => ({
    useEntities: (params: unknown) => mockUseEntities(params),
}));

/**
 * Creates a fresh QueryClient for each test to prevent cache leaks.
 */
function createTestQueryClient() {
    return new QueryClient({
        defaultOptions: {
            queries: { retry: false },
        },
    });
}

/**
 * Renders the component wrapped in the required providers.
 */
function renderComponent(props: Partial<React.ComponentProps<typeof EntityAutocomplete>> = {}) {
    const queryClient = createTestQueryClient();
    const defaultProps = {
        campaignId: 1,
        onSelect: vi.fn(),
        ...props,
    };

    return {
        ...render(
            <QueryClientProvider client={queryClient}>
                <EntityAutocomplete {...defaultProps} />
            </QueryClientProvider>,
        ),
        onSelect: defaultProps.onSelect,
    };
}

describe('EntityAutocomplete', () => {
    beforeEach(() => {
        mockUseEntities.mockReturnValue({
            data: mockEntities,
            isLoading: false,
        });
    });

    it('renders an autocomplete input with default label', () => {
        renderComponent();

        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(screen.getByLabelText('Search entities...')).toBeInTheDocument();
    });

    it('renders with a custom label', () => {
        renderComponent({ label: 'Reassign to entity' });

        expect(screen.getByLabelText('Reassign to entity')).toBeInTheDocument();
    });

    it('passes campaignId and searchTerm to useEntities', async () => {
        const user = userEvent.setup();
        renderComponent();

        const input = screen.getByRole('combobox');
        await user.type(input, 'Sher');

        await waitFor(() => {
            expect(mockUseEntities).toHaveBeenCalledWith(
                expect.objectContaining({
                    campaignId: 1,
                    searchTerm: 'Sher',
                    pageSize: 20,
                }),
            );
        });
    });

    it('shows options when the autocomplete is opened', async () => {
        const user = userEvent.setup();
        renderComponent();

        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByText('Sherlock Holmes')).toBeInTheDocument();
            expect(screen.getByText('Baker Street')).toBeInTheDocument();
            expect(screen.getByText('Magnifying Glass')).toBeInTheDocument();
        });
    });

    it('displays entity type chips alongside entity names', async () => {
        const user = userEvent.setup();
        renderComponent();

        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByText('Npc')).toBeInTheDocument();
            expect(screen.getByText('Location')).toBeInTheDocument();
            expect(screen.getByText('Item')).toBeInTheDocument();
        });
    });

    it('calls onSelect with the chosen entity', async () => {
        const user = userEvent.setup();
        const { onSelect } = renderComponent();

        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByText('Baker Street')).toBeInTheDocument();
        });

        await user.click(screen.getByText('Baker Street'));

        expect(onSelect).toHaveBeenCalledWith({
            id: 2,
            name: 'Baker Street',
            entityType: 'location',
        });
    });

    it('excludes entity matching excludeEntityId from results', async () => {
        const user = userEvent.setup();
        renderComponent({ excludeEntityId: 1 });

        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByText('Baker Street')).toBeInTheDocument();
            expect(screen.getByText('Magnifying Glass')).toBeInTheDocument();
        });

        expect(screen.queryByText('Sherlock Holmes')).not.toBeInTheDocument();
    });

    it('shows "Type to search..." when input is empty', async () => {
        mockUseEntities.mockReturnValue({
            data: [],
            isLoading: false,
        });
        const user = userEvent.setup();
        renderComponent();

        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByText('Type to search...')).toBeInTheDocument();
        });
    });

    it('shows loading indicator when popup is open and data is loading', async () => {
        mockUseEntities.mockReturnValue({
            data: undefined,
            isLoading: true,
        });
        const user = userEvent.setup();
        renderComponent();

        // Open the autocomplete popup so MUI renders the loading indicator.
        const input = screen.getByRole('combobox');
        await user.click(input);

        await waitFor(() => {
            expect(screen.getByRole('progressbar')).toBeInTheDocument();
        });
    });

    it('debounces input before querying', async () => {
        const user = userEvent.setup();
        renderComponent();

        // Clear call history after the initial render call.
        mockUseEntities.mockClear();

        const input = screen.getByRole('combobox');
        await user.type(input, 'abc');

        // The hook is called on every render with the current
        // debouncedInput. Immediately after typing, the debounced
        // value should still be empty (undefined searchTerm).
        const immediateSearchTermCalls = mockUseEntities.mock.calls.filter(
            (call: unknown[]) => {
                const params = call[0] as { searchTerm?: string };
                return params.searchTerm === 'abc';
            },
        );
        expect(immediateSearchTermCalls.length).toBe(0);

        // After waiting for the debounce period, the search term
        // should propagate to the hook.
        await waitFor(() => {
            expect(mockUseEntities).toHaveBeenCalledWith(
                expect.objectContaining({ searchTerm: 'abc' }),
            );
        }, { timeout: 2000 });
    });
});
