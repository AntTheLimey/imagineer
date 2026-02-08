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
import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import { Markdown } from 'tiptap-markdown';
import WikiLinkNode from './WikiLinkNode';

/**
 * Represents the JSON shape of a wikiLink node returned by
 * TipTap's `editor.getJSON()`.
 */
interface WikiLinkJSON {
    type: string;
    attrs?: {
        entityName?: string;
        displayText?: string | null;
    };
}

/**
 * Shape of the tiptap-markdown storage object used to access
 * the getMarkdown() helper.
 */
interface MarkdownStorage {
    markdown: {
        getMarkdown: () => string;
    };
}

/**
 * Creates a headless TipTap editor configured with StarterKit,
 * the Markdown extension, and the WikiLinkNode extension.
 *
 * @param content - Initial Markdown content for the editor.
 * @returns A configured TipTap Editor instance.
 */
function createEditor(content: string): Editor {
    return new Editor({
        extensions: [
            StarterKit,
            Markdown.configure({
                html: false,
                transformCopiedText: true,
                transformPastedText: true,
            }),
            WikiLinkNode,
        ],
        content,
    });
}

/**
 * Extracts the Markdown string from a TipTap editor instance.
 *
 * @param editor - The TipTap editor.
 * @returns The current Markdown content.
 */
function getMarkdown(editor: Editor): string {
    const storage = editor.storage as unknown as MarkdownStorage;
    return storage.markdown.getMarkdown();
}

describe('WikiLinkNode', () => {
    describe('Markdown parsing', () => {
        it('parses [[Entity Name]] into a wikiLink node', () => {
            const editor = createEditor('Hello [[Arkham]] world');
            const doc = editor.getJSON();

            const paragraph = doc.content?.[0];
            expect(paragraph?.type).toBe('paragraph');

            const wikiLinkNode = paragraph?.content?.find(
                (node: { type: string }) => node.type === 'wikiLink',
            ) as WikiLinkJSON | undefined;
            expect(wikiLinkNode).toBeDefined();
            expect(wikiLinkNode?.attrs?.entityName).toBe('Arkham');
            expect(wikiLinkNode?.attrs?.displayText).toBeNull();

            editor.destroy();
        });

        it('parses [[Entity Name|display text]] with alias', () => {
            const editor = createEditor(
                'Visit [[Arkham|the cursed city]] soon',
            );
            const doc = editor.getJSON();

            const paragraph = doc.content?.[0];
            const wikiLinkNode = paragraph?.content?.find(
                (node: { type: string }) => node.type === 'wikiLink',
            ) as WikiLinkJSON | undefined;
            expect(wikiLinkNode).toBeDefined();
            expect(wikiLinkNode?.attrs?.entityName).toBe('Arkham');
            expect(wikiLinkNode?.attrs?.displayText).toBe(
                'the cursed city',
            );

            editor.destroy();
        });

        it('parses multiple wiki links in one paragraph', () => {
            const editor = createEditor(
                '[[Alice]] met [[Bob]] at [[The Tavern]]',
            );
            const doc = editor.getJSON();

            const paragraph = doc.content?.[0];
            const wikiLinks = (
                paragraph?.content?.filter(
                    (node: { type: string }) =>
                        node.type === 'wikiLink',
                ) ?? []
            ) as WikiLinkJSON[];

            expect(wikiLinks).toHaveLength(3);
            expect(wikiLinks[0]?.attrs?.entityName).toBe('Alice');
            expect(wikiLinks[1]?.attrs?.entityName).toBe('Bob');
            expect(wikiLinks[2]?.attrs?.entityName).toBe(
                'The Tavern',
            );

            editor.destroy();
        });

        it('does not parse empty brackets [[]]', () => {
            const editor = createEditor('Empty [[]] here');
            const doc = editor.getJSON();

            const paragraph = doc.content?.[0];
            const wikiLinks =
                paragraph?.content?.filter(
                    (node: { type: string }) =>
                        node.type === 'wikiLink',
                ) ?? [];

            expect(wikiLinks).toHaveLength(0);

            editor.destroy();
        });

        it('does not parse unclosed brackets [[no close', () => {
            const editor = createEditor('Unclosed [[no close');
            const doc = editor.getJSON();

            const paragraph = doc.content?.[0];
            const wikiLinks =
                paragraph?.content?.filter(
                    (node: { type: string }) =>
                        node.type === 'wikiLink',
                ) ?? [];

            expect(wikiLinks).toHaveLength(0);

            editor.destroy();
        });
    });

    describe('Markdown serialization', () => {
        it('serializes a wikiLink node to [[Entity Name]]', () => {
            const editor = createEditor('Hello [[Arkham]] world');
            const markdown = getMarkdown(editor);

            expect(markdown).toContain('[[Arkham]]');

            editor.destroy();
        });

        it('serializes a wikiLink node with alias to [[Entity Name|display text]]', () => {
            const editor = createEditor(
                'Visit [[Arkham|the cursed city]] soon',
            );
            const markdown = getMarkdown(editor);

            expect(markdown).toContain('[[Arkham|the cursed city]]');

            editor.destroy();
        });

        it('preserves surrounding text during serialization', () => {
            const editor = createEditor(
                'Before [[Arkham]] after',
            );
            const markdown = getMarkdown(editor);

            expect(markdown).toContain('Before');
            expect(markdown).toContain('[[Arkham]]');
            expect(markdown).toContain('after');

            editor.destroy();
        });

        it('serializes multiple wiki links correctly', () => {
            const editor = createEditor(
                '[[Alice]] and [[Bob]]',
            );
            const markdown = getMarkdown(editor);

            expect(markdown).toContain('[[Alice]]');
            expect(markdown).toContain('[[Bob]]');

            editor.destroy();
        });
    });

    describe('round-trip', () => {
        it('round-trips [[Entity Name]] through parse and serialize', () => {
            const original = 'Hello [[Arkham]] world';
            const editor = createEditor(original);
            const markdown = getMarkdown(editor);

            expect(markdown).toContain('[[Arkham]]');
            expect(markdown).toContain('Hello');
            expect(markdown).toContain('world');

            // Parse the serialized output again to verify
            // structural equivalence.
            const editor2 = createEditor(markdown);
            const doc2 = editor2.getJSON();
            const paragraph2 = doc2.content?.[0];
            const wikiLink = paragraph2?.content?.find(
                (node: { type: string }) => node.type === 'wikiLink',
            ) as WikiLinkJSON | undefined;
            expect(wikiLink?.attrs?.entityName).toBe('Arkham');

            editor.destroy();
            editor2.destroy();
        });

        it('round-trips [[Entity|Display]] through parse and serialize', () => {
            const original =
                'Visit [[Arkham|the cursed city]] soon';
            const editor = createEditor(original);
            const markdown = getMarkdown(editor);

            expect(markdown).toContain(
                '[[Arkham|the cursed city]]',
            );

            const editor2 = createEditor(markdown);
            const doc2 = editor2.getJSON();
            const paragraph2 = doc2.content?.[0];
            const wikiLink = paragraph2?.content?.find(
                (node: { type: string }) => node.type === 'wikiLink',
            ) as WikiLinkJSON | undefined;
            expect(wikiLink?.attrs?.entityName).toBe('Arkham');
            expect(wikiLink?.attrs?.displayText).toBe(
                'the cursed city',
            );

            editor.destroy();
            editor2.destroy();
        });
    });
});
