/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { visit } from 'unist-util-visit';
import type { Root, Text, PhrasingContent } from 'mdast';

/**
 * Regex that matches wiki-link syntax: [[Entity Name]] or
 * [[Entity Name|display text]].
 *
 * Captures:
 *   Group 1: entity name (required, non-empty)
 *   Group 2: display text (optional, after pipe)
 */
const WIKI_LINK_RE = /\[\[([^\]|]+?)(?:\|([^\]]*?))?\]\]/g;

/**
 * Custom mdast node representing a wiki link.
 *
 * Uses `data.hName` and `data.hProperties` to instruct mdast-util-to-hast
 * to produce a `<wiki-link>` element with entity metadata as properties,
 * which react-markdown then maps to the WikiLinkInline component.
 */
export interface WikiLinkNode {
    type: 'wikiLink';
    data: {
        hName: 'wiki-link';
        hProperties: {
            entityName: string;
            displayText: string;
        };
    };
    children: Array<{ type: 'text'; value: string }>;
}

/**
 * Remark plugin that transforms `[[Entity Name]]` and
 * `[[Entity Name|display text]]` patterns in Markdown text into custom
 * wiki-link mdast nodes.
 *
 * The plugin walks the mdast tree and splits text nodes that contain
 * wiki-link syntax into a sequence of plain text and wiki-link nodes.
 * Text inside `code` and `inlineCode` nodes is left untouched.
 *
 * @returns A unified transformer function.
 *
 * @example
 * ```tsx
 * <ReactMarkdown remarkPlugins={[remarkWikiLinks]}>
 *     {content}
 * </ReactMarkdown>
 * ```
 */
export default function remarkWikiLinks() {
    return (tree: Root) => {
        visit(tree, 'text', (node: Text, index, parent) => {
            if (!parent || index === undefined) {
                return;
            }

            // Skip text inside code blocks and inline code.
            if ((parent.type as string) === 'code' || (parent.type as string) === 'inlineCode') {
                return;
            }

            const value = node.value;
            WIKI_LINK_RE.lastIndex = 0;

            if (!WIKI_LINK_RE.test(value)) {
                return;
            }

            // Reset regex and build replacement nodes.
            WIKI_LINK_RE.lastIndex = 0;
            const nodes: PhrasingContent[] = [];
            let lastIndex = 0;
            let match: RegExpExecArray | null;

            while ((match = WIKI_LINK_RE.exec(value)) !== null) {
                const entityName = match[1].trim();
                const displayText = match[2]?.trim() || entityName;

                // Skip empty entity names (e.g., [[]] or [[|text]]).
                if (!entityName) {
                    continue;
                }

                // Add preceding plain text if any.
                if (match.index > lastIndex) {
                    nodes.push({
                        type: 'text',
                        value: value.slice(lastIndex, match.index),
                    });
                }

                // Create the wiki-link node.
                const wikiLinkNode: WikiLinkNode = {
                    type: 'wikiLink',
                    data: {
                        hName: 'wiki-link',
                        hProperties: {
                            entityName,
                            displayText,
                        },
                    },
                    children: [{ type: 'text', value: displayText }],
                };

                nodes.push(wikiLinkNode as unknown as PhrasingContent);
                lastIndex = match.index + match[0].length;
            }

            // If no nodes were produced, the regex matched but all entries
            // were empty; leave the original text node untouched.
            if (nodes.length === 0) {
                return;
            }

            // Add trailing plain text if any.
            if (lastIndex < value.length) {
                nodes.push({
                    type: 'text',
                    value: value.slice(lastIndex),
                });
            }

            // Replace the original text node with the new sequence.
            parent.children.splice(index, 1, ...nodes);
        });
    };
}
