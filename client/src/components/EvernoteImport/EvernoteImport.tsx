// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * EvernoteImport component - workflow for importing notes from the local
 * Evernote application on MacOS.
 */

import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Alert,
    Box,
    Button,
    CircularProgress,
    Typography,
} from '@mui/material';
import { Refresh as RefreshIcon } from '@mui/icons-material';
import {
    useEvernoteStatus,
    useEvernoteNotebooks,
    useEvernoteNotes,
    useImportEvernoteLocal,
} from '../../hooks/useEvernoteImport';
import type { Notebook } from '../../api/evernoteImport';
import type { ImportResult } from '../../types';
import NotebookList from './NotebookList';
import NotesList from './NotesList';
import ImportProgress from './ImportProgress';

export interface EvernoteImportProps {
    campaignId: string;
}

type ImportStep = 'status' | 'notebooks' | 'notes' | 'results';

/**
 * Main component for importing notes from the local Evernote application.
 * Implements a multi-step workflow: status check, notebook selection,
 * notes preview, and import results.
 */
export default function EvernoteImport({ campaignId }: EvernoteImportProps) {
    const navigate = useNavigate();
    const [step, setStep] = useState<ImportStep>('status');
    const [selectedNotebook, setSelectedNotebook] = useState<Notebook | null>(
        null
    );
    const [importResult, setImportResult] = useState<ImportResult | null>(null);

    // Detect MacOS - memoize to avoid recalculating
    const isMacOS = useMemo(() => {
        return navigator.platform.toLowerCase().includes('mac');
    }, []);

    // Query hooks - always call them to follow Rules of Hooks
    const {
        data: status,
        isLoading: statusLoading,
        refetch: refetchStatus,
        error: statusError,
    } = useEvernoteStatus({
        enabled: isMacOS,
    });

    const {
        data: notebooks,
        isLoading: notebooksLoading,
        refetch: refetchNotebooks,
    } = useEvernoteNotebooks({
        enabled: isMacOS && (step === 'notebooks' || step === 'notes'),
    });

    const {
        data: notes,
        isLoading: notesLoading,
    } = useEvernoteNotes(selectedNotebook?.name ?? '', {
        enabled: isMacOS && step === 'notes' && !!selectedNotebook,
    });

    const importMutation = useImportEvernoteLocal();

    // Handle status check on mount and when status changes
    useEffect(() => {
        if (isMacOS && status?.available && step === 'status') {
            setStep('notebooks');
        }
    }, [isMacOS, status, step]);

    // Handle notebook selection
    const handleSelectNotebook = (notebook: Notebook) => {
        setSelectedNotebook(notebook);
        setStep('notes');
    };

    // Handle back to notebooks
    const handleBackToNotebooks = () => {
        setSelectedNotebook(null);
        setStep('notebooks');
    };

    // Handle import
    const handleImport = async () => {
        if (!selectedNotebook) return;

        try {
            const result = await importMutation.mutateAsync({
                campaignId,
                notebookName: selectedNotebook.name,
                autoDetectEntities: true,
            });
            setImportResult(result);
            setStep('results');
        } catch (error) {
            // Error is handled by the mutation
            console.error('Import failed:', error);
        }
    };

    // Handle view entities
    const handleViewEntities = () => {
        navigate(`/campaigns/${campaignId}/entities`);
    };

    // Handle import more
    const handleImportMore = () => {
        setImportResult(null);
        setSelectedNotebook(null);
        setStep('notebooks');
    };

    // Check if not on MacOS - render after all hooks are called
    if (!isMacOS) {
        return (
            <Alert severity="info">
                Evernote local import is only available on MacOS. On other
                platforms, please export your notes as .enex files and use the
                file import option.
            </Alert>
        );
    }

    // Render status check step
    if (step === 'status') {
        if (statusLoading) {
            return (
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        gap: 2,
                        py: 4,
                    }}
                >
                    <CircularProgress />
                    <Typography>Checking Evernote status...</Typography>
                </Box>
            );
        }

        if (statusError || !status?.available) {
            return (
                <Box>
                    <Alert
                        severity="warning"
                        sx={{ mb: 2 }}
                        action={
                            <Button
                                color="inherit"
                                size="small"
                                startIcon={<RefreshIcon />}
                                onClick={() => refetchStatus()}
                            >
                                Retry
                            </Button>
                        }
                    >
                        {status?.error ||
                            'Evernote is not running. Please open Evernote and try again.'}
                    </Alert>
                    <Typography variant="body2" color="text.secondary">
                        Make sure Evernote is installed and running on your Mac.
                        This feature uses AppleScript to communicate with the
                        Evernote application.
                    </Typography>
                </Box>
            );
        }
    }

    // Render notebooks step
    if (step === 'notebooks') {
        return (
            <NotebookList
                notebooks={notebooks ?? []}
                isLoading={notebooksLoading}
                onSelect={handleSelectNotebook}
                onRefresh={() => refetchNotebooks()}
            />
        );
    }

    // Render notes step
    if (step === 'notes' && selectedNotebook) {
        return (
            <Box>
                {importMutation.error && (
                    <Alert
                        severity="error"
                        sx={{ mb: 2 }}
                        onClose={() => importMutation.reset()}
                    >
                        Failed to import notes. Please try again.
                    </Alert>
                )}
                <NotesList
                    notebookName={selectedNotebook.name}
                    notes={notes ?? []}
                    isLoading={notesLoading}
                    onBack={handleBackToNotebooks}
                    onImport={handleImport}
                    isImporting={importMutation.isPending}
                />
            </Box>
        );
    }

    // Render results step
    if (step === 'results' && importResult) {
        return (
            <ImportProgress
                result={importResult}
                onViewEntities={handleViewEntities}
                onImportMore={handleImportMore}
            />
        );
    }

    // Fallback - should not reach here
    return null;
}
