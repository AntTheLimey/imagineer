/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useRef, useState } from 'react';
import {
    Box,
    Chip,
    CircularProgress,
    IconButton,
    List,
    ListItemButton,
    ListItemText,
    Popover,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import { InsertLink } from '@mui/icons-material';
import type { Editor } from '@tiptap/react';
import { useEntityResolve } from '../../hooks/useEntityResolve';
import {
    entityTypeColors,
    formatEntityType,
} from '../EntitySelector/entityConstants';

/**
 * Props for the WikiLinkToolbarButton component.
 */
interface WikiLinkToolbarButtonProps {
    /** The TipTap editor instance. */
    editor: Editor;
    /** Campaign ID for entity resolution. */
    campaignId?: number;
}

/**
 * A toolbar button that opens a popover for searching and inserting
 * wiki links into the editor.
 *
 * Provides an alternative to the `[[` autocomplete approach by
 * presenting a search interface anchored to a toolbar icon button.
 * Uses the useEntityResolve hook with built-in debouncing to query
 * matching entities. Selecting a result inserts a WikiLink node at
 * the current cursor position.
 *
 * @param props - The component props.
 * @returns A React element containing the wiki link toolbar button
 *     and its search popover, or null if no campaignId is provided.
 */
export default function WikiLinkToolbarButton({
    editor,
    campaignId,
}: WikiLinkToolbarButtonProps) {
    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const searchFieldRef = useRef<HTMLInputElement>(null);

    const { data: results, isLoading } = useEntityResolve(
        campaignId ?? 0,
        searchTerm,
    );

    if (!campaignId) {
        return null;
    }

    const open = Boolean(anchorEl);

    const handleOpen = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const handleClose = () => {
        setAnchorEl(null);
        setSearchTerm('');
    };

    const handleSelectEntity = (entityName: string) => {
        editor
            .chain()
            .focus()
            .insertContent({
                type: 'wikiLink',
                attrs: {
                    entityName,
                    displayText: null,
                },
            })
            .run();
        handleClose();
    };

    const showHint = searchTerm.length > 0 && searchTerm.length < 3;
    const showNoResults =
        searchTerm.length >= 3 &&
        !isLoading &&
        (!results || results.length === 0);
    const showResults =
        searchTerm.length >= 3 && results && results.length > 0;

    return (
        <>
            <Tooltip title="Insert Wiki Link">
                <IconButton
                    size="small"
                    onClick={handleOpen}
                    color={open ? 'primary' : 'default'}
                    aria-label="Insert wiki link"
                >
                    <InsertLink fontSize="small" />
                </IconButton>
            </Tooltip>
            <Popover
                open={open}
                anchorEl={anchorEl}
                onClose={handleClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'left',
                }}
                slotProps={{
                    paper: {
                        sx: { width: 320, maxHeight: 400 },
                    },
                }}
                TransitionProps={{
                    onEntered: () => {
                        searchFieldRef.current?.focus();
                    },
                }}
            >
                <Box sx={{ p: 2, pb: 1 }}>
                    <TextField
                        inputRef={searchFieldRef}
                        fullWidth
                        size="small"
                        placeholder="Search entities..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        autoComplete="off"
                    />
                </Box>
                <Box
                    sx={{
                        px: 2,
                        pb: 2,
                        maxHeight: 300,
                        overflowY: 'auto',
                    }}
                >
                    {isLoading && searchTerm.length >= 3 && (
                        <Box
                            sx={{
                                display: 'flex',
                                justifyContent: 'center',
                                py: 2,
                            }}
                        >
                            <CircularProgress size={24} />
                        </Box>
                    )}
                    {showHint && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ py: 1 }}
                        >
                            Type 3+ characters...
                        </Typography>
                    )}
                    {showNoResults && (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ py: 1 }}
                        >
                            No entities found
                        </Typography>
                    )}
                    {showResults && (
                        <List dense disablePadding>
                            {results.map((entity) => (
                                <ListItemButton
                                    key={entity.id}
                                    onClick={() =>
                                        handleSelectEntity(entity.name)
                                    }
                                    sx={{ borderRadius: 1 }}
                                >
                                    <ListItemText
                                        primary={entity.name}
                                    />
                                    <Chip
                                        label={formatEntityType(
                                            entity.entityType,
                                        )}
                                        color={
                                            entityTypeColors[
                                                entity.entityType
                                            ] ?? 'default'
                                        }
                                        size="small"
                                        sx={{ ml: 1 }}
                                    />
                                </ListItemButton>
                            ))}
                        </List>
                    )}
                </Box>
            </Popover>
        </>
    );
}
