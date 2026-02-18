// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/* eslint-disable react-refresh/only-export-components */

/**
 * Hook for tracking dirty state and preventing navigation with unsaved changes.
 *
 * Uses the beforeunload event for browser navigation/refresh. For in-app
 * navigation, provides a confirmation mechanism that can be used with
 * navigation handlers.
 *
 * Note: React Router's useBlocker requires a data router (createBrowserRouter),
 * which would require refactoring App.tsx. This implementation uses beforeunload
 * for browser navigation and provides manual confirmation for in-app navigation.
 */

import { useState, useEffect, useCallback, ReactNode } from 'react';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
} from '@mui/material';

/**
 * Props for the confirmation dialog component.
 */
interface ConfirmDialogProps {
    /** Whether the dialog is open */
    open: boolean;
    /** Title for the dialog */
    title?: string;
    /** Message to display in the dialog */
    message?: string;
    /** Text for the confirm button */
    confirmText?: string;
    /** Text for the cancel button */
    cancelText?: string;
    /** Handler for confirming navigation */
    onConfirm: () => void;
    /** Handler for canceling navigation */
    onCancel: () => void;
}

/**
 * Renders a modal confirmation dialog that prompts the user about unsaved changes.
 *
 * @param open - Whether the dialog is visible
 * @param title - Dialog title (defaults to `"Unsaved Changes"`)
 * @param message - Body text shown in the dialog (defaults to a warning about lost changes)
 * @param confirmText - Text for the confirm action button (defaults to `"Leave"`)
 * @param cancelText - Text for the cancel action button (defaults to `"Stay"`)
 * @param onConfirm - Called when the user confirms navigation
 * @param onCancel - Called when the user cancels or closes the dialog
 * @returns The confirmation dialog React element
 */
function ConfirmNavigationDialog({
    open,
    title = 'Unsaved Changes',
    message = 'You have unsaved changes that have not been saved to the server. If you leave, a draft will be kept in your browser so you can recover it next time you edit.',
    confirmText = 'Leave Without Saving',
    cancelText = 'Stay and Keep Editing',
    onConfirm,
    onCancel,
}: ConfirmDialogProps) {
    return (
        <Dialog open={open} onClose={onCancel}>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <DialogContentText>{message}</DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel} variant="outlined">
                    {cancelText}
                </Button>
                <Button onClick={onConfirm} variant="contained" color="error">
                    {confirmText}
                </Button>
            </DialogActions>
        </Dialog>
    );
}

/**
 * Options for the useUnsavedChanges hook.
 */
interface UseUnsavedChangesOptions {
    /** Initial dirty state (default: false) */
    initialDirty?: boolean;
    /** Message to show in the confirmation dialog */
    message?: string;
    /** Whether blocking is enabled (default: true when dirty) */
    enabled?: boolean;
}

/**
 * Return type for the useUnsavedChanges hook.
 */
interface UseUnsavedChangesReturn {
    /** Whether there are unsaved changes */
    isDirty: boolean;
    /** Set the dirty state */
    setIsDirty: (dirty: boolean) => void;
    /** Mark as dirty */
    markDirty: () => void;
    /** Clear dirty state */
    clearDirty: () => void;
    /** Show the confirmation dialog */
    showConfirmDialog: () => void;
    /** Hide the confirmation dialog */
    hideConfirmDialog: () => void;
    /** Check if should block and show dialog if needed. Returns true if blocked. */
    checkUnsavedChanges: (onConfirm?: () => void) => boolean;
    /** The confirmation dialog component to render */
    ConfirmDialog: ReactNode;
}

/**
 * Track unsaved changes and block navigation by showing a confirmation dialog when needed.
 *
 * Provides utilities to mark/clear dirty state, programmatically show or hide the in-app
 * confirmation dialog, check and optionally block navigation when there are unsaved changes,
 * and a React node to render the confirmation dialog. When enabled and dirty, browser unloads
 * are also prevented.
 *
 * @param options - Configuration options
 * @param options.initialDirty - Whether the state starts as dirty (default: `false`)
 * @param options.message - Message to display inside the confirmation dialog
 * @param options.enabled - Whether blocking behavior is active (default: `true`)
 * @returns An object containing:
 * - `isDirty` — whether there are unsaved changes
 * - `setIsDirty` — setter to update the dirty state
 * - `markDirty` — convenience to set the state to dirty
 * - `clearDirty` — convenience to clear the dirty state
 * - `showConfirmDialog` / `hideConfirmDialog` — control the in-app confirmation dialog
 * - `checkUnsavedChanges` — when called with an optional `onConfirm` callback, opens the dialog and returns `true` if navigation was blocked (dialog shown), `false` otherwise
 * - `ConfirmDialog` — a React node that renders the confirmation dialog and should be included in the component tree
 */
export function useUnsavedChanges({
    initialDirty = false,
    message,
    enabled = true,
}: UseUnsavedChangesOptions = {}): UseUnsavedChangesReturn {
    const [isDirty, setIsDirtyState] = useState(initialDirty);
    const [dialogOpen, setDialogOpen] = useState(false);
    const [pendingAction, setPendingAction] = useState<(() => void) | null>(
        null
    );

    /**
     * Set the dirty state.
     */
    const setIsDirty = useCallback((dirty: boolean) => {
        setIsDirtyState(dirty);
    }, []);

    /**
     * Mark as dirty.
     */
    const markDirty = useCallback(() => {
        setIsDirtyState(true);
    }, []);

    /**
     * Clear dirty state.
     */
    const clearDirty = useCallback(() => {
        setIsDirtyState(false);
    }, []);

    /**
     * Show the confirmation dialog.
     */
    const showConfirmDialog = useCallback(() => {
        setDialogOpen(true);
    }, []);

    /**
     * Hide the confirmation dialog.
     */
    const hideConfirmDialog = useCallback(() => {
        setDialogOpen(false);
        setPendingAction(null);
    }, []);

    /**
     * Check if should block navigation and show dialog if needed.
     * Returns true if blocked (dialog shown), false if can proceed.
     */
    const checkUnsavedChanges = useCallback(
        (onConfirm?: () => void): boolean => {
            if (!enabled || !isDirty) {
                return false;
            }
            setPendingAction(() => onConfirm || null);
            setDialogOpen(true);
            return true;
        },
        [enabled, isDirty]
    );

    /**
     * Handle dialog confirmation.
     */
    const handleConfirm = useCallback(() => {
        setDialogOpen(false);
        setIsDirtyState(false);
        if (pendingAction) {
            pendingAction();
        }
        setPendingAction(null);
    }, [pendingAction]);

    /**
     * Handle dialog cancellation.
     */
    const handleCancel = useCallback(() => {
        setDialogOpen(false);
        setPendingAction(null);
    }, []);

    // Handle browser beforeunload event
    useEffect(() => {
        if (!enabled || !isDirty) {
            return;
        }

        const handleBeforeUnload = (event: BeforeUnloadEvent) => {
            event.preventDefault();
            // Modern browsers ignore custom messages, but we still need to set
            // returnValue for the dialog to appear
            event.returnValue = '';
            return '';
        };

        window.addEventListener('beforeunload', handleBeforeUnload);

        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, [enabled, isDirty]);

    // Build the confirmation dialog
    const ConfirmDialog = (
        <ConfirmNavigationDialog
            open={dialogOpen}
            message={message}
            onConfirm={handleConfirm}
            onCancel={handleCancel}
        />
    );

    return {
        isDirty,
        setIsDirty,
        markDirty,
        clearDirty,
        showConfirmDialog,
        hideConfirmDialog,
        checkUnsavedChanges,
        ConfirmDialog,
    };
}