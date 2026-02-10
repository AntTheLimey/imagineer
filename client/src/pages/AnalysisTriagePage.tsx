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
 * AnalysisTriagePage - Full-screen triage page for reviewing and resolving
 * content analysis items (wiki links, untagged mentions, misspellings).
 */

import React, { useState, useMemo, useCallback } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import {
    Accordion,
    AccordionDetails,
    AccordionSummary,
    Autocomplete,
    Box,
    Typography,
    Chip,
    Button,
    Checkbox,
    IconButton,
    List,
    ListItemButton,
    ListItemText,
    Divider,
    CircularProgress,
    FormControlLabel,
    LinearProgress,
    Paper,
    TextField,
    Select,
    MenuItem,
    FormControl,
    InputLabel,
    Stack,
    Alert,
    Tooltip,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogContentText,
    DialogActions,
} from '@mui/material';
import { MarkdownEditor } from '../components/MarkdownEditor';
import {
    Check,
    AddCircle,
    Close,
    DoneAll,
    PersonAdd,
    Undo,
    ExpandMore,
} from '@mui/icons-material';
import FullScreenLayout from '../layouts/FullScreenLayout';
import {
    useAnalysisJob,
    useAnalysisItems,
    useResolveItem,
    useBatchResolve,
    useRevertItem,
    useEnrichmentStream,
    useTriggerEnrichment,
    useCancelEnrichment,
} from '../hooks/useContentAnalysis';
import { useUserSettings } from '../hooks/useUserSettings';
import {
    useRelationshipTypes,
    useCreateRelationshipType,
} from '../hooks/useRelationshipTypes';
import {
    entityTypeColors,
    formatEntityType,
} from '../components/EntitySelector/entityConstants';
import type { EntityType, RelationshipType } from '../types';
import type { ContentAnalysisItem } from '../api/contentAnalysis';

/**
 * Detection type grouping configuration.
 */
const DETECTION_GROUPS = [
    {
        key: 'wiki_link_resolved' as const,
        label: 'Wiki Links (Resolved)',
        color: '#4caf50',
    },
    {
        key: 'wiki_link_unresolved' as const,
        label: 'Wiki Links (Unresolved)',
        color: '#ff9800',
    },
    {
        key: 'untagged_mention' as const,
        label: 'Untagged Mentions',
        color: '#2196f3',
    },
    {
        key: 'potential_alias' as const,
        label: 'Potential Aliases',
        color: '#9c27b0',
    },
    {
        key: 'misspelling' as const,
        label: 'Misspellings',
        color: '#ffc107',
    },
];

/**
 * Enrichment detection type grouping configuration (Phase 2).
 */
const ENRICHMENT_GROUPS = [
    {
        key: 'description_update' as const,
        label: 'Description Updates',
        color: '#00897b',
    },
    {
        key: 'log_entry' as const,
        label: 'Log Entries',
        color: '#5c6bc0',
    },
    {
        key: 'relationship_suggestion' as const,
        label: 'Relationships',
        color: '#ef6c00',
    },
    {
        key: 'new_entity_suggestion' as const,
        label: 'New Entity Suggestions',
        color: '#7b1fa2',
    },
];

/**
 * All available entity types for the new entity creation form.
 */
const ENTITY_TYPES: EntityType[] = [
    'npc',
    'location',
    'item',
    'faction',
    'clue',
    'creature',
    'organization',
    'event',
    'document',
    'other',
];

/**
 * Props for the RelationshipTypeAutocomplete component.
 */
interface RelTypeAutocompleteProps {
    value: string;
    onChange: (typeName: string) => void;
    relationshipTypes: RelationshipType[];
    campaignId: number;
}

/**
 * Convert a string to snake_case.
 */
function toSnakeCase(str: string): string {
    return str
        .trim()
        .replace(/([a-z])([A-Z])/g, '$1_$2')
        .replace(/[\s-]+/g, '_')
        .toLowerCase();
}

/**
 * Convert a string to Title Case.
 */
function toTitleCase(str: string): string {
    return str
        .trim()
        .replace(/[_-]/g, ' ')
        .replace(/\b\w/g, (c) => c.toUpperCase());
}

/**
 * Autocomplete for selecting or creating a relationship type.
 *
 * Supports freeSolo input so LLM-suggested types that do not yet exist
 * in the system can be displayed. When the user enters a value that
 * does not match any existing type, a fuzzy-match confirmation dialog
 * is shown before optionally creating a new type.
 */
function RelationshipTypeAutocomplete({
    value,
    onChange,
    relationshipTypes,
    campaignId,
}: RelTypeAutocompleteProps) {
    const createRelType = useCreateRelationshipType();

    // Fuzzy-match confirmation state.
    const [fuzzyOpen, setFuzzyOpen] = useState(false);
    const [fuzzyMatches, setFuzzyMatches] = useState<RelationshipType[]>([]);
    const [pendingCustomValue, setPendingCustomValue] = useState('');

    // Create-new-type dialog state.
    const [createOpen, setCreateOpen] = useState(false);
    const [createName, setCreateName] = useState('');
    const [createDisplayLabel, setCreateDisplayLabel] = useState('');
    const [createInverseName, setCreateInverseName] = useState('');
    const [createInverseDisplayLabel, setCreateInverseDisplayLabel] =
        useState('');
    const [createIsSymmetric, setCreateIsSymmetric] = useState(false);

    /**
     * Find the RelationshipType object matching the current value so we
     * can render its displayLabel in the input.
     */
    const selectedOption =
        relationshipTypes.find((rt) => rt.name === value) ?? null;

    /**
     * Check whether a custom string closely matches any existing type.
     */
    const findFuzzyMatches = useCallback(
        (typed: string): RelationshipType[] => {
            const lower = typed.toLowerCase();
            return relationshipTypes.filter(
                (rt) =>
                    rt.name.toLowerCase().includes(lower) ||
                    rt.displayLabel.toLowerCase().includes(lower) ||
                    lower.includes(rt.name.toLowerCase()) ||
                    lower.includes(rt.displayLabel.toLowerCase()) ||
                    rt.name.toLowerCase().startsWith(lower) ||
                    rt.displayLabel.toLowerCase().startsWith(lower),
            );
        },
        [relationshipTypes],
    );

    /**
     * Handle the user confirming a value that is not an existing option.
     */
    const handleCustomValue = useCallback(
        (typed: string) => {
            const matches = findFuzzyMatches(typed);
            if (matches.length > 0) {
                setFuzzyMatches(matches);
                setPendingCustomValue(typed);
                setFuzzyOpen(true);
            } else {
                // No close matches -- go straight to create dialog.
                setPendingCustomValue(typed);
                setCreateName(toSnakeCase(typed));
                setCreateDisplayLabel(toTitleCase(typed));
                setCreateInverseName('');
                setCreateInverseDisplayLabel('');
                setCreateIsSymmetric(false);
                setCreateOpen(true);
            }
        },
        [findFuzzyMatches],
    );

    /**
     * User chose "create new type" from the fuzzy-match dialog.
     */
    const handleFuzzyCreateNew = useCallback(() => {
        setFuzzyOpen(false);
        setCreateName(toSnakeCase(pendingCustomValue));
        setCreateDisplayLabel(toTitleCase(pendingCustomValue));
        setCreateInverseName('');
        setCreateInverseDisplayLabel('');
        setCreateIsSymmetric(false);
        setCreateOpen(true);
    }, [pendingCustomValue]);

    /**
     * Submit the create-new-type form.
     */
    const handleCreateSubmit = useCallback(() => {
        createRelType.mutate(
            {
                campaignId,
                input: {
                    name: createName,
                    inverseName: createInverseName,
                    isSymmetric: createIsSymmetric,
                    displayLabel: createDisplayLabel,
                    inverseDisplayLabel: createInverseDisplayLabel,
                },
            },
            {
                onSuccess: (created) => {
                    onChange(created.name);
                    setCreateOpen(false);
                },
            },
        );
    }, [
        campaignId,
        createName,
        createInverseName,
        createIsSymmetric,
        createDisplayLabel,
        createInverseDisplayLabel,
        createRelType,
        onChange,
    ]);

    return (
        <>
            <Autocomplete
                freeSolo
                size="small"
                sx={{ minWidth: 180 }}
                options={relationshipTypes}
                getOptionLabel={(option) =>
                    typeof option === 'string'
                        ? option
                        : option.displayLabel
                }
                isOptionEqualToValue={(option, val) =>
                    option.name === val.name
                }
                value={selectedOption}
                inputValue={
                    selectedOption
                        ? selectedOption.displayLabel
                        : value
                }
                onInputChange={(_event, newInputValue, reason) => {
                    // Only track typing -- ignore reset/clear events so the
                    // input text stays in sync with the selected option.
                    if (reason === 'input') {
                        onChange(newInputValue);
                    }
                }}
                onChange={(_event, newValue) => {
                    if (newValue === null) {
                        onChange('');
                    } else if (typeof newValue === 'string') {
                        // Free-solo string entered -- check if it matches
                        // an existing type by name before treating as custom.
                        const existing = relationshipTypes.find(
                            (rt) =>
                                rt.name === newValue ||
                                rt.displayLabel.toLowerCase() ===
                                    newValue.toLowerCase(),
                        );
                        if (existing) {
                            onChange(existing.name);
                        } else {
                            handleCustomValue(newValue);
                        }
                    } else {
                        onChange(newValue.name);
                    }
                }}
                onBlur={() => {
                    // On blur with a custom (non-matched) value, trigger
                    // the create flow.
                    if (
                        value &&
                        !relationshipTypes.find((rt) => rt.name === value)
                    ) {
                        handleCustomValue(value);
                    }
                }}
                renderInput={(params) => (
                    <TextField
                        {...params}
                        placeholder="Relationship type"
                        sx={{
                            '& .MuiInputBase-input': {
                                fontWeight: 600,
                                fontSize: '0.875rem',
                            },
                        }}
                    />
                )}
                renderOption={(props, option) => {
                    const { key, ...otherProps } = props;
                    return (
                        <li key={key} {...otherProps}>
                            {option.displayLabel}
                        </li>
                    );
                }}
            />

            {/* Fuzzy-match confirmation dialog */}
            <Dialog
                open={fuzzyOpen}
                onClose={() => setFuzzyOpen(false)}
                maxWidth="xs"
                fullWidth
            >
                <DialogTitle>Similar types found</DialogTitle>
                <DialogContent>
                    <DialogContentText sx={{ mb: 2 }}>
                        Did you mean one of these existing types?
                    </DialogContentText>
                    <Stack spacing={1}>
                        {fuzzyMatches.map((rt) => (
                            <Button
                                key={rt.id}
                                variant="outlined"
                                size="small"
                                onClick={() => {
                                    onChange(rt.name);
                                    setFuzzyOpen(false);
                                }}
                            >
                                {rt.displayLabel}
                            </Button>
                        ))}
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setFuzzyOpen(false)}>
                        Cancel
                    </Button>
                    <Button
                        variant="contained"
                        onClick={handleFuzzyCreateNew}
                    >
                        Create &quot;{pendingCustomValue}&quot;
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Create new relationship type dialog */}
            <Dialog
                open={createOpen}
                onClose={() => setCreateOpen(false)}
                maxWidth="sm"
                fullWidth
            >
                <DialogTitle>Create Relationship Type</DialogTitle>
                <DialogContent>
                    <Stack spacing={2} sx={{ mt: 1 }}>
                        <TextField
                            label="Name (snake_case)"
                            size="small"
                            fullWidth
                            value={createName}
                            onChange={(e) =>
                                setCreateName(e.target.value)
                            }
                        />
                        <TextField
                            label="Display Label"
                            size="small"
                            fullWidth
                            value={createDisplayLabel}
                            onChange={(e) =>
                                setCreateDisplayLabel(e.target.value)
                            }
                        />
                        <TextField
                            label="Inverse Name (snake_case)"
                            size="small"
                            fullWidth
                            required
                            value={createInverseName}
                            onChange={(e) =>
                                setCreateInverseName(e.target.value)
                            }
                        />
                        <TextField
                            label="Inverse Display Label"
                            size="small"
                            fullWidth
                            required
                            value={createInverseDisplayLabel}
                            onChange={(e) =>
                                setCreateInverseDisplayLabel(
                                    e.target.value,
                                )
                            }
                        />
                        <FormControlLabel
                            control={
                                <Checkbox
                                    checked={createIsSymmetric}
                                    onChange={(e) =>
                                        setCreateIsSymmetric(
                                            e.target.checked,
                                        )
                                    }
                                />
                            }
                            label="Is Symmetric"
                        />
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setCreateOpen(false)}>
                        Cancel
                    </Button>
                    <Button
                        variant="contained"
                        onClick={handleCreateSubmit}
                        disabled={
                            !createName ||
                            !createInverseName ||
                            !createInverseDisplayLabel ||
                            createRelType.isPending
                        }
                    >
                        {createRelType.isPending
                            ? 'Creating...'
                            : 'Create'}
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
}

/**
 * A left-panel entry is either an individual item or an entity group
 * (for enrichment types that are grouped by entity in the left panel).
 */
type LeftPanelEntry =
    | { kind: 'item'; item: ContentAnalysisItem }
    | {
          kind: 'entityGroup';
          entityId: number;
          detectionType: string;
          itemIds: number[];
      };

/**
 * Full-screen triage page for reviewing content analysis results.
 *
 * Displays pending items grouped by detection type in a left panel and
 * a detail view with resolution actions in a right panel. Resolved
 * items are collected into a collapsible "Accepted Changes" section
 * at the bottom of the left panel.
 */
export default function AnalysisTriagePage() {
    const { campaignId, jobId } = useParams<{
        campaignId: string;
        jobId: string;
    }>();
    const navigate = useNavigate();

    const numericCampaignId = Number(campaignId);
    const numericJobId = Number(jobId);

    const { data: job, isLoading: jobLoading } = useAnalysisJob(
        numericCampaignId,
        numericJobId
    );
    const { data: items, isLoading: itemsLoading } = useAnalysisItems(
        numericCampaignId,
        numericJobId
    );
    const resolveItem = useResolveItem(numericCampaignId);
    const batchResolve = useBatchResolve(numericCampaignId);
    const revertItem = useRevertItem(numericCampaignId);
    const { data: userSettings } = useUserSettings();
    // Available for future use: trigger enrichment manually
    useTriggerEnrichment(numericCampaignId);
    const cancelEnrichment = useCancelEnrichment(numericCampaignId);

    // Determine if enrichment is in progress
    const isEnriching = job?.status === 'enriching';

    // Poll when all Phase 1 items are resolved but enrichment has not yet
    // started producing items. The backend sets status = 'enriching' after
    // the HTTP response returns, so without this the UI misses the
    // transition.
    const awaitingEnrichment = !!(
        job &&
        job.resolvedItems === job.totalItems &&
        job.totalItems > 0 &&
        job.enrichmentTotal === 0 &&
        job.status !== 'enriching' &&
        job.status !== 'completed'
    );

    // Enable enrichment stream polling when enriching OR awaiting
    useEnrichmentStream(
        numericCampaignId,
        numericJobId,
        isEnriching || awaitingEnrichment,
    );

    // Check if LLM is configured
    const hasLLMConfigured = !!userSettings?.contentGenService && !!userSettings?.contentGenApiKey;

    const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
    const [newEntityName, setNewEntityName] = useState('');
    const [newEntityType, setNewEntityType] = useState<EntityType>('npc');
    const [showNewEntityForm, setShowNewEntityForm] = useState(false);
    const [showDoneDialog, setShowDoneDialog] = useState(false);
    const [showCancelEnrichDialog, setShowCancelEnrichDialog] = useState(false);
    const [, setEditedDescription] = useState<string>('');
    const [selectedEntityGroup, setSelectedEntityGroup] = useState<{
        entityId: number;
        detectionType: string;
    } | null>(null);

    // Relationship type overrides keyed by analysis item ID.
    const [editedRelTypes, setEditedRelTypes] = useState<Map<number, string>>(
        () => new Map(),
    );

    // Fetch available relationship types for the campaign.
    const { data: relationshipTypes } = useRelationshipTypes(
        numericCampaignId,
    );

    const isLoading = jobLoading || itemsLoading;

    /**
     * Split items by phase and resolution status.
     */
    const identificationItems = useMemo(() => {
        if (!items) return [];
        return items.filter((i) => !i.phase || i.phase === 'identification');
    }, [items]);

    const enrichmentItems = useMemo(() => {
        if (!items) return [];
        return items.filter((i) => i.phase === 'enrichment');
    }, [items]);

    // Enrichment status
    const enrichmentStatus = useMemo(() => {
        if (!job || job.totalItems === 0) return null;
        if (!hasLLMConfigured) return null;

        if (isEnriching) {
            const resolved = job.enrichmentResolved ?? 0;
            const total = job.enrichmentTotal ?? 0;
            return {
                type: 'enriching' as const,
                message: `Analysing entities... (${resolved} of ${total} complete)`,
            };
        }

        if (job.resolvedItems < job.totalItems) {
            const remaining = job.totalItems - job.resolvedItems;
            return {
                type: 'waiting' as const,
                message: `Resolve ${remaining} remaining item${remaining === 1 ? '' : 's'} to trigger enrichment`,
            };
        }

        if (job.enrichmentTotal > 0 && job.enrichmentResolved === job.enrichmentTotal && enrichmentItems.length === 0) {
            return {
                type: 'complete' as const,
                message: 'Enrichment complete — all suggestions reviewed',
            };
        }

        if (enrichmentItems.length > 0) {
            const pending = enrichmentItems.filter(i => i.resolution === 'pending').length;
            if (pending > 0) {
                return {
                    type: 'ready' as const,
                    message: `${pending} enrichment suggestion${pending === 1 ? '' : 's'} to review`,
                };
            }
            return {
                type: 'complete' as const,
                message: 'Enrichment complete — all suggestions reviewed',
            };
        }

        return null;
    }, [job, hasLLMConfigured, isEnriching, enrichmentItems]);

    // Current stage label for the title bar
    const stageLabel = useMemo(() => {
        if (!job) return undefined;
        const source = `${job.sourceTable} — ${job.sourceField}`;

        if (isEnriching || enrichmentItems.length > 0) {
            return `Step 2 of 2: Entity Enrichment — ${source}`;
        }

        if (job.resolvedItems === job.totalItems && job.totalItems > 0 && hasLLMConfigured) {
            return `Step 2 of 2: Entity Enrichment — ${source}`;
        }

        if (!hasLLMConfigured) {
            return `${source}`;
        }

        return `Step 1 of 2: Entity Detection — ${source}`;
    }, [job, isEnriching, enrichmentItems, hasLLMConfigured]);

    const pendingIdentificationItems = useMemo(() => {
        return identificationItems.filter((i) => i.resolution === 'pending');
    }, [identificationItems]);

    const pendingEnrichmentItems = useMemo(() => {
        return enrichmentItems.filter((i) => i.resolution === 'pending');
    }, [enrichmentItems]);

    const resolvedItems = useMemo(() => {
        if (!items) return [];
        return items.filter((i) => i.resolution !== 'pending');
    }, [items]);

    const pendingItems = useMemo(() => {
        return [...pendingIdentificationItems, ...pendingEnrichmentItems];
    }, [pendingIdentificationItems, pendingEnrichmentItems]);

    /**
     * Group only pending identification items by detection type for the
     * Phase 1 sections.
     */
    const groupedPendingItems = useMemo(() => {
        const groups: Record<string, ContentAnalysisItem[]> = {};
        for (const group of DETECTION_GROUPS) {
            const groupItems = pendingIdentificationItems.filter(
                (item) => item.detectionType === group.key
            );
            if (groupItems.length > 0) {
                groups[group.key] = groupItems;
            }
        }
        return groups;
    }, [pendingIdentificationItems]);

    const selectedItem = useMemo(() => {
        if (!items || selectedItemId === null) return null;
        return items.find((item) => item.id === selectedItemId) ?? null;
    }, [items, selectedItemId]);

    /**
     * Derive the items for the currently selected entity group from
     * live data so the list updates automatically when items are resolved.
     */
    const selectedGroupItems = useMemo(() => {
        if (!selectedEntityGroup || !items) return [];
        return items.filter(
            (i) =>
                i.phase === 'enrichment' &&
                i.entityId === selectedEntityGroup.entityId &&
                i.detectionType === selectedEntityGroup.detectionType &&
                i.resolution === 'pending',
        );
    }, [selectedEntityGroup, items]);

    const selectedGroupEntityName =
        selectedGroupItems.length > 0
            ? selectedGroupItems[0].matchedText
            : '';
    const selectedGroupEntityType =
        selectedGroupItems.length > 0
            ? selectedGroupItems[0].entityType
            : undefined;

    /**
     * Derive the initial edited-description value from the selected item.
     * The MarkdownEditor already carries key={selectedItem.id} so it
     * remounts whenever the selection changes, picking up this value
     * via its value prop.  User edits are tracked separately in the
     * editedDescription state (via onChange).
     */
    const initialEditedDescription = useMemo(() => {
        if (
            selectedItem?.phase === 'enrichment' &&
            selectedItem?.detectionType === 'description_update'
        ) {
            const suggestion = selectedItem.suggestedContent as Record<string, unknown> | null;
            return String(suggestion?.suggestedDescription ?? '');
        }
        return '';
    }, [selectedItem]);

    const resolvedCount = resolvedItems.length;
    const totalCount = items ? items.length : 0;
    const allResolved = resolvedCount === totalCount && totalCount > 0;

    // System status message for the title bar
    const systemStatus = useMemo(() => {
        if (!job) return null;

        // Phase 1: still reviewing
        if (job.resolvedItems < job.totalItems) {
            if (hasLLMConfigured) {
                return {
                    message: 'Review detections below, then enrichment will analyse linked entities',
                    color: 'text.secondary' as const,
                    showSpinner: false,
                };
            }
            return null; // No extra status needed when no LLM
        }

        // All Phase 1 done, enrichment starting/waiting
        if (isEnriching) {
            const resolved = job.enrichmentResolved ?? 0;
            const total = job.enrichmentTotal ?? 0;
            return {
                message: total > 0
                    ? `Analysing entities... (${resolved} of ${total})`
                    : 'Starting entity analysis...',
                color: 'info.main' as const,
                showSpinner: true,
            };
        }

        // All Phase 1 done, waiting for enrichment to start
        if (hasLLMConfigured && job.enrichmentTotal === 0 && job.status !== 'completed') {
            return {
                message: 'Starting entity analysis...',
                color: 'info.main' as const,
                showSpinner: true,
            };
        }

        // Enrichment items ready to review
        if (enrichmentItems.length > 0) {
            const pendingEnrichment = enrichmentItems.filter(i => i.resolution === 'pending').length;
            if (pendingEnrichment > 0) {
                return {
                    message: `${pendingEnrichment} enrichment suggestion${pendingEnrichment === 1 ? '' : 's'} ready for review`,
                    color: 'success.main' as const,
                    showSpinner: false,
                };
            }
        }

        // Everything done
        if (allResolved) {
            return {
                message: 'All items reviewed',
                color: 'success.main' as const,
                showSpinner: false,
            };
        }

        return null;
    }, [job, isEnriching, enrichmentItems, hasLLMConfigured, allResolved]);

    /**
     * Build the return path based on the job's source table and ID.
     */
    const getReturnPath = useCallback(() => {
        const base = `/campaigns/${campaignId}`;
        if (!job) return `${base}/overview`;
        switch (job.sourceTable) {
            case 'campaigns':
                return `${base}/overview`;
            case 'entities':
                return `${base}/entities/${job.sourceId}/edit`;
            case 'chapters':
                return `${base}/chapters/${job.sourceId}/edit`;
            case 'sessions':
                return `${base}/sessions`;
            default:
                return `${base}/overview`;
        }
    }, [campaignId, job]);

    /**
     * Build a list representing the visual order of left-panel entries.
     *
     * Each entry is either:
     * - An individual item (Phase 1 items, description_update,
     *   new_entity_suggestion), or
     * - An entity group (log_entry, relationship_suggestion grouped by
     *   entityId).
     *
     * This mirrors the rendering order: DETECTION_GROUPS first, then
     * ENRICHMENT_GROUPS, each containing their items/groups in order.
     */
    const leftPanelOrder = useMemo((): LeftPanelEntry[] => {
        const entries: LeftPanelEntry[] = [];

        // Phase 1: identification items listed individually per
        // detection group.
        for (const group of DETECTION_GROUPS) {
            const groupItems = groupedPendingItems[group.key];
            if (!groupItems || groupItems.length === 0) continue;
            for (const item of groupItems) {
                entries.push({ kind: 'item', item });
            }
        }

        // Phase 2: enrichment items. log_entry and
        // relationship_suggestion are grouped by entity; others
        // are listed individually.
        for (const group of ENRICHMENT_GROUPS) {
            const groupItems = pendingEnrichmentItems.filter(
                (item) => item.detectionType === group.key,
            );
            if (groupItems.length === 0) continue;

            if (
                group.key === 'log_entry' ||
                group.key === 'relationship_suggestion'
            ) {
                const entityGroups = new Map<
                    number,
                    ContentAnalysisItem[]
                >();
                for (const item of groupItems) {
                    const eid = item.entityId ?? 0;
                    if (!entityGroups.has(eid))
                        entityGroups.set(eid, []);
                    entityGroups.get(eid)!.push(item);
                }
                for (const [entityId, entityItems] of entityGroups) {
                    entries.push({
                        kind: 'entityGroup',
                        entityId,
                        detectionType: group.key,
                        itemIds: entityItems.map((i) => i.id),
                    });
                }
            } else {
                for (const item of groupItems) {
                    entries.push({ kind: 'item', item });
                }
            }
        }

        return entries;
    }, [groupedPendingItems, pendingEnrichmentItems]);

    /**
     * Advance the selection to the next pending item after a resolution,
     * excluding the item(s) that were just resolved.
     *
     * Respects the left panel's visual order and grouping:
     * - If the resolved item belongs to an entity group that still has
     *   remaining pending items, stay on that entity group.
     * - Otherwise, advance to the next visible entry in the left panel
     *   (which may be an individual item or an entity group).
     */
    const advanceToNextPending = useCallback(
        (resolvedIds: number | number[]) => {
            const excluded = new Set(
                Array.isArray(resolvedIds) ? resolvedIds : [resolvedIds],
            );

            // If we are currently viewing an entity group, check
            // whether it still has pending items after excluding the
            // resolved ones.
            if (selectedEntityGroup) {
                const remainingInGroup = leftPanelOrder.find(
                    (entry) =>
                        entry.kind === 'entityGroup' &&
                        entry.entityId ===
                            selectedEntityGroup.entityId &&
                        entry.detectionType ===
                            selectedEntityGroup.detectionType &&
                        entry.itemIds.some(
                            (id) => !excluded.has(id),
                        ),
                );
                if (remainingInGroup) {
                    // Group still has pending items -- stay on it.
                    // Clear selectedItemId to ensure the group view
                    // is shown rather than a single-item view.
                    setSelectedItemId(null);
                    return;
                }
            }

            // Find the first left-panel entry that is not fully
            // covered by the excluded set.
            const nextEntry = leftPanelOrder.find((entry) => {
                if (entry.kind === 'item') {
                    return !excluded.has(entry.item.id);
                }
                // Entity group: at least one item must remain.
                return entry.itemIds.some(
                    (id) => !excluded.has(id),
                );
            });

            if (nextEntry) {
                if (nextEntry.kind === 'item') {
                    setSelectedItemId(nextEntry.item.id);
                    setSelectedEntityGroup(null);
                } else {
                    setSelectedItemId(null);
                    setSelectedEntityGroup({
                        entityId: nextEntry.entityId,
                        detectionType: nextEntry.detectionType,
                    });
                }
            } else {
                setSelectedItemId(null);
                setSelectedEntityGroup(null);
            }
        },
        [leftPanelOrder, selectedEntityGroup],
    );

    /**
     * Handle resolving an enrichment item (Phase 2). For dismissed items,
     * simply mark as dismissed. For accepted items, mark as accepted; the
     * backend applies the suggestion when resolving.
     */
    const handleEnrichmentResolve = useCallback(
        async (
            itemId: number,
            resolution: 'accepted' | 'dismissed',
            item: ContentAnalysisItem,
        ) => {
            const onSuccessAdvance = {
                onSuccess: () => advanceToNextPending(itemId),
            };

            if (resolution === 'dismissed') {
                resolveItem.mutate(
                    {
                        itemId,
                        req: { resolution: 'dismissed' },
                    },
                    onSuccessAdvance,
                );
                return;
            }

            try {
                const suggestion = item.suggestedContent;
                if (!suggestion) {
                    resolveItem.mutate(
                        {
                            itemId,
                            req: { resolution: 'accepted' },
                        },
                        onSuccessAdvance,
                    );
                    return;
                }

                // If the user overrode the relationship type, include
                // the override so the backend can use the corrected value.
                const overriddenType = editedRelTypes.get(itemId);
                const override =
                    item.detectionType === 'relationship_suggestion' &&
                    overriddenType
                        ? {
                              relationshipType: overriddenType,
                          }
                        : undefined;

                resolveItem.mutate(
                    {
                        itemId,
                        req: {
                            resolution: 'accepted',
                            suggestedContentOverride: override,
                        },
                    },
                    onSuccessAdvance,
                );
            } catch (error) {
                console.error(
                    'Failed to apply enrichment suggestion:',
                    error,
                );
            }
        },
        [resolveItem, editedRelTypes, advanceToNextPending],
    );

    if (Number.isNaN(numericCampaignId) || Number.isNaN(numericJobId)) {
        return (
            <Box sx={{ p: 4 }}>
                <Alert severity="error">
                    Invalid campaign or job ID.
                </Alert>
            </Box>
        );
    }

    /**
     * Handle the Done button: navigate immediately if all items are
     * resolved, otherwise show the confirmation dialog.
     */
    const handleDone = () => {
        // If enrichment is running, show confirmation dialog
        if (isEnriching) {
            setShowCancelEnrichDialog(true);
            return;
        }
        if (hasLLMConfigured && job &&
            job.resolvedItems === job.totalItems &&
            job.totalItems > 0 &&
            job.enrichmentTotal === 0 &&
            job.status !== 'completed') {
            return; // Stay — enrichment is about to trigger
        }

        if (allResolved) {
            navigate(getReturnPath());
        } else {
            setShowDoneDialog(true);
        }
    };

    /**
     * Handle confirming cancellation of enrichment and leaving.
     */
    const handleCancelEnrichAndLeave = async () => {
        setShowCancelEnrichDialog(false);
        try {
            await cancelEnrichment.mutateAsync(numericJobId);
        } catch (error) {
            console.error('Failed to cancel enrichment:', error);
        }
        navigate(getReturnPath());
    };

    /**
     * Handle resolving an item with the given resolution.
     */
    const handleResolve = (
        itemId: number,
        resolution: 'accepted' | 'new_entity' | 'dismissed',
        entityName?: string,
        entityType?: EntityType
    ) => {
        resolveItem.mutate(
            {
                itemId,
                req: {
                    resolution,
                    ...(entityName && { entityName }),
                    ...(entityType && { entityType }),
                },
            },
            {
                onSuccess: () => advanceToNextPending(itemId),
            },
        );
        setShowNewEntityForm(false);
    };

    /**
     * Handle batch-resolving all pending items of a detection type.
     */
    const handleBatchResolve = (detectionType: string) => {
        const resolvedIds = pendingItems
            .filter((item) => item.detectionType === detectionType)
            .map((item) => item.id);
        batchResolve.mutate(
            {
                jobId: numericJobId,
                detectionType,
                resolution: 'accepted',
            },
            {
                onSuccess: () => advanceToNextPending(resolvedIds),
            },
        );
    };

    /**
     * Handle reverting a resolved item back to pending.
     */
    const handleRevert = (itemId: number) => {
        revertItem.mutate({ itemId });
    };

    /**
     * Handle selecting an item and pre-filling the new entity form.
     */
    const handleSelectItem = (item: ContentAnalysisItem) => {
        setSelectedItemId(item.id);
        setSelectedEntityGroup(null);
        setNewEntityName(item.matchedText);
        setNewEntityType('npc');
        setShowNewEntityForm(false);
    };

    /**
     * Highlight matched text within a context snippet.
     */
    const highlightContext = (
        context: string,
        matchedText: string
    ) => {
        const index = context.toLowerCase().indexOf(matchedText.toLowerCase());
        if (index === -1) {
            return <Typography variant="body2">{context}</Typography>;
        }
        const before = context.slice(0, index);
        const match = context.slice(index, index + matchedText.length);
        const after = context.slice(index + matchedText.length);
        return (
            <Typography variant="body2">
                {before}
                <Box
                    component="mark"
                    sx={{
                        bgcolor: 'warning.light',
                        px: 0.5,
                        borderRadius: 0.5,
                    }}
                >
                    {match}
                </Box>
                {after}
            </Typography>
        );
    };

    if (isLoading) {
        return (
            <FullScreenLayout
                title="Content Analysis"
                backPath={`/campaigns/${campaignId}/overview`}
                showSaveButtons={false}
            >
                <Box
                    sx={{
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'center',
                        height: '100%',
                    }}
                >
                    <CircularProgress />
                </Box>
            </FullScreenLayout>
        );
    }

    if (!items || items.length === 0) {
        const jobInProgress = job?.status === 'enriching'
            || job?.status === 'analyzing'
            || awaitingEnrichment;

        return (
            <FullScreenLayout
                title="Content Analysis"
                subtitle={
                    job
                        ? `${job.sourceTable} — ${job.sourceField}`
                        : undefined
                }
                backPath={`/campaigns/${campaignId}/overview`}
                showSaveButtons={false}
            >
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        justifyContent: 'center',
                        alignItems: 'center',
                        height: '100%',
                        gap: 2,
                    }}
                >
                    {jobInProgress ? (
                        <>
                            <Alert
                                severity="info"
                                icon={<CircularProgress size={20} />}
                                sx={{ maxWidth: 520 }}
                            >
                                Enrichment in progress — analysing
                                content with AI. New suggestions will
                                appear automatically.
                            </Alert>
                            <LinearProgress
                                sx={{ width: '100%', maxWidth: 520 }}
                            />
                        </>
                    ) : (
                        <Typography color="text.secondary">
                            No items to review
                        </Typography>
                    )}
                </Box>
            </FullScreenLayout>
        );
    }

    return (
        <FullScreenLayout
            title="Content Analysis"
            subtitle={stageLabel}
            backPath={`/campaigns/${campaignId}/overview`}
            showSaveButtons={false}
            actions={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    {systemStatus && (
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            {systemStatus.showSpinner && (
                                <CircularProgress size={16} />
                            )}
                            <Typography
                                variant="body2"
                                color={systemStatus.color}
                                sx={{ whiteSpace: 'nowrap' }}
                            >
                                {systemStatus.message}
                            </Typography>
                        </Box>
                    )}
                    <Chip
                        label={`${resolvedCount} of ${totalCount} reviewed`}
                        size="small"
                        color={allResolved ? 'success' : 'default'}
                    />
                    <Button
                        variant="contained"
                        size="small"
                        color={allResolved ? 'success' : 'primary'}
                        onClick={handleDone}
                    >
                        Done
                    </Button>
                </Box>
            }
        >
            <Box
                sx={{
                    display: 'flex',
                    gap: 2,
                    height: '100%',
                    minHeight: 0,
                }}
            >
                {/* Left panel - Item list */}
                <Paper
                    sx={{
                        width: '40%',
                        overflow: 'auto',
                        flexShrink: 0,
                    }}
                >
                    <List disablePadding>
                        {/* No LLM configured notice */}
                        {!hasLLMConfigured && job && job.totalItems > 0 && (
                            <Box sx={{ px: 2, pt: 1.5, pb: 0.5 }}>
                                <Alert
                                    severity="info"
                                    sx={{ fontSize: '0.8125rem' }}
                                >
                                    Configure an LLM service in{' '}
                                    <Link
                                        to="/settings"
                                        style={{ color: 'inherit' }}
                                    >
                                        Account Settings
                                    </Link>{' '}
                                    to enable entity enrichment suggestions.
                                </Alert>
                            </Box>
                        )}
                        {DETECTION_GROUPS.map((group) => {
                            const groupItems =
                                groupedPendingItems[group.key];
                            if (!groupItems || groupItems.length === 0) {
                                return null;
                            }
                            return (
                                <Box key={group.key}>
                                    <Box
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 1,
                                            px: 2,
                                            py: 1.5,
                                            bgcolor: 'background.default',
                                        }}
                                    >
                                        <Box
                                            sx={{
                                                width: 10,
                                                height: 10,
                                                borderRadius: '50%',
                                                bgcolor: group.color,
                                            }}
                                        />
                                        <Typography
                                            variant="subtitle2"
                                            sx={{
                                                fontWeight: 600,
                                                flexGrow: 1,
                                            }}
                                        >
                                            {group.label} (
                                            {groupItems.length})
                                        </Typography>
                                        <Tooltip title="Accept all pending items in this group">
                                            <span>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    startIcon={
                                                        <DoneAll
                                                            fontSize="small"
                                                        />
                                                    }
                                                    onClick={() =>
                                                        handleBatchResolve(
                                                            group.key
                                                        )
                                                    }
                                                    disabled={
                                                        batchResolve.isPending
                                                    }
                                                    sx={{
                                                        textTransform: 'none',
                                                        py: 0,
                                                        minHeight: 28,
                                                    }}
                                                >
                                                    Accept All
                                                </Button>
                                            </span>
                                        </Tooltip>
                                    </Box>
                                    <Divider />
                                    {groupItems.map((item) => (
                                        <ListItemButton
                                            key={item.id}
                                            selected={
                                                selectedItemId === item.id
                                            }
                                            onClick={() =>
                                                handleSelectItem(item)
                                            }
                                            sx={{
                                                bgcolor:
                                                    selectedItemId ===
                                                    item.id
                                                        ? 'action.selected'
                                                        : undefined,
                                            }}
                                        >
                                            <ListItemText
                                                primary={
                                                    <Box
                                                        sx={{
                                                            display: 'flex',
                                                            alignItems:
                                                                'center',
                                                            gap: 1,
                                                        }}
                                                    >
                                                        <Typography
                                                            variant="body2"
                                                            sx={{
                                                                fontWeight: 600,
                                                            }}
                                                        >
                                                            {
                                                                item.matchedText
                                                            }
                                                        </Typography>
                                                        {item.entityId &&
                                                            item.entityName &&
                                                            item.entityType && (
                                                            <Chip
                                                                label={
                                                                    item.entityName
                                                                }
                                                                size="small"
                                                                color={
                                                                    entityTypeColors[
                                                                        item
                                                                            .entityType
                                                                    ] ??
                                                                        'default'
                                                                }
                                                                sx={{
                                                                    height: 20,
                                                                }}
                                                            />
                                                        )}
                                                        {item.similarity !==
                                                            undefined &&
                                                            item.detectionType ===
                                                                'misspelling' && (
                                                            <Typography
                                                                variant="caption"
                                                                color="text.secondary"
                                                            >
                                                                {Math.round(
                                                                    item.similarity *
                                                                        100
                                                                )}
                                                                %
                                                            </Typography>
                                                        )}
                                                    </Box>
                                                }
                                                secondary={
                                                    item.contextSnippet
                                                        ? item.contextSnippet.length >
                                                          80
                                                            ? `${item.contextSnippet.slice(0, 80)}...`
                                                            : item.contextSnippet
                                                        : undefined
                                                }
                                            />
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 0.5,
                                                    ml: 1,
                                                }}
                                            >
                                                <Tooltip title="Accept — link to matched entity">
                                                    <IconButton
                                                        size="small"
                                                        color="success"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleResolve(
                                                                item.id,
                                                                'accepted'
                                                            );
                                                        }}
                                                    >
                                                        <Check fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                                <Tooltip title="Create new entity from this text">
                                                    <IconButton
                                                        size="small"
                                                        color="primary"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleSelectItem(
                                                                item
                                                            );
                                                            setShowNewEntityForm(
                                                                true
                                                            );
                                                        }}
                                                    >
                                                        <AddCircle fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                                <Tooltip title="Dismiss — ignore this detection">
                                                    <IconButton
                                                        size="small"
                                                        color="error"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleResolve(
                                                                item.id,
                                                                'dismissed'
                                                            );
                                                        }}
                                                    >
                                                        <Close fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                            </Box>
                                        </ListItemButton>
                                    ))}
                                </Box>
                            );
                        })}
                    </List>

                    {/* Enrichment status */}
                    {enrichmentStatus && (
                        <Box
                            sx={{
                                px: 2,
                                py: 1.5,
                                bgcolor: 'background.default',
                                borderTop: 1,
                                borderColor: 'divider',
                            }}
                        >
                            <Typography
                                variant="subtitle2"
                                sx={{
                                    fontWeight: 700,
                                    color: 'text.primary',
                                    mb: enrichmentStatus.type === 'enriching' ? 1 : 0,
                                }}
                            >
                                Entity Enrichment
                            </Typography>
                            <Typography
                                variant="body2"
                                color={
                                    enrichmentStatus.type === 'waiting'
                                        ? 'text.secondary'
                                        : enrichmentStatus.type === 'complete'
                                        ? 'success.main'
                                        : 'text.primary'
                                }
                                sx={{ mt: 0.5 }}
                            >
                                {enrichmentStatus.message}
                            </Typography>
                            {enrichmentStatus.type === 'enriching' && (
                                <LinearProgress sx={{ mt: 1 }} />
                            )}
                        </Box>
                    )}

                    {/* Enrichment groups */}
                    {ENRICHMENT_GROUPS.map((group) => {
                        const groupItems = pendingEnrichmentItems.filter(
                            (item) => item.detectionType === group.key,
                        );
                        if (groupItems.length === 0) return null;
                        return (
                            <Box key={group.key}>
                                <Box
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 1,
                                        px: 2,
                                        py: 1.5,
                                        bgcolor: 'background.default',
                                    }}
                                >
                                    <Box
                                        sx={{
                                            width: 10,
                                            height: 10,
                                            borderRadius: '50%',
                                            bgcolor: group.color,
                                        }}
                                    />
                                    <Typography
                                        variant="subtitle2"
                                        sx={{
                                            fontWeight: 600,
                                            flexGrow: 1,
                                        }}
                                    >
                                        {group.label} (
                                        {groupItems.length})
                                    </Typography>
                                </Box>
                                <Divider />
                                {group.key === 'log_entry' ||
                                group.key === 'relationship_suggestion' ? (
                                    (() => {
                                        const entityGroups = new Map<
                                            number,
                                            ContentAnalysisItem[]
                                        >();
                                        for (const item of groupItems) {
                                            const eid =
                                                item.entityId ?? 0;
                                            if (
                                                !entityGroups.has(eid)
                                            )
                                                entityGroups.set(
                                                    eid,
                                                    [],
                                                );
                                            entityGroups
                                                .get(eid)!
                                                .push(item);
                                        }
                                        return Array.from(
                                            entityGroups.entries(),
                                        ).map(
                                            ([
                                                entityId,
                                                entityItems,
                                            ]) => {
                                                const entityName =
                                                    entityItems[0]
                                                        .matchedText;
                                                const entityType =
                                                    entityItems[0]
                                                        .entityType;
                                                return (
                                                    <ListItemButton
                                                        key={`${group.key}-${entityId}`}
                                                        selected={
                                                            selectedEntityGroup?.entityId ===
                                                                entityId &&
                                                            selectedEntityGroup?.detectionType ===
                                                                group.key
                                                        }
                                                        onClick={() => {
                                                            setSelectedItemId(
                                                                null,
                                                            );
                                                            setSelectedEntityGroup(
                                                                {
                                                                    entityId,
                                                                    detectionType:
                                                                        group.key,
                                                                },
                                                            );
                                                        }}
                                                    >
                                                        <ListItemText
                                                            primary={
                                                                <Box
                                                                    sx={{
                                                                        display:
                                                                            'flex',
                                                                        alignItems:
                                                                            'center',
                                                                        gap: 1,
                                                                    }}
                                                                >
                                                                    <Typography
                                                                        variant="body2"
                                                                        sx={{
                                                                            fontWeight: 600,
                                                                        }}
                                                                    >
                                                                        {
                                                                            entityName
                                                                        }
                                                                    </Typography>
                                                                    {entityType && (
                                                                        <Chip
                                                                            label={formatEntityType(
                                                                                entityType,
                                                                            )}
                                                                            size="small"
                                                                            color={
                                                                                entityTypeColors[
                                                                                    entityType
                                                                                ] ??
                                                                                'default'
                                                                            }
                                                                            sx={{
                                                                                height: 20,
                                                                            }}
                                                                        />
                                                                    )}
                                                                </Box>
                                                            }
                                                            secondary={`${entityItems.length} suggestion${entityItems.length === 1 ? '' : 's'}`}
                                                        />
                                                    </ListItemButton>
                                                );
                                            },
                                        );
                                    })()
                                ) : (
                                    groupItems.map((item) => (
                                        <ListItemButton
                                            key={item.id}
                                            selected={
                                                selectedItemId === item.id
                                            }
                                            onClick={() =>
                                                handleSelectItem(item)
                                            }
                                        >
                                            <ListItemText
                                                primary={
                                                    <Box
                                                        sx={{
                                                            display: 'flex',
                                                            alignItems:
                                                                'center',
                                                            gap: 1,
                                                        }}
                                                    >
                                                        <Typography
                                                            variant="body2"
                                                            sx={{
                                                                fontWeight: 600,
                                                            }}
                                                        >
                                                            {
                                                                item.matchedText
                                                            }
                                                        </Typography>
                                                        {item.entityType && (
                                                            <Chip
                                                                label={
                                                                    item.entityName ??
                                                                    formatEntityType(
                                                                        item.entityType,
                                                                    )
                                                                }
                                                                size="small"
                                                                color={
                                                                    entityTypeColors[
                                                                        item
                                                                            .entityType
                                                                    ] ??
                                                                    'default'
                                                                }
                                                                sx={{
                                                                    height: 20,
                                                                }}
                                                            />
                                                        )}
                                                        {!item.entityType &&
                                                            item.detectionType === 'new_entity_suggestion' &&
                                                            !!item.suggestedContent?.entity_type && (
                                                            <Chip
                                                                label={formatEntityType(
                                                                    item.suggestedContent.entity_type as EntityType,
                                                                )}
                                                                size="small"
                                                                sx={{
                                                                    height: 20,
                                                                    backgroundColor: entityTypeColors[item.suggestedContent.entity_type as EntityType] || '#757575',
                                                                    color: 'white',
                                                                }}
                                                            />
                                                        )}
                                                    </Box>
                                                }
                                                secondary={
                                                    item.detectionType === 'new_entity_suggestion'
                                                        ? 'New entity suggestion'
                                                        : 'Description update suggestion'
                                                }
                                            />
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 0.5,
                                                    ml: 1,
                                                }}
                                            >
                                                <Tooltip title={item.detectionType === 'new_entity_suggestion' ? 'Create Entity' : 'Accept'}>
                                                    <IconButton
                                                        size="small"
                                                        color="success"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            if (item.detectionType === 'new_entity_suggestion') {
                                                                handleResolve(
                                                                    item.id,
                                                                    'new_entity',
                                                                    item.matchedText,
                                                                    (item.suggestedContent?.entity_type as EntityType) || 'npc',
                                                                );
                                                            } else {
                                                                handleEnrichmentResolve(
                                                                    item.id,
                                                                    'accepted',
                                                                    item,
                                                                );
                                                            }
                                                        }}
                                                    >
                                                        {item.detectionType === 'new_entity_suggestion'
                                                            ? <PersonAdd fontSize="small" />
                                                            : <Check fontSize="small" />
                                                        }
                                                    </IconButton>
                                                </Tooltip>
                                                <Tooltip title="Dismiss">
                                                    <IconButton
                                                        size="small"
                                                        color="error"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            if (item.detectionType === 'new_entity_suggestion') {
                                                                handleResolve(item.id, 'dismissed');
                                                            } else {
                                                                handleEnrichmentResolve(
                                                                    item.id,
                                                                    'dismissed',
                                                                    item,
                                                                );
                                                            }
                                                        }}
                                                    >
                                                        <Close fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                            </Box>
                                        </ListItemButton>
                                    ))
                                )}
                            </Box>
                        );
                    })}

                    {/* Accepted Changes - collapsible section */}
                    {resolvedItems.length > 0 && (
                        <Accordion
                            defaultExpanded={false}
                            disableGutters
                            elevation={0}
                            square
                            sx={{
                                '&:before': { display: 'none' },
                                borderTop: 1,
                                borderColor: 'divider',
                            }}
                        >
                            <AccordionSummary
                                expandIcon={<ExpandMore />}
                                sx={{
                                    bgcolor: 'background.default',
                                    minHeight: 44,
                                    '& .MuiAccordionSummary-content': {
                                        my: 0,
                                    },
                                }}
                            >
                                <Typography
                                    variant="subtitle2"
                                    sx={{ fontWeight: 600 }}
                                >
                                    Accepted Changes (
                                    {resolvedItems.length})
                                </Typography>
                            </AccordionSummary>
                            <AccordionDetails sx={{ p: 0 }}>
                                <List disablePadding>
                                    {resolvedItems.map((item) => (
                                        <ListItemButton
                                            key={item.id}
                                            selected={
                                                selectedItemId === item.id
                                            }
                                            onClick={() =>
                                                handleSelectItem(item)
                                            }
                                            sx={{
                                                opacity: 0.7,
                                                bgcolor:
                                                    selectedItemId ===
                                                    item.id
                                                        ? 'action.selected'
                                                        : undefined,
                                            }}
                                        >
                                            <ListItemText
                                                primary={
                                                    <Box
                                                        sx={{
                                                            display: 'flex',
                                                            alignItems:
                                                                'center',
                                                            gap: 1,
                                                        }}
                                                    >
                                                        <Typography
                                                            variant="body2"
                                                            sx={{
                                                                fontWeight: 600,
                                                            }}
                                                        >
                                                            {
                                                                item.matchedText
                                                            }
                                                        </Typography>
                                                        {item.entityId &&
                                                            item.entityName &&
                                                            item.entityType && (
                                                            <Chip
                                                                label={
                                                                    item.entityName
                                                                }
                                                                size="small"
                                                                color={
                                                                    entityTypeColors[
                                                                        item
                                                                            .entityType
                                                                    ] ??
                                                                        'default'
                                                                }
                                                                sx={{
                                                                    height: 20,
                                                                }}
                                                            />
                                                        )}
                                                        <Chip
                                                            label={
                                                                item.resolution
                                                            }
                                                            size="small"
                                                            variant="outlined"
                                                            sx={{
                                                                height: 20,
                                                            }}
                                                        />
                                                    </Box>
                                                }
                                            />
                                            <Tooltip title="Revert">
                                                <span>
                                                    <IconButton
                                                        size="small"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleRevert(
                                                                item.id
                                                            );
                                                        }}
                                                        disabled={
                                                            revertItem.isPending
                                                        }
                                                    >
                                                        <Undo fontSize="small" />
                                                    </IconButton>
                                                </span>
                                            </Tooltip>
                                        </ListItemButton>
                                    ))}
                                </List>
                            </AccordionDetails>
                        </Accordion>
                    )}
                </Paper>

                {/* Right panel - Detail view */}
                <Paper
                    sx={{
                        width: '60%',
                        overflow: 'auto',
                        p: 3,
                    }}
                >
                    {selectedEntityGroup &&
                    selectedGroupItems.length > 0 ? (
                        /* Entity group detail view (grouped log entries / relationships) */
                        <Stack spacing={3}>
                            {/* Entity header */}
                            <Box>
                                <Box
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 1,
                                        mb: 1,
                                    }}
                                >
                                    <Typography
                                        variant="h6"
                                        sx={{ fontWeight: 600 }}
                                    >
                                        {selectedGroupEntityName}
                                    </Typography>
                                    {selectedGroupEntityType && (
                                        <Chip
                                            label={formatEntityType(
                                                selectedGroupEntityType,
                                            )}
                                            size="small"
                                            color={
                                                entityTypeColors[
                                                    selectedGroupEntityType
                                                ] ?? 'default'
                                            }
                                        />
                                    )}
                                </Box>
                                <Typography
                                    variant="body2"
                                    color="text.secondary"
                                >
                                    {selectedEntityGroup.detectionType ===
                                    'log_entry'
                                        ? `${selectedGroupItems.length} suggested log entr${selectedGroupItems.length === 1 ? 'y' : 'ies'}`
                                        : `${selectedGroupItems.length} suggested relationship${selectedGroupItems.length === 1 ? '' : 's'}`}
                                </Typography>
                            </Box>
                            <Divider />
                            {selectedGroupItems.map((item) => {
                                const suggestion =
                                    item.suggestedContent as Record<
                                        string,
                                        unknown
                                    > | null;
                                if (!suggestion) return null;

                                if (
                                    selectedEntityGroup.detectionType ===
                                    'log_entry'
                                ) {
                                    return (
                                        <Paper
                                            key={item.id}
                                            variant="outlined"
                                            sx={{ p: 2 }}
                                        >
                                            <Typography
                                                variant="body2"
                                                sx={{ mb: 1.5 }}
                                            >
                                                {String(
                                                    suggestion.content ??
                                                        '',
                                                )}
                                            </Typography>
                                            {!!suggestion.occurredAt && (
                                                <Typography
                                                    variant="caption"
                                                    color="text.secondary"
                                                    sx={{
                                                        display: 'block',
                                                        mb: 1.5,
                                                    }}
                                                >
                                                    Occurred:{' '}
                                                    {String(
                                                        suggestion.occurredAt,
                                                    )}
                                                </Typography>
                                            )}
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 1,
                                                }}
                                            >
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    color="success"
                                                    startIcon={
                                                        <Check />
                                                    }
                                                    onClick={() =>
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'accepted',
                                                            item,
                                                        )
                                                    }
                                                >
                                                    Accept
                                                </Button>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    color="error"
                                                    startIcon={
                                                        <Close />
                                                    }
                                                    onClick={() =>
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'dismissed',
                                                            item,
                                                        )
                                                    }
                                                >
                                                    Dismiss
                                                </Button>
                                            </Box>
                                        </Paper>
                                    );
                                }

                                if (
                                    selectedEntityGroup.detectionType ===
                                    'relationship_suggestion'
                                ) {
                                    return (
                                        <Paper
                                            key={item.id}
                                            variant="outlined"
                                            sx={{ p: 2 }}
                                        >
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    alignItems:
                                                        'center',
                                                    gap: 1,
                                                    mb: 1,
                                                    flexWrap: 'wrap',
                                                }}
                                            >
                                                <Chip
                                                    label={String(
                                                        suggestion.sourceEntityName ??
                                                            'Source',
                                                    )}
                                                    size="small"
                                                />
                                                <RelationshipTypeAutocomplete
                                                    value={
                                                        editedRelTypes.get(
                                                            item.id,
                                                        ) ??
                                                        String(
                                                            suggestion.relationshipType ??
                                                                '',
                                                        )
                                                    }
                                                    onChange={(
                                                        typeName,
                                                    ) => {
                                                        setEditedRelTypes(
                                                            (prev) => {
                                                                const next =
                                                                    new Map(
                                                                        prev,
                                                                    );
                                                                next.set(
                                                                    item.id,
                                                                    typeName,
                                                                );
                                                                return next;
                                                            },
                                                        );
                                                    }}
                                                    relationshipTypes={
                                                        relationshipTypes ??
                                                        []
                                                    }
                                                    campaignId={
                                                        numericCampaignId
                                                    }
                                                />
                                                <Chip
                                                    label={String(
                                                        suggestion.targetEntityName ??
                                                            'Target',
                                                    )}
                                                    size="small"
                                                />
                                            </Box>
                                            {!!suggestion.description && (
                                                <Typography
                                                    variant="body2"
                                                    color="text.secondary"
                                                    sx={{ mb: 1.5 }}
                                                >
                                                    {String(
                                                        suggestion.description,
                                                    )}
                                                </Typography>
                                            )}
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    gap: 1,
                                                }}
                                            >
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    color="success"
                                                    startIcon={
                                                        <Check />
                                                    }
                                                    onClick={() =>
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'accepted',
                                                            item,
                                                        )
                                                    }
                                                >
                                                    Accept
                                                </Button>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    color="error"
                                                    startIcon={
                                                        <Close />
                                                    }
                                                    onClick={() =>
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'dismissed',
                                                            item,
                                                        )
                                                    }
                                                >
                                                    Dismiss
                                                </Button>
                                            </Box>
                                        </Paper>
                                    );
                                }

                                return null;
                            })}
                        </Stack>
                    ) : selectedItem ? (
                        selectedItem.phase === 'enrichment' ? (
                            /* Enrichment detail view (Phase 2) */
                            <Stack spacing={3}>
                                {/* Entity header */}
                                <Box>
                                    <Box
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 1,
                                            mb: 1,
                                        }}
                                    >
                                        <Typography
                                            variant="h6"
                                            sx={{ fontWeight: 600 }}
                                        >
                                            {selectedItem.matchedText}
                                        </Typography>
                                        {selectedItem.entityType && (
                                            <Chip
                                                label={formatEntityType(
                                                    selectedItem.entityType,
                                                )}
                                                size="small"
                                                color={
                                                    entityTypeColors[
                                                        selectedItem.entityType
                                                    ] ?? 'default'
                                                }
                                            />
                                        )}
                                    </Box>
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        {selectedItem.detectionType ===
                                        'description_update'
                                            ? 'Suggested description update'
                                            : selectedItem.detectionType ===
                                              'log_entry'
                                              ? 'Suggested log entry'
                                              : selectedItem.detectionType ===
                                                'relationship_suggestion'
                                                ? 'Suggested relationship'
                                                : selectedItem.detectionType ===
                                                  'new_entity_suggestion'
                                                  ? 'New entity suggestion'
                                                  : 'Enrichment suggestion'}
                                    </Typography>
                                </Box>

                                <Divider />

                                {/* Description update detail */}
                                {selectedItem.detectionType ===
                                    'description_update' &&
                                    ((): React.ReactNode => {
                                        let suggestion: Record<
                                            string,
                                            unknown
                                        > | null = null;
                                        try {
                                            suggestion =
                                                selectedItem.suggestedContent ??
                                                null;
                                        } catch {
                                            suggestion = null;
                                        }
                                        if (!suggestion) {
                                            return (
                                                <Typography
                                                    variant="body2"
                                                    color="text.secondary"
                                                >
                                                    No suggestion data
                                                    available.
                                                </Typography>
                                            );
                                        }
                                        return (
                                            <Stack spacing={2}>
                                                <Box>
                                                    <Typography
                                                        variant="subtitle2"
                                                        gutterBottom
                                                    >
                                                        Current
                                                        Description
                                                    </Typography>
                                                    <Paper
                                                        variant="outlined"
                                                        sx={{
                                                            p: 2,
                                                            bgcolor:
                                                                'action.hover',
                                                        }}
                                                    >
                                                        <Typography variant="body2">
                                                            {suggestion.currentDescription !==
                                                                undefined &&
                                                            suggestion.currentDescription !==
                                                                null &&
                                                            String(
                                                                suggestion.currentDescription,
                                                            ) !== ''
                                                                ? String(
                                                                      suggestion.currentDescription,
                                                                  )
                                                                : '(No current description)'}
                                                        </Typography>
                                                    </Paper>
                                                </Box>
                                                <Box>
                                                    <Typography
                                                        variant="subtitle2"
                                                        gutterBottom
                                                    >
                                                        Suggested
                                                        Description
                                                    </Typography>
                                                    <MarkdownEditor
                                                        key={selectedItem.id}
                                                        value={initialEditedDescription}
                                                        onChange={setEditedDescription}
                                                        placeholder="Edit the suggested description..."
                                                        minHeight={120}
                                                    />
                                                </Box>
                                                {!!suggestion.rationale && (
                                                    <Box>
                                                        <Typography
                                                            variant="subtitle2"
                                                            gutterBottom
                                                        >
                                                            Rationale
                                                        </Typography>
                                                        <Typography
                                                            variant="body2"
                                                            color="text.secondary"
                                                        >
                                                            {String(
                                                                suggestion.rationale,
                                                            )}
                                                        </Typography>
                                                    </Box>
                                                )}
                                            </Stack>
                                        );
                                    })()}

                                {/* Log entry detail */}
                                {selectedItem.detectionType ===
                                    'log_entry' &&
                                    ((): React.ReactNode => {
                                        let suggestion: Record<
                                            string,
                                            unknown
                                        > | null = null;
                                        try {
                                            suggestion =
                                                selectedItem.suggestedContent ??
                                                null;
                                        } catch {
                                            suggestion = null;
                                        }
                                        if (!suggestion) {
                                            return (
                                                <Typography
                                                    variant="body2"
                                                    color="text.secondary"
                                                >
                                                    No suggestion data
                                                    available.
                                                </Typography>
                                            );
                                        }
                                        return (
                                            <Stack spacing={2}>
                                                <Box>
                                                    <Typography
                                                        variant="subtitle2"
                                                        gutterBottom
                                                    >
                                                        Suggested Log
                                                        Entry
                                                    </Typography>
                                                    <Paper
                                                        variant="outlined"
                                                        sx={{ p: 2 }}
                                                    >
                                                        <Typography variant="body2">
                                                            {String(
                                                                suggestion.content ??
                                                                    '',
                                                            )}
                                                        </Typography>
                                                    </Paper>
                                                </Box>
                                                {!!suggestion.occurredAt && (
                                                    <Box>
                                                        <Typography
                                                            variant="subtitle2"
                                                            gutterBottom
                                                        >
                                                            Occurred At
                                                        </Typography>
                                                        <Typography variant="body2">
                                                            {String(
                                                                suggestion.occurredAt,
                                                            )}
                                                        </Typography>
                                                    </Box>
                                                )}
                                                {!!suggestion.rationale && (
                                                    <Box>
                                                        <Typography
                                                            variant="subtitle2"
                                                            gutterBottom
                                                        >
                                                            Rationale
                                                        </Typography>
                                                        <Typography
                                                            variant="body2"
                                                            color="text.secondary"
                                                        >
                                                            {String(
                                                                suggestion.rationale,
                                                            )}
                                                        </Typography>
                                                    </Box>
                                                )}
                                            </Stack>
                                        );
                                    })()}

                                {/* Relationship suggestion detail */}
                                {selectedItem.detectionType ===
                                    'relationship_suggestion' &&
                                    ((): React.ReactNode => {
                                        let suggestion: Record<
                                            string,
                                            unknown
                                        > | null = null;
                                        try {
                                            suggestion =
                                                selectedItem.suggestedContent ??
                                                null;
                                        } catch {
                                            suggestion = null;
                                        }
                                        if (!suggestion) {
                                            return (
                                                <Typography
                                                    variant="body2"
                                                    color="text.secondary"
                                                >
                                                    No suggestion data
                                                    available.
                                                </Typography>
                                            );
                                        }
                                        return (
                                            <Stack spacing={2}>
                                                <Box>
                                                    <Typography
                                                        variant="subtitle2"
                                                        gutterBottom
                                                    >
                                                        Suggested
                                                        Relationship
                                                    </Typography>
                                                    <Box
                                                        sx={{
                                                            display: 'flex',
                                                            alignItems:
                                                                'center',
                                                            gap: 1,
                                                            flexWrap:
                                                                'wrap',
                                                        }}
                                                    >
                                                        <Chip
                                                            label={String(
                                                                suggestion.sourceEntityName ??
                                                                    'Source',
                                                            )}
                                                            size="small"
                                                            color={
                                                                suggestion.sourceEntityType
                                                                    ? entityTypeColors[
                                                                          suggestion.sourceEntityType as EntityType
                                                                      ] ??
                                                                      'default'
                                                                    : 'default'
                                                            }
                                                        />
                                                        <RelationshipTypeAutocomplete
                                                            value={
                                                                editedRelTypes.get(
                                                                    selectedItem.id,
                                                                ) ??
                                                                String(
                                                                    suggestion.relationshipType ??
                                                                        '',
                                                                )
                                                            }
                                                            onChange={(
                                                                typeName,
                                                            ) => {
                                                                setEditedRelTypes(
                                                                    (
                                                                        prev,
                                                                    ) => {
                                                                        const next =
                                                                            new Map(
                                                                                prev,
                                                                            );
                                                                        next.set(
                                                                            selectedItem.id,
                                                                            typeName,
                                                                        );
                                                                        return next;
                                                                    },
                                                                );
                                                            }}
                                                            relationshipTypes={
                                                                relationshipTypes ??
                                                                []
                                                            }
                                                            campaignId={
                                                                numericCampaignId
                                                            }
                                                        />
                                                        <Chip
                                                            label={String(
                                                                suggestion.targetEntityName ??
                                                                    'Target',
                                                            )}
                                                            size="small"
                                                            color={
                                                                suggestion.targetEntityType
                                                                    ? entityTypeColors[
                                                                          suggestion.targetEntityType as EntityType
                                                                      ] ??
                                                                      'default'
                                                                    : 'default'
                                                            }
                                                        />
                                                    </Box>
                                                </Box>
                                                {!!suggestion.description && (
                                                    <Box>
                                                        <Typography
                                                            variant="subtitle2"
                                                            gutterBottom
                                                        >
                                                            Description
                                                        </Typography>
                                                        <Typography variant="body2">
                                                            {String(
                                                                suggestion.description,
                                                            )}
                                                        </Typography>
                                                    </Box>
                                                )}
                                                {!!suggestion.rationale && (
                                                    <Box>
                                                        <Typography
                                                            variant="subtitle2"
                                                            gutterBottom
                                                        >
                                                            Rationale
                                                        </Typography>
                                                        <Typography
                                                            variant="body2"
                                                            color="text.secondary"
                                                        >
                                                            {String(
                                                                suggestion.rationale,
                                                            )}
                                                        </Typography>
                                                    </Box>
                                                )}
                                            </Stack>
                                        );
                                    })()}

                                {/* New entity suggestion detail */}
                                {selectedItem.detectionType ===
                                    'new_entity_suggestion' && (
                                    <Box sx={{ mt: 1 }}>
                                        <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                                            <PersonAdd fontSize="small" sx={{ color: '#7b1fa2' }} />
                                            <Typography variant="subtitle2">
                                                {selectedItem.matchedText}
                                            </Typography>
                                            {!!selectedItem.suggestedContent?.entity_type && (
                                                <Chip
                                                    label={formatEntityType(selectedItem.suggestedContent.entity_type as EntityType)}
                                                    size="small"
                                                    sx={{
                                                        backgroundColor: entityTypeColors[selectedItem.suggestedContent.entity_type as EntityType] || '#757575',
                                                        color: 'white',
                                                        fontSize: '0.7rem',
                                                    }}
                                                />
                                            )}
                                        </Stack>
                                        {!!selectedItem.suggestedContent?.description && (
                                            <Typography variant="body2" sx={{ mb: 0.5 }}>
                                                {String(selectedItem.suggestedContent.description)}
                                            </Typography>
                                        )}
                                        {!!selectedItem.suggestedContent?.reasoning && (
                                            <Typography variant="body2" sx={{ fontStyle: 'italic', color: 'text.secondary', mb: 1 }}>
                                                {String(selectedItem.suggestedContent.reasoning)}
                                            </Typography>
                                        )}
                                    </Box>
                                )}

                                {/* Resolution status */}
                                {selectedItem.resolution !== 'pending' && (
                                    <Alert severity="info">
                                        This item has been resolved as{' '}
                                        <strong>
                                            {selectedItem.resolution}
                                        </strong>
                                        .
                                    </Alert>
                                )}

                                {/* Enrichment action buttons */}
                                {selectedItem.resolution === 'pending' &&
                                 selectedItem.detectionType !== 'new_entity_suggestion' && (
                                    <Stack
                                        direction="row"
                                        spacing={1}
                                    >
                                        <Button
                                            variant="contained"
                                            color="success"
                                            startIcon={<Check />}
                                            onClick={() =>
                                                handleEnrichmentResolve(
                                                    selectedItem.id,
                                                    'accepted',
                                                    selectedItem,
                                                )
                                            }
                                            disabled={
                                                resolveItem.isPending
                                            }
                                        >
                                            Accept
                                        </Button>
                                        <Button
                                            variant="outlined"
                                            color="error"
                                            startIcon={<Close />}
                                            onClick={() =>
                                                handleEnrichmentResolve(
                                                    selectedItem.id,
                                                    'dismissed',
                                                    selectedItem,
                                                )
                                            }
                                            disabled={
                                                resolveItem.isPending
                                            }
                                        >
                                            Dismiss
                                        </Button>
                                    </Stack>
                                )}

                                {/* New entity suggestion action buttons */}
                                {selectedItem.resolution === 'pending' &&
                                 selectedItem.detectionType === 'new_entity_suggestion' && (
                                    <Stack direction="row" spacing={1}>
                                        <Button
                                            size="small"
                                            variant="contained"
                                            color="success"
                                            startIcon={<Check />}
                                            onClick={() => handleResolve(
                                                selectedItem.id,
                                                'new_entity',
                                                selectedItem.matchedText,
                                                (selectedItem.suggestedContent?.entity_type as EntityType) || 'npc',
                                            )}
                                            disabled={resolveItem.isPending}
                                        >
                                            Create Entity
                                        </Button>
                                        <Button
                                            size="small"
                                            variant="outlined"
                                            color="error"
                                            startIcon={<Close />}
                                            onClick={() => handleResolve(selectedItem.id, 'dismissed')}
                                            disabled={resolveItem.isPending}
                                        >
                                            Dismiss
                                        </Button>
                                    </Stack>
                                )}
                            </Stack>
                        ) : (
                        /* Identification detail view (Phase 1) */
                        <Stack spacing={3}>
                            {/* Context with highlighted match */}
                            <Box>
                                <Typography
                                    variant="subtitle2"
                                    gutterBottom
                                >
                                    Context
                                </Typography>
                                {selectedItem.contextSnippet ? (
                                    highlightContext(
                                        selectedItem.contextSnippet,
                                        selectedItem.matchedText
                                    )
                                ) : (
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        No context available
                                    </Typography>
                                )}
                            </Box>

                            <Divider />

                            {/* Entity match info */}
                            {selectedItem.entityId &&
                                selectedItem.entityName && (
                                <Box>
                                    <Typography
                                        variant="subtitle2"
                                        gutterBottom
                                    >
                                        Matched Entity
                                    </Typography>
                                    <Box
                                        sx={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 1,
                                        }}
                                    >
                                        <Typography variant="body1">
                                            {selectedItem.entityName}
                                        </Typography>
                                        {selectedItem.entityType && (
                                            <Chip
                                                label={formatEntityType(
                                                    selectedItem.entityType
                                                )}
                                                size="small"
                                                color={
                                                    entityTypeColors[
                                                        selectedItem
                                                            .entityType
                                                    ] ?? 'default'
                                                }
                                            />
                                        )}
                                        {selectedItem.similarity !==
                                            undefined && (
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                            >
                                                Similarity:{' '}
                                                {Math.round(
                                                    selectedItem.similarity *
                                                        100
                                                )}
                                                %
                                            </Typography>
                                        )}
                                    </Box>
                                </Box>
                            )}

                            {/* Resolution status for already-resolved items */}
                            {selectedItem.resolution !== 'pending' && (
                                <Alert severity="info">
                                    This item has been resolved as{' '}
                                    <strong>{selectedItem.resolution}</strong>.
                                </Alert>
                            )}

                            {/* Action buttons */}
                            {selectedItem.resolution === 'pending' && (
                                <Box>
                                    <Typography
                                        variant="subtitle2"
                                        gutterBottom
                                    >
                                        Actions
                                    </Typography>
                                    <Stack spacing={2}>
                                        <Stack
                                            direction="row"
                                            spacing={1}
                                        >
                                            {selectedItem.entityId && (
                                                <Button
                                                    variant="contained"
                                                    color="success"
                                                    startIcon={<Check />}
                                                    onClick={() =>
                                                        handleResolve(
                                                            selectedItem.id,
                                                            'accepted'
                                                        )
                                                    }
                                                    disabled={
                                                        resolveItem.isPending
                                                    }
                                                >
                                                    Link to{' '}
                                                    {selectedItem.entityName}
                                                </Button>
                                            )}
                                            <Button
                                                variant="outlined"
                                                color="primary"
                                                startIcon={<AddCircle />}
                                                onClick={() =>
                                                    setShowNewEntityForm(
                                                        !showNewEntityForm
                                                    )
                                                }
                                            >
                                                New Entity
                                            </Button>
                                            <Button
                                                variant="outlined"
                                                color="error"
                                                startIcon={<Close />}
                                                onClick={() =>
                                                    handleResolve(
                                                        selectedItem.id,
                                                        'dismissed'
                                                    )
                                                }
                                                disabled={
                                                    resolveItem.isPending
                                                }
                                            >
                                                Dismiss
                                            </Button>
                                        </Stack>

                                        {/* New entity form */}
                                        {showNewEntityForm && (
                                            <Paper
                                                variant="outlined"
                                                sx={{ p: 2 }}
                                            >
                                                <Typography
                                                    variant="subtitle2"
                                                    gutterBottom
                                                >
                                                    Create New Entity
                                                </Typography>
                                                <Stack spacing={2}>
                                                    <TextField
                                                        label="Entity Name"
                                                        value={newEntityName}
                                                        onChange={(e) =>
                                                            setNewEntityName(
                                                                e.target.value
                                                            )
                                                        }
                                                        size="small"
                                                        fullWidth
                                                    />
                                                    <FormControl
                                                        size="small"
                                                        fullWidth
                                                    >
                                                        <InputLabel>
                                                            Entity Type
                                                        </InputLabel>
                                                        <Select
                                                            value={
                                                                newEntityType
                                                            }
                                                            label="Entity Type"
                                                            onChange={(e) =>
                                                                setNewEntityType(
                                                                    e.target
                                                                        .value as EntityType
                                                                )
                                                            }
                                                        >
                                                            {ENTITY_TYPES.map(
                                                                (type) => (
                                                                    <MenuItem
                                                                        key={
                                                                            type
                                                                        }
                                                                        value={
                                                                            type
                                                                        }
                                                                    >
                                                                        {formatEntityType(
                                                                            type
                                                                        )}
                                                                    </MenuItem>
                                                                )
                                                            )}
                                                        </Select>
                                                    </FormControl>
                                                    <Button
                                                        variant="contained"
                                                        onClick={() =>
                                                            handleResolve(
                                                                selectedItem.id,
                                                                'new_entity',
                                                                newEntityName,
                                                                newEntityType
                                                            )
                                                        }
                                                        disabled={
                                                            !newEntityName.trim() ||
                                                            resolveItem.isPending
                                                        }
                                                    >
                                                        Create & Link
                                                    </Button>
                                                </Stack>
                                            </Paper>
                                        )}
                                    </Stack>
                                </Box>
                            )}
                        </Stack>
                        )
                    ) : (
                        <Box
                            sx={{
                                display: 'flex',
                                justifyContent: 'center',
                                alignItems: 'center',
                                height: '100%',
                            }}
                        >
                            <Typography color="text.secondary">
                                Select an item to review
                            </Typography>
                        </Box>
                    )}
                </Paper>
            </Box>

            {/* Done confirmation dialog */}
            <Dialog
                open={showDoneDialog}
                onClose={() => setShowDoneDialog(false)}
            >
                <DialogTitle>Unresolved Items</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        There are {pendingItems.length} items that
                        haven&apos;t been reviewed. Unresolved items
                        will be saved for later review.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => setShowDoneDialog(false)}
                    >
                        Continue Reviewing
                    </Button>
                    <Button
                        variant="contained"
                        onClick={() => navigate(getReturnPath())}
                    >
                        Finish Anyway
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Cancel enrichment confirmation dialog */}
            <Dialog
                open={showCancelEnrichDialog}
                onClose={() => setShowCancelEnrichDialog(false)}
            >
                <DialogTitle>Enrichment in Progress</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        LLM enrichment is still running. Leaving now will
                        cancel enrichment to save tokens. Any suggestions
                        already generated will be kept.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => setShowCancelEnrichDialog(false)}
                    >
                        Stay
                    </Button>
                    <Button
                        onClick={handleCancelEnrichAndLeave}
                        color="warning"
                        variant="contained"
                    >
                        Cancel Enrichment &amp; Leave
                    </Button>
                </DialogActions>
            </Dialog>
        </FullScreenLayout>
    );
}
