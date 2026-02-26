/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { describe, it, expect } from 'vitest';
import type { ContentAnalysisJob } from './contentAnalysis';

describe('ContentAnalysisJob type', () => {
    it('includes failureReason field', () => {
        const job: ContentAnalysisJob = {
            id: 1,
            campaignId: 1,
            sourceTable: 'chapters',
            sourceId: 1,
            sourceField: 'overview',
            status: 'failed',
            totalItems: 5,
            resolvedItems: 0,
            enrichmentTotal: 0,
            enrichmentResolved: 0,
            phases: ['identify'],
            currentPhase: 'identify',
            failureReason: 'API quota exceeded',
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
        };
        expect(job.failureReason).toBe('API quota exceeded');
    });

    it('allows undefined failureReason', () => {
        const job: ContentAnalysisJob = {
            id: 1,
            campaignId: 1,
            sourceTable: 'chapters',
            sourceId: 1,
            sourceField: 'overview',
            status: 'completed',
            totalItems: 5,
            resolvedItems: 5,
            enrichmentTotal: 0,
            enrichmentResolved: 0,
            phases: ['identify'],
            currentPhase: null,
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
        };
        expect(job.failureReason).toBeUndefined();
    });
});
