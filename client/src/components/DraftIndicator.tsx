// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * DraftIndicator - Small chip indicating that an unsaved draft exists.
 *
 * Renders a warning-coloured MUI Chip with an EditNote icon when
 * `hasDraft` is true. Returns null otherwise. Designed for inline
 * placement in list items, table rows, or next to headings.
 */

import { Chip, Tooltip } from '@mui/material';
import { EditNote as EditNoteIcon } from '@mui/icons-material';

/**
 * Props for the DraftIndicator component.
 */
interface DraftIndicatorProps {
    /** Whether a draft exists for this item. */
    hasDraft: boolean;
    /** Size of the chip (default: 'small'). */
    size?: 'small' | 'medium';
    /** Custom tooltip text (default: 'Unsaved draft'). */
    tooltip?: string;
}

/**
 * Display a small warning chip indicating that an unsaved draft exists
 * for the associated item. Returns null when no draft is present.
 *
 * @param props - The draft indicator configuration.
 * @returns A Tooltip-wrapped Chip element, or null.
 */
export function DraftIndicator({
    hasDraft,
    size = 'small',
    tooltip = 'Unsaved draft',
}: DraftIndicatorProps) {
    if (!hasDraft) {
        return null;
    }

    return (
        <Tooltip title={tooltip}>
            <Chip
                icon={<EditNoteIcon />}
                label="Draft"
                color="warning"
                size={size}
                variant="outlined"
            />
        </Tooltip>
    );
}
