// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import { useCampaignOwnership, isCampaignOwner } from './useCampaignOwnership';
import * as AuthContext from '../contexts/AuthContext';
import * as CampaignsHooks from './useCampaigns';
import type { Campaign } from '../types';

// Mock the auth context
vi.mock('../contexts/AuthContext', () => ({
    useAuth: vi.fn(),
}));

// Mock the useCampaign hook
vi.mock('./useCampaigns', () => ({
    useCampaign: vi.fn(),
}));

describe('useCampaignOwnership', () => {
    const mockUseAuth = AuthContext.useAuth as ReturnType<typeof vi.fn>;
    const mockUseCampaign = CampaignsHooks.useCampaign as ReturnType<typeof vi.fn>;

    const mockCampaign: Campaign = {
        id: 1,
        name: 'Test Campaign',
        systemId: 1,
        ownerId: 123,
        description: 'A test campaign',
        settings: {},
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    };

    const mockUser = {
        id: '123',
        email: 'test@example.com',
        name: 'Test User',
    };

    let queryClient: QueryClient;

    const createWrapper = () => {
        return function Wrapper({ children }: { children: ReactNode }) {
            return (
                <QueryClientProvider client={queryClient}>
                    {children}
                </QueryClientProvider>
            );
        };
    };

    beforeEach(() => {
        vi.clearAllMocks();
        queryClient = new QueryClient({
            defaultOptions: {
                queries: {
                    retry: false,
                },
            },
        });
    });

    it('should return isOwner true when user is campaign owner', async () => {
        mockUseAuth.mockReturnValue({
            user: mockUser,
            isAuthenticated: true,
            token: 'test-token',
            isLoading: false,
            login: vi.fn(),
            logout: vi.fn(),
        });

        mockUseCampaign.mockReturnValue({
            data: mockCampaign,
            isLoading: false,
            error: null,
        });

        const { result } = renderHook(
            () => useCampaignOwnership(1),
            { wrapper: createWrapper() }
        );

        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        expect(result.current.isOwner).toBe(true);
        expect(result.current.campaign).toEqual(mockCampaign);
    });

    it('should return isOwner false when user is not campaign owner', async () => {
        mockUseAuth.mockReturnValue({
            user: { ...mockUser, id: '999' },
            isAuthenticated: true,
            token: 'test-token',
            isLoading: false,
            login: vi.fn(),
            logout: vi.fn(),
        });

        mockUseCampaign.mockReturnValue({
            data: mockCampaign,
            isLoading: false,
            error: null,
        });

        const { result } = renderHook(
            () => useCampaignOwnership(1),
            { wrapper: createWrapper() }
        );

        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        expect(result.current.isOwner).toBe(false);
    });

    it('should return isOwner false when user is not authenticated', async () => {
        mockUseAuth.mockReturnValue({
            user: null,
            isAuthenticated: false,
            token: null,
            isLoading: false,
            login: vi.fn(),
            logout: vi.fn(),
        });

        mockUseCampaign.mockReturnValue({
            data: mockCampaign,
            isLoading: false,
            error: null,
        });

        const { result } = renderHook(
            () => useCampaignOwnership(1),
            { wrapper: createWrapper() }
        );

        expect(result.current.isOwner).toBe(false);
    });

    it('should return isLoading true while campaign is loading', async () => {
        mockUseAuth.mockReturnValue({
            user: mockUser,
            isAuthenticated: true,
            token: 'test-token',
            isLoading: false,
            login: vi.fn(),
            logout: vi.fn(),
        });

        mockUseCampaign.mockReturnValue({
            data: undefined,
            isLoading: true,
            error: null,
        });

        const { result } = renderHook(
            () => useCampaignOwnership(1),
            { wrapper: createWrapper() }
        );

        expect(result.current.isLoading).toBe(true);
        expect(result.current.isOwner).toBe(false);
    });

    it('should return isOwner false when campaign has no ownerId', async () => {
        const campaignWithoutOwner: Campaign = {
            ...mockCampaign,
            ownerId: undefined,
        };

        mockUseAuth.mockReturnValue({
            user: mockUser,
            isAuthenticated: true,
            token: 'test-token',
            isLoading: false,
            login: vi.fn(),
            logout: vi.fn(),
        });

        mockUseCampaign.mockReturnValue({
            data: campaignWithoutOwner,
            isLoading: false,
            error: null,
        });

        const { result } = renderHook(
            () => useCampaignOwnership(1),
            { wrapper: createWrapper() }
        );

        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        expect(result.current.isOwner).toBe(false);
    });
});

describe('isCampaignOwner', () => {
    const mockCampaign: Campaign = {
        id: 1,
        name: 'Test Campaign',
        systemId: 1,
        ownerId: 123,
        description: 'A test campaign',
        settings: {},
        createdAt: '2025-01-01T00:00:00Z',
        updatedAt: '2025-01-01T00:00:00Z',
    };

    it('should return true when userId matches campaign ownerId', () => {
        expect(isCampaignOwner(123, mockCampaign)).toBe(true);
    });

    it('should return false when userId does not match campaign ownerId', () => {
        expect(isCampaignOwner(999, mockCampaign)).toBe(false);
    });

    it('should return false when userId is undefined', () => {
        expect(isCampaignOwner(undefined, mockCampaign)).toBe(false);
    });

    it('should return false when campaign is undefined', () => {
        expect(isCampaignOwner(123, undefined)).toBe(false);
    });

    it('should return false when campaign has no ownerId', () => {
        const campaignWithoutOwner: Campaign = {
            ...mockCampaign,
            ownerId: undefined,
        };
        expect(isCampaignOwner(123, campaignWithoutOwner)).toBe(false);
    });
});
