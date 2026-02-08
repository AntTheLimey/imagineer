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
import { unified } from 'unified';
import remarkParse from 'remark-parse';
import remarkWikiLinks from './remarkWikiLinks';
import type { Root, Text } from 'mdast';
import type { WikiLinkNode } from './remarkWikiLinks';

/**
 * Helper that parses Markdown through remark with the wiki-links
 * plugin and returns the resulting mdast tree.
 */
function parseWithWikiLinks(markdown: string): Root {
    const processor = unified()
        .use(remarkParse)
        .use(remarkWikiLinks);

    return processor.runSync(processor.parse(markdown)) as Root;
}

/**
 * Recursively collects all nodes of a given type from the mdast tree.
 */
function collectNodes<T extends { type: string }>(
    node: { type: string; children?: Array<{ type: string }> },
    type: string
): T[] {
    const results: T[] = [];
    if (node.type === type) {
        results.push(node as T);
    }
    if ('children' in node && Array.isArray(node.children)) {
        for (const child of node.children) {
            results.push(...collectNodes<T>(child, type));
        }
    }
    return results;
}

describe('remarkWikiLinks', () => {
    it('transforms [[Entity Name]] into a wikiLink node', () => {
        const tree = parseWithWikiLinks('Hello [[Inspector Legrasse]] world');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hName).toBe('wiki-link');
        expect(wikiLinks[0].data.hProperties.entityName).toBe(
            'Inspector Legrasse'
        );
        expect(wikiLinks[0].data.hProperties.displayText).toBe(
            'Inspector Legrasse'
        );
        expect(wikiLinks[0].children[0].value).toBe(
            'Inspector Legrasse'
        );
    });

    it('extracts display text from [[Entity Name|display text]]', () => {
        const tree = parseWithWikiLinks(
            'See [[Inspector Legrasse|the Inspector]] for details'
        );
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hProperties.entityName).toBe(
            'Inspector Legrasse'
        );
        expect(wikiLinks[0].data.hProperties.displayText).toBe(
            'the Inspector'
        );
        expect(wikiLinks[0].children[0].value).toBe('the Inspector');
    });

    it('handles multiple wiki links in one text block', () => {
        const tree = parseWithWikiLinks(
            '[[Alpha]] met [[Beta]] at [[Gamma]]'
        );
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(3);
        expect(wikiLinks[0].data.hProperties.entityName).toBe('Alpha');
        expect(wikiLinks[1].data.hProperties.entityName).toBe('Beta');
        expect(wikiLinks[2].data.hProperties.entityName).toBe('Gamma');
    });

    it('preserves surrounding text around wiki links', () => {
        const tree = parseWithWikiLinks(
            'Before [[Entity]] after'
        );
        // The paragraph should contain: text("Before "), wikiLink, text(" after")
        const paragraph = tree.children[0];
        expect(paragraph.type).toBe('paragraph');

        if (paragraph.type === 'paragraph' && 'children' in paragraph) {
            const children = paragraph.children;
            expect(children).toHaveLength(3);

            const before = children[0] as Text;
            expect(before.type).toBe('text');
            expect(before.value).toBe('Before ');

            expect(children[1].type).toBe('wikiLink');

            const after = children[2] as Text;
            expect(after.type).toBe('text');
            expect(after.value).toBe(' after');
        }
    });

    it('handles wiki links inside bold text', () => {
        const tree = parseWithWikiLinks('**bold [[Entity]] text**');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hProperties.entityName).toBe('Entity');
    });

    it('handles wiki links inside italic text', () => {
        const tree = parseWithWikiLinks('*italic [[Entity]] text*');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hProperties.entityName).toBe('Entity');
    });

    it('handles wiki links inside list items', () => {
        const tree = parseWithWikiLinks(
            '- First [[Alpha]]\n- Second [[Beta]]'
        );
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(2);
        expect(wikiLinks[0].data.hProperties.entityName).toBe('Alpha');
        expect(wikiLinks[1].data.hProperties.entityName).toBe('Beta');
    });

    it('leaves unclosed [[ as regular text', () => {
        const tree = parseWithWikiLinks('This has [[ unclosed brackets');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(0);

        // The text should still be present unmodified.
        const texts = collectNodes<Text>(tree, 'text');
        const fullText = texts.map((t) => t.value).join('');
        expect(fullText).toContain('[[');
    });

    it('leaves empty [[]] as regular text', () => {
        const tree = parseWithWikiLinks('Empty [[]] link');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        // The regex requires at least one non-pipe, non-bracket character,
        // so [[]] should not produce a wiki link node.
        expect(wikiLinks).toHaveLength(0);
    });

    it('does not process wiki links inside inline code', () => {
        const tree = parseWithWikiLinks('Use `[[Entity]]` in code');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(0);
    });

    it('does not process wiki links inside fenced code blocks', () => {
        const tree = parseWithWikiLinks(
            '```\n[[Entity]]\n```'
        );
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(0);
    });

    it('trims whitespace from entity names', () => {
        const tree = parseWithWikiLinks('[[  Spaced Entity  ]]');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hProperties.entityName).toBe(
            'Spaced Entity'
        );
    });

    it('trims whitespace from display text', () => {
        const tree = parseWithWikiLinks(
            '[[Entity|  display text  ]]'
        );
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(1);
        expect(wikiLinks[0].data.hProperties.displayText).toBe(
            'display text'
        );
    });

    it('handles adjacent wiki links without space', () => {
        const tree = parseWithWikiLinks('[[Alpha]][[Beta]]');
        const wikiLinks = collectNodes<WikiLinkNode>(tree, 'wikiLink');

        expect(wikiLinks).toHaveLength(2);
        expect(wikiLinks[0].data.hProperties.entityName).toBe('Alpha');
        expect(wikiLinks[1].data.hProperties.entityName).toBe('Beta');
    });
});
