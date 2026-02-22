/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { Box, Typography } from '@mui/material';
import { useWizardContext } from '../contexts/AnalysisWizardContext';

/**
 * Placeholder for the Revise phase page.
 *
 * Displays a summary of the items and pending count for the current
 * phase. This placeholder will be replaced with the full Revise
 * phase implementation in Task D.
 */
export default function RevisePhasePage() {
    const { phaseItems, pendingCount } = useWizardContext();

    return (
        <Box sx={{ p: 2 }}>
            <Typography variant="h5">
                Revise Phase
            </Typography>
            <Typography>
                {phaseItems.length} items ({pendingCount} pending)
            </Typography>
        </Box>
    );
}
