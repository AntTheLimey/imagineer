// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Import API service - import operations for Evernote, Google Docs, and file uploads.
 */

import { apiClient } from './client';
import type { ImportResult } from '../types';

/**
 * Input for Google Docs import.
 */
export interface GoogleDocsImportInput {
    campaignId: string;
    documentUrl: string;
}

/**
 * Input for file upload import.
 */
export interface FileUploadImportInput {
    campaignId: string;
    file: File;
}

/**
 * Import API service.
 */
export const importApi = {
    /**
     * Import content from an Evernote .enex file.
     */
    async importEvernote(campaignId: string, file: File): Promise<ImportResult> {
        return apiClient.upload<ImportResult>(
            `/campaigns/${campaignId}/import/evernote`,
            file
        );
    },

    /**
     * Import content from a Google Docs document.
     */
    importGoogleDocs(input: GoogleDocsImportInput): Promise<ImportResult> {
        const { campaignId, documentUrl } = input;
        return apiClient.post<ImportResult>(
            `/campaigns/${campaignId}/import/google-docs`,
            { documentUrl }
        );
    },

    /**
     * Import content from uploaded files (text, markdown, docx, pdf).
     */
    async importFile(campaignId: string, file: File): Promise<ImportResult> {
        return apiClient.upload<ImportResult>(
            `/campaigns/${campaignId}/import/file`,
            file
        );
    },

    /**
     * Import multiple files at once.
     */
    async importFiles(campaignId: string, files: File[]): Promise<ImportResult[]> {
        const results = await Promise.all(
            files.map((file) => importApi.importFile(campaignId, file))
        );
        return results;
    },
};

export default importApi;
