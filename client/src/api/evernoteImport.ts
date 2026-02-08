// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Evernote Local Import API service - imports notes from the local Evernote
 * application on MacOS via AppleScript automation.
 */

import { apiClient } from './client';
import type { ImportResult } from '../types';

/**
 * Status of the local Evernote application.
 */
export interface EvernoteStatus {
    available: boolean;
    error?: string;
}

/**
 * A notebook from the local Evernote application.
 */
export interface Notebook {
    name: string;
    noteCount: number;
}

/**
 * Summary of a note from the local Evernote application.
 */
export interface NoteSummary {
    noteLink: string;
    title: string;
    created: string;
    modified: string;
    tags: string[];
}

/**
 * Request body for importing notes from a notebook.
 */
export interface EvernoteLocalImportRequest {
    campaignId: number;
    notebookName: string;
    autoDetectEntities: boolean;
}

/**
 * Evernote Local Import API service.
 */
export const evernoteImportApi = {
    /**
     * Check if the local Evernote application is available and running.
     */
    getStatus(): Promise<EvernoteStatus> {
        return apiClient.get<EvernoteStatus>('/import/evernote/status');
    },

    /**
     * List all notebooks in the local Evernote application.
     */
    listNotebooks(): Promise<Notebook[]> {
        return apiClient.get<Notebook[]>('/import/evernote/notebooks');
    },

    /**
     * List all notes in a specific notebook.
     */
    listNotes(notebookName: string): Promise<NoteSummary[]> {
        const encodedName = encodeURIComponent(notebookName);
        return apiClient.get<NoteSummary[]>(
            `/import/evernote/notebooks/${encodedName}/notes`
        );
    },

    /**
     * Import all notes from a notebook into a campaign.
     */
    importNotebook(request: EvernoteLocalImportRequest): Promise<ImportResult> {
        return apiClient.post<ImportResult>(
            '/import/evernote/import',
            request
        );
    },
};

export default evernoteImportApi;
