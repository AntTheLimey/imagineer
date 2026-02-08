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

// Mock campaigns data for tests
const mockCampaigns = [
    { id: 1, name: 'Campaign 1', systemId: 1, createdAt: '2024-01-01' },
    { id: 2, name: 'Campaign 2', systemId: 1, createdAt: '2024-01-02' },
    { id: 3, name: 'Stored Campaign', systemId: 1, createdAt: '2024-01-03' },
    { id: 4, name: 'Existing Campaign', systemId: 1, createdAt: '2024-01-04' },
    { id: 5, name: 'To Remove', systemId: 1, createdAt: '2024-01-05' },
];

// Mock the hooks
vi.mock('../hooks', () => ({
    useCampaign: vi.fn((id: number, options?: { enabled?: boolean }) => {
        if (!options?.enabled || !id) {
            return { data: null, isLoading: false, error: null };
        }
        return {
            data: { id, name: `Campaign ${id}`, systemId: 1 },
            isLoading: false,
            error: null,
        };
    }),
    useCampaigns: vi.fn(() => ({
        data: mockCampaigns,
        isLoading: false,
        error: null,
    })),
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
        it('should initialize with null campaign when nothing in localStorage', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization to complete
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            expect(result.current.currentCampaignId).toBeNull();
            expect(result.current.currentCampaign).toBeNull();
        });

        it('should initialize from localStorage if campaign ID is stored and valid', async () => {
            // Use a campaign ID that exists in mockCampaigns
            localStorage.setItem(STORAGE_KEY, '3');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization to complete
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            expect(result.current.currentCampaignId).toBe(3);
        });

        it('should clear invalid campaign ID from localStorage', async () => {
            // Use a campaign ID that does NOT exist in mockCampaigns
            localStorage.setItem(STORAGE_KEY, '999');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization to complete
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            // Invalid ID should be cleared
            expect(result.current.currentCampaignId).toBeNull();
            expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
        });
    });

    describe('setCurrentCampaignId', () => {
        it('should update current campaign ID', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            act(() => {
                result.current.setCurrentCampaignId(1);
            });

            expect(result.current.currentCampaignId).toBe(1);
        });

        it('should persist campaign ID to localStorage', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            act(() => {
                result.current.setCurrentCampaignId(2);
            });

            await waitFor(() => {
                expect(localStorage.getItem(STORAGE_KEY)).toBe('2');
            });
        });
    });

    describe('clearCurrentCampaign', () => {
        it('should clear the current campaign', async () => {
            localStorage.setItem(STORAGE_KEY, '4');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            act(() => {
                result.current.clearCurrentCampaign();
            });

            expect(result.current.currentCampaignId).toBeNull();
        });

        it('should remove campaign ID from localStorage', async () => {
            localStorage.setItem(STORAGE_KEY, '5');

            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            act(() => {
                result.current.clearCurrentCampaign();
            });

            await waitFor(() => {
                expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
            });
        });
    });

    describe('getLatestCampaign', () => {
        it('should return the most recently created campaign', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            const latest = result.current.getLatestCampaign();
            // id=5 (To Remove) has the latest createdAt date in mockCampaigns
            expect(latest?.id).toBe(5);
        });

        it('should return latest campaign when campaigns exist', async () => {
            const { result } = renderHook(() => useCampaignContext(), {
                wrapper: createWrapper(),
            });

            // Wait for initialization
            await waitFor(() => {
                expect(result.current.isInitializing).toBe(false);
            });

            // The getLatestCampaign function handles undefined/empty arrays
            // Testing that it returns a campaign when campaigns exist (which it does)
            const latest = result.current.getLatestCampaign();
            expect(latest).not.toBeNull();
            // Verify it returns the most recent one (id=5 has latest createdAt)
            expect(latest?.id).toBe(5);
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
