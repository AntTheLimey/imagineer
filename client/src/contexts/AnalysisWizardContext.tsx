// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/* eslint-disable react-refresh/only-export-components */

import {
    createContext,
    useContext,
    ReactNode,
} from 'react';
import type {
    AnalysisWizardState,
} from '../hooks/useAnalysisWizard';

const AnalysisWizardContext = createContext<
    AnalysisWizardState | null
>(null);

/**
 * Provides the analysis wizard state to descendant phase page components.
 *
 * The AnalysisWizard shell wraps its Outlet in this provider, passing the
 * return value of useAnalysisWizard so that phase pages (Identify, Revise,
 * Enrich) can access wizard state without calling the hook directly.
 */
export function AnalysisWizardProvider({
    value,
    children,
}: {
    value: AnalysisWizardState;
    children: ReactNode;
}) {
    return (
        <AnalysisWizardContext.Provider value={value}>
            {children}
        </AnalysisWizardContext.Provider>
    );
}

/**
 * Access the analysis wizard state from the nearest AnalysisWizardProvider.
 *
 * @returns The AnalysisWizardState containing job data, items, phase
 *          navigation, and actions.
 * @throws Error if called outside of an AnalysisWizardProvider.
 */
export function useWizardContext(): AnalysisWizardState {
    const ctx = useContext(AnalysisWizardContext);
    if (!ctx) {
        throw new Error(
            'useWizardContext must be used within AnalysisWizardProvider'
        );
    }
    return ctx;
}

export default AnalysisWizardContext;
