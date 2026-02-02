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
import { useCampaign, useCampaigns } from '../hooks';
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
    /** Whether we are still initializing (validating stored campaign) */
    isInitializing: boolean;
    /** Error if campaign fetch failed */
    error: Error | null;
    /** All available campaigns for the user */
    campaigns: Campaign[] | undefined;
    /** Whether campaigns list is loading */
    campaignsLoading: boolean;
    /** Set the current campaign by ID (persists to localStorage) */
    setCurrentCampaignId: (id: string | null) => void;
    /** Clear the current campaign selection */
    clearCurrentCampaign: () => void;
    /** Get the most recently created campaign */
    getLatestCampaign: () => Campaign | null;
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
 *
 * On initialization, validates that the stored campaign ID is still valid
 * (exists in the user's campaign list). If not valid, clears the stored ID.
 */
export function CampaignProvider({ children }: CampaignProviderProps) {
    const [currentCampaignId, setCurrentCampaignIdState] = useState<
        string | null
    >(() => {
        // Initialize from localStorage
        const stored = localStorage.getItem(CAMPAIGN_STORAGE_KEY);
        return stored ?? null;
    });

    // Track whether we've completed initial validation
    const [isInitializing, setIsInitializing] = useState(true);

    // Fetch all campaigns to validate stored ID and provide list
    const {
        data: campaigns,
        isLoading: campaignsLoading,
    } = useCampaigns();

    // Fetch the full campaign object when ID changes
    const {
        data: currentCampaign,
        isLoading,
        error,
    } = useCampaign(currentCampaignId ?? '', {
        enabled: !!currentCampaignId,
    });

    // Validate stored campaign ID against user's campaigns
    useEffect(() => {
        if (campaignsLoading) {
            return;
        }

        // Once campaigns are loaded, validate the stored ID
        if (currentCampaignId && campaigns) {
            const isValid = campaigns.some(c => c.id === currentCampaignId);
            if (!isValid) {
                // Stored campaign no longer valid, clear it
                setCurrentCampaignIdState(null);
                localStorage.removeItem(CAMPAIGN_STORAGE_KEY);
            }
        }

        setIsInitializing(false);
    }, [campaigns, campaignsLoading, currentCampaignId]);

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

    /**
     * Get the most recently created campaign.
     * Returns null if no campaigns exist.
     */
    const getLatestCampaign = useCallback((): Campaign | null => {
        if (!campaigns || campaigns.length === 0) {
            return null;
        }
        // Sort by createdAt descending and return the first one
        const sorted = [...campaigns].sort((a, b) =>
            new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
        return sorted[0];
    }, [campaigns]);

    const value: CampaignContextValue = {
        currentCampaignId,
        currentCampaign: currentCampaign ?? null,
        isLoading,
        isInitializing,
        error: error as Error | null,
        campaigns,
        campaignsLoading,
        setCurrentCampaignId,
        clearCurrentCampaign,
        getLatestCampaign,
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
