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
    Box,
    Typography,
    Chip,
    Button,
    IconButton,
    List,
    ListItemButton,
    ListItemText,
    Divider,
    CircularProgress,
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
import {
    Check,
    AddCircle,
    Close,
    DoneAll,
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
} from '../hooks/useContentAnalysis';
import { useUserSettings } from '../hooks/useUserSettings';
import {
    entityTypeColors,
    formatEntityType,
} from '../components/EntitySelector/entityConstants';
import type { EntityType } from '../types';
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

    // Determine if enrichment is in progress
    const isEnriching = job?.status === 'enriching';

    // Enable enrichment stream polling when enriching
    useEnrichmentStream(numericCampaignId, numericJobId, isEnriching);

    // Check if LLM is configured
    const hasLLMConfigured = !!userSettings?.contentGenService && !!userSettings?.contentGenApiKey;

    const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
    const [newEntityName, setNewEntityName] = useState('');
    const [newEntityType, setNewEntityType] = useState<EntityType>('npc');
    const [showNewEntityForm, setShowNewEntityForm] = useState(false);
    const [showDoneDialog, setShowDoneDialog] = useState(false);

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

    const resolvedCount = resolvedItems.length;
    const totalCount = items ? items.length : 0;
    const allResolved = resolvedCount === totalCount && totalCount > 0;

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
        if (allResolved) {
            navigate(getReturnPath());
        } else {
            setShowDoneDialog(true);
        }
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
        resolveItem.mutate({
            itemId,
            req: {
                resolution,
                ...(entityName && { entityName }),
                ...(entityType && { entityType }),
            },
        });
        setShowNewEntityForm(false);
    };

    /**
     * Handle batch-resolving all pending items of a detection type.
     */
    const handleBatchResolve = (detectionType: string) => {
        batchResolve.mutate({
            jobId: numericJobId,
            detectionType,
            resolution: 'accepted',
        });
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
        setNewEntityName(item.matchedText);
        setNewEntityType('npc');
        setShowNewEntityForm(false);
    };

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
            if (resolution === 'dismissed') {
                resolveItem.mutate({
                    itemId,
                    req: { resolution: 'dismissed' },
                });
                return;
            }

            try {
                const suggestion = item.suggestedContent;
                if (!suggestion) {
                    resolveItem.mutate({
                        itemId,
                        req: { resolution: 'accepted' },
                    });
                    return;
                }

                resolveItem.mutate({
                    itemId,
                    req: { resolution: 'accepted' },
                });
            } catch (error) {
                console.error(
                    'Failed to apply enrichment suggestion:',
                    error,
                );
            }
        },
        [resolveItem],
    );

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
                        justifyContent: 'center',
                        alignItems: 'center',
                        height: '100%',
                    }}
                >
                    <Typography color="text.secondary">
                        No items to review
                    </Typography>
                </Box>
            </FullScreenLayout>
        );
    }

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
            actions={
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
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

                    {/* Enrichment section divider */}
                    {(enrichmentItems.length > 0 || (job && job.resolvedItems === job.totalItems && job.totalItems > 0)) && (
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
                                }}
                            >
                                Entity Enrichment
                            </Typography>
                        </Box>
                    )}

                    {/* Enrichment loading indicator */}
                    {isEnriching && (
                        <Box sx={{ px: 2, py: 1.5 }}>
                            <Typography
                                variant="body2"
                                color="text.secondary"
                                gutterBottom
                            >
                                Analyzing{' '}
                                {job?.enrichmentTotal ?? 0} entities...
                            </Typography>
                            <LinearProgress />
                        </Box>
                    )}

                    {/* No LLM configured message */}
                    {!hasLLMConfigured &&
                        job &&
                        job.resolvedItems === job.totalItems &&
                        job.totalItems > 0 &&
                        enrichmentItems.length === 0 &&
                        !isEnriching && (
                        <Box sx={{ px: 2, py: 1.5 }}>
                            <Alert
                                severity="info"
                                sx={{ fontSize: '0.875rem' }}
                            >
                                Configure an LLM service in{' '}
                                <Link
                                    to="/settings"
                                    style={{ color: 'inherit' }}
                                >
                                    Account Settings
                                </Link>{' '}
                                to enable entity enrichment.
                            </Alert>
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
                                {groupItems.map((item) => (
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
                                                        {item.matchedText}
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
                                                                    item.entityType
                                                                ] ?? 'default'
                                                            }
                                                            sx={{
                                                                height: 20,
                                                            }}
                                                        />
                                                    )}
                                                </Box>
                                            }
                                            secondary={
                                                item.detectionType ===
                                                'description_update'
                                                    ? 'Description update suggestion'
                                                    : item.detectionType ===
                                                      'log_entry'
                                                      ? 'New log entry suggestion'
                                                      : item.detectionType ===
                                                        'relationship_suggestion'
                                                        ? 'Relationship suggestion'
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
                                            <Tooltip title="Accept">
                                                <IconButton
                                                    size="small"
                                                    color="success"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'accepted',
                                                            item,
                                                        );
                                                    }}
                                                >
                                                    <Check fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip title="Dismiss">
                                                <IconButton
                                                    size="small"
                                                    color="error"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        handleEnrichmentResolve(
                                                            item.id,
                                                            'dismissed',
                                                            item,
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

                    {/* Accepted Changes - collapsible section */}
                    {resolvedItems.length > 0 && (
                        <Accordion
                            defaultExpanded={false}
                            disableGutters
                            elevation={0}
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
                    {selectedItem ? (
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
                                                {!!suggestion.currentDescription && (
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
                                                                    'grey.50',
                                                            }}
                                                        >
                                                            <Typography variant="body2">
                                                                {String(
                                                                    suggestion.currentDescription,
                                                                )}
                                                            </Typography>
                                                        </Paper>
                                                    </Box>
                                                )}
                                                <Box>
                                                    <Typography
                                                        variant="subtitle2"
                                                        gutterBottom
                                                    >
                                                        Suggested
                                                        Description
                                                    </Typography>
                                                    <Paper
                                                        variant="outlined"
                                                        sx={{
                                                            p: 2,
                                                            bgcolor:
                                                                'success.50',
                                                            borderColor:
                                                                'success.light',
                                                        }}
                                                    >
                                                        <Typography variant="body2">
                                                            {String(
                                                                suggestion.suggestedDescription ??
                                                                    '',
                                                            )}
                                                        </Typography>
                                                    </Paper>
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
                                                                suggestion.logText ??
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
                                                        <Typography
                                                            variant="body2"
                                                            sx={{
                                                                fontWeight: 600,
                                                            }}
                                                        >
                                                            {String(
                                                                suggestion.relationshipType ??
                                                                    'related to',
                                                            )}
                                                        </Typography>
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
                                {selectedItem.resolution === 'pending' && (
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
        </FullScreenLayout>
    );
}
