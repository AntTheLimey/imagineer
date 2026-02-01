// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/* eslint-disable react-refresh/only-export-components */

import {
    createContext,
    useContext,
    useCallback,
    ReactNode,
} from 'react';

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
 * Draft context value providing localStorage draft management.
 */
interface DraftContextValue {
    /** Save a draft to localStorage */
    saveDraft: <T>(key: string, data: T, serverVersion?: number) => void;
    /** Retrieve a draft from localStorage */
    getDraft: <T>(key: string) => DraftData<T> | null;
    /** Delete a draft from localStorage */
    deleteDraft: (key: string) => void;
    /** Check if a draft exists */
    hasDraft: (key: string) => boolean;
    /** List all draft keys */
    listDrafts: () => string[];
    /** Clear all drafts */
    clearAllDrafts: () => void;
}

const DraftContext = createContext<DraftContextValue | undefined>(undefined);

interface DraftProviderProps {
    children: ReactNode;
}

/**
 * Construct the full localStorage key for a draft by prepending the draft prefix.
 *
 * @param key - The draft identifier (without the prefix)
 * @returns The storage key formed by concatenating the draft prefix and `key`
 */
function getDraftKey(key: string): string {
    return `${DRAFT_PREFIX}${key}`;
}

/**
 * Provides draft management via React context to descendant components.
 *
 * Drafts are persisted to localStorage under a namespaced key and include metadata:
 * an ISO `savedAt` timestamp and an optional `serverVersion` for conflict detection.
 *
 * @returns The provider element that exposes draft management APIs to descendants via context.
 */
export function DraftProvider({ children }: DraftProviderProps) {
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
     * Delete a draft from localStorage.
     */
    const deleteDraft = useCallback((key: string) => {
        try {
            localStorage.removeItem(getDraftKey(key));
        } catch (error) {
            console.error('Failed to delete draft:', error);
        }
    }, []);

    /**
     * Check if a draft exists.
     */
    const hasDraft = useCallback((key: string): boolean => {
        try {
            return localStorage.getItem(getDraftKey(key)) !== null;
        } catch (error) {
            console.error('Failed to check draft:', error);
            return false;
        }
    }, []);

    /**
     * List all draft keys (without the prefix).
     */
    const listDrafts = useCallback((): string[] => {
        try {
            const drafts: string[] = [];
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(DRAFT_PREFIX)) {
                    drafts.push(key.slice(DRAFT_PREFIX.length));
                }
            }
            return drafts;
        } catch (error) {
            console.error('Failed to list drafts:', error);
            return [];
        }
    }, []);

    /**
     * Clear all drafts from localStorage.
     */
    const clearAllDrafts = useCallback(() => {
        try {
            const keysToRemove: string[] = [];
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (key && key.startsWith(DRAFT_PREFIX)) {
                    keysToRemove.push(key);
                }
            }
            keysToRemove.forEach((key) => {
                localStorage.removeItem(key);
            });
        } catch (error) {
            console.error('Failed to clear drafts:', error);
        }
    }, []);

    const value: DraftContextValue = {
        saveDraft,
        getDraft,
        deleteDraft,
        hasDraft,
        listDrafts,
        clearAllDrafts,
    };

    return (
        <DraftContext.Provider value={value}>
            {children}
        </DraftContext.Provider>
    );
}

/**
 * Retrieve the draft management API from the nearest DraftProvider.
 *
 * @returns The DraftContextValue providing draft operations (save, get, delete, etc.).
 * @throws Error if the hook is called outside of a DraftProvider.
 */
export function useDraftContext(): DraftContextValue {
    const context = useContext(DraftContext);
    if (context === undefined) {
        throw new Error(
            'useDraftContext must be used within a DraftProvider'
        );
    }
    return context;
}