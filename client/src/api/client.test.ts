// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiClient, ApiError } from './client';

describe('ApiError', () => {
    it('should create an error with message and status', () => {
        const error = new ApiError('Not Found', 404);

        expect(error.message).toBe('Not Found');
        expect(error.status).toBe(404);
        expect(error.name).toBe('ApiError');
        expect(error.body).toBeUndefined();
    });

    it('should create an error with body', () => {
        const body = { detail: 'Resource not found' };
        const error = new ApiError('Not Found', 404, body);

        expect(error.message).toBe('Not Found');
        expect(error.status).toBe(404);
        expect(error.body).toEqual(body);
    });

    it('should be an instance of Error', () => {
        const error = new ApiError('Test error', 500);

        expect(error).toBeInstanceOf(Error);
        expect(error).toBeInstanceOf(ApiError);
    });
});

describe('apiClient', () => {
    const originalFetch = global.fetch;

    beforeEach(() => {
        global.fetch = vi.fn();
    });

    afterEach(() => {
        global.fetch = originalFetch;
        vi.resetAllMocks();
    });

    describe('get', () => {
        it('should make a GET request and return JSON response', async () => {
            const mockData = { id: '123', name: 'Test' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve(mockData),
            } as Response);

            const result = await apiClient.get('/test');

            expect(global.fetch).toHaveBeenCalledTimes(1);
            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('/api/test');
            expect(options?.method).toBe('GET');
            expect(options?.headers).toEqual({ 'Content-Type': 'application/json' });
            expect(result).toEqual(mockData);
        });

        it('should include query parameters in the URL', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve({}),
            } as Response);

            await apiClient.get('/test', { page: 1, limit: 10, active: true });

            const [url] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('page=1');
            expect(url).toContain('limit=10');
            expect(url).toContain('active=true');
        });

        it('should skip undefined query parameters', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve({}),
            } as Response);

            await apiClient.get('/test', { page: 1, limit: undefined });

            const [url] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('page=1');
            expect(url).not.toContain('limit');
        });
    });

    describe('post', () => {
        it('should make a POST request with JSON body', async () => {
            const requestBody = { name: 'New Item' };
            const responseData = { id: '456', name: 'New Item' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 201,
                json: () => Promise.resolve(responseData),
            } as Response);

            const result = await apiClient.post('/items', requestBody);

            expect(global.fetch).toHaveBeenCalledTimes(1);
            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('/api/items');
            expect(options?.method).toBe('POST');
            expect(options?.body).toBe(JSON.stringify(requestBody));
            expect(result).toEqual(responseData);
        });
    });

    describe('put', () => {
        it('should make a PUT request with JSON body', async () => {
            const requestBody = { name: 'Updated Item' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve(requestBody),
            } as Response);

            const result = await apiClient.put('/items/123', requestBody);

            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('/api/items/123');
            expect(options?.method).toBe('PUT');
            expect(options?.body).toBe(JSON.stringify(requestBody));
            expect(result).toEqual(requestBody);
        });
    });

    describe('patch', () => {
        it('should make a PATCH request with JSON body', async () => {
            const requestBody = { name: 'Patched Item' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve(requestBody),
            } as Response);

            const result = await apiClient.patch('/items/123', requestBody);

            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('/api/items/123');
            expect(options?.method).toBe('PATCH');
            expect(options?.body).toBe(JSON.stringify(requestBody));
            expect(result).toEqual(requestBody);
        });
    });

    describe('delete', () => {
        it('should make a DELETE request', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 204,
            } as Response);

            await apiClient.delete('/items/123');

            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toContain('/api/items/123');
            expect(options?.method).toBe('DELETE');
        });

        it('should handle 204 No Content response', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 204,
            } as Response);

            const result = await apiClient.delete('/items/123');

            expect(result).toBeUndefined();
        });
    });

    describe('error handling', () => {
        it('should throw ApiError on non-2xx response with JSON body', async () => {
            const errorBody = { message: 'Not found', code: 'NOT_FOUND' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: false,
                status: 404,
                statusText: 'Not Found',
                json: () => Promise.resolve(errorBody),
            } as Response);

            await expect(apiClient.get('/not-found')).rejects.toThrow(ApiError);

            try {
                await apiClient.get('/not-found');
            } catch (error) {
                expect(error).toBeInstanceOf(ApiError);
                const apiError = error as ApiError;
                expect(apiError.status).toBe(404);
                expect(apiError.message).toBe('Not found');
                expect(apiError.body).toEqual(errorBody);
            }
        });

        it('should throw ApiError with status text when response body is not JSON', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: false,
                status: 500,
                statusText: 'Internal Server Error',
                json: () => Promise.reject(new Error('Not JSON')),
                text: () => Promise.resolve('Server error occurred'),
            } as unknown as Response);

            try {
                await apiClient.get('/server-error');
            } catch (error) {
                expect(error).toBeInstanceOf(ApiError);
                const apiError = error as ApiError;
                expect(apiError.status).toBe(500);
                expect(apiError.message).toBe('HTTP 500: Internal Server Error');
            }
        });

        it('should throw ApiError on 401 Unauthorized', async () => {
            vi.mocked(global.fetch).mockResolvedValue({
                ok: false,
                status: 401,
                statusText: 'Unauthorized',
                json: () => Promise.resolve({ message: 'Authentication required' }),
            } as Response);

            await expect(apiClient.get('/protected')).rejects.toThrow(ApiError);

            try {
                await apiClient.get('/protected');
            } catch (error) {
                const apiError = error as ApiError;
                expect(apiError.status).toBe(401);
                expect(apiError.message).toBe('Authentication required');
            }
        });
    });

    describe('upload', () => {
        it('should upload a file using FormData', async () => {
            const mockFile = new File(['test content'], 'test.txt', { type: 'text/plain' });
            const responseData = { id: '789', filename: 'test.txt' };
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve(responseData),
            } as Response);

            const result = await apiClient.upload('/upload', mockFile);

            expect(global.fetch).toHaveBeenCalledTimes(1);
            const [url, options] = vi.mocked(global.fetch).mock.calls[0];
            expect(url).toBe('/api/upload');
            expect(options?.method).toBe('POST');
            expect(options?.body).toBeInstanceOf(FormData);
            expect(result).toEqual(responseData);
        });

        it('should include additional data in FormData', async () => {
            const mockFile = new File(['test'], 'test.txt', { type: 'text/plain' });
            vi.mocked(global.fetch).mockResolvedValue({
                ok: true,
                status: 200,
                json: () => Promise.resolve({}),
            } as Response);

            await apiClient.upload('/upload', mockFile, { campaignId: '123' });

            const [, options] = vi.mocked(global.fetch).mock.calls[0];
            const formData = options?.body as FormData;
            expect(formData.get('campaignId')).toBe('123');
        });

        it('should throw ApiError on upload failure', async () => {
            const mockFile = new File(['test'], 'test.txt', { type: 'text/plain' });
            vi.mocked(global.fetch).mockResolvedValue({
                ok: false,
                status: 400,
                statusText: 'Bad Request',
                json: () => Promise.resolve({ message: 'Invalid file type' }),
            } as Response);

            await expect(apiClient.upload('/upload', mockFile)).rejects.toThrow(ApiError);

            try {
                await apiClient.upload('/upload', mockFile);
            } catch (error) {
                const apiError = error as ApiError;
                expect(apiError.status).toBe(400);
                expect(apiError.message).toBe('Invalid file type');
            }
        });
    });
});
