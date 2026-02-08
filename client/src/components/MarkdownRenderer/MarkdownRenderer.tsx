/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import ReactMarkdown from 'react-markdown';
import { Box } from '@mui/material';
import type { SxProps, Theme } from '@mui/material';

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
}

/**
 * Allowed HTML elements that ReactMarkdown may render.
 *
 * This whitelist prevents rendering of potentially dangerous elements
 * such as images, links, and iframes while permitting standard text
 * formatting and structural elements.
 */
const ALLOWED_ELEMENTS = [
    'p', 'strong', 'em', 'del',
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'ul', 'ol', 'li',
    'blockquote', 'hr', 'br',
    'code', 'pre',
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
 * @param props - The component props.
 * @returns A React element containing the rendered Markdown, or null
 *          if content is falsy.
 *
 * @example
 * ```tsx
 * <MarkdownRenderer content={entity.description} />
 *
 * <MarkdownRenderer content={entity.description} maxLines={3} />
 * ```
 */
export default function MarkdownRenderer({
    content,
    maxLines,
}: MarkdownRendererProps) {
    if (!content) {
        return null;
    }

    const clampSx: SxProps<Theme> = maxLines
        ? {
            display: '-webkit-box',
            WebkitLineClamp: maxLines,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
        }
        : {};

    return (
        <Box sx={[typographySx, clampSx] as SxProps<Theme>}>
            <ReactMarkdown
                allowedElements={ALLOWED_ELEMENTS}
                unwrapDisallowed
            >
                {content}
            </ReactMarkdown>
        </Box>
    );
}
