// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hook for localStorage draft management.
 *
 * Provides functions to save, retrieve, and delete drafts stored in
 * localStorage. Each draft includes metadata about when it was saved
 * and optionally the server version for conflict detection.
 */

import { useCallback } from 'react';

const DRAFT_PREFIX = 'imagineer_draft_';

/**
 * Metadata stored with each draft.
 */
export interface DraftData<T> {
    /** The draft data */
    data: T;
    /** ISO timestamp when the draft was saved */
    savedAt: string;
    /** Server version of the data (for conflict detection) */
    serverVersion?: number;
}

/**
 * Builds the namespaced localStorage key for a draft.
 *
 * @param key - The draft's identifier to append to the draft key prefix
 * @returns The full localStorage key used to store the draft
 */
function getDraftKey(key: string): string {
    return `${DRAFT_PREFIX}${key}`;
}

/**
 * Return type for the useDraft hook.
 */
interface UseDraftReturn {
    /** Retrieve a draft from localStorage */
    getDraft: <T>(key: string) => DraftData<T> | null;
    /** Save a draft to localStorage */
    saveDraft: <T>(key: string, data: T, serverVersion?: number) => void;
    /** Delete a draft from localStorage */
    deleteDraft: (key: string) => void;
    /** Check if a draft exists */
    hasDraft: (key: string) => boolean;
}

/**
 * Hook for managing drafts in localStorage.
 *
 * @returns Object containing draft management functions.
 *
 * @example
 * ```typescript
 * const { getDraft, saveDraft, deleteDraft, hasDraft } = useDraft();
 *
 * // Check for existing draft
 * const existing = getDraft<EntityFormData>('entity-new');
 * if (existing) {
 *   setFormData(existing.data);
 * }
 *
 * // Save draft
 * saveDraft('entity-new', formData, entity.version);
 *
 * // Clean up after successful save
 * deleteDraft('entity-new');
 * ```
 */
export function useDraft(): UseDraftReturn {
    /**
     * Retrieve a draft from localStorage.
     */
    const getDraft = useCallback(<T,>(key: string): DraftData<T> | null => {
        try {
            const stored = localStorage.getItem(getDraftKey(key));
            if (!stored) {
                return null;
            }
            return JSON.parse(stored) as DraftData<T>;
        } catch (error) {
            console.error('Failed to retrieve draft:', error);
            return null;
        }
    }, []);

    /**
     * Save a draft to localStorage.
     */
    const saveDraft = useCallback(
        <T,>(key: string, data: T, serverVersion?: number) => {
            const draftData: DraftData<T> = {
                data,
                savedAt: new Date().toISOString(),
                serverVersion,
            };
            try {
                localStorage.setItem(
                    getDraftKey(key),
                    JSON.stringify(draftData)
                );
            } catch (error) {
                console.error('Failed to save draft:', error);
            }
        },
        []
    );

    /**
     * Delete a draft from localStorage.
     */
    const deleteDraft = useCallback((key: string) => {
        localStorage.removeItem(getDraftKey(key));
    }, []);

    /**
     * Check if a draft exists.
     */
    const hasDraft = useCallback((key: string): boolean => {
        return localStorage.getItem(getDraftKey(key)) !== null;
    }, []);

    return {
        getDraft,
        saveDraft,
        deleteDraft,
        hasDraft,
    };
}