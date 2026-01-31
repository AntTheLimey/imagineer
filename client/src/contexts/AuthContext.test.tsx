// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AuthProvider, useAuth } from './AuthContext';
import { getStoredToken, clearStoredAuth } from './authUtils';

// Test component to access auth context
function TestComponent() {
    const { user, token, isAuthenticated, isLoading, login, logout } = useAuth();

    return (
        <div>
            <div data-testid="loading">{isLoading ? 'loading' : 'ready'}</div>
            <div data-testid="authenticated">
                {isAuthenticated ? 'authenticated' : 'not-authenticated'}
            </div>
            <div data-testid="user">{user ? user.name : 'no-user'}</div>
            <div data-testid="token">{token || 'no-token'}</div>
            <button
                onClick={() =>
                    login('test-token', {
                        id: '123',
                        email: 'test@example.com',
                        name: 'Test User',
                    })
                }
            >
                Login
            </button>
            <button onClick={logout}>Logout</button>
        </div>
    );
}

describe('AuthContext', () => {
    beforeEach(() => {
        localStorage.clear();
    });

    afterEach(() => {
        localStorage.clear();
        vi.restoreAllMocks();
    });

    describe('AuthProvider', () => {
        it('should provide auth context to children', async () => {
            render(
                <AuthProvider>
                    <TestComponent />
                </AuthProvider>
            );

            await waitFor(() => {
                expect(screen.getByTestId('loading')).toHaveTextContent('ready');
            });

            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'not-authenticated'
            );
            expect(screen.getByTestId('user')).toHaveTextContent('no-user');
        });

        it('should load auth state from localStorage on mount', async () => {
            // Create a valid non-expired token (expires in 1 hour)
            const exp = Math.floor(Date.now() / 1000) + 3600;
            const payload = { exp };
            const token = `header.${btoa(JSON.stringify(payload))}.signature`;
            const user = {
                id: '123',
                email: 'test@example.com',
                name: 'Stored User',
            };

            localStorage.setItem('imagineer_token', token);
            localStorage.setItem('imagineer_user', JSON.stringify(user));

            render(
                <AuthProvider>
                    <TestComponent />
                </AuthProvider>
            );

            await waitFor(() => {
                expect(screen.getByTestId('loading')).toHaveTextContent('ready');
            });

            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'authenticated'
            );
            expect(screen.getByTestId('user')).toHaveTextContent('Stored User');
        });

        it('should clear expired tokens on mount', async () => {
            // Create an expired token (expired 1 hour ago)
            const exp = Math.floor(Date.now() / 1000) - 3600;
            const payload = { exp };
            const token = `header.${btoa(JSON.stringify(payload))}.signature`;
            const user = {
                id: '123',
                email: 'test@example.com',
                name: 'Stored User',
            };

            localStorage.setItem('imagineer_token', token);
            localStorage.setItem('imagineer_user', JSON.stringify(user));

            render(
                <AuthProvider>
                    <TestComponent />
                </AuthProvider>
            );

            await waitFor(() => {
                expect(screen.getByTestId('loading')).toHaveTextContent('ready');
            });

            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'not-authenticated'
            );
            expect(localStorage.getItem('imagineer_token')).toBeNull();
            expect(localStorage.getItem('imagineer_user')).toBeNull();
        });
    });

    describe('login', () => {
        it('should store token and user in state and localStorage', async () => {
            const user = userEvent.setup();

            render(
                <AuthProvider>
                    <TestComponent />
                </AuthProvider>
            );

            await waitFor(() => {
                expect(screen.getByTestId('loading')).toHaveTextContent('ready');
            });

            await user.click(screen.getByText('Login'));

            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'authenticated'
            );
            expect(screen.getByTestId('user')).toHaveTextContent('Test User');
            expect(screen.getByTestId('token')).toHaveTextContent('test-token');
            expect(localStorage.getItem('imagineer_token')).toBe('test-token');
            expect(
                JSON.parse(localStorage.getItem('imagineer_user') || '{}')
            ).toEqual({
                id: '123',
                email: 'test@example.com',
                name: 'Test User',
            });
        });
    });

    describe('logout', () => {
        it('should clear token and user from state and localStorage', async () => {
            const user = userEvent.setup();

            render(
                <AuthProvider>
                    <TestComponent />
                </AuthProvider>
            );

            await waitFor(() => {
                expect(screen.getByTestId('loading')).toHaveTextContent('ready');
            });

            // First login
            await user.click(screen.getByText('Login'));
            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'authenticated'
            );

            // Then logout
            await user.click(screen.getByText('Logout'));
            expect(screen.getByTestId('authenticated')).toHaveTextContent(
                'not-authenticated'
            );
            expect(screen.getByTestId('user')).toHaveTextContent('no-user');
            expect(localStorage.getItem('imagineer_token')).toBeNull();
            expect(localStorage.getItem('imagineer_user')).toBeNull();
        });
    });

    describe('useAuth', () => {
        it('should throw error when used outside AuthProvider', () => {
            // Suppress console.error for this test
            const consoleSpy = vi.spyOn(console, 'error');
            consoleSpy.mockImplementation(() => {});

            expect(() => {
                render(<TestComponent />);
            }).toThrow('useAuth must be used within an AuthProvider');

            consoleSpy.mockRestore();
        });
    });

    describe('getStoredToken', () => {
        it('should return token when valid and not expired', () => {
            const exp = Math.floor(Date.now() / 1000) + 3600;
            const payload = { exp };
            const token = `header.${btoa(JSON.stringify(payload))}.signature`;

            localStorage.setItem('imagineer_token', token);

            expect(getStoredToken()).toBe(token);
        });

        it('should return null when token is expired', () => {
            const exp = Math.floor(Date.now() / 1000) - 3600;
            const payload = { exp };
            const token = `header.${btoa(JSON.stringify(payload))}.signature`;

            localStorage.setItem('imagineer_token', token);

            expect(getStoredToken()).toBeNull();
        });

        it('should return null when no token stored', () => {
            expect(getStoredToken()).toBeNull();
        });

        it('should return null when token is invalid', () => {
            localStorage.setItem('imagineer_token', 'invalid-token');

            expect(getStoredToken()).toBeNull();
        });
    });

    describe('clearStoredAuth', () => {
        it('should remove token and user from localStorage', () => {
            localStorage.setItem('imagineer_token', 'test-token');
            localStorage.setItem('imagineer_user', '{"name":"Test"}');

            clearStoredAuth();

            expect(localStorage.getItem('imagineer_token')).toBeNull();
            expect(localStorage.getItem('imagineer_user')).toBeNull();
        });
    });
});
