// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hook for fetching draft indicators in a single batch query.
 *
 * Provides O(1) lookups to check whether a draft exists for a given
 * source ID, and whether any "new item" drafts exist for the filtered
 * source table.
 */

import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { draftsApi } from '../api/drafts';

/**
 * Return type for the useDraftIndicators hook.
 */
export interface UseDraftIndicatorsReturn {
    /** Check whether a draft exists for the given source ID. */
    hasDraft: (sourceId: number) => boolean;
    /** Whether any indicator represents a new (unsaved) item. */
    hasNewItemDraft: boolean;
    /** Whether the indicators query is loading. */
    isLoading: boolean;
}

/**
 * Fetch draft indicators for a campaign in a single batch query and
 * expose O(1) lookups for checking draft existence.
 *
 * @param campaignId - The campaign to fetch indicators for.
 * @param sourceTable - Optional source table filter (e.g. 'entities').
 * @returns An object with a `hasDraft` lookup function, a flag for
 *          new-item drafts, and loading state.
 *
 * @example
 * ```typescript
 * const { hasDraft, hasNewItemDraft, isLoading } = useDraftIndicators(
 *     campaignId,
 *     'entities'
 * );
 *
 * // In a list item:
 * <DraftIndicator hasDraft={hasDraft(entity.id)} />
 * ```
 */
export function useDraftIndicators(
    campaignId: number,
    sourceTable?: string
): UseDraftIndicatorsReturn {
    const { data: indicators, isLoading } = useQuery({
        queryKey: ['draftIndicators', campaignId, sourceTable],
        queryFn: () => draftsApi.listIndicators(campaignId, sourceTable),
        enabled: !!campaignId,
        staleTime: 30000,
    });

    // Convert the array to a Set for O(1) lookups
    const draftIdSet = useMemo(() => {
        if (!indicators) return new Set<number>();
        return new Set(indicators.map((ind) => ind.sourceId));
    }, [indicators]);

    // Check if any indicator is for a new item
    const hasNewItemDraft = useMemo(() => {
        if (!indicators) return false;
        return indicators.some((ind) => ind.isNew);
    }, [indicators]);

    const hasDraft = useMemo(
        () => (sourceId: number): boolean => draftIdSet.has(sourceId),
        [draftIdSet]
    );

    return {
        hasDraft,
        hasNewItemDraft,
        isLoading,
    };
}
