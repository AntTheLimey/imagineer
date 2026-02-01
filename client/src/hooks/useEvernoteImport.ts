// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for Evernote local import operations.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    evernoteImportApi,
    EvernoteStatus,
    Notebook,
    NoteSummary,
    EvernoteLocalImportRequest,
} from '../api/evernoteImport';
import type { ImportResult } from '../types';
import { entityKeys } from './useEntities';
import { relationshipKeys } from './useRelationships';
import { timelineKeys } from './useTimeline';
import { statsKeys } from './useStats';

/**
 * Query keys for Evernote import queries.
 */
export const evernoteImportKeys = {
    all: ['evernoteImport'] as const,
    status: () => [...evernoteImportKeys.all, 'status'] as const,
    notebooks: () => [...evernoteImportKeys.all, 'notebooks'] as const,
    notes: (notebookName: string) =>
        [...evernoteImportKeys.all, 'notes', notebookName] as const,
};

/**
 * Hook to check if the local Evernote application is available.
 */
export function useEvernoteStatus(options?: { enabled?: boolean }) {
    return useQuery<EvernoteStatus>({
        queryKey: evernoteImportKeys.status(),
        queryFn: () => evernoteImportApi.getStatus(),
        staleTime: 30 * 1000, // 30 seconds
        retry: false,
        ...options,
    });
}

/**
 * Hook to list notebooks from the local Evernote application.
 */
export function useEvernoteNotebooks(options?: { enabled?: boolean }) {
    return useQuery<Notebook[]>({
        queryKey: evernoteImportKeys.notebooks(),
        queryFn: () => evernoteImportApi.listNotebooks(),
        staleTime: 60 * 1000, // 1 minute
        ...options,
    });
}

/**
 * Hook to list notes in a specific notebook.
 */
export function useEvernoteNotes(
    notebookName: string,
    options?: { enabled?: boolean }
) {
    return useQuery<NoteSummary[]>({
        queryKey: evernoteImportKeys.notes(notebookName),
        queryFn: () => evernoteImportApi.listNotes(notebookName),
        staleTime: 60 * 1000, // 1 minute
        ...options,
        // Combine notebookName guard with any external enabled option
        enabled: !!notebookName && (options?.enabled ?? true),
    });
}

/**
 * Hook to import notes from a notebook into a campaign.
 */
export function useImportEvernoteLocal() {
    const queryClient = useQueryClient();

    return useMutation<ImportResult, Error, EvernoteLocalImportRequest>({
        mutationFn: (request: EvernoteLocalImportRequest) =>
            evernoteImportApi.importNotebook(request),
        onSuccess: (data: ImportResult, variables: EvernoteLocalImportRequest) => {
            if (data.success) {
                // Invalidate all related queries for this campaign
                invalidateCampaignData(queryClient, variables.campaignId);
            }
        },
    });
}

/**
 * Helper to invalidate all campaign-related data after import.
 */
function invalidateCampaignData(
    queryClient: ReturnType<typeof useQueryClient>,
    campaignId: string
) {
    // Invalidate entities
    queryClient.invalidateQueries({
        queryKey: [...entityKeys.lists()],
        predicate: (query) => {
            const key = query.queryKey as unknown[];
            if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                const params = key[2] as { campaignId?: string };
                return params.campaignId === campaignId;
            }
            return false;
        },
    });

    // Invalidate relationships
    queryClient.invalidateQueries({
        queryKey: [...relationshipKeys.lists()],
        predicate: (query) => {
            const key = query.queryKey as unknown[];
            if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                const params = key[2] as { campaignId?: string };
                return params.campaignId === campaignId;
            }
            return false;
        },
    });

    // Invalidate timeline events
    queryClient.invalidateQueries({
        queryKey: [...timelineKeys.lists()],
        predicate: (query) => {
            const key = query.queryKey as unknown[];
            if (key.length >= 3 && typeof key[2] === 'object' && key[2] !== null) {
                const params = key[2] as { campaignId?: string };
                return params.campaignId === campaignId;
            }
            return false;
        },
    });

    // Invalidate stats
    queryClient.invalidateQueries({
        queryKey: statsKeys.dashboard(),
    });
    queryClient.invalidateQueries({
        queryKey: statsKeys.campaign(campaignId),
    });
}
