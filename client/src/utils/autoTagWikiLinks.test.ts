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
import { autoTagWikiLinks } from './autoTagWikiLinks';

describe('autoTagWikiLinks', () => {
    it('wraps exact name matches', () => {
        const result = autoTagWikiLinks(
            'Viktor met Elara at the tavern.',
            ['Viktor', 'Elara'],
        );
        expect(result).toBe(
            '[[Viktor]] met [[Elara]] at the tavern.',
        );
    });

    it('skips already tagged names', () => {
        const result = autoTagWikiLinks(
            '[[Viktor]] met Elara.',
            ['Viktor', 'Elara'],
        );
        expect(result).toBe(
            '[[Viktor]] met [[Elara]].',
        );
    });

    it('handles case-insensitive match', () => {
        const result = autoTagWikiLinks(
            'viktor arrived.',
            ['Viktor'],
        );
        expect(result).toBe(
            '[[viktor]] arrived.',
        );
    });

    it('handles no matches', () => {
        const result = autoTagWikiLinks(
            'Nothing to tag here.',
            ['Viktor'],
        );
        expect(result).toBe(
            'Nothing to tag here.',
        );
    });

    it('handles empty entity list', () => {
        const result = autoTagWikiLinks(
            'Some text.',
            [],
        );
        expect(result).toBe('Some text.');
    });
});
