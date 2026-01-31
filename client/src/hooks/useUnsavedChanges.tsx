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
 * Confirmation dialog for unsaved changes.
 */
function ConfirmNavigationDialog({
    open,
    title = 'Unsaved Changes',
    message = 'You have unsaved changes. Are you sure you want to leave? Your changes will be lost.',
    confirmText = 'Leave',
    cancelText = 'Stay',
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
 * Hook for tracking dirty state and preventing navigation with unsaved changes.
 *
 * @param options - Configuration options
 * @returns Object containing dirty state, setters, and confirmation dialog
 *
 * @example
 * ```typescript
 * const {
 *   isDirty,
 *   setIsDirty,
 *   markDirty,
 *   clearDirty,
 *   checkUnsavedChanges,
 *   ConfirmDialog,
 * } = useUnsavedChanges({
 *   message: 'Your entity changes will be lost.',
 * });
 *
 * // Track changes
 * const handleFieldChange = (value) => {
 *   setFormData({ ...formData, field: value });
 *   markDirty();
 * };
 *
 * // Clear on successful save
 * const handleSave = async () => {
 *   await saveEntity(formData);
 *   clearDirty();
 * };
 *
 * // Handle back navigation
 * const handleBack = () => {
 *   if (!checkUnsavedChanges(() => navigate(-1))) {
 *     navigate(-1);
 *   }
 * };
 *
 * // Render the dialog
 * return (
 *   <>
 *     <Form />
 *     {ConfirmDialog}
 *   </>
 * );
 * ```
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
