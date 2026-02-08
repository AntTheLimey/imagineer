/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { Node, mergeAttributes } from '@tiptap/core';
import { ReactNodeViewRenderer } from '@tiptap/react';
import WikiLinkComponent from './WikiLinkComponent';

/**
 * Minimal structural type for a markdown-it instance, covering
 * only the parts needed by the wiki link parser registration.
 *
 * The full `markdown-it` type declarations are not available at
 * the top level (they are nested inside tiptap-markdown), so a
 * structural type avoids the unresolvable import.
 */
interface MarkdownItInstance {
    inline: {
        ruler: {
            push: (
                name: string,
                rule: (
                    state: WikiLinkParserState,
                    silent: boolean,
                ) => boolean,
            ) => void;
        };
    };
    renderer: {
        rules: Record<
            string,
            (
                tokens: Array<{
                    meta: {
                        entityName: string;
                        displayText: string | null;
                    };
                }>,
                idx: number,
            ) => string
        >;
    };
}

/**
 * Structural type for the markdown-it inline parser state, covering
 * only the properties used by the wiki link rule.
 */
interface WikiLinkParserState {
    src: string;
    pos: number;
    posMax: number;
    push: (
        type: string,
        tag: string,
        nesting: number,
    ) => {
        markup: string;
        meta: Record<string, unknown>;
    };
}

/**
 * A custom TipTap node extension for wiki-style links.
 *
 * Represents `[[Entity Name]]` and `[[Entity Name|display text]]`
 * links in the editor. The node is inline, atomic (not editable
 * in-place), and renders via the WikiLinkComponent React component.
 *
 * Integrates with tiptap-markdown for round-trip Markdown
 * serialization and parsing:
 *
 * - Serialization converts the node back to `[[entityName]]` or
 *   `[[entityName|displayText]]` syntax.
 * - Parsing uses a custom markdown-it inline rule to detect
 *   `[[...]]` patterns and convert them into wikiLink nodes.
 */
const WikiLinkNode = Node.create({
    name: 'wikiLink',

    group: 'inline',

    inline: true,

    atom: true,

    addAttributes() {
        return {
            entityName: {
                default: '',
            },
            displayText: {
                default: null,
            },
        };
    },

    parseHTML() {
        return [
            {
                tag: 'span[data-wiki-link]',
                getAttrs: (element) => {
                    const el = element as HTMLElement;
                    return {
                        entityName:
                            el.getAttribute('data-entity-name') || '',
                        displayText:
                            el.getAttribute('data-display-text') ||
                            null,
                    };
                },
            },
        ];
    },

    renderHTML({ HTMLAttributes }) {
        return [
            'span',
            mergeAttributes(HTMLAttributes, {
                'data-wiki-link': '',
                'data-entity-name': HTMLAttributes.entityName,
                'data-display-text': HTMLAttributes.displayText || '',
            }),
            HTMLAttributes.displayText || HTMLAttributes.entityName,
        ];
    },

    addNodeView() {
        return ReactNodeViewRenderer(WikiLinkComponent);
    },

    addStorage() {
        return {
            markdown: {
                serialize(
                    state: {
                        text: (
                            text: string,
                            escape?: boolean,
                        ) => void;
                    },
                    node: {
                        attrs: {
                            entityName: string;
                            displayText: string | null;
                        };
                    },
                ) {
                    const { entityName, displayText } = node.attrs;
                    if (displayText) {
                        state.text(
                            `[[${entityName}|${displayText}]]`,
                            false,
                        );
                    } else {
                        state.text(
                            `[[${entityName}]]`,
                            false,
                        );
                    }
                },
                parse: {
                    setup(markdownit: MarkdownItInstance) {
                        // Guard against duplicate registration.
                        // tiptap-markdown calls setup() on every
                        // parse() invocation, so without this check
                        // the rule would accumulate duplicates.
                        if (
                            !markdownit.renderer.rules['wiki_link']
                        ) {
                            markdownit.inline.ruler.push(
                                'wiki_link',
                                wikiLinkRule,
                            );
                        }
                        markdownit.renderer.rules['wiki_link'] =
                            renderWikiLinkToken;
                    },
                },
            },
        };
    },
});

/**
 * Markdown-it inline rule that detects `[[...]]` wiki link syntax.
 *
 * Scans for the `[[` opening delimiter, captures all text until
 * the closing `]]`, and optionally splits on `|` to separate
 * the entity name from display text. Produces a `wiki_link`
 * token with `entityName` and `displayText` metadata.
 *
 * @param state - The markdown-it inline parser state.
 * @param silent - If true, only validate without producing tokens.
 * @returns True if a wiki link was found, false otherwise.
 */
function wikiLinkRule(
    state: WikiLinkParserState,
    silent: boolean,
): boolean {
    const { src, pos, posMax } = state;

    // Must start with [[
    if (
        pos + 3 >= posMax ||
        src.charCodeAt(pos) !== 0x5b ||
        src.charCodeAt(pos + 1) !== 0x5b
    ) {
        return false;
    }

    // Find the closing ]]
    const closingPos = src.indexOf(']]', pos + 2);
    if (closingPos === -1) {
        return false;
    }

    // Extract content between [[ and ]]
    const content = src.slice(pos + 2, closingPos);
    if (content.length === 0) {
        return false;
    }

    if (!silent) {
        const pipeIndex = content.indexOf('|');
        let entityName: string;
        let displayText: string | null = null;

        if (pipeIndex !== -1) {
            entityName = content.slice(0, pipeIndex).trim();
            displayText = content.slice(pipeIndex + 1).trim();
        } else {
            entityName = content.trim();
        }

        const token = state.push('wiki_link', '', 0);
        token.markup = `[[${content}]]`;
        token.meta = { entityName, displayText };
    }

    state.pos = closingPos + 2;
    return true;
}

/**
 * Markdown-it renderer for wiki_link tokens.
 *
 * Converts the parsed token into an HTML `<span>` element with
 * data attributes that the TipTap parseHTML rule can pick up to
 * reconstruct the wikiLink node.
 *
 * @param tokens - Array of markdown-it tokens.
 * @param idx - Index of the current token.
 * @returns An HTML string representing the wiki link.
 */
function renderWikiLinkToken(
    tokens: Array<{
        meta: { entityName: string; displayText: string | null };
    }>,
    idx: number,
): string {
    const { entityName, displayText } = tokens[idx].meta;
    const displayAttr = displayText
        ? ` data-display-text="${escapeHtml(displayText)}"`
        : '';
    const label = displayText || entityName;
    return (
        `<span data-wiki-link` +
        ` data-entity-name="${escapeHtml(entityName)}"` +
        `${displayAttr}>${escapeHtml(label)}</span>`
    );
}

/**
 * Escapes HTML special characters in a string.
 *
 * @param str - The string to escape.
 * @returns The escaped string safe for use in HTML attributes
 *     and content.
 */
function escapeHtml(str: string): string {
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

export default WikiLinkNode;
