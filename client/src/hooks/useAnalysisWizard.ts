// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Wizard state management for the analysis wizard.
 *
 * Provides job data, items grouped by phase, navigation between
 * phases, and auto-advancement when all items in a phase are
 * resolved.
 */

import {
    useMemo,
    useCallback,
    useState,
    useEffect,
} from 'react';
import { useNavigate } from 'react-router-dom';
import {
    useAnalysisJob,
    useAnalysisItems,
} from './useContentAnalysis';
import type {
    ContentAnalysisItem,
} from '../api/contentAnalysis';

/** Detection types for the Identify phase. */
export const DETECTION_GROUPS = [
    'wiki_link_resolved',
    'wiki_link_unresolved',
    'untagged_mention',
    'potential_alias',
    'misspelling',
] as const;

/** Detection types for the Revise phase. */
export const ANALYSIS_GROUPS = [
    'analysis_report',
    'content_suggestion',
    'mechanics_warning',
    'investigation_gap',
    'pacing_note',
    'canon_contradiction',
    'temporal_inconsistency',
    'character_inconsistency',
] as const;

/** Detection types for the Enrich phase. */
export const ENRICHMENT_GROUPS = [
    'description_update',
    'log_entry',
    'relationship_suggestion',
    'new_entity_suggestion',
    'graph_warning',
    'redundant_edge',
    'invalid_type_pair',
    'orphan_warning',
] as const;

/** Phase key to route segment mapping. */
export const PHASE_ROUTES: Record<string, string> = {
    identification: 'identify',
    analysis: 'revise',
    enrichment: 'enrich',
};

/** Route segment to phase key mapping (inverse of PHASE_ROUTES). */
const ROUTE_TO_PHASE: Record<string, string> =
    Object.fromEntries(
        Object.entries(PHASE_ROUTES).map(([k, v]) => [v, k]),
    );

/**
 * Filter items by phase using detection type group membership.
 */
export function getPhaseItems(
    items: ContentAnalysisItem[],
    phase: string,
): ContentAnalysisItem[] {
    let types: readonly string[];
    switch (phase) {
        case 'identification':
            types = DETECTION_GROUPS;
            break;
        case 'analysis':
            types = ANALYSIS_GROUPS;
            break;
        case 'enrichment':
            types = ENRICHMENT_GROUPS;
            break;
        default:
            return [];
    }
    return items.filter((item) =>
        types.includes(item.detectionType),
    );
}

/** Return type for the useAnalysisWizard hook. */
export interface AnalysisWizardState {
    // -- data --
    job: ReturnType<typeof useAnalysisJob>['data'];
    items: ContentAnalysisItem[];
    isLoading: boolean;
    error: Error | null;

    // -- phase navigation --
    phases: string[];
    currentPhase: string | null;
    currentPhaseIndex: number;
    phaseItems: ContentAnalysisItem[];
    pendingCount: number;
    canAdvance: boolean;
    canGoBack: boolean;
    nextPhaseLabel: string | null;

    // -- actions --
    goToPhase: (phase: string) => void;
    goToNextPhase: () => void;
}

/**
 * The useAnalysisWizard hook provides job data, items, phase
 * navigation state, and auto-advance logic for the AnalysisWizard
 * shell.
 */
export function useAnalysisWizard(
    campaignId: number,
    jobId: number,
    currentRoutePhase?: string,
): AnalysisWizardState {
    const navigate = useNavigate();
    const jobQuery = useAnalysisJob(campaignId, jobId);
    const itemsQuery = useAnalysisItems(campaignId, jobId);
    const [autoAdvanced, setAutoAdvanced] = useState(false);

    const job = jobQuery.data;
    const items = useMemo(
        () => itemsQuery.data ?? [],
        [itemsQuery.data],
    );
    const phases = useMemo(
        () => job?.phases ?? [],
        [job?.phases],
    );

    const currentPhase = currentRoutePhase
        ? ROUTE_TO_PHASE[currentRoutePhase] ?? currentRoutePhase
        : job?.currentPhase ?? phases[0] ?? null;

    const currentPhaseIndex = currentPhase
        ? phases.indexOf(currentPhase)
        : -1;

    const phaseItems = useMemo(
        () => currentPhase ? getPhaseItems(items, currentPhase) : [],
        [items, currentPhase],
    );

    const pendingCount = useMemo(
        () => phaseItems.filter((i) => i.resolution === 'pending').length,
        [phaseItems],
    );

    const canAdvance =
        currentPhaseIndex >= 0 &&
        currentPhaseIndex < phases.length - 1;
    const canGoBack = currentPhaseIndex > 0;

    const nextPhaseLabel = canAdvance
        ? PHASE_ROUTES[phases[currentPhaseIndex + 1]] ?? null
        : null;

    const goToPhase = useCallback(
        (phase: string) => {
            const segment = PHASE_ROUTES[phase] ?? phase;
            navigate(
                `/campaigns/${campaignId}/analysis/${jobId}/${segment}`,
            );
        },
        [navigate, campaignId, jobId],
    );

    const goToNextPhase = useCallback(() => {
        if (canAdvance) {
            goToPhase(phases[currentPhaseIndex + 1]);
        }
    }, [canAdvance, goToPhase, phases, currentPhaseIndex]);

    // Auto-advance when all items are resolved.
    useEffect(() => {
        if (
            pendingCount === 0 &&
            phaseItems.length > 0 &&
            canAdvance &&
            !autoAdvanced
        ) {
            setAutoAdvanced(true);
            const timer = setTimeout(() => {
                goToNextPhase();
                setAutoAdvanced(false);
            }, 1500);
            return () => clearTimeout(timer);
        }
        return undefined;
    }, [pendingCount, phaseItems.length, canAdvance, autoAdvanced, goToNextPhase]);

    return {
        job,
        items,
        isLoading: jobQuery.isLoading || itemsQuery.isLoading,
        error: jobQuery.error ?? itemsQuery.error,
        phases,
        currentPhase,
        currentPhaseIndex,
        phaseItems,
        pendingCount,
        canAdvance,
        canGoBack,
        nextPhaseLabel,
        goToPhase,
        goToNextPhase,
    };
}
