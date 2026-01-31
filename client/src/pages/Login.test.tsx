// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Login from './Login';

// Mock the AuthContext
const mockUseAuth = vi.fn();

vi.mock('../contexts/AuthContext', () => ({
    useAuth: () => mockUseAuth(),
}));

// Mock window.location
const mockLocation = {
    href: '',
};

Object.defineProperty(window, 'location', {
    value: mockLocation,
    writable: true,
});

const renderLogin = () => {
    return render(
        <MemoryRouter>
            <Login />
        </MemoryRouter>
    );
};

describe('Login', () => {
    beforeEach(() => {
        mockLocation.href = '';
        mockUseAuth.mockReturnValue({
            isAuthenticated: false,
            isLoading: false,
        });
    });

    it('should render the Imagineer branding', () => {
        renderLogin();

        expect(screen.getByText('Imagineer')).toBeInTheDocument();
        expect(
            screen.getByText('TTRPG Campaign Intelligence Platform')
        ).toBeInTheDocument();
    });

    it('should render the sign in description', () => {
        renderLogin();

        expect(
            screen.getByText(/Sign in to manage your campaigns/i)
        ).toBeInTheDocument();
    });

    it('should render the Google sign in button', () => {
        renderLogin();

        expect(
            screen.getByRole('button', { name: /sign in with google/i })
        ).toBeInTheDocument();
    });

    it('should show loading state when auth is loading', () => {
        mockUseAuth.mockReturnValue({
            isAuthenticated: false,
            isLoading: true,
        });

        renderLogin();

        // Should show loading indicator, not the login form
        expect(screen.queryByText('Imagineer')).not.toBeInTheDocument();
    });

    it('should redirect to OAuth endpoint when sign in button is clicked', async () => {
        const user = userEvent.setup();

        renderLogin();

        const signInButton = screen.getByRole('button', {
            name: /sign in with google/i,
        });
        await user.click(signInButton);

        expect(mockLocation.href).toBe('/api/auth/google');
    });

    it('should show redirecting state after clicking sign in', async () => {
        const user = userEvent.setup();

        renderLogin();

        const signInButton = screen.getByRole('button', {
            name: /sign in with google/i,
        });
        await user.click(signInButton);

        expect(
            screen.getByRole('button', { name: /redirecting/i })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: /redirecting/i })
        ).toBeDisabled();
    });
});
