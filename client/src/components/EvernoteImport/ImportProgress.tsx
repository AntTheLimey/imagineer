// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * ImportProgress component - displays import progress and results.
 */

import {
    Box,
    Button,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Paper,
    Typography,
} from '@mui/material';
import {
    CheckCircle as SuccessIcon,
    Error as ErrorIcon,
    Warning as WarningIcon,
} from '@mui/icons-material';
import type { ImportResult } from '../../types';

export interface ImportProgressProps {
    result: ImportResult;
    onViewEntities: () => void;
    onImportMore: () => void;
}

/**
 * Renders import results with entity counts, warnings, and errors.
 */
export default function ImportProgress({
    result,
    onViewEntities,
    onImportMore,
}: ImportProgressProps) {
    const hasWarnings = result.warnings.length > 0;
    const hasErrors = result.errors.length > 0;

    return (
        <Box>
            <Paper
                sx={{
                    p: 3,
                    mb: 2,
                    bgcolor: result.success ? 'success.light' : 'error.light',
                }}
            >
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 1,
                        mb: 2,
                    }}
                >
                    {result.success ? (
                        <SuccessIcon color="success" fontSize="large" />
                    ) : (
                        <ErrorIcon color="error" fontSize="large" />
                    )}
                    <Typography variant="h5">
                        {result.success ? 'Import Successful' : 'Import Failed'}
                    </Typography>
                </Box>

                <Typography variant="body1" sx={{ mb: 2 }}>
                    Created {result.entitiesCreated} entit
                    {result.entitiesCreated !== 1 ? 'ies' : 'y'} and{' '}
                    {result.relationshipsCreated} relationship
                    {result.relationshipsCreated !== 1 ? 's' : ''}.
                </Typography>

                {hasWarnings && (
                    <Box sx={{ mb: 2 }}>
                        <Typography
                            variant="subtitle2"
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 0.5,
                                mb: 1,
                            }}
                        >
                            <WarningIcon fontSize="small" color="warning" />
                            Warnings ({result.warnings.length})
                        </Typography>
                        <List dense disablePadding>
                            {result.warnings.map((warning, index) => (
                                <ListItem key={index} sx={{ py: 0, pl: 3 }}>
                                    <ListItemIcon sx={{ minWidth: 28 }}>
                                        <WarningIcon
                                            fontSize="small"
                                            color="warning"
                                        />
                                    </ListItemIcon>
                                    <ListItemText primary={warning} />
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                )}

                {hasErrors && (
                    <Box sx={{ mb: 2 }}>
                        <Typography
                            variant="subtitle2"
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 0.5,
                                mb: 1,
                            }}
                        >
                            <ErrorIcon fontSize="small" color="error" />
                            Errors ({result.errors.length})
                        </Typography>
                        <List dense disablePadding>
                            {result.errors.map((error, index) => (
                                <ListItem key={index} sx={{ py: 0, pl: 3 }}>
                                    <ListItemIcon sx={{ minWidth: 28 }}>
                                        <ErrorIcon
                                            fontSize="small"
                                            color="error"
                                        />
                                    </ListItemIcon>
                                    <ListItemText primary={error} />
                                </ListItem>
                            ))}
                        </List>
                    </Box>
                )}
            </Paper>

            <Box sx={{ display: 'flex', gap: 2 }}>
                <Button variant="contained" onClick={onViewEntities}>
                    View Entities
                </Button>
                <Button variant="outlined" onClick={onImportMore}>
                    Import More
                </Button>
            </Box>
        </Box>
    );
}
