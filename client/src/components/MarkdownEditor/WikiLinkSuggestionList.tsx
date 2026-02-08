/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import {
    forwardRef,
    useEffect,
    useImperativeHandle,
    useState,
} from 'react';
import {
    Box,
    Chip,
    CircularProgress,
    List,
    ListItemButton,
    ListItemText,
    Paper,
    Typography,
} from '@mui/material';
import type { EntityResolveResult } from '../../api/entityResolve';
import {
    entityTypeColors,
    formatEntityType,
} from '../EntitySelector/entityConstants';

/**
 * Props for the WikiLinkSuggestionList component.
 */
export interface WikiLinkSuggestionListProps {
    /** The list of matching entities to display. */
    items: EntityResolveResult[];
    /** Callback fired when an item is selected. */
    command: (item: EntityResolveResult) => void;
    /** Whether the API request is currently loading. */
    loading: boolean;
    /** The current search query (before any pipe character). */
    query: string;
}

/**
 * Handle exposed to the parent via useImperativeHandle for keyboard
 * navigation support required by @tiptap/suggestion.
 */
export interface WikiLinkSuggestionListHandle {
    onKeyDown: (props: { event: KeyboardEvent }) => boolean;
}

/**
 * Dropdown list component for wiki link entity suggestions.
 *
 * Renders matching entities with their type chips and supports
 * keyboard navigation (arrow keys, Enter, Escape). Displays
 * contextual messages when the query is too short, results are
 * loading, or no matches are found.
 *
 * Uses forwardRef and useImperativeHandle to expose a keyboard
 * handler to the @tiptap/suggestion plugin.
 */
const WikiLinkSuggestionList = forwardRef<
    WikiLinkSuggestionListHandle,
    WikiLinkSuggestionListProps
>(function WikiLinkSuggestionList(
    { items, command, loading, query },
    ref,
) {
    const [selectedIndex, setSelectedIndex] = useState(0);

    // Reset selection when items change
    useEffect(() => {
        setSelectedIndex(0);
    }, [items]);

    useImperativeHandle(ref, () => ({
        onKeyDown: ({ event }: { event: KeyboardEvent }) => {
            if (event.key === 'ArrowUp') {
                setSelectedIndex((prev) =>
                    prev <= 0 ? items.length - 1 : prev - 1,
                );
                return true;
            }

            if (event.key === 'ArrowDown') {
                setSelectedIndex((prev) =>
                    prev >= items.length - 1 ? 0 : prev + 1,
                );
                return true;
            }

            if (event.key === 'Enter') {
                if (items.length > 0 && items[selectedIndex]) {
                    command(items[selectedIndex]);
                }
                return true;
            }

            return false;
        },
    }));

    const selectItem = (index: number) => {
        if (items[index]) {
            command(items[index]);
        }
    };

    // Extract the entity search portion (before any pipe)
    const entityQuery = query.includes('|')
        ? query.split('|')[0].trim()
        : query;
    const tooShort = entityQuery.length < 3;

    return (
        <Paper
            elevation={8}
            sx={{
                minWidth: 280,
                maxWidth: 400,
                maxHeight: 300,
                overflowY: 'auto',
                py: 0.5,
            }}
        >
            {tooShort && (
                <Box sx={{ px: 2, py: 1.5 }}>
                    <Typography
                        variant="body2"
                        color="text.secondary"
                    >
                        Type 3+ characters to search...
                    </Typography>
                </Box>
            )}

            {!tooShort && loading && (
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 1,
                        px: 2,
                        py: 1.5,
                    }}
                >
                    <CircularProgress size={16} />
                    <Typography
                        variant="body2"
                        color="text.secondary"
                    >
                        Searching...
                    </Typography>
                </Box>
            )}

            {!tooShort && !loading && items.length === 0 && (
                <Box sx={{ px: 2, py: 1.5 }}>
                    <Typography
                        variant="body2"
                        color="text.secondary"
                    >
                        No matches found
                    </Typography>
                </Box>
            )}

            {!tooShort && items.length > 0 && (
                <List dense disablePadding>
                    {items.map((item, index) => (
                        <ListItemButton
                            key={item.id}
                            selected={index === selectedIndex}
                            onClick={() => selectItem(index)}
                            sx={{ px: 2, py: 0.75 }}
                        >
                            <ListItemText
                                primary={item.name}
                                primaryTypographyProps={{
                                    variant: 'body2',
                                }}
                            />
                            <Chip
                                label={formatEntityType(
                                    item.entityType,
                                )}
                                size="small"
                                color={
                                    entityTypeColors[
                                        item.entityType
                                    ]
                                }
                                variant="outlined"
                                sx={{ ml: 1, height: 20 }}
                            />
                        </ListItemButton>
                    ))}
                </List>
            )}
        </Paper>
    );
});

export default WikiLinkSuggestionList;
