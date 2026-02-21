// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * PhaseStrip - A horizontal workflow strip with three sequential phases
 * and a save/action button.
 *
 * Replaces the SaveSplitButton on CampaignOverview, giving the GM
 * explicit control over which pipeline phases run after saving.
 */

import { useCallback, useEffect, useState } from 'react';
import {
    Box,
    Button,
    Checkbox,
    CircularProgress,
    FormControlLabel,
    Tooltip,
    Typography,
} from '@mui/material';

/** Which pipeline phases are selected by the user. */
export interface PhaseSelection {
    identify: boolean;
    revise: boolean;
    enrich: boolean;
}

export interface PhaseStripProps {
    /** Callback fired when the user clicks the save/action button. */
    onSave: (phases: PhaseSelection) => void;
    /** Whether the form has unsaved changes. */
    isDirty: boolean;
    /** Whether a save operation is currently in progress. */
    isSaving: boolean;
}

/** localStorage key for persisting the default phase selection. */
const STORAGE_KEY = 'imagineer:phaseSelection';

/** Phase metadata displayed in the strip. */
interface PhaseDefinition {
    key: keyof PhaseSelection;
    label: string;
    subtitle: string;
    tooltip: string;
}

const PHASES: PhaseDefinition[] = [
    {
        key: 'identify',
        label: 'Identify',
        subtitle: 'Pattern detection',
        tooltip: 'Detect wiki links, mentions, and misspellings (no AI)',
    },
    {
        key: 'revise',
        label: 'Revise',
        subtitle: 'AI advisory',
        tooltip: 'TTRPG Expert + Canon Expert review (iterative)',
    },
    {
        key: 'enrich',
        label: 'Enrich',
        subtitle: 'Structural',
        tooltip: 'Generate descriptions, relationships, and graph hygiene',
    },
];

const DEFAULT_PHASES: PhaseSelection = {
    identify: true,
    revise: false,
    enrich: false,
};

/**
 * Load phase selection from localStorage, falling back to defaults.
 */
function loadPhases(): PhaseSelection {
    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
            const parsed = JSON.parse(stored);
            return {
                identify: Boolean(parsed.identify),
                revise: Boolean(parsed.revise),
                enrich: Boolean(parsed.enrich),
            };
        }
    } catch {
        // Ignore corrupt data
    }
    return { ...DEFAULT_PHASES };
}

/**
 * A horizontal strip of three labelled checkboxes representing sequential
 * workflow phases, plus a "Save & Go" action button.
 *
 * Phase selection persists to localStorage so the GM's preferred workflow
 * is remembered across sessions.
 */
export default function PhaseStrip({
    onSave,
    isDirty,
    isSaving,
}: PhaseStripProps) {
    const [phases, setPhases] = useState<PhaseSelection>(loadPhases);

    // Persist to localStorage whenever phases change.
    useEffect(() => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(phases));
    }, [phases]);

    const hasAnyPhase = phases.identify || phases.revise || phases.enrich;
    const buttonLabel = hasAnyPhase ? 'Save & Go' : 'Save';
    const isDisabled = !isDirty || isSaving;

    const handleToggle = useCallback(
        (key: keyof PhaseSelection) => {
            setPhases((prev) => ({ ...prev, [key]: !prev[key] }));
        },
        [],
    );

    const handleClick = useCallback(() => {
        onSave(phases);
    }, [onSave, phases]);

    return (
        <Box
            sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 1,
                px: 2,
                py: 1,
                borderRadius: 1,
                bgcolor: 'action.hover',
            }}
        >
            {PHASES.map((phase) => (
                <Tooltip key={phase.key} title={phase.tooltip} arrow>
                    <Box
                        sx={{
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                            minWidth: 100,
                        }}
                    >
                        <FormControlLabel
                            control={
                                <Checkbox
                                    checked={phases[phase.key]}
                                    onChange={() => handleToggle(phase.key)}
                                    size="small"
                                />
                            }
                            label={
                                <Typography
                                    variant="body2"
                                    sx={{ fontWeight: 'bold', lineHeight: 1.2 }}
                                >
                                    {phase.label}
                                </Typography>
                            }
                            sx={{ mr: 0 }}
                        />
                        <Typography
                            variant="caption"
                            color="text.secondary"
                            sx={{ mt: -0.5 }}
                        >
                            {phase.subtitle}
                        </Typography>
                    </Box>
                </Tooltip>
            ))}

            <Button
                variant="contained"
                onClick={handleClick}
                disabled={isDisabled}
                startIcon={
                    isSaving ? (
                        <CircularProgress size={20} color="inherit" />
                    ) : null
                }
                sx={{ ml: 1, whiteSpace: 'nowrap' }}
            >
                {buttonLabel}
            </Button>
        </Box>
    );
}
