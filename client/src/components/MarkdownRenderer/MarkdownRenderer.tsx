/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
import { Box } from '@mui/material';
import type { SxProps, Theme } from '@mui/material';
import remarkWikiLinks from './remarkWikiLinks';
import WikiLinkInline from './WikiLinkInline';
import type { WikiLinkEntity } from './WikiLinkInline';

/**
 * Props for the MarkdownRenderer component.
 */
export interface MarkdownRendererProps {
    /** The Markdown content to render. */
    content: string | null | undefined;
    /**
     * Maximum number of visible lines before clamping.
     *
     * When set, applies CSS line clamping so the rendered output is
     * truncated with an ellipsis after the specified number of lines.
     */
    maxLines?: number;
    /**
     * Callback fired when a wiki link `[[Entity Name]]` is clicked.
     *
     * Receives the canonical entity name extracted from the wiki-link
     * syntax. This is the fallback handler used when `onEntityNavigate`
     * is not provided or the entity is not found in `entities`.
     */
    onEntityClick?: (name: string) => void;
    /**
     * Array of entities available for wiki link matching.
     *
     * When provided alongside `onEntityNavigate`, wiki links whose
     * names match an entity (case-insensitive) will navigate directly
     * by entity ID and display a hover popover preview.
     */
    entities?: WikiLinkEntity[];
    /**
     * Callback fired when a matched entity wiki link is clicked or
     * the popover "View" link is activated. Receives the entity ID.
     */
    onEntityNavigate?: (entityId: number) => void;
}

/**
 * Allowed HTML elements that ReactMarkdown may render.
 *
 * This whitelist prevents rendering of potentially dangerous elements
 * such as images, links, and iframes while permitting standard text
 * formatting and structural elements. The `wiki-link` entry allows
 * the custom element produced by the remarkWikiLinks plugin.
 */
const ALLOWED_ELEMENTS = [
    'p', 'strong', 'em', 'del',
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'ul', 'ol', 'li',
    'blockquote', 'hr', 'br',
    'code', 'pre',
    'wiki-link',
];

/**
 * Typography and element styles applied to the rendered Markdown.
 *
 * These styles ensure consistent appearance across the application
 * by mapping Markdown elements to MUI-compatible spacing and colors.
 */
const typographySx: SxProps<Theme> = {
    '& h1': { fontSize: '1.5rem', mt: 2, mb: 1 },
    '& h2': { fontSize: '1.25rem', mt: 2, mb: 1 },
    '& h3': { fontSize: '1.1rem', mt: 1.5, mb: 0.5 },
    '& p': { mt: 0, mb: 1 },
    '& ul, & ol': { mt: 0, mb: 1, pl: 3 },
    '& blockquote': {
        borderLeft: 3,
        borderColor: 'divider',
        pl: 2,
        ml: 0,
        color: 'text.secondary',
        fontStyle: 'italic',
    },
    '& hr': { my: 2, borderColor: 'divider' },
    '& code': {
        backgroundColor: 'action.hover',
        px: 0.5,
        borderRadius: 0.5,
        fontSize: '0.875em',
    },
};

/**
 * Renders Markdown content as styled HTML using react-markdown.
 *
 * Replaces the previous pattern of `sanitizeHtml()` with
 * `dangerouslySetInnerHTML` by parsing Markdown at render time with a
 * strict element whitelist. Disallowed elements are unwrapped so their
 * text content is preserved. An optional `maxLines` prop applies CSS
 * line clamping for preview contexts such as entity cards and search
 * snippets.
 *
 * Wiki-link syntax (`[[Entity Name]]` or `[[Entity Name|display text]]`)
 * is transformed by the remarkWikiLinks plugin into clickable inline
 * elements. When `onEntityClick` is provided, clicking a wiki link
 * invokes the callback with the entity name.
 *
 * When `entities` and `onEntityNavigate` are provided, wiki links that
 * match an entity name (case-insensitive) navigate directly by entity
 * ID and display a hover popover with the entity name, type, and a
 * description snippet.
 *
 * @param props - The component props.
 * @returns A React element containing the rendered Markdown, or null
 *          if content is falsy.
 *
 * @example
 * ```tsx
 * <MarkdownRenderer content={entity.description} />
 *
 * <MarkdownRenderer content={entity.description} maxLines={3} />
 *
 * <MarkdownRenderer
 *     content={session.notes}
 *     onEntityClick={(name) => console.log('Navigate to', name)}
 * />
 *
 * <MarkdownRenderer
 *     content={session.notes}
 *     entities={campaignEntities}
 *     onEntityNavigate={(id) => navigate(`/entities/${id}`)}
 * />
 * ```
 */
export default function MarkdownRenderer({
    content,
    maxLines,
    onEntityClick,
    entities,
    onEntityNavigate,
}: MarkdownRendererProps) {
    /**
     * Custom component mapping for react-markdown.
     *
     * The `wiki-link` element is produced by the remarkWikiLinks remark
     * plugin and rendered by WikiLinkInline. The callbacks and entity
     * data are captured via closure so that each wiki link instance
     * receives them as props.
     *
     * The explicit `any` in the return type is required because
     * react-markdown's `Components` type does not include custom
     * element names such as `wiki-link`.
     */
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const components: Record<string, React.ComponentType<any>> = useMemo(() => ({
        'wiki-link': ({
            entityName,
            displayText,
            children,
        }: {
            entityName?: string;
            displayText?: string;
            children?: React.ReactNode;
        }) => (
            <WikiLinkInline
                entityName={entityName}
                displayText={displayText}
                onEntityClick={onEntityClick}
                entities={entities}
                onEntityNavigate={onEntityNavigate}
            >
                {children}
            </WikiLinkInline>
        ),
    }), [onEntityClick, entities, onEntityNavigate]);

    const clampSx: SxProps<Theme> = maxLines
        ? {
            display: '-webkit-box',
            WebkitLineClamp: maxLines,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
        }
        : {};

    if (!content) {
        return null;
    }

    return (
        <Box sx={[typographySx, clampSx] as SxProps<Theme>}>
            <ReactMarkdown
                allowedElements={ALLOWED_ELEMENTS}
                unwrapDisallowed
                remarkPlugins={[remarkWikiLinks]}
                components={components}
            >
                {content}
            </ReactMarkdown>
        </Box>
    );
}
