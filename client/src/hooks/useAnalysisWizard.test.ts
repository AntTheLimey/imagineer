// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { describe, it, expect } from 'vitest';
import {
    getPhaseItems,
    DETECTION_GROUPS,
    ANALYSIS_GROUPS,
    ENRICHMENT_GROUPS,
} from './useAnalysisWizard';
import type {
    ContentAnalysisItem,
} from '../api/contentAnalysis';

function makeItem(
    detectionType: string,
    phase: string,
    resolution = 'pending',
): ContentAnalysisItem {
    return {
        id: Math.random() * 1000,
        jobId: 1,
        detectionType,
        matchedText: 'test',
        resolution,
        phase,
        createdAt: '2026-01-01T00:00:00Z',
    } as ContentAnalysisItem;
}

describe('getPhaseItems', () => {
    it('filters identification items', () => {
        const items = [
            makeItem('wiki_link_resolved', 'identification'),
            makeItem('untagged_mention', 'identification'),
            makeItem('description_update', 'enrichment'),
            makeItem('analysis_report', 'analysis'),
        ];
        const result = getPhaseItems(items, 'identification');
        expect(result).toHaveLength(2);
        expect(
            result.every((i) =>
                (DETECTION_GROUPS as readonly string[]).includes(
                    i.detectionType,
                ),
            ),
        ).toBe(true);
    });

    it('filters analysis items', () => {
        const items = [
            makeItem('analysis_report', 'analysis'),
            makeItem('canon_contradiction', 'analysis'),
            makeItem('wiki_link_resolved', 'identification'),
        ];
        const result = getPhaseItems(items, 'analysis');
        expect(result).toHaveLength(2);
        expect(
            result.every((i) =>
                (ANALYSIS_GROUPS as readonly string[]).includes(
                    i.detectionType,
                ),
            ),
        ).toBe(true);
    });

    it('filters enrichment items', () => {
        const items = [
            makeItem('description_update', 'enrichment'),
            makeItem('log_entry', 'enrichment'),
            makeItem('analysis_report', 'analysis'),
        ];
        const result = getPhaseItems(items, 'enrichment');
        expect(result).toHaveLength(2);
        expect(
            result.every((i) =>
                (ENRICHMENT_GROUPS as readonly string[]).includes(
                    i.detectionType,
                ),
            ),
        ).toBe(true);
    });

    it('returns empty array for unknown phase', () => {
        const items = [
            makeItem('wiki_link_resolved', 'identification'),
        ];
        const result = getPhaseItems(items, 'unknown_phase');
        expect(result).toHaveLength(0);
    });
});
