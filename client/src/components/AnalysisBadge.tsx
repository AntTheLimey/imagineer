// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * AnalysisBadge - Reusable badge showing pending content analysis items.
 *
 * Supports two display variants:
 *
 * - "chip" (default): A compact warning chip with the count. Suitable for
 *   inline placement next to headings or in toolbars.
 *
 * - "banner": A full-width Alert with an info severity, a count summary,
 *   and a "Review Now" button. Suitable for prominent placement at the
 *   top of a page.
 *
 * When sourceTable and sourceId are omitted, the component shows the
 * campaign-wide pending count across all source tables.
 */

import { Alert, Button, Chip } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { usePendingAnalysisCount, useAnalysisJobs } from '../hooks';

/**
 * Props for the AnalysisBadge component.
 */
interface AnalysisBadgeProps {
    /** The campaign to query pending analysis items for. */
    campaignId: number;
    /** Optional source table filter. Omit for campaign-wide count. */
    sourceTable?: string;
    /** Optional source ID filter. Omit for campaign-wide count. */
    sourceId?: number;
    /** Display variant: "chip" for inline badge, "banner" for full-width alert. */
    variant?: 'chip' | 'banner';
}

/**
 * Display a warning chip or info banner with the count of pending analysis
 * items. Clicking the chip or the banner's "Review Now" button navigates
 * to the triage page for the latest matching analysis job.
 *
 * Returns null when the count is zero or data is still loading.
 *
 * @param props - The campaign, optional source filter, and display variant.
 * @returns A Chip or Alert element, or null.
 */
export function AnalysisBadge({
    campaignId,
    sourceTable,
    sourceId,
    variant = 'chip',
}: AnalysisBadgeProps) {
    const navigate = useNavigate();

    const { data: pendingData } = usePendingAnalysisCount(
        campaignId,
        sourceTable,
        sourceId
    );

    const { data: jobs } = useAnalysisJobs(campaignId);

    const count = pendingData?.count ?? 0;

    if (count === 0) {
        return null;
    }

    // Find the latest job matching the source filter (or the latest job
    // overall when no filter is provided). Spread the array before sorting
    // to avoid mutating the React Query cache.
    const latestJob = [...(jobs ?? [])]
        .filter((job) => {
            if (sourceTable && sourceId !== undefined) {
                return (
                    job.sourceTable === sourceTable &&
                    job.sourceId === sourceId
                );
            }
            return true;
        })
        .sort((a, b) => b.id - a.id)[0];

    if (!latestJob) {
        return null;
    }

    const triageUrl = `/campaigns/${campaignId}/analysis/${latestJob.id}`;

    if (variant === 'banner') {
        const itemWord = count === 1 ? 'item' : 'items';
        return (
            <Alert
                severity="info"
                action={
                    <Button
                        color="inherit"
                        size="small"
                        onClick={() => navigate(triageUrl)}
                    >
                        Review Now
                    </Button>
                }
            >
                Content analysis found {count} {itemWord} to review.
            </Alert>
        );
    }

    return (
        <Chip
            label={`${count} to review`}
            color="warning"
            size="small"
            onClick={() => navigate(triageUrl)}
            sx={{ cursor: 'pointer' }}
        />
    );
}
