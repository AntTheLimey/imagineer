// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * React Query hooks for user settings management.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userSettingsApi, UserSettingsUpdateRequest, UserSettingsResponse } from '../api/userSettings';

/**
 * Query keys for user settings queries.
 */
export const userSettingsKeys = {
    all: ['userSettings'] as const,
    detail: () => [...userSettingsKeys.all, 'detail'] as const,
};

/**
 * Hook to fetch the current user's settings.
 */
export function useUserSettings() {
    return useQuery({
        queryKey: userSettingsKeys.detail(),
        queryFn: () => userSettingsApi.get(),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
}

/**
 * Hook to update the current user's settings.
 */
export function useUpdateUserSettings() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (input: UserSettingsUpdateRequest) => userSettingsApi.update(input),
        onSuccess: (data) => {
            // Update the cache with the new settings
            queryClient.setQueryData<UserSettingsResponse>(
                userSettingsKeys.detail(),
                data
            );
        },
    });
}
