// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import AuthCallback from './AuthCallback';

// Mock react-router-dom's useNavigate
const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

// Mock the AuthContext
const mockLogin = vi.fn();

vi.mock('../contexts/AuthContext', () => ({
    useAuth: () => ({
        login: mockLogin,
    }),
}));

const renderAuthCallback = (search: string = '') => {
    return render(
        <MemoryRouter initialEntries={[`/auth/callback${search}`]}>
            <AuthCallback />
        </MemoryRouter>
    );
};

describe('AuthCallback', () => {
    beforeEach(() => {
        mockNavigate.mockClear();
        mockLogin.mockClear();
    });

    it('should display error when no params are provided', async () => {
        renderAuthCallback();

        // Without params, it shows missing token error
        await waitFor(() => {
            expect(
                screen.getByText('Missing authentication token')
            ).toBeInTheDocument();
        });

        expect(screen.getByText('Return to login')).toBeInTheDocument();
    });

    it('should display error when error parameter is present', async () => {
        renderAuthCallback('?error=Authentication%20failed');

        await waitFor(() => {
            expect(screen.getByText('Authentication failed')).toBeInTheDocument();
        });

        expect(screen.getByText('Return to login')).toBeInTheDocument();
    });

    it('should display error when token is missing', async () => {
        const user = encodeURIComponent(
            JSON.stringify({ id: '123', email: 'test@example.com', name: 'Test' })
        );
        renderAuthCallback(`?user=${user}`);

        await waitFor(() => {
            expect(
                screen.getByText('Missing authentication token')
            ).toBeInTheDocument();
        });
    });

    it('should display error when user is missing', async () => {
        renderAuthCallback('?token=test-token');

        await waitFor(() => {
            expect(
                screen.getByText('Missing user information')
            ).toBeInTheDocument();
        });
    });

    it('should display error when user data is invalid JSON', async () => {
        renderAuthCallback('?token=test-token&user=invalid-json');

        await waitFor(() => {
            expect(screen.getByText('Invalid user data')).toBeInTheDocument();
        });
    });

    it('should display error when user is missing required fields', async () => {
        const user = encodeURIComponent(JSON.stringify({ id: '123' }));
        renderAuthCallback(`?token=test-token&user=${user}`);

        await waitFor(() => {
            expect(
                screen.getByText('Incomplete user information')
            ).toBeInTheDocument();
        });
    });

    it('should call login and navigate on successful auth', async () => {
        const userData = {
            id: '123',
            email: 'test@example.com',
            name: 'Test User',
        };
        const user = encodeURIComponent(JSON.stringify(userData));

        renderAuthCallback(`?token=test-token&user=${user}`);

        await waitFor(() => {
            expect(mockLogin).toHaveBeenCalledWith('test-token', userData);
        });

        expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true });
    });

    it('should handle user with optional avatarUrl', async () => {
        const userData = {
            id: '123',
            email: 'test@example.com',
            name: 'Test User',
            avatarUrl: 'https://example.com/avatar.jpg',
        };
        const user = encodeURIComponent(JSON.stringify(userData));

        renderAuthCallback(`?token=test-token&user=${user}`);

        await waitFor(() => {
            expect(mockLogin).toHaveBeenCalledWith('test-token', userData);
        });
    });
});
