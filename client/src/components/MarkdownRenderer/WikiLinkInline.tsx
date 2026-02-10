/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useRef, useState, useCallback } from 'react';
import { useTheme } from '@mui/material/styles';
import type { SxProps, Theme } from '@mui/material';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Link from '@mui/material/Link';
import Popover from '@mui/material/Popover';
import Typography from '@mui/material/Typography';
import type { ReactNode } from 'react';

/**
 * Lightweight entity record supplied by the parent MarkdownRenderer.
 *
 * Contains just enough data for direct navigation and the hover
 * popover preview.
 */
export interface WikiLinkEntity {
    id: number;
    name: string;
    entityType: string;
    description?: string | null;
}

/**
 * Props for the WikiLinkInline component.
 *
 * The `entityName` and `displayText` properties are injected by the
 * remark-wiki-links plugin via `data.hProperties`. The `onEntityClick`
 * callback is passed through from the parent MarkdownRenderer.
 *
 * When `entities` and `onEntityNavigate` are supplied, the component
 * supports direct navigation by entity ID and a hover popover preview.
 */
export interface WikiLinkInlineProps {
    /** The canonical entity name extracted from `[[Entity Name]]`. */
    entityName?: string;
    /** Display text shown to the user; defaults to entityName. */
    displayText?: string;
    /** Callback fired when the wiki link is clicked (fallback). */
    onEntityClick?: (name: string) => void;
    /**
     * Array of entities available for matching. When provided alongside
     * `onEntityNavigate`, enables direct navigation and hover popover.
     */
    entities?: WikiLinkEntity[];
    /**
     * Callback fired when a matched entity link is clicked or the
     * popover "View" link is activated. Receives the entity ID.
     */
    onEntityNavigate?: (entityId: number) => void;
    /** Children rendered by react-markdown (the display text). */
    children?: ReactNode;
}

/** Delay in milliseconds before the popover opens on hover. */
const POPOVER_OPEN_DELAY = 300;

/** Delay in milliseconds before the popover closes on mouse leave. */
const POPOVER_CLOSE_DELAY = 200;

/** Maximum number of characters shown in the description snippet. */
const DESCRIPTION_SNIPPET_LENGTH = 150;

/**
 * Truncates a description string to approximately the given length,
 * breaking at a word boundary where possible.
 */
function truncateDescription(
    text: string,
    maxLength: number
): string {
    if (text.length <= maxLength) {
        return text;
    }
    const truncated = text.slice(0, maxLength);
    const lastSpace = truncated.lastIndexOf(' ');
    const breakPoint = lastSpace > maxLength * 0.5
        ? lastSpace
        : maxLength;
    return truncated.slice(0, breakPoint) + '\u2026';
}

/**
 * Renders a wiki link as a styled inline element in read mode.
 *
 * Displays a clickable span with a subtle background tint indicating
 * it is an entity reference. Clicking the element invokes the
 * `onEntityNavigate` callback with the matched entity ID when the
 * entity is found in the `entities` array, or falls back to the
 * `onEntityClick` callback with the entity name.
 *
 * When `entities` is provided and the wiki link text matches an entity
 * name (case-insensitive), hovering over the link displays a popover
 * with the entity name, type, and a description snippet.
 *
 * Styling adapts to both light and dark themes using MUI's theme
 * palette.
 *
 * @param props - The component props.
 * @returns A React element containing the styled wiki link.
 *
 * @example
 * ```tsx
 * <WikiLinkInline
 *     entityName="Inspector Legrasse"
 *     entities={[{ id: 1, name: 'Inspector Legrasse', entityType: 'npc' }]}
 *     onEntityNavigate={(id) => navigate(`/entities/${id}`)}
 * >
 *     Inspector Legrasse
 * </WikiLinkInline>
 * ```
 */
export default function WikiLinkInline({
    entityName,
    onEntityClick,
    entities,
    onEntityNavigate,
    children,
}: WikiLinkInlineProps) {
    const theme = useTheme();
    const isDark = theme.palette.mode === 'dark';

    const [popoverOpen, setPopoverOpen] = useState(false);
    const anchorRef = useRef<HTMLSpanElement | null>(null);
    const openTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
        null
    );
    const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
        null
    );

    /**
     * Find the matching entity for this wiki link. The match is
     * case-insensitive on the entity name.
     */
    const matchedEntity = entities?.find(
        (e) => e.name.toLowerCase() === entityName?.toLowerCase()
    );

    const hasClickHandler = !!(
        (matchedEntity && onEntityNavigate) ||
        onEntityClick
    );

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
        cursor: hasClickHandler ? 'pointer' : 'default',
        fontWeight: 500,
        fontSize: 'inherit',
        lineHeight: 'inherit',
        transition: 'background-color 0.15s ease',
        '&:hover': hasClickHandler
            ? {
                backgroundColor: isDark
                    ? 'rgba(144, 202, 249, 0.28)'
                    : 'rgba(25, 118, 210, 0.16)',
            }
            : undefined,
    };

    const handleClick = () => {
        if (matchedEntity && onEntityNavigate) {
            onEntityNavigate(matchedEntity.id);
        } else if (onEntityClick && entityName) {
            onEntityClick(entityName);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            handleClick();
        }
    };

    /* ----------------------------------------------------------------
     * Popover hover handlers
     * ----------------------------------------------------------------
     * Both the trigger element and the popover itself share enter/leave
     * handlers. A delayed open prevents flicker on casual mouse-overs,
     * and a delayed close allows the user to move the cursor from the
     * trigger into the popover without it disappearing.
     * --------------------------------------------------------------- */

    const cancelTimers = useCallback(() => {
        if (openTimerRef.current) {
            clearTimeout(openTimerRef.current);
            openTimerRef.current = null;
        }
        if (closeTimerRef.current) {
            clearTimeout(closeTimerRef.current);
            closeTimerRef.current = null;
        }
    }, []);

    const scheduleOpen = useCallback(() => {
        cancelTimers();
        openTimerRef.current = setTimeout(() => {
            setPopoverOpen(true);
        }, POPOVER_OPEN_DELAY);
    }, [cancelTimers]);

    const scheduleClose = useCallback(() => {
        cancelTimers();
        closeTimerRef.current = setTimeout(() => {
            setPopoverOpen(false);
        }, POPOVER_CLOSE_DELAY);
    }, [cancelTimers]);

    const handleTriggerMouseEnter = () => {
        if (matchedEntity) {
            scheduleOpen();
        }
    };

    const handleTriggerMouseLeave = () => {
        scheduleClose();
    };

    const handlePopoverMouseEnter = () => {
        cancelTimers();
    };

    const handlePopoverMouseLeave = () => {
        scheduleClose();
    };

    const handleViewClick = (
        e: React.MouseEvent<HTMLAnchorElement>
    ) => {
        e.preventDefault();
        if (matchedEntity && onEntityNavigate) {
            setPopoverOpen(false);
            onEntityNavigate(matchedEntity.id);
        }
    };

    return (
        <>
            <Box
                ref={anchorRef}
                component="span"
                role={hasClickHandler ? 'button' : undefined}
                tabIndex={hasClickHandler ? 0 : undefined}
                data-entity-name={entityName}
                data-testid="wiki-link"
                sx={sx}
                onMouseEnter={handleTriggerMouseEnter}
                onMouseLeave={handleTriggerMouseLeave}
                {...(hasClickHandler
                    ? {
                        onClick: handleClick,
                        onKeyDown: handleKeyDown,
                    }
                    : {})}
            >
                {children}
            </Box>
            {matchedEntity && (
                <Popover
                    open={popoverOpen}
                    anchorEl={anchorRef.current}
                    anchorOrigin={{
                        vertical: 'bottom',
                        horizontal: 'left',
                    }}
                    transformOrigin={{
                        vertical: 'top',
                        horizontal: 'left',
                    }}
                    disableRestoreFocus
                    sx={{ pointerEvents: 'none' }}
                    slotProps={{
                        paper: {
                            sx: {
                                pointerEvents: 'auto',
                                p: 1.5,
                                maxWidth: 320,
                            },
                            onMouseEnter: handlePopoverMouseEnter,
                            onMouseLeave: handlePopoverMouseLeave,
                        },
                    }}
                >
                    <Box
                        sx={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 1,
                            mb:
                                matchedEntity.description
                                    ? 0.75
                                    : 0,
                        }}
                    >
                        <Typography
                            variant="subtitle2"
                            sx={{ fontWeight: 700 }}
                        >
                            {matchedEntity.name}
                        </Typography>
                        <Chip
                            label={matchedEntity.entityType}
                            size="small"
                            variant="outlined"
                        />
                    </Box>
                    {matchedEntity.description && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ mb: 0.75 }}
                        >
                            {truncateDescription(
                                matchedEntity.description,
                                DESCRIPTION_SNIPPET_LENGTH
                            )}
                        </Typography>
                    )}
                    {onEntityNavigate && (
                        <Link
                            component="a"
                            href="#"
                            variant="body2"
                            onClick={handleViewClick}
                            data-testid="wiki-link-popover-view"
                            sx={{ cursor: 'pointer' }}
                        >
                            View
                        </Link>
                    )}
                </Popover>
            )}
        </>
    );
}
