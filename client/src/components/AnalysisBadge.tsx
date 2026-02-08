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
 * Displays a warning chip with the count of pending analysis items for a
 * given source. Clicking it navigates to the analysis triage page for the
 * latest matching job.
 */

import { Chip } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { usePendingAnalysisCount, useAnalysisJobs } from '../hooks';

/**
 * Props for the AnalysisBadge component.
 */
interface AnalysisBadgeProps {
    campaignId: number;
    sourceTable: string;
    sourceId: number;
}

/**
 * Display a warning chip with the count of pending analysis items for a
 * specific source record. Clicking the chip navigates to the triage page
 * for the latest analysis job that matches the source.
 *
 * Returns null when the count is zero or data is still loading.
 *
 * @param props - The campaign, source table, and source ID to query.
 * @returns A warning Chip element, or null.
 */
export function AnalysisBadge({ campaignId, sourceTable, sourceId }: AnalysisBadgeProps) {
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

    // Find the latest job matching this source
    const latestJob = jobs
        ?.filter(
            (job) =>
                job.sourceTable === sourceTable &&
                job.sourceId === sourceId
        )
        .sort((a, b) => b.id - a.id)[0];

    if (!latestJob) {
        return null;
    }

    return (
        <Chip
            label={`${count} to review`}
            color="warning"
            size="small"
            onClick={() =>
                navigate(
                    `/campaigns/${campaignId}/analysis/${latestJob.id}`
                )
            }
            sx={{ cursor: 'pointer' }}
        />
    );
}
