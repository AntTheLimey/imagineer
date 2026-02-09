// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * EntityLogSection - Displays a chronological list of log entries for an
 * entity with inline create, edit, and delete capabilities.
 *
 * Renders below the Relationships section in the EntityEditor.
 */

import { useState, useCallback } from 'react';
import {
    Box,
    Typography,
    Button,
    TextField,
    IconButton,
    List,
    ListItem,
    ListItemText,
    Paper,
    Stack,
    Divider,
    Tooltip,
    CircularProgress,
} from '@mui/material';
import {
    Add,
    Edit,
    Delete,
    Save,
    Cancel,
} from '@mui/icons-material';
import {
    useEntityLogs,
    useCreateEntityLog,
    useUpdateEntityLog,
    useDeleteEntityLog,
} from '../../hooks/useEntityLog';

export interface EntityLogSectionProps {
    campaignId: number;
    entityId: number;
}

/**
 * Display and manage log entries for a campaign entity.
 *
 * Provides an inline form for adding new entries and inline editing for
 * existing entries, with delete support. Each log entry can optionally
 * include an in-game date via the `occurredAt` field.
 *
 * @param props.campaignId - The campaign the entity belongs to.
 * @param props.entityId - The entity to display log entries for.
 * @returns The React element for the entity log section.
 */
export default function EntityLogSection({ campaignId, entityId }: EntityLogSectionProps) {
    const { data: logs, isLoading } = useEntityLogs(campaignId, entityId);
    const createLog = useCreateEntityLog(campaignId, entityId);
    const updateLog = useUpdateEntityLog(campaignId, entityId);
    const deleteLog = useDeleteEntityLog(campaignId, entityId);

    // State for inline add form
    const [showAddForm, setShowAddForm] = useState(false);
    const [newContent, setNewContent] = useState('');
    const [newOccurredAt, setNewOccurredAt] = useState('');

    // State for inline edit
    const [editingId, setEditingId] = useState<number | null>(null);
    const [editContent, setEditContent] = useState('');
    const [editOccurredAt, setEditOccurredAt] = useState('');

    const handleAdd = useCallback(() => {
        if (!newContent.trim()) return;
        createLog.mutate(
            {
                content: newContent.trim(),
                occurredAt: newOccurredAt.trim() || undefined,
            },
            {
                onSuccess: () => {
                    setNewContent('');
                    setNewOccurredAt('');
                    setShowAddForm(false);
                },
            }
        );
    }, [newContent, newOccurredAt, createLog]);

    const handleStartEdit = useCallback((log: { id: number; content: string; occurredAt?: string }) => {
        setEditingId(log.id);
        setEditContent(log.content);
        setEditOccurredAt(log.occurredAt ?? '');
    }, []);

    const handleSaveEdit = useCallback(() => {
        if (editingId === null || !editContent.trim()) return;
        updateLog.mutate(
            {
                logId: editingId,
                input: {
                    content: editContent.trim(),
                    occurredAt: editOccurredAt.trim() || undefined,
                },
            },
            {
                onSuccess: () => {
                    setEditingId(null);
                    setEditContent('');
                    setEditOccurredAt('');
                },
            }
        );
    }, [editingId, editContent, editOccurredAt, updateLog]);

    const handleCancelEdit = useCallback(() => {
        setEditingId(null);
        setEditContent('');
        setEditOccurredAt('');
    }, []);

    const handleDelete = useCallback((logId: number) => {
        deleteLog.mutate(logId);
    }, [deleteLog]);

    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    return (
        <Box>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="subtitle1" fontWeight="medium">
                    Event Log
                </Typography>
                <Button
                    size="small"
                    startIcon={<Add />}
                    onClick={() => setShowAddForm(!showAddForm)}
                    sx={{ textTransform: 'none' }}
                >
                    Add Entry
                </Button>
            </Box>

            {/* Add form */}
            {showAddForm && (
                <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
                    <Stack spacing={2}>
                        <TextField
                            label="Log Entry"
                            value={newContent}
                            onChange={(e) => setNewContent(e.target.value)}
                            multiline
                            rows={2}
                            size="small"
                            fullWidth
                            autoFocus
                        />
                        <TextField
                            label="Occurred At (optional, in-game date)"
                            value={newOccurredAt}
                            onChange={(e) => setNewOccurredAt(e.target.value)}
                            size="small"
                            fullWidth
                            placeholder="e.g., Spring 1925, Day 3"
                        />
                        <Stack direction="row" spacing={1}>
                            <Button
                                variant="contained"
                                size="small"
                                startIcon={<Save />}
                                onClick={handleAdd}
                                disabled={!newContent.trim() || createLog.isPending}
                            >
                                Save
                            </Button>
                            <Button
                                size="small"
                                startIcon={<Cancel />}
                                onClick={() => {
                                    setShowAddForm(false);
                                    setNewContent('');
                                    setNewOccurredAt('');
                                }}
                            >
                                Cancel
                            </Button>
                        </Stack>
                    </Stack>
                </Paper>
            )}

            {/* Log entries list */}
            {logs && logs.length > 0 ? (
                <List disablePadding>
                    {logs.map((log, index) => (
                        <Box key={log.id}>
                            {index > 0 && <Divider />}
                            <ListItem
                                sx={{ px: 0 }}
                                secondaryAction={
                                    editingId !== log.id ? (
                                        <Box sx={{ display: 'flex', gap: 0.5 }}>
                                            <Tooltip title="Edit">
                                                <IconButton
                                                    size="small"
                                                    onClick={() => handleStartEdit(log)}
                                                >
                                                    <Edit fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip title="Delete">
                                                <IconButton
                                                    size="small"
                                                    color="error"
                                                    onClick={() => handleDelete(log.id)}
                                                    disabled={deleteLog.isPending}
                                                >
                                                    <Delete fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                        </Box>
                                    ) : undefined
                                }
                            >
                                {editingId === log.id ? (
                                    <Box sx={{ width: '100%', pr: 2 }}>
                                        <Stack spacing={1}>
                                            <TextField
                                                value={editContent}
                                                onChange={(e) => setEditContent(e.target.value)}
                                                multiline
                                                rows={2}
                                                size="small"
                                                fullWidth
                                                autoFocus
                                            />
                                            <TextField
                                                label="Occurred At"
                                                value={editOccurredAt}
                                                onChange={(e) => setEditOccurredAt(e.target.value)}
                                                size="small"
                                                fullWidth
                                            />
                                            <Stack direction="row" spacing={1}>
                                                <Button
                                                    variant="contained"
                                                    size="small"
                                                    startIcon={<Save />}
                                                    onClick={handleSaveEdit}
                                                    disabled={!editContent.trim() || updateLog.isPending}
                                                >
                                                    Save
                                                </Button>
                                                <Button
                                                    size="small"
                                                    startIcon={<Cancel />}
                                                    onClick={handleCancelEdit}
                                                >
                                                    Cancel
                                                </Button>
                                            </Stack>
                                        </Stack>
                                    </Box>
                                ) : (
                                    <ListItemText
                                        primary={log.content}
                                        secondary={
                                            log.occurredAt
                                                ? `Occurred: ${log.occurredAt}`
                                                : undefined
                                        }
                                        sx={{ pr: 8 }}
                                    />
                                )}
                            </ListItem>
                        </Box>
                    ))}
                </List>
            ) : (
                <Typography variant="body2" color="text.secondary">
                    No log entries yet.
                </Typography>
            )}
        </Box>
    );
}
