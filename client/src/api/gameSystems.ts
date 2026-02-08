// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Game Systems API service - list available game systems.
 */

import { apiClient } from './client';
import type { GameSystem } from '../types';

/**
 * Game Systems API service.
 */
export const gameSystemsApi = {
    /**
     * List all available game systems.
     */
    list(): Promise<GameSystem[]> {
        return apiClient.get<GameSystem[]>('/game-systems');
    },

    /**
     * Get a single game system by ID.
     */
    get(id: number): Promise<GameSystem> {
        return apiClient.get<GameSystem>(`/game-systems/${id}`);
    },

    /**
     * Get a game system by code (e.g., 'coc-7e', 'gurps-4e', 'fitd').
     */
    getByCode(code: string): Promise<GameSystem> {
        return apiClient.get<GameSystem>(`/game-systems/code/${code}`);
    },
};

export default gameSystemsApi;
