/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

/**
 * AnalysisWizard shell component.
 *
 * Replaces the monolithic AnalysisTriagePage as the route target for
 * /campaigns/:campaignId/analysis/:jobId. It renders a MUI Stepper,
 * error banner, navigation buttons, and an Outlet wrapped in
 * AnalysisWizardProvider so that nested phase pages (Identify, Revise,
 * Enrich) can access wizard state via context.
 */

import {
    useParams,
    Outlet,
    useLocation,
    useNavigate,
} from 'react-router-dom';
import {
    Box,
    Stepper,
    Step,
    StepLabel,
    Button,
    Alert,
    Snackbar,
    Typography,
    CircularProgress,
} from '@mui/material';
import { useState, useEffect, useRef } from 'react';
import { useAnalysisWizard } from '../hooks/useAnalysisWizard';
import { AnalysisWizardProvider } from '../contexts/AnalysisWizardContext';

/** Human-readable labels for each wizard phase. */
const PHASE_LABELS: Record<string, string> = {
    identification: 'Identify',
    analysis: 'Revise',
    enrichment: 'Enrich',
};

/**
 * Map a failure reason string to a user-friendly error message.
 *
 * - Quota errors get a specific message about API token limits.
 * - Rate-limit errors get a temporary-unavailability message.
 * - All other errors display the raw failure reason.
 */
function getErrorMessage(reason: string): string {
    if (reason.toLowerCase().includes('quota')) {
        return 'Your API token limit has been reached. Please check your account and try again later.';
    }
    if (reason.toLowerCase().includes('rate limit')) {
        return 'The AI service is temporarily unavailable. Please try again in a few minutes.';
    }
    return `Enrichment encountered an error: ${reason}`;
}

export default function AnalysisWizard() {
    const { campaignId, jobId } = useParams<{
        campaignId: string;
        jobId: string;
    }>();
    const location = useLocation();
    const navigate = useNavigate();

    // Extract the current route phase segment from the URL pathname.
    // For /campaigns/1/analysis/1/identify the last segment is "identify".
    const pathSegments = location.pathname.split('/');
    const currentRoutePhase =
        pathSegments[pathSegments.length - 1];

    const cId = Number(campaignId);
    const jId = Number(jobId);

    const wizard = useAnalysisWizard(cId, jId, currentRoutePhase);

    // Snackbar state for the auto-advance notification.
    const [snackOpen, setSnackOpen] = useState(false);
    const prevPending = useRef(wizard.pendingCount);

    // Show snackbar when pending count drops to 0 and the wizard can
    // advance to the next phase.
    useEffect(() => {
        if (
            prevPending.current > 0 &&
            wizard.pendingCount === 0 &&
            wizard.phaseItems.length > 0 &&
            wizard.canAdvance
        ) {
            setSnackOpen(true);
        }
        prevPending.current = wizard.pendingCount;
    }, [wizard.pendingCount, wizard.phaseItems.length, wizard.canAdvance]);

    // -- NaN guard --------------------------------------------------------

    if (Number.isNaN(cId) || Number.isNaN(jId)) {
        return (
            <Box sx={{ p: 4 }}>
                <Alert severity="error">
                    Invalid campaign or job ID.
                </Alert>
            </Box>
        );
    }

    // -- Loading state ----------------------------------------------------

    if (wizard.isLoading) {
        return (
            <Box
                sx={{
                    display: 'flex',
                    justifyContent: 'center',
                    p: 4,
                }}
            >
                <CircularProgress />
            </Box>
        );
    }

    // -- Derive display values --------------------------------------------

    const nextPhaseName =
        wizard.currentPhaseIndex >= 0 &&
        wizard.currentPhaseIndex < wizard.phases.length - 1
            ? PHASE_LABELS[
                  wizard.phases[wizard.currentPhaseIndex + 1]
              ] ?? 'Next'
            : 'Next';

    // -- Render -----------------------------------------------------------

    return (
        <AnalysisWizardProvider value={wizard}>
            <Box sx={{ p: 3, maxWidth: 1400, mx: 'auto' }}>
                {/* Header with job title / source info */}
                <Typography variant="h5" gutterBottom>
                    Content Analysis
                </Typography>
                {wizard.job?.sourceTable && wizard.job?.sourceField && (
                    <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mb: 2 }}
                    >
                        Source: {wizard.job.sourceTable} &mdash;{' '}
                        {wizard.job.sourceField}
                    </Typography>
                )}

                {/* MUI Stepper */}
                <Stepper
                    activeStep={wizard.currentPhaseIndex}
                    sx={{ mb: 3 }}
                >
                    {wizard.phases.map((phase) => (
                        <Step key={phase}>
                            <StepLabel>
                                {PHASE_LABELS[phase] ?? phase}
                            </StepLabel>
                        </Step>
                    ))}
                </Stepper>

                {/* Error banner for failed jobs */}
                {wizard.job?.status === 'failed' &&
                    wizard.job?.failureReason && (
                        <Alert severity="error" sx={{ mb: 2 }}>
                            {getErrorMessage(wizard.job.failureReason)}
                        </Alert>
                    )}

                {/* Nested route outlet */}
                <Outlet />

                {/* Navigation buttons */}
                <Box
                    sx={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        mt: 3,
                    }}
                >
                    <Button
                        onClick={() => navigate(`/campaigns/${cId}`)}
                    >
                        Back to Overview
                    </Button>
                    {wizard.canAdvance && (
                        <Button
                            variant="contained"
                            onClick={wizard.goToNextPhase}
                        >
                            Continue to {nextPhaseName}
                        </Button>
                    )}
                </Box>

                {/* Snackbar for auto-advance notification */}
                <Snackbar
                    open={snackOpen}
                    autoHideDuration={3000}
                    onClose={() => setSnackOpen(false)}
                    message="All items resolved! Moving to next phase..."
                />
            </Box>
        </AnalysisWizardProvider>
    );
}
