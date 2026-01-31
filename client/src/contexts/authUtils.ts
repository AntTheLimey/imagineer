// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

export const TOKEN_KEY = 'imagineer_token';
export const USER_KEY = 'imagineer_user';

/**
 * Decode the JWT payload from a token string.
 * Returns the parsed payload object or null if the token is invalid or
 * malformed. This function does NOT perform expiration checks; use
 * isTokenExpired for that purpose.
 */
export function parseJWT(token: string): { exp: number } | null {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            return null;
        }
        // JWTs use base64url encoding which differs from standard base64:
        // - Uses '-' instead of '+'
        // - Uses '_' instead of '/'
        // - Does not require padding with '='
        // Convert base64url to standard base64 before decoding
        let base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
        // Add padding if needed
        const padding = base64.length % 4;
        if (padding) {
            base64 += '='.repeat(4 - padding);
        }
        const payload = JSON.parse(atob(base64));
        return payload;
    } catch {
        return null;
    }
}

/**
 * Check if a JWT token is expired.
 */
export function isTokenExpired(token: string): boolean {
    const payload = parseJWT(token);
    if (!payload || !payload.exp) {
        return true;
    }
    // Add 60 second buffer to avoid edge cases
    return payload.exp * 1000 < Date.now() + 60000;
}

/**
 * Get the stored token directly from localStorage.
 * Useful for API client that needs token outside of React context.
 */
export function getStoredToken(): string | null {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token && !isTokenExpired(token)) {
        return token;
    }
    return null;
}

/**
 * Clear stored auth data.
 * Useful for API client to clear auth on 401 responses.
 */
export function clearStoredAuth(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
}
