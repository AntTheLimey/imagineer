// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for import operations.
 */

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { importApi, GoogleDocsImportInput } from '../api';
import type { ImportResult } from '../types';
import { entityKeys } from './useEntities';
import { relationshipKeys } from './useRelationships';
import { timelineKeys } from './useTimeline';
import { statsKeys } from './useStats';

/**
 * Hook to import content from an Evernote .enex file.
 */
export function useImportEvernote() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ campaignId, file }: { campaignId: string; file: File }) =>
            importApi.importEvernote(campaignId, file),
        onSuccess: (data: ImportResult, variables: { campaignId: string; file: File }) => {
            if (data.success) {
                // Invalidate all related queries for this campaign
                invalidateCampaignData(queryClient, variables.campaignId);
            }
        },
    });
}

/**
 * Hook to import content from a Google Docs document.
 */
export function useImportGoogleDocs() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: GoogleDocsImportInput) =>
            importApi.importGoogleDocs(input),
        onSuccess: (data: ImportResult, variables: GoogleDocsImportInput) => {
            if (data.success) {
                // Invalidate all related queries for this campaign
                invalidateCampaignData(queryClient, variables.campaignId);
            }
        },
    });
}

/**
 * Hook to import content from uploaded files.
 */
export function useImportFile() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ campaignId, file }: { campaignId: string; file: File }) =>
            importApi.importFile(campaignId, file),
        onSuccess: (data: ImportResult, variables: { campaignId: string; file: File }) => {
            if (data.success) {
                // Invalidate all related queries for this campaign
                invalidateCampaignData(queryClient, variables.campaignId);
            }
        },
    });
}

/**
 * Hook to import multiple files at once.
 */
export function useImportFiles() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ campaignId, files }: { campaignId: string; files: File[] }) =>
            importApi.importFiles(campaignId, files),
        onSuccess: (
            data: ImportResult[],
            variables: { campaignId: string; files: File[] }
        ) => {
            // Check if any import was successful
            const anySuccess = data.some((result) => result.success);
            if (anySuccess) {
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
