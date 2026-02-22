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
 * Placeholder for the Identify phase page.
 *
 * Displays a summary of the items and pending count for the current
 * phase. This placeholder will be replaced with the full Identify
 * phase implementation in Task C.
 */
export default function IdentifyPhasePage() {
    const { phaseItems, pendingCount } = useWizardContext();

    return (
        <Box sx={{ p: 2 }}>
            <Typography variant="h5">
                Identify Phase
            </Typography>
            <Typography>
                {phaseItems.length} items ({pendingCount} pending)
            </Typography>
        </Box>
    );
}
