// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hook for server-side draft persistence.
 *
 * Replaces the combination of useDraft + useAutosave + useUnsavedChanges
 * for editor pages that need server-backed draft storage. Provides
 * debounced autosave to the server, localStorage fallback, dirty state
 * tracking, and reliable save-on-unmount via fetch with keepalive.
 */

import {
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react';
import {
    useQuery,
    useMutation,
    useQueryClient,
} from '@tanstack/react-query';
import { draftsApi } from '../api/drafts';
import type { Draft, SaveDraftInput } from '../api/drafts';

const AUTOSAVE_DEBOUNCE_MS = 500;
const LOCAL_DRAFT_PREFIX = 'imagineer_server_draft_';

/**
 * Options for the useServerDraft hook.
 */
export interface UseServerDraftOptions<T> {
    /** The campaign this draft belongs to. */
    campaignId: number;
    /** The source table (e.g. 'entities', 'sessions'). */
    sourceTable: string;
    /** The source row ID. Use 0 for new items. */
    sourceId: number;
    /** Whether this draft is for a new (unsaved) item. */
    isNew: boolean;
    /** The last committed/saved data from the server. */
    committedData: T | undefined;
    /** The current form data to compare and save. */
    currentData: T;
    /** Server version for conflict detection. */
    serverVersion?: number;
    /** Whether the hook is enabled (default: true). */
    enabled?: boolean;
}

/**
 * Return type for the useServerDraft hook.
 */
export interface UseServerDraftReturn<T> {
    /** The draft data loaded from the server, or null if none. */
    serverDraft: T | null;
    /** Whether a server draft exists. */
    hasDraft: boolean;
    /** Manually trigger a save to the server. */
    saveDraftToServer: () => Promise<void>;
    /** Delete the draft from the server and invalidate caches. */
    deleteDraftFromServer: () => Promise<void>;
    /** Whether currentData differs from committedData. */
    isDirty: boolean;
    /** ISO timestamp of the last successful server save, or null. */
    lastSaved: string | null;
    /** Whether the draft query is loading. */
    isLoading: boolean;
    /** ISO timestamp of when the draft was last updated on the server, or null. */
    draftUpdatedAt: string | null;
}

/**
 * Produce a stable JSON string for deep comparison.
 * Sorts object keys to ensure consistent ordering.
 */
function stableStringify(value: unknown): string {
    return JSON.stringify(value, (_key, val) => {
        if (val && typeof val === 'object' && !Array.isArray(val)) {
            return Object.keys(val)
                .sort()
                .reduce<Record<string, unknown>>((sorted, k) => {
                    sorted[k] = (val as Record<string, unknown>)[k];
                    return sorted;
                }, {});
        }
        return val;
    });
}

/**
 * Build the localStorage key for a server-draft fallback.
 */
function localDraftKey(
    campaignId: number,
    sourceTable: string,
    sourceId: number
): string {
    return `${LOCAL_DRAFT_PREFIX}${campaignId}_${sourceTable}_${sourceId}`;
}

/**
 * Hook for server-side draft persistence with debounced autosave,
 * localStorage fallback, and reliable save-on-unmount.
 *
 * @param options - Configuration for the draft hook.
 * @returns An object with the server draft, dirty state, save/delete
 *          functions, and loading state.
 *
 * @example
 * ```typescript
 * const {
 *     serverDraft,
 *     hasDraft,
 *     saveDraftToServer,
 *     deleteDraftFromServer,
 *     isDirty,
 *     lastSaved,
 *     isLoading,
 * } = useServerDraft({
 *     campaignId,
 *     sourceTable: 'entities',
 *     sourceId: entityId,
 *     isNew: false,
 *     committedData: entity,
 *     currentData: formData,
 *     serverVersion: entity?.version,
 * });
 * ```
 */
export function useServerDraft<T>({
    campaignId,
    sourceTable,
    sourceId,
    isNew,
    committedData,
    currentData,
    serverVersion,
    enabled = true,
}: UseServerDraftOptions<T>): UseServerDraftReturn<T> {
    const queryClient = useQueryClient();
    const [lastSaved, setLastSaved] = useState<string | null>(null);

    // Refs for unmount cleanup to avoid stale closures
    const currentDataRef = useRef(currentData);
    const committedDataRef = useRef(committedData);
    const campaignIdRef = useRef(campaignId);
    const sourceTableRef = useRef(sourceTable);
    const sourceIdRef = useRef(sourceId);
    const isNewRef = useRef(isNew);
    const serverVersionRef = useRef(serverVersion);
    const enabledRef = useRef(enabled);
    const isDirtyRef = useRef(false);

    // Keep refs current
    useEffect(() => { currentDataRef.current = currentData; }, [currentData]);
    useEffect(() => { committedDataRef.current = committedData; }, [committedData]);
    useEffect(() => { campaignIdRef.current = campaignId; }, [campaignId]);
    useEffect(() => { sourceTableRef.current = sourceTable; }, [sourceTable]);
    useEffect(() => { sourceIdRef.current = sourceId; }, [sourceId]);
    useEffect(() => { isNewRef.current = isNew; }, [isNew]);
    useEffect(() => { serverVersionRef.current = serverVersion; }, [serverVersion]);
    useEffect(() => { enabledRef.current = enabled; }, [enabled]);

    // Compute dirty state via stable JSON comparison
    const committedStr = useMemo(
        () => stableStringify(committedData),
        [committedData]
    );
    const currentStr = useMemo(
        () => stableStringify(currentData),
        [currentData]
    );
    const isDirty = committedStr !== currentStr;

    // Keep the dirty ref in sync
    useEffect(() => { isDirtyRef.current = isDirty; }, [isDirty]);

    // Query key for this specific draft
    const draftQueryKey = useMemo(
        () => ['drafts', campaignId, sourceTable, sourceId] as const,
        [campaignId, sourceTable, sourceId]
    );

    // Fetch the existing draft from the server
    const { data: draftResponse, isLoading } = useQuery({
        queryKey: draftQueryKey,
        queryFn: () => draftsApi.getDraft(campaignId, sourceTable, sourceId),
        enabled: enabled && (isNew || sourceId > 0),
        retry: false,
        // Do not refetch on window focus for drafts
        refetchOnWindowFocus: false,
    });

    const serverDraft = useMemo<T | null>(() => {
        if (draftResponse?.draftData) {
            return draftResponse.draftData as T;
        }
        return null;
    }, [draftResponse]);

    const hasDraft = serverDraft !== null;

    const draftUpdatedAt = draftResponse?.updatedAt ?? null;

    // Build the save input from current state
    const buildSaveInput = useCallback((): SaveDraftInput => {
        return {
            sourceTable: sourceTableRef.current,
            sourceId: sourceIdRef.current,
            isNew: isNewRef.current,
            draftData: currentDataRef.current as Record<string, unknown>,
            serverVersion: serverVersionRef.current,
        };
    }, []);

    // Mutation for saving a draft
    const saveMutation = useMutation({
        mutationFn: (input: SaveDraftInput) =>
            draftsApi.saveDraft(campaignIdRef.current, input),
        onSuccess: (data: Draft) => {
            setLastSaved(new Date().toISOString());
            // Update the draft query cache
            queryClient.setQueryData(draftQueryKey, data);
            // Invalidate the indicators cache
            queryClient.invalidateQueries({
                queryKey: ['draftIndicators', campaignIdRef.current],
            });
        },
    });

    // Mutation for deleting a draft
    const deleteMutation = useMutation({
        mutationFn: () =>
            draftsApi.deleteDraft(
                campaignIdRef.current,
                sourceTableRef.current,
                sourceIdRef.current
            ),
        onSuccess: () => {
            // Clear the draft from cache
            queryClient.removeQueries({ queryKey: draftQueryKey });
            // Invalidate indicators
            queryClient.invalidateQueries({
                queryKey: ['draftIndicators', campaignIdRef.current],
            });
            // Clear localStorage fallback
            const key = localDraftKey(
                campaignIdRef.current,
                sourceTableRef.current,
                sourceIdRef.current
            );
            try {
                localStorage.removeItem(key);
            } catch {
                // Ignore localStorage errors
            }
        },
    });

    /**
     * Manually trigger a save to the server.
     */
    const saveDraftToServer = useCallback(async () => {
        if (!enabledRef.current) return;
        const input = buildSaveInput();
        await saveMutation.mutateAsync(input);
    }, [buildSaveInput, saveMutation]);

    /**
     * Delete the draft from the server.
     */
    const deleteDraftFromServer = useCallback(async () => {
        await deleteMutation.mutateAsync();
    }, [deleteMutation]);

    // Debounced autosave: schedule a server save when dirty
    useEffect(() => {
        if (!enabled || !isDirty) {
            return;
        }

        const timer = setTimeout(() => {
            const input = buildSaveInput();
            saveMutation.mutate(input);

            // Also save to localStorage as fallback
            const key = localDraftKey(campaignId, sourceTable, sourceId);
            try {
                localStorage.setItem(key, JSON.stringify({
                    data: currentDataRef.current,
                    savedAt: new Date().toISOString(),
                    serverVersion: serverVersionRef.current,
                }));
            } catch {
                // Ignore localStorage errors
            }
        }, AUTOSAVE_DEBOUNCE_MS);

        return () => clearTimeout(timer);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [currentStr, enabled, isDirty, campaignId, sourceTable, sourceId]);

    // Save on unmount using fetch with keepalive
    useEffect(() => {
        return () => {
            if (!enabledRef.current || !isDirtyRef.current) {
                return;
            }

            const input: SaveDraftInput = {
                sourceTable: sourceTableRef.current,
                sourceId: sourceIdRef.current,
                isNew: isNewRef.current,
                draftData: currentDataRef.current as Record<string, unknown>,
                serverVersion: serverVersionRef.current,
            };

            // Use the beacon endpoint for reliable delivery
            draftsApi.beaconSave(campaignIdRef.current, input);

            // Also save to localStorage as final fallback
            const key = localDraftKey(
                campaignIdRef.current,
                sourceTableRef.current,
                sourceIdRef.current
            );
            try {
                localStorage.setItem(key, JSON.stringify({
                    data: currentDataRef.current,
                    savedAt: new Date().toISOString(),
                    serverVersion: serverVersionRef.current,
                }));
            } catch {
                // Ignore localStorage errors
            }
        };
    }, []);

    return {
        serverDraft,
        hasDraft,
        saveDraftToServer,
        deleteDraftFromServer,
        isDirty,
        lastSaved,
        isLoading,
        draftUpdatedAt,
    };
}
