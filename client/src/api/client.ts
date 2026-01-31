// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { getStoredToken, clearStoredAuth } from '../contexts/AuthContext';

/**
 * API client with fetch wrapper for communicating with the backend.
 * Base URL points to /api - Vite proxy handles forwarding to :8080.
 * Includes JWT authentication and automatic 401 handling.
 */

const BASE_URL = '/api';

/**
 * Custom error class for API errors with status code and response body.
 */
export class ApiError extends Error {
    constructor(
        message: string,
        public status: number,
        public body?: unknown
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

/**
 * Response type for paginated list endpoints.
 */
export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    pageSize: number;
}

/**
 * Type for query parameters - allows any object with primitive values.
 */
export type QueryParams = {
    [key: string]: string | number | boolean | undefined;
};

/**
 * Options for API requests.
 */
interface RequestOptions {
    method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
    body?: unknown;
    headers?: Record<string, string>;
    params?: QueryParams;
    skipAuth?: boolean;
}

/**
 * Build a URL with query parameters.
 */
function buildUrl(path: string, params?: QueryParams): string {
    const url = new URL(`${BASE_URL}${path}`, window.location.origin);

    if (params) {
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined) {
                url.searchParams.append(key, String(value));
            }
        });
    }

    return url.toString();
}

/**
 * Get authorization headers with JWT token if available.
 */
function getAuthHeaders(): Record<string, string> {
    const token = getStoredToken();
    if (token) {
        return { Authorization: `Bearer ${token}` };
    }
    return {};
}

/**
 * Handle 401 Unauthorized responses by clearing auth and redirecting to login.
 */
function handleUnauthorized(): void {
    clearStoredAuth();
    // Only redirect if not already on the login or auth callback pages
    if (!window.location.pathname.startsWith('/login') &&
        !window.location.pathname.startsWith('/auth/callback')) {
        window.location.href = '/login';
    }
}

/**
 * Make an API request with JSON content-type, authentication, and error handling.
 * Throws ApiError on non-2xx responses.
 * Automatically handles 401 responses by clearing auth and redirecting to login.
 */
async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const { method = 'GET', body, headers = {}, params, skipAuth = false } = options;

    const url = buildUrl(path, params);

    const requestHeaders: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(skipAuth ? {} : getAuthHeaders()),
        ...headers,
    };

    const fetchOptions: RequestInit = {
        method,
        headers: requestHeaders,
    };

    if (body !== undefined && method !== 'GET') {
        fetchOptions.body = JSON.stringify(body);
    }

    const response = await fetch(url, fetchOptions);

    // Handle 401 Unauthorized
    if (response.status === 401) {
        handleUnauthorized();
        throw new ApiError('Unauthorized', 401);
    }

    // Handle non-2xx responses
    if (!response.ok) {
        let errorBody: unknown;
        try {
            errorBody = await response.json();
        } catch {
            // Response body is not JSON
            errorBody = await response.text();
        }

        const message =
            typeof errorBody === 'object' &&
            errorBody !== null &&
            'message' in errorBody
                ? String((errorBody as { message: unknown }).message)
                : `HTTP ${response.status}: ${response.statusText}`;

        throw new ApiError(message, response.status, errorBody);
    }

    // Handle 204 No Content
    if (response.status === 204) {
        return undefined as T;
    }

    // Parse JSON response
    return response.json() as Promise<T>;
}

/**
 * API client with typed methods for each HTTP verb.
 * All requests include JWT authentication when available.
 */
export const apiClient = {
    /**
     * Make a GET request.
     */
    get<T, P extends QueryParams = QueryParams>(path: string, params?: P): Promise<T> {
        return request<T>(path, { method: 'GET', params: params as QueryParams });
    },

    /**
     * Make a POST request.
     */
    post<T, P extends QueryParams = QueryParams>(path: string, body?: unknown, params?: P): Promise<T> {
        return request<T>(path, { method: 'POST', body, params: params as QueryParams });
    },

    /**
     * Make a PUT request.
     */
    put<T, P extends QueryParams = QueryParams>(path: string, body?: unknown, params?: P): Promise<T> {
        return request<T>(path, { method: 'PUT', body, params: params as QueryParams });
    },

    /**
     * Make a PATCH request.
     */
    patch<T, P extends QueryParams = QueryParams>(path: string, body?: unknown, params?: P): Promise<T> {
        return request<T>(path, { method: 'PATCH', body, params: params as QueryParams });
    },

    /**
     * Make a DELETE request.
     */
    delete<T, P extends QueryParams = QueryParams>(path: string, params?: P): Promise<T> {
        return request<T>(path, { method: 'DELETE', params: params as QueryParams });
    },

    /**
     * Upload a file using multipart/form-data.
     * Includes JWT authentication.
     */
    async upload<T>(path: string, file: File, additionalData?: Record<string, string>): Promise<T> {
        const formData = new FormData();
        formData.append('file', file);

        if (additionalData) {
            Object.entries(additionalData).forEach(([key, value]) => {
                formData.append(key, value);
            });
        }

        const response = await fetch(`${BASE_URL}${path}`, {
            method: 'POST',
            body: formData,
            headers: getAuthHeaders(),
            // Note: Do not set Content-Type header - browser will set it with boundary
        });

        // Handle 401 Unauthorized
        if (response.status === 401) {
            handleUnauthorized();
            throw new ApiError('Unauthorized', 401);
        }

        if (!response.ok) {
            let errorBody: unknown;
            try {
                errorBody = await response.json();
            } catch {
                errorBody = await response.text();
            }

            const message =
                typeof errorBody === 'object' &&
                errorBody !== null &&
                'message' in errorBody
                    ? String((errorBody as { message: unknown }).message)
                    : `HTTP ${response.status}: ${response.statusText}`;

            throw new ApiError(message, response.status, errorBody);
        }

        return response.json() as Promise<T>;
    },
};

export default apiClient;
