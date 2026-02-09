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

import { useState, useMemo, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
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
} from '../hooks/useContentAnalysis';
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

    const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
    const [newEntityName, setNewEntityName] = useState('');
    const [newEntityType, setNewEntityType] = useState<EntityType>('npc');
    const [showNewEntityForm, setShowNewEntityForm] = useState(false);
    const [showDoneDialog, setShowDoneDialog] = useState(false);

    const isLoading = jobLoading || itemsLoading;

    /**
     * Split items into pending and resolved arrays.
     */
    const pendingItems = useMemo(() => {
        if (!items) return [];
        return items.filter((i) => i.resolution === 'pending');
    }, [items]);

    const resolvedItems = useMemo(() => {
        if (!items) return [];
        return items.filter((i) => i.resolution !== 'pending');
    }, [items]);

    /**
     * Group only pending items by detection type for the main sections.
     */
    const groupedPendingItems = useMemo(() => {
        const groups: Record<string, ContentAnalysisItem[]> = {};
        for (const group of DETECTION_GROUPS) {
            const groupItems = pendingItems.filter(
                (item) => item.detectionType === group.key
            );
            if (groupItems.length > 0) {
                groups[group.key] = groupItems;
            }
        }
        return groups;
    }, [pendingItems]);

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
