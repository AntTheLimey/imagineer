/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { createRoot, type Root } from 'react-dom/client';
import { createElement } from 'react';
import { PluginKey } from '@tiptap/pm/state';
import { type ResolvedPos } from '@tiptap/pm/model';
import type { SuggestionOptions, SuggestionProps } from '@tiptap/suggestion';
import { entityResolveApi } from '../../api/entityResolve';
import type { EntityResolveResult } from '../../api/entityResolve';
import WikiLinkSuggestionList from './WikiLinkSuggestionList';
import type { WikiLinkSuggestionListHandle } from './WikiLinkSuggestionList';

/**
 * Plugin key for the wiki link suggestion plugin so it does not
 * collide with other suggestion instances (such as mentions).
 */
export const wikiLinkSuggestionPluginKey = new PluginKey(
    'wikiLinkSuggestion',
);

/**
 * Custom match function for `[[` double-bracket triggers.
 *
 * The standard @tiptap/suggestion `findSuggestionMatch` only
 * supports single-character triggers. This implementation detects
 * `[[` as the opening delimiter, captures everything after it as
 * the query, and allows spaces within entity names.
 *
 * @param config - The trigger configuration from the suggestion
 *     plugin (only `$position` is used).
 * @returns A match object with range and query, or null if the
 *     cursor is not inside a `[[` trigger.
 */
function findDoubleBracketMatch(config: {
    $position: ResolvedPos;
    char: string;
    allowSpaces: boolean;
    allowToIncludeChar: boolean;
    allowedPrefixes: string[] | null;
    startOfLine: boolean;
}) {
    const { $position } = config;
    const text =
        $position.nodeBefore?.isText && $position.nodeBefore.text;

    if (!text) {
        return null;
    }

    // Find the last occurrence of [[
    const triggerIndex = text.lastIndexOf('[[');
    if (triggerIndex === -1) {
        return null;
    }

    // Make sure there is no ]] closing bracket between the trigger
    // and the end of the text (the cursor position). If there is,
    // the wiki link is already closed and we should not trigger.
    const afterTrigger = text.slice(triggerIndex + 2);
    if (afterTrigger.includes(']]')) {
        return null;
    }

    const textFrom = $position.pos - text.length;
    const from = textFrom + triggerIndex;
    const to = $position.pos;
    const query = afterTrigger;

    return {
        range: { from, to },
        query,
        text: text.slice(triggerIndex),
    };
}

/**
 * Builds the suggestion configuration for wiki link autocomplete.
 *
 * The returned object is compatible with `@tiptap/suggestion` and
 * can be passed to `Suggestion()` to create a ProseMirror plugin.
 * The configuration:
 *
 * - Triggers on `[[` using a custom `findSuggestionMatch`.
 * - Debounces entity resolve API calls by 300ms.
 * - Requires at least 3 characters before querying.
 * - Supports `|` pipe syntax to separate entity name from display
 *   text (only the entity name portion is used for searching).
 * - Inserts a WikiLink node on selection.
 * - Renders an MUI dropdown via a React portal.
 *
 * @param campaignId - The campaign to search entities within.
 * @returns A partial SuggestionOptions object (without `editor`,
 *     which is injected when the plugin is created).
 */
export function buildWikiLinkSuggestion(
    campaignId: number,
): Omit<SuggestionOptions<EntityResolveResult>, 'editor'> {
    let debounceTimer: ReturnType<typeof setTimeout> | null = null;

    return {
        pluginKey: wikiLinkSuggestionPluginKey,

        char: '[[',

        allowSpaces: true,

        findSuggestionMatch: findDoubleBracketMatch,

        items: ({ query }) => {
            return new Promise<EntityResolveResult[]>((resolve) => {
                if (debounceTimer) {
                    clearTimeout(debounceTimer);
                }

                // Extract entity name portion (before pipe)
                const entityQuery = query.includes('|')
                    ? query.split('|')[0].trim()
                    : query.trim();

                // Require at least 3 characters
                if (entityQuery.length < 3) {
                    resolve([]);
                    return;
                }

                debounceTimer = setTimeout(async () => {
                    try {
                        const results =
                            await entityResolveApi.resolve(
                                campaignId,
                                entityQuery,
                            );
                        resolve(results);
                    } catch {
                        resolve([]);
                    }
                }, 300);
            });
        },

        command: ({ editor, range, props: item }) => {
            // Check for pipe syntax in the current query
            const state = wikiLinkSuggestionPluginKey.getState(
                editor.state,
            );
            const rawQuery: string = state?.query ?? '';
            const pipeIndex = rawQuery.indexOf('|');
            const displayText =
                pipeIndex !== -1
                    ? rawQuery.slice(pipeIndex + 1).trim() || null
                    : null;

            editor
                .chain()
                .focus()
                .deleteRange(range)
                .insertContent({
                    type: 'wikiLink',
                    attrs: {
                        entityName: item.name,
                        displayText,
                    },
                })
                .run();
        },

        render: () => {
            let container: HTMLDivElement | null = null;
            let root: Root | null = null;
            let listRef: WikiLinkSuggestionListHandle | null = null;

            /**
             * Renders or updates the suggestion list by mounting a
             * React component into a portal container.
             */
            function renderList(
                props: SuggestionProps<EntityResolveResult>,
                loading: boolean,
            ) {
                if (!container || !root) return;

                // Position the container near the cursor
                const rect = props.clientRect?.();
                if (rect) {
                    container.style.position = 'fixed';
                    container.style.left = `${rect.left}px`;
                    container.style.top = `${rect.bottom + 4}px`;
                    container.style.zIndex = '1300';
                }

                root.render(
                    createElement(WikiLinkSuggestionList, {
                        items: props.items,
                        command: props.command,
                        loading,
                        query: props.query,
                        ref: (handle: WikiLinkSuggestionListHandle | null) => {
                            listRef = handle;
                        },
                    }),
                );
            }

            return {
                onStart: (props) => {
                    container = document.createElement('div');
                    container.setAttribute(
                        'data-wiki-link-suggestion',
                        '',
                    );
                    document.body.appendChild(container);
                    root = createRoot(container);

                    // Initial render shows "type 3+ characters"
                    renderList(props, false);
                },

                onUpdate: (props) => {
                    renderList(props, false);
                },

                onKeyDown: ({ event }) => {
                    if (event.key === 'Escape') {
                        return true;
                    }

                    return (
                        listRef?.onKeyDown({ event }) ?? false
                    );
                },

                onExit: () => {
                    if (root) {
                        root.unmount();
                        root = null;
                    }
                    if (container) {
                        container.remove();
                        container = null;
                    }
                    listRef = null;

                    if (debounceTimer) {
                        clearTimeout(debounceTimer);
                        debounceTimer = null;
                    }
                },
            };
        },
    };
}
