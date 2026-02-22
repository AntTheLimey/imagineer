/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * Wraps untagged entity name occurrences in wiki-link syntax ([[name]]).
 * Names that already appear inside wiki-links are skipped.
 *
 * @param text - The text to scan for entity names.
 * @param entityNames - The entity names to match.
 * @returns The text with matching names wrapped in wiki-links.
 */
export function autoTagWikiLinks(
    text: string,
    entityNames: string[],
): string {
    if (entityNames.length === 0) return text;

    let result = text;
    for (const name of entityNames) {
        // Match the name when it is NOT already inside [[ ]].
        const escaped = name.replace(
            /[.*+?^${}()|[\]\\]/g, '\\$&',
        );
        const pattern = new RegExp(
            `(?<!\\[\\[)\\b(${escaped})\\b(?!\\]\\])`,
            'gi',
        );
        result = result.replace(pattern, '[[$1]]');
    }
    return result;
}
