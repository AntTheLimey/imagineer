// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * User Settings API service - CRUD operations for user settings including
 * LLM API key configuration.
 */

import { apiClient } from './client';

/**
 * Content generation service options.
 */
export type ContentGenService = 'anthropic' | 'openai' | 'gemini';

/**
 * Embedding service options.
 */
export type EmbeddingService = 'ollama' | 'voyage' | 'openai' | 'gemini';

/**
 * Image generation service options.
 */
export type ImageGenService = 'openai' | 'stability';

/**
 * User settings response from the API.
 * API keys are masked, showing only last 4 characters.
 */
export interface UserSettingsResponse {
    contentGenService: ContentGenService | null;
    contentGenApiKey: string | null;
    embeddingService: EmbeddingService | null;
    embeddingApiKey: string | null;
    imageGenService: ImageGenService | null;
    imageGenApiKey: string | null;
}

/**
 * Request body for updating user settings.
 * Only include fields that should be updated.
 * Do not send API key fields if they still contain masked values.
 */
export interface UserSettingsUpdateRequest {
    contentGenService?: ContentGenService | null;
    contentGenApiKey?: string;
    embeddingService?: EmbeddingService | null;
    embeddingApiKey?: string;
    imageGenService?: ImageGenService | null;
    imageGenApiKey?: string;
}

/**
 * User Settings API service.
 */
export const userSettingsApi = {
    /**
     * Get the current user's settings.
     * API keys in the response are masked (showing only last 4 chars).
     */
    get(): Promise<UserSettingsResponse> {
        return apiClient.get<UserSettingsResponse>('/user/settings');
    },

    /**
     * Update the current user's settings.
     * Only send API key fields if the user has changed them from the masked value.
     */
    update(input: UserSettingsUpdateRequest): Promise<UserSettingsResponse> {
        return apiClient.put<UserSettingsResponse>('/user/settings', input);
    },
};

export default userSettingsApi;
