// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Hook for automatic draft saving at regular intervals.
 *
 * Automatically saves data to localStorage drafts at a configurable
 * interval (default 30 seconds). Tracks the last save time and can
 * be enabled/disabled dynamically.
 */

import { useEffect, useRef, useState, useCallback, MutableRefObject } from 'react';
import { useDraft } from './useDraft';

const DEFAULT_INTERVAL = 30000; // 30 seconds

/**
 * Options for the useAutosave hook.
 */
interface UseAutosaveOptions<T> {
    /** The data to autosave */
    data: T;
    /** The key to store the draft under */
    key: string;
    /** Autosave interval in milliseconds (default: 30000) */
    interval?: number;
    /** Whether autosave is enabled (default: true) */
    enabled?: boolean;
    /** Server version for conflict detection */
    serverVersion?: number;
    /** Callback when autosave occurs */
    onSave?: () => void;
}

/**
 * Return type for the useAutosave hook.
 */
interface UseAutosaveReturn {
    /** ISO timestamp of the last autosave, or null if never saved */
    lastSaved: string | null;
    /** Manually trigger a save */
    saveNow: () => void;
}

/**
 * Hook for automatic draft saving at regular intervals.
 *
 * @param options - Configuration options for autosave behavior
 * @returns Object containing last save timestamp and manual save function
 *
 * @example
 * ```typescript
 * const { lastSaved, saveNow } = useAutosave({
 *   data: formData,
 *   key: `entity-${entityId}`,
 *   interval: 30000,
 *   enabled: isDirty,
 *   serverVersion: entity.version,
 * });
 *
 * // Display last save time to user
 * {lastSaved && <span>Saved {formatRelativeTime(lastSaved)}</span>}
 *
 * // Manual save on blur
 * onBlur={() => saveNow()}
 * ```
 */
export function useAutosave<T>({
    data,
    key,
    interval = DEFAULT_INTERVAL,
    enabled = true,
    serverVersion,
    onSave,
}: UseAutosaveOptions<T>): UseAutosaveReturn {
    const { saveDraft } = useDraft();
    const [lastSaved, setLastSaved] = useState<string | null>(null);

    // Use refs to avoid stale closures in the interval
    const dataRef = useRef(data);
    const serverVersionRef = useRef(serverVersion);
    const onSaveRef = useRef(onSave);

    // Refs for enabled and key to allow unmount-only effect
    const enabledRef: MutableRefObject<boolean> = useRef(enabled);
    const keyRef: MutableRefObject<string> = useRef(key);

    // Keep refs updated
    useEffect(() => {
        dataRef.current = data;
    }, [data]);

    useEffect(() => {
        serverVersionRef.current = serverVersion;
    }, [serverVersion]);

    useEffect(() => {
        onSaveRef.current = onSave;
    }, [onSave]);

    useEffect(() => {
        enabledRef.current = enabled;
    }, [enabled]);

    useEffect(() => {
        keyRef.current = key;
    }, [key]);

    /**
     * Perform the save operation.
     */
    const performSave = useCallback(() => {
        saveDraft(key, dataRef.current, serverVersionRef.current);
        const now = new Date().toISOString();
        setLastSaved(now);
        onSaveRef.current?.();
    }, [key, saveDraft]);

    /**
     * Manually trigger a save.
     */
    const saveNow = useCallback(() => {
        if (enabled) {
            performSave();
        }
    }, [enabled, performSave]);

    // Set up the autosave interval
    useEffect(() => {
        if (!enabled) {
            return;
        }

        const intervalId = setInterval(() => {
            performSave();
        }, interval);

        return () => {
            clearInterval(intervalId);
        };
    }, [enabled, interval, performSave]);

    // Save on unmount if enabled (uses refs to ensure this only runs on actual unmount)
    useEffect(() => {
        return () => {
            if (enabledRef.current) {
                // Note: We use the refs here since the cleanup runs with
                // potentially stale closure values
                saveDraft(keyRef.current, dataRef.current, serverVersionRef.current);
            }
        };
    }, [saveDraft]);

    return {
        lastSaved,
        saveNow,
    };
}
