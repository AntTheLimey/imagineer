// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDraft, DraftData } from './useDraft';

describe('useDraft', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.clearAllMocks();
    });

    describe('saveDraft', () => {
        it('should save draft to localStorage with metadata', () => {
            const { result } = renderHook(() => useDraft());

            const testData = { name: 'Test Entity', type: 'npc' };

            act(() => {
                result.current.saveDraft('test-key', testData, 1);
            });

            const stored = localStorage.getItem('imagineer_draft_test-key');
            expect(stored).not.toBeNull();

            const parsed = JSON.parse(stored!) as DraftData<typeof testData>;
            expect(parsed.data).toEqual(testData);
            expect(parsed.serverVersion).toBe(1);
            expect(parsed.savedAt).toBeDefined();
        });

        it('should save draft without server version', () => {
            const { result } = renderHook(() => useDraft());

            const testData = { name: 'Test' };

            act(() => {
                result.current.saveDraft('no-version', testData);
            });

            const stored = localStorage.getItem('imagineer_draft_no-version');
            const parsed = JSON.parse(stored!) as DraftData<typeof testData>;
            expect(parsed.serverVersion).toBeUndefined();
        });
    });

    describe('getDraft', () => {
        it('should retrieve existing draft', () => {
            const { result } = renderHook(() => useDraft());

            const testData = { name: 'Test Entity' };
            const draftData: DraftData<typeof testData> = {
                data: testData,
                savedAt: new Date().toISOString(),
                serverVersion: 2,
            };
            localStorage.setItem(
                'imagineer_draft_get-test',
                JSON.stringify(draftData)
            );

            const retrieved = result.current.getDraft<typeof testData>('get-test');

            expect(retrieved).not.toBeNull();
            expect(retrieved?.data).toEqual(testData);
            expect(retrieved?.serverVersion).toBe(2);
        });

        it('should return null for non-existent draft', () => {
            const { result } = renderHook(() => useDraft());

            const retrieved = result.current.getDraft('non-existent');

            expect(retrieved).toBeNull();
        });

        it('should return null for invalid JSON', () => {
            localStorage.setItem('imagineer_draft_invalid', 'not-json');

            const { result } = renderHook(() => useDraft());

            const retrieved = result.current.getDraft('invalid');

            expect(retrieved).toBeNull();
        });
    });

    describe('deleteDraft', () => {
        it('should remove draft from localStorage', () => {
            localStorage.setItem(
                'imagineer_draft_to-delete',
                JSON.stringify({ data: 'test' })
            );

            const { result } = renderHook(() => useDraft());

            act(() => {
                result.current.deleteDraft('to-delete');
            });

            expect(localStorage.getItem('imagineer_draft_to-delete')).toBeNull();
        });

        it('should not throw for non-existent draft', () => {
            const { result } = renderHook(() => useDraft());

            expect(() => {
                act(() => {
                    result.current.deleteDraft('does-not-exist');
                });
            }).not.toThrow();
        });
    });

    describe('hasDraft', () => {
        it('should return true for existing draft', () => {
            localStorage.setItem(
                'imagineer_draft_exists',
                JSON.stringify({ data: 'test' })
            );

            const { result } = renderHook(() => useDraft());

            expect(result.current.hasDraft('exists')).toBe(true);
        });

        it('should return false for non-existent draft', () => {
            const { result } = renderHook(() => useDraft());

            expect(result.current.hasDraft('does-not-exist')).toBe(false);
        });
    });

    describe('round-trip', () => {
        it('should save and retrieve draft correctly', () => {
            const { result } = renderHook(() => useDraft());

            const testData = {
                name: 'Complex Entity',
                tags: ['tag1', 'tag2'],
                nested: { value: 123 },
            };

            act(() => {
                result.current.saveDraft('round-trip', testData, 5);
            });

            const retrieved = result.current.getDraft<typeof testData>('round-trip');

            expect(retrieved?.data).toEqual(testData);
            expect(retrieved?.serverVersion).toBe(5);

            act(() => {
                result.current.deleteDraft('round-trip');
            });

            expect(result.current.hasDraft('round-trip')).toBe(false);
        });
    });
});
