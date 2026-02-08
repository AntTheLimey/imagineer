// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for game system operations.
 */

import { useQuery } from '@tanstack/react-query';
import { gameSystemsApi } from '../api';

/**
 * Query keys for game system-related queries.
 */
export const gameSystemKeys = {
    all: ['gameSystems'] as const,
    lists: () => [...gameSystemKeys.all, 'list'] as const,
    details: () => [...gameSystemKeys.all, 'detail'] as const,
    detail: (id: number) => [...gameSystemKeys.details(), id] as const,
    byCode: (code: string) => [...gameSystemKeys.all, 'code', code] as const,
};

/**
 * Hook to fetch list of all available game systems.
 */
export function useGameSystems() {
    return useQuery({
        queryKey: gameSystemKeys.lists(),
        queryFn: () => gameSystemsApi.list(),
        // Game systems rarely change, so use longer stale time
        staleTime: 30 * 60 * 1000, // 30 minutes
    });
}

/**
 * Hook to fetch a single game system by ID.
 */
export function useGameSystem(id: number, options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: gameSystemKeys.detail(id),
        queryFn: () => gameSystemsApi.get(id),
        enabled: options?.enabled ?? !!id,
        staleTime: 30 * 60 * 1000,
    });
}

/**
 * Hook to fetch a game system by its code (e.g., 'coc-7e').
 */
export function useGameSystemByCode(code: string, options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: gameSystemKeys.byCode(code),
        queryFn: () => gameSystemsApi.getByCode(code),
        enabled: options?.enabled ?? !!code,
        staleTime: 30 * 60 * 1000,
    });
}
