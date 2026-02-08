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

import { useState, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
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
} from '@mui/material';
import {
    Check,
    AddCircle,
    Close,
} from '@mui/icons-material';
import FullScreenLayout from '../layouts/FullScreenLayout';
import {
    useAnalysisJob,
    useAnalysisItems,
    useResolveItem,
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
 * Displays items grouped by detection type in a left panel and a detail
 * view with resolution actions in a right panel.
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

    const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
    const [newEntityName, setNewEntityName] = useState('');
    const [newEntityType, setNewEntityType] = useState<EntityType>('npc');
    const [showNewEntityForm, setShowNewEntityForm] = useState(false);

    const isLoading = jobLoading || itemsLoading;

    /**
     * Group items by detection type, with resolved items sorted to the
     * bottom of each group.
     */
    const groupedItems = useMemo(() => {
        if (!items) return {};
        const groups: Record<string, ContentAnalysisItem[]> = {};
        for (const group of DETECTION_GROUPS) {
            const groupItems = items
                .filter((item) => item.detectionType === group.key)
                .sort((a, b) => {
                    const aResolved = a.resolution !== 'pending' ? 1 : 0;
                    const bResolved = b.resolution !== 'pending' ? 1 : 0;
                    return aResolved - bResolved;
                });
            if (groupItems.length > 0) {
                groups[group.key] = groupItems;
            }
        }
        return groups;
    }, [items]);

    const selectedItem = useMemo(() => {
        if (!items || selectedItemId === null) return null;
        return items.find((item) => item.id === selectedItemId) ?? null;
    }, [items, selectedItemId]);

    const resolvedCount = items
        ? items.filter((item) => item.resolution !== 'pending').length
        : 0;
    const totalCount = items ? items.length : 0;

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
                        label={`${resolvedCount}/${totalCount} resolved`}
                        size="small"
                        color={
                            resolvedCount === totalCount
                                ? 'success'
                                : 'default'
                        }
                    />
                    <Button
                        variant="outlined"
                        size="small"
                        onClick={() =>
                            navigate(
                                `/campaigns/${campaignId}/overview`
                            )
                        }
                    >
                        Skip for now
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
                            const groupItems = groupedItems[group.key];
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
                                            sx={{ fontWeight: 600 }}
                                        >
                                            {group.label} ({groupItems.length})
                                        </Typography>
                                    </Box>
                                    <Divider />
                                    {groupItems.map((item) => {
                                        const isResolved =
                                            item.resolution !== 'pending';
                                        return (
                                            <ListItemButton
                                                key={item.id}
                                                selected={
                                                    selectedItemId === item.id
                                                }
                                                onClick={() =>
                                                    handleSelectItem(item)
                                                }
                                                sx={{
                                                    opacity: isResolved
                                                        ? 0.5
                                                        : 1,
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
                                                            {isResolved && (
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
                                                {!isResolved && (
                                                    <Box
                                                        sx={{
                                                            display: 'flex',
                                                            gap: 0.5,
                                                            ml: 1,
                                                        }}
                                                    >
                                                        <IconButton
                                                            size="small"
                                                            color="success"
                                                            title="Accept"
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
                                                        <IconButton
                                                            size="small"
                                                            color="primary"
                                                            title="New entity"
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
                                                        <IconButton
                                                            size="small"
                                                            color="error"
                                                            title="Dismiss"
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
                                                    </Box>
                                                )}
                                            </ListItemButton>
                                        );
                                    })}
                                </Box>
                            );
                        })}
                    </List>
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
        </FullScreenLayout>
    );
}
