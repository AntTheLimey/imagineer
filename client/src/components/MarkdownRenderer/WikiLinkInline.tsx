/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useTheme } from '@mui/material/styles';
import type { SxProps, Theme } from '@mui/material';
import Box from '@mui/material/Box';
import type { ReactNode } from 'react';

/**
 * Props for the WikiLinkInline component.
 *
 * The `entityName` and `displayText` properties are injected by the
 * remark-wiki-links plugin via `data.hProperties`. The `onEntityClick`
 * callback is passed through from the parent MarkdownRenderer.
 */
export interface WikiLinkInlineProps {
    /** The canonical entity name extracted from `[[Entity Name]]`. */
    entityName?: string;
    /** Display text shown to the user; defaults to entityName. */
    displayText?: string;
    /** Callback fired when the wiki link is clicked. */
    onEntityClick?: (name: string) => void;
    /** Children rendered by react-markdown (the display text). */
    children?: ReactNode;
}

/**
 * Renders a wiki link as a styled inline element in read mode.
 *
 * Displays a clickable span with a subtle background tint indicating
 * it is an entity reference. Clicking the element invokes the
 * `onEntityClick` callback with the entity name. Styling adapts to
 * both light and dark themes using MUI's theme palette.
 *
 * @param props - The component props.
 * @returns A React element containing the styled wiki link.
 *
 * @example
 * ```tsx
 * <WikiLinkInline
 *     entityName="Inspector Legrasse"
 *     onEntityClick={(name) => navigate(`/entities?search=${name}`)}
 * >
 *     Inspector Legrasse
 * </WikiLinkInline>
 * ```
 */
export default function WikiLinkInline({
    entityName,
    onEntityClick,
    children,
}: WikiLinkInlineProps) {
    const theme = useTheme();

    const isDark = theme.palette.mode === 'dark';

    const sx: SxProps<Theme> = {
        display: 'inline',
        backgroundColor: isDark
            ? 'rgba(144, 202, 249, 0.16)'
            : 'rgba(25, 118, 210, 0.08)',
        color: isDark
            ? theme.palette.primary.light
            : theme.palette.primary.dark,
        borderRadius: '4px',
        px: 0.5,
        py: 0.125,
        cursor: onEntityClick ? 'pointer' : 'default',
        fontWeight: 500,
        fontSize: 'inherit',
        lineHeight: 'inherit',
        transition: 'background-color 0.15s ease',
        '&:hover': onEntityClick
            ? {
                backgroundColor: isDark
                    ? 'rgba(144, 202, 249, 0.28)'
                    : 'rgba(25, 118, 210, 0.16)',
            }
            : undefined,
    };

    const handleClick = () => {
        if (onEntityClick && entityName) {
            onEntityClick(entityName);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleClick();
        }
    };

    return (
        <Box
            component="span"
            role={onEntityClick ? "button" : undefined}
            tabIndex={onEntityClick ? 0 : undefined}
            data-entity-name={entityName}
            data-testid="wiki-link"
            sx={sx}
            {...(onEntityClick
                ? {
                    onClick: handleClick,
                    onKeyDown: handleKeyDown,
                }
                : {})}
        >
            {children}
        </Box>
    );
}
