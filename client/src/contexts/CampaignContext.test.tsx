// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import { CampaignProvider, useCampaignContext } from './CampaignContext';

// Mock the useCampaign hook
vi.mock('../hooks', () => ({
    useCampaign: vi.fn((id: string, options?: { enabled?: boolean }) => {
        if (!options?.enabled || !id) {
            return { data: null, isLoading: false, error: null };
        }
        return {
            data: { id, name: `Campaign ${id}`, systemId: 'coc7' },
            isLoading: false,
            error: null,
        };
    }),
}));

const STORAGE_KEY = 'imagineer_current_campaign_id';

function createWrapper() {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: { retry: false },
        },
    });

    return function Wrapper({ children }: { children: ReactNode }) {
        return (
            <QueryClientProvider client={queryClient}>
                <CampaignProvider>{children}</CampaignProvider>
            </QueryClientProvider>
        );
    };
}

describe('CampaignContext', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.clearAllMocks();
    });

    describe('initialization', () => {
        it('should initialize with null campaign when nothing in localStorage', () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            expect(result.current.currentCampaignId).toBeNull();
            expect(result.current.currentCampaign).toBeNull();
        });

        it('should initialize from localStorage if campaign ID is stored', () => {
            localStorage.setItem(STORAGE_KEY, 'stored-campaign-id');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            expect(result.current.currentCampaignId).toBe('stored-campaign-id');
        });
    });

    describe('setCurrentCampaignId', () => {
        it('should update current campaign ID', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            act(() => {
                result.current.setCurrentCampaignId('new-campaign-id');
            });

            expect(result.current.currentCampaignId).toBe('new-campaign-id');
        });

        it('should persist campaign ID to localStorage', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            act(() => {
                result.current.setCurrentCampaignId('persisted-id');
            });

            await waitFor(() => {
                expect(localStorage.getItem(STORAGE_KEY)).toBe('persisted-id');
            });
        });
    });

    describe('clearCurrentCampaign', () => {
        it('should clear the current campaign', async () => {
            localStorage.setItem(STORAGE_KEY, 'existing-id');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            act(() => {
                result.current.clearCurrentCampaign();
            });

            expect(result.current.currentCampaignId).toBeNull();
        });

        it('should remove campaign ID from localStorage', async () => {
            localStorage.setItem(STORAGE_KEY, 'to-remove');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            act(() => {
                result.current.clearCurrentCampaign();
            });

            await waitFor(() => {
                expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
            });
        });
    });

    describe('error handling', () => {
        it('should throw error when used outside provider', () => {
            // Suppress console.error for this test
            const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            expect(() => {
                renderHook(() => useCampaignContext());
            }).toThrow('useCampaignContext must be used within a CampaignProvider');

            consoleSpy.mockRestore();
        });
    });
});
