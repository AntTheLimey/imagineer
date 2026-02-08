/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { Chip } from '@mui/material';
import { NodeViewWrapper } from '@tiptap/react';
import type { NodeViewProps } from '@tiptap/react';

/**
 * React component for rendering WikiLink nodes within the TipTap
 * editor.
 *
 * Displays the wiki link as a compact MUI Chip with the display
 * text (or entity name as fallback). The chip uses an outlined
 * variant for a subtle appearance that integrates well with
 * surrounding text.
 *
 * @param props - The TipTap node view props containing node
 *     attributes.
 * @returns A React element rendering the wiki link inline.
 */
export default function WikiLinkComponent({ node }: NodeViewProps) {
    const entityName = node.attrs.entityName as string;
    const displayText = node.attrs.displayText as string | null;

    return (
        <NodeViewWrapper
            as="span"
            style={{ display: 'inline' }}
        >
            <Chip
                label={displayText || entityName}
                size="small"
                variant="outlined"
                color="info"
                sx={{
                    cursor: 'pointer',
                    verticalAlign: 'text-bottom',
                    height: '1.4em',
                    fontSize: '0.9em',
                    '& .MuiChip-label': {
                        px: 1,
                    },
                }}
            />
        </NodeViewWrapper>
    );
}
