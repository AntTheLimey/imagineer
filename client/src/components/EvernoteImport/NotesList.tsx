// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * NotesList component - displays a table of notes from a selected notebook
 * with an import button.
 */

import {
    Box,
    Button,
    Chip,
    CircularProgress,
    Paper,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import { ArrowBack as BackIcon } from '@mui/icons-material';
import type { NoteSummary } from '../../api/evernoteImport';

export interface NotesListProps {
    notebookName: string;
    notes: NoteSummary[];
    isLoading: boolean;
    onBack: () => void;
    onImport: () => void;
    isImporting: boolean;
}

/**
 * Formats an ISO date string to a readable format.
 */
function formatDate(dateString: string): string {
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
        });
    } catch {
        return dateString;
    }
}

/**
 * Renders a table of notes with title, dates, and tags.
 */
export default function NotesList({
    notebookName,
    notes,
    isLoading,
    onBack,
    onImport,
    isImporting,
}: NotesListProps) {
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
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Button
                        startIcon={<BackIcon />}
                        onClick={onBack}
                        disabled={isImporting}
                    >
                        Back
                    </Button>
                    <Typography variant="h6">{notebookName}</Typography>
                </Box>
                <Typography variant="body2" color="text.secondary">
                    {notes.length} note{notes.length !== 1 ? 's' : ''}
                </Typography>
            </Box>

            {notes.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                    This notebook is empty. Select a different notebook or add
                    notes to this notebook in Evernote.
                </Typography>
            ) : (
                <>
                    <TableContainer component={Paper} sx={{ mb: 2 }}>
                        <Table size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell>Title</TableCell>
                                    <TableCell>Created</TableCell>
                                    <TableCell>Modified</TableCell>
                                    <TableCell>Tags</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {notes.map((note) => (
                                    <TableRow key={note.noteLink}>
                                        <TableCell>
                                            <Typography
                                                variant="body2"
                                                sx={{
                                                    maxWidth: 300,
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    whiteSpace: 'nowrap',
                                                }}
                                            >
                                                {note.title}
                                            </Typography>
                                        </TableCell>
                                        <TableCell>
                                            {formatDate(note.created)}
                                        </TableCell>
                                        <TableCell>
                                            {formatDate(note.modified)}
                                        </TableCell>
                                        <TableCell>
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 0.5,
                                                    flexWrap: 'wrap',
                                                    maxWidth: 200,
                                                }}
                                            >
                                                {note.tags.slice(0, 3).map((tag) => (
                                                    <Chip
                                                        key={tag}
                                                        label={tag}
                                                        size="small"
                                                        variant="outlined"
                                                    />
                                                ))}
                                                {note.tags.length > 3 && (
                                                    <Chip
                                                        label={`+${note.tags.length - 3}`}
                                                        size="small"
                                                        variant="outlined"
                                                    />
                                                )}
                                            </Box>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>

                    <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                        <Button
                            variant="contained"
                            onClick={onImport}
                            disabled={isImporting || notes.length === 0}
                            startIcon={
                                isImporting ? (
                                    <CircularProgress size={20} />
                                ) : undefined
                            }
                        >
                            {isImporting ? 'Importing...' : 'Import All Notes'}
                        </Button>
                    </Box>
                </>
            )}
        </Box>
    );
}
