// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React hook for entity detection in text segments.
 * Manages suggestions in local state and tracks dismissed entities.
 */

import { useState, useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';
import { entityDetectionApi, EntitySuggestion } from '../api/entityDetection';

/**
 * Options for the useEntityDetection hook.
 */
export interface UseEntityDetectionOptions {
    campaignId: number;
    excludeEntityIds?: number[];
}

/**
 * Return type for the useEntityDetection hook.
 */
export interface UseEntityDetectionReturn {
    suggestions: EntitySuggestion[];
    isConfigured: boolean;
    isAnalyzing: boolean;
    error: Error | null;
    analyze: (textSegments: string[]) => Promise<void>;
    clearSuggestions: () => void;
    dismissSuggestion: (entityId: number) => void;
}

/**
 * Hook for detecting entities in text segments.
 *
 * Manages entity suggestions in local state rather than the React Query cache,
 * allowing for dismissal of individual suggestions and manual clearing.
 *
 * @param options - Configuration options including campaignId and excluded entities
 * @returns Object with suggestions, loading state, and control functions
 */
export function useEntityDetection(
    options: UseEntityDetectionOptions
): UseEntityDetectionReturn {
    const { campaignId, excludeEntityIds = [] } = options;

    // Local state for suggestions
    const [suggestions, setSuggestions] = useState<EntitySuggestion[]>([]);

    // Track whether the backend is configured for entity detection
    const [isConfigured, setIsConfigured] = useState(true);

    // Track dismissed entity IDs (entities user has rejected)
    const [dismissedIds, setDismissedIds] = useState<Set<number>>(new Set());

    // Mutation for calling the detection API
    const mutation = useMutation({
        mutationFn: (textSegments: string[]) =>
            entityDetectionApi.detect(campaignId, {
                textSegments,
                excludeEntityIds: [
                    ...excludeEntityIds,
                    ...Array.from(dismissedIds),
                ],
            }),
        onSuccess: (data) => {
            setIsConfigured(data.configured);
            // Filter out any dismissed entities from the response
            const filteredSuggestions = data.suggestions.filter(
                (suggestion) => !dismissedIds.has(suggestion.entity.id)
            );
            setSuggestions(filteredSuggestions);
        },
    });

    /**
     * Analyze text segments to detect entity mentions.
     */
    const analyze = useCallback(
        async (textSegments: string[]): Promise<void> => {
            if (!campaignId || textSegments.length === 0) {
                return;
            }
            await mutation.mutateAsync(textSegments);
        },
        [campaignId, mutation]
    );

    /**
     * Clear all suggestions.
     */
    const clearSuggestions = useCallback((): void => {
        setSuggestions([]);
    }, []);

    /**
     * Dismiss a suggestion by entity ID.
     * The dismissed entity will be filtered from current and future suggestions.
     */
    const dismissSuggestion = useCallback((entityId: number): void => {
        setDismissedIds((prev) => new Set([...prev, entityId]));
        setSuggestions((prev) =>
            prev.filter((suggestion) => suggestion.entity.id !== entityId)
        );
    }, []);

    return {
        suggestions,
        isConfigured,
        isAnalyzing: mutation.isPending,
        error: mutation.error,
        analyze,
        clearSuggestions,
        dismissSuggestion,
    };
}
