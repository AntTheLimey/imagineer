// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * NotebookList component - displays a list of notebooks from the local
 * Evernote application for selection.
 */

import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    CircularProgress,
    Grid,
    IconButton,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Book as NotebookIcon,
    Refresh as RefreshIcon,
} from '@mui/icons-material';
import type { Notebook } from '../../api/evernoteImport';

export interface NotebookListProps {
    notebooks: Notebook[];
    isLoading: boolean;
    onSelect: (notebook: Notebook) => void;
    onRefresh: () => void;
}

/**
 * Renders a grid of notebook cards that can be selected.
 */
export default function NotebookList({
    notebooks,
    isLoading,
    onSelect,
    onRefresh,
}: NotebookListProps) {
    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                <CircularProgress />
            </Box>
        );
    }

    return (
        <Box>
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    mb: 2,
                }}
            >
                <Typography variant="h6">Select a Notebook</Typography>
                <Tooltip title="Refresh notebooks">
                    <IconButton onClick={onRefresh} disabled={isLoading}>
                        <RefreshIcon />
                    </IconButton>
                </Tooltip>
            </Box>

            {notebooks.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                    No notebooks found in Evernote. Create a notebook in Evernote
                    and try refreshing.
                </Typography>
            ) : (
                <Grid container spacing={2}>
                    {notebooks.map((notebook) => (
                        <Grid item xs={12} sm={6} md={4} key={notebook.name}>
                            <Card>
                                <CardActionArea
                                    onClick={() => onSelect(notebook)}
                                >
                                    <CardContent>
                                        <Box
                                            sx={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 1,
                                            }}
                                        >
                                            <NotebookIcon color="primary" />
                                            <Box sx={{ minWidth: 0, flex: 1 }}>
                                                <Typography
                                                    variant="subtitle1"
                                                    noWrap
                                                >
                                                    {notebook.name}
                                                </Typography>
                                                <Typography
                                                    variant="body2"
                                                    color="text.secondary"
                                                >
                                                    {notebook.noteCount} note
                                                    {notebook.noteCount !== 1
                                                        ? 's'
                                                        : ''}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </CardActionArea>
                            </Card>
                        </Grid>
                    ))}
                </Grid>
            )}
        </Box>
    );
}
