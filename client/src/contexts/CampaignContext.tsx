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
    useState,
    useEffect,
    useCallback,
    ReactNode,
} from 'react';
import { useCampaign } from '../hooks';
import type { Campaign } from '../types';

const CAMPAIGN_STORAGE_KEY = 'imagineer_current_campaign_id';

/**
 * Campaign context value providing current campaign state and selection.
 */
interface CampaignContextValue {
    /** ID of the currently selected campaign */
    currentCampaignId: string | null;
    /** Full campaign object for the current campaign */
    currentCampaign: Campaign | null;
    /** Whether the campaign data is currently loading */
    isLoading: boolean;
    /** Error if campaign fetch failed */
    error: Error | null;
    /** Set the current campaign by ID (persists to localStorage) */
    setCurrentCampaignId: (id: string | null) => void;
    /** Clear the current campaign selection */
    clearCurrentCampaign: () => void;
}

const CampaignContext = createContext<CampaignContextValue | undefined>(
    undefined
);

interface CampaignProviderProps {
    children: ReactNode;
}

/**
 * Provides campaign selection state to descendant components via React context.
 *
 * Persists the selected campaign ID to localStorage and fetches the full
 * campaign object using React Query. The context value includes the current
 * campaign ID and object, loading state, and functions to change selection.
 */
export function CampaignProvider({ children }: CampaignProviderProps) {
    const [currentCampaignId, setCurrentCampaignIdState] = useState<
        string | null
    >(() => {
        // Initialize from localStorage
        const stored = localStorage.getItem(CAMPAIGN_STORAGE_KEY);
        return stored ?? null;
    });

    // Fetch the full campaign object when ID changes
    const {
        data: currentCampaign,
        isLoading,
        error,
    } = useCampaign(currentCampaignId ?? '', {
        enabled: !!currentCampaignId,
    });

    // Persist campaign ID to localStorage when it changes
    useEffect(() => {
        if (currentCampaignId) {
            localStorage.setItem(CAMPAIGN_STORAGE_KEY, currentCampaignId);
        } else {
            localStorage.removeItem(CAMPAIGN_STORAGE_KEY);
        }
    }, [currentCampaignId]);

    /**
     * Set the current campaign by ID.
     */
    const setCurrentCampaignId = useCallback((id: string | null) => {
        setCurrentCampaignIdState(id);
    }, []);

    /**
     * Clear the current campaign selection.
     */
    const clearCurrentCampaign = useCallback(() => {
        setCurrentCampaignIdState(null);
    }, []);

    const value: CampaignContextValue = {
        currentCampaignId,
        currentCampaign: currentCampaign ?? null,
        isLoading,
        error: error as Error | null,
        setCurrentCampaignId,
        clearCurrentCampaign,
    };

    return (
        <CampaignContext.Provider value={value}>
            {children}
        </CampaignContext.Provider>
    );
}

/**
 * Accesses the current campaign context.
 *
 * @returns The CampaignContextValue containing campaign state and selection functions.
 * @throws Error if called outside of a CampaignProvider.
 */
export function useCampaignContext(): CampaignContextValue {
    const context = useContext(CampaignContext);
    if (context === undefined) {
        throw new Error(
            'useCampaignContext must be used within a CampaignProvider'
        );
    }
    return context;
}
