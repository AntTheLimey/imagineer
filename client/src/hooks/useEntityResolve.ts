// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hook for resolving entity names to matching entities.
 * Includes built-in debouncing to avoid excessive API calls during typing.
 */

import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { entityResolveApi } from '../api/entityResolve';
import type { EntityResolveResult } from '../api/entityResolve';

/**
 * Query keys for entity resolve queries.
 */
export const entityResolveKeys = {
    all: ['entities', 'resolve'] as const,
    resolve: (campaignId: number, term: string) =>
        [...entityResolveKeys.all, campaignId, term] as const,
};

/**
 * Hook to resolve entity names with built-in debouncing.
 *
 * Debounces the search term by 300ms before issuing the API request.
 * Only queries when the campaign ID is valid and the debounced term
 * is at least 3 characters long.
 *
 * @param campaignId - The campaign to search within.
 * @param searchTerm - The raw search term (will be debounced internally).
 * @returns Object with resolve results and loading state.
 */
export function useEntityResolve(campaignId: number, searchTerm: string) {
    const [debouncedTerm, setDebouncedTerm] = useState(searchTerm);

    // Debounce the search term
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedTerm(searchTerm);
        }, 300);

        return () => clearTimeout(timer);
    }, [searchTerm]);

    const { data, isLoading } = useQuery<EntityResolveResult[]>({
        queryKey: entityResolveKeys.resolve(campaignId, debouncedTerm),
        queryFn: () => entityResolveApi.resolve(campaignId, debouncedTerm),
        enabled: !!campaignId && debouncedTerm.length >= 3,
    });

    return { data, isLoading };
}
