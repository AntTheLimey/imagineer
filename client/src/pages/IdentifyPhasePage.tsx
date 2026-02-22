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
 * IdentifyPhasePage -- the first phase of the analysis wizard.
 *
 * Two-column layout:
 *  - Left panel (~40%): items grouped by detection type with counts.
 *  - Right panel (~60%): detail view for the selected item with context,
 *    entity info, and resolution actions.
 */

import { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import {
    Box,
    Grid,
    List,
    ListItemButton,
    ListItemText,
    Typography,
    Chip,
    IconButton,
    Button,
    Stack,
    Divider,
    Alert,
    TextField,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    Paper,
    Tooltip,
} from '@mui/material';
import { Check, Close, AddCircle, DoneAll } from '@mui/icons-material';
import { useWizardContext } from '../contexts/AnalysisWizardContext';
import { useResolveItem, useBatchResolve } from '../hooks/useContentAnalysis';
import {
    entityTypeColors,
    formatEntityType,
} from '../components/EntitySelector/entityConstants';
import type { ContentAnalysisItem } from '../api/contentAnalysis';
import type { EntityType } from '../types';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Display metadata for each detection group in the Identify phase. */
const IDENTIFY_GROUPS = [
    {
        key: 'wiki_link_resolved',
        label: 'Wiki Links (Resolved)',
        color: '#4caf50',
    },
    {
        key: 'wiki_link_unresolved',
        label: 'Wiki Links (Unresolved)',
        color: '#ff9800',
    },
    {
        key: 'untagged_mention',
        label: 'Untagged Mentions',
        color: '#2196f3',
    },
    {
        key: 'potential_alias',
        label: 'Potential Aliases',
        color: '#9c27b0',
    },
    {
        key: 'misspelling',
        label: 'Misspellings',
        color: '#ffc107',
    },
] as const;

/** Entity types derived from the shared colour map. */
const ENTITY_TYPES = Object.keys(entityTypeColors) as EntityType[];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Highlight matchedText within contextSnippet by wrapping the first
 * occurrence in a <strong> tag.
 */
function highlightContext(
    snippet: string,
    matchedText: string,
): React.ReactNode {
    if (!matchedText) {
        return snippet;
    }
    const idx = snippet.toLowerCase().indexOf(matchedText.toLowerCase());
    if (idx === -1) {
        return snippet;
    }
    const before = snippet.slice(0, idx);
    const match = snippet.slice(idx, idx + matchedText.length);
    const after = snippet.slice(idx + matchedText.length);
    return (
        <>
            {before}
            <strong>{match}</strong>
            {after}
        </>
    );
}

/**
 * Truncate a string to maxLen characters, adding an ellipsis if
 * truncated.
 */
function truncate(text: string, maxLen: number): string {
    if (text.length <= maxLen) return text;
    return text.slice(0, maxLen) + '\u2026';
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export default function IdentifyPhasePage() {
    const { campaignId } = useParams<{ campaignId: string }>();
    const { phaseItems, pendingCount, job } = useWizardContext();
    const resolveItem = useResolveItem(Number(campaignId));
    const batchResolve = useBatchResolve(Number(campaignId));

    // -- Local state -------------------------------------------------------

    const [selectedItemId, setSelectedItemId] = useState<
        number | null
    >(null);
    const [showNewEntityForm, setShowNewEntityForm] = useState(false);
    const [newEntityName, setNewEntityName] = useState('');
    const [newEntityType, setNewEntityType] = useState<EntityType>(
        'npc',
    );

    // -- Derived data ------------------------------------------------------

    /** Group items by detection type, preserving the order of IDENTIFY_GROUPS. */
    const groupedItems = useMemo(() => {
        const grouped = phaseItems.reduce<
            Record<string, ContentAnalysisItem[]>
        >((acc, item) => {
            const key = item.detectionType;
            if (!acc[key]) acc[key] = [];
            acc[key].push(item);
            return acc;
        }, {});

        return IDENTIFY_GROUPS.filter(
            (g) => grouped[g.key] && grouped[g.key].length > 0,
        ).map((g) => ({
            ...g,
            items: grouped[g.key],
        }));
    }, [phaseItems]);

    const selectedItem = useMemo(
        () =>
            selectedItemId !== null
                ? phaseItems.find((i) => i.id === selectedItemId) ??
                  null
                : null,
        [phaseItems, selectedItemId],
    );

    // -- Handlers ----------------------------------------------------------

    const handleResolve = (
        itemId: number,
        resolution: 'accepted' | 'new_entity' | 'dismissed',
        entityName?: string,
        entityType?: EntityType,
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
                onSuccess: () => {
                    // NOTE: phaseItems may be stale here since the
                    // query invalidation triggered by the mutation
                    // hasn't updated the context yet. The
                    // i.id !== itemId guard ensures we skip the
                    // just-resolved item regardless. For rapid
                    // sequential resolutions, there is a small
                    // window where a previously resolved item could
                    // be selected, but the next render cycle
                    // corrects this.
                    const pendingItems = phaseItems.filter(
                        (i) =>
                            i.resolution === 'pending' &&
                            i.id !== itemId,
                    );
                    if (pendingItems.length > 0) {
                        setSelectedItemId(pendingItems[0].id);
                    } else {
                        setSelectedItemId(null);
                    }
                    setShowNewEntityForm(false);
                },
                onError: (err: Error) => {
                    console.error(
                        'Failed to resolve item:',
                        err.message,
                    );
                },
            },
        );
    };

    const handleSelectItem = (item: ContentAnalysisItem) => {
        setSelectedItemId(item.id);
        setShowNewEntityForm(false);
        setNewEntityName(item.matchedText);
        setNewEntityType('npc');
    };

    const handleToggleNewEntity = () => {
        setShowNewEntityForm((prev) => !prev);
        if (selectedItem) {
            setNewEntityName(selectedItem.matchedText);
        }
    };

    const handleBatchResolve = (detectionType: string) => {
        if (!job) return;
        batchResolve.mutate({
            jobId: job.id,
            detectionType,
            resolution: 'accepted',
        }, {
            onSuccess: () => {
                // Clear selection since the items are now resolved
                setSelectedItemId(null);
            },
            onError: (err: Error) => {
                console.error('Batch resolve failed:', err.message);
            },
        });
    };

    // -- Render ------------------------------------------------------------

    return (
        <Grid container spacing={2} sx={{ height: '100%' }}>
            {/* Left panel -- item list */}
            <Grid item xs={12} md={5}>
                <Paper
                    variant="outlined"
                    sx={{ height: '100%', overflow: 'auto' }}
                >
                    <Box sx={{ p: 2, pb: 1 }}>
                        <Typography variant="subtitle2" color="text.secondary">
                            {phaseItems.length} items ({pendingCount}{' '}
                            pending)
                        </Typography>
                    </Box>
                    <Divider />
                    <List disablePadding>
                        {groupedItems.map((group) => {
                            const pendingInGroup = group.items.filter(
                                (i) => i.resolution === 'pending',
                            ).length;
                            const allHaveEntity = group.items
                                .filter((i) => i.resolution === 'pending')
                                .every((i) => !!i.entityId);

                            return (
                            <Box key={group.key}>
                                {/* Group header */}
                                <Box
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 1,
                                        px: 2,
                                        py: 1,
                                        bgcolor: 'action.hover',
                                    }}
                                >
                                    <Box
                                        sx={{
                                            width: 10,
                                            height: 10,
                                            borderRadius: '50%',
                                            bgcolor: group.color,
                                            flexShrink: 0,
                                        }}
                                    />
                                    <Typography
                                        variant="subtitle2"
                                        sx={{ flexGrow: 1 }}
                                    >
                                        {group.label}
                                    </Typography>
                                    {allHaveEntity && pendingInGroup > 0 && (
                                        <Tooltip title={`Accept all ${pendingInGroup} items in this group`}>
                                            <span>
                                                <Button
                                                    size="small"
                                                    variant="outlined"
                                                    startIcon={<DoneAll fontSize="small" />}
                                                    onClick={() => handleBatchResolve(group.key)}
                                                    disabled={batchResolve.isPending}
                                                    sx={{ textTransform: 'none', py: 0, minHeight: 28 }}
                                                >
                                                    Accept All
                                                </Button>
                                            </span>
                                        </Tooltip>
                                    )}
                                    <Chip
                                        label={group.items.length}
                                        size="small"
                                    />
                                </Box>
                                {/* Items in this group */}
                                {group.items.map((item) => (
                                    <ListItemButton
                                        key={item.id}
                                        selected={
                                            selectedItemId === item.id
                                        }
                                        onClick={() =>
                                            handleSelectItem(item)
                                        }
                                        sx={{ pl: 4 }}
                                    >
                                        <ListItemText
                                            primary={
                                                <Stack
                                                    direction="row"
                                                    spacing={1}
                                                    alignItems="center"
                                                >
                                                    <Typography
                                                        variant="body2"
                                                        fontWeight="bold"
                                                        noWrap
                                                    >
                                                        {
                                                            item.matchedText
                                                        }
                                                    </Typography>
                                                    {item.entityName && (
                                                        <Chip
                                                            label={
                                                                item.entityName
                                                            }
                                                            size="small"
                                                            color={
                                                                item.entityType
                                                                    ? entityTypeColors[
                                                                          item
                                                                              .entityType
                                                                      ]
                                                                    : 'default'
                                                            }
                                                            variant="outlined"
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
                                                                        100,
                                                                )}
                                                                %
                                                            </Typography>
                                                        )}
                                                </Stack>
                                            }
                                            secondary={
                                                item.contextSnippet
                                                    ? truncate(
                                                          item.contextSnippet,
                                                          60,
                                                      )
                                                    : undefined
                                            }
                                        />
                                        {/* Quick action buttons */}
                                        {item.resolution ===
                                            'pending' && (
                                            <Stack
                                                direction="row"
                                                spacing={0.5}
                                                sx={{ ml: 1 }}
                                            >
                                                <Tooltip title="Accept">
                                                    <IconButton
                                                        size="small"
                                                        color="success"
                                                        disabled={resolveItem.isPending}
                                                        onClick={(
                                                            e,
                                                        ) => {
                                                            e.stopPropagation();
                                                            handleResolve(
                                                                item.id,
                                                                'accepted',
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
                                                        disabled={resolveItem.isPending}
                                                        onClick={(
                                                            e,
                                                        ) => {
                                                            e.stopPropagation();
                                                            handleResolve(
                                                                item.id,
                                                                'dismissed',
                                                            );
                                                        }}
                                                    >
                                                        <Close fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                            </Stack>
                                        )}
                                    </ListItemButton>
                                ))}
                            </Box>
                        );
                        })}
                        {groupedItems.length === 0 && (
                            <Box sx={{ p: 3, textAlign: 'center' }}>
                                <Typography
                                    variant="body2"
                                    color="text.secondary"
                                >
                                    No items in this phase.
                                </Typography>
                            </Box>
                        )}
                    </List>
                </Paper>
            </Grid>

            {/* Right panel -- detail view */}
            <Grid item xs={12} md={7}>
                <Paper
                    variant="outlined"
                    sx={{ height: '100%', overflow: 'auto', p: 3 }}
                >
                    {selectedItem ? (
                        <DetailPanel
                            item={selectedItem}
                            isResolving={resolveItem.isPending}
                            showNewEntityForm={showNewEntityForm}
                            newEntityName={newEntityName}
                            newEntityType={newEntityType}
                            onNewEntityNameChange={setNewEntityName}
                            onNewEntityTypeChange={setNewEntityType}
                            onToggleNewEntity={handleToggleNewEntity}
                            onResolve={handleResolve}
                        />
                    ) : (
                        <Box
                            sx={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                height: '100%',
                                minHeight: 200,
                            }}
                        >
                            <Typography
                                variant="body1"
                                color="text.secondary"
                            >
                                Select an item to view details
                            </Typography>
                        </Box>
                    )}
                </Paper>
            </Grid>
        </Grid>
    );
}

// ---------------------------------------------------------------------------
// Detail panel sub-component
// ---------------------------------------------------------------------------

interface DetailPanelProps {
    item: ContentAnalysisItem;
    isResolving: boolean;
    showNewEntityForm: boolean;
    newEntityName: string;
    newEntityType: EntityType;
    onNewEntityNameChange: (name: string) => void;
    onNewEntityTypeChange: (type: EntityType) => void;
    onToggleNewEntity: () => void;
    onResolve: (
        itemId: number,
        resolution: 'accepted' | 'new_entity' | 'dismissed',
        entityName?: string,
        entityType?: EntityType,
    ) => void;
}

function DetailPanel({
    item,
    isResolving,
    showNewEntityForm,
    newEntityName,
    newEntityType,
    onNewEntityNameChange,
    onNewEntityTypeChange,
    onToggleNewEntity,
    onResolve,
}: DetailPanelProps) {
    const isPending = item.resolution === 'pending';

    return (
        <Stack spacing={3}>
            {/* Context section */}
            {item.contextSnippet && (
                <Box>
                    <Typography
                        variant="subtitle2"
                        gutterBottom
                    >
                        Context
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{
                            p: 2,
                            bgcolor: 'grey.50',
                            fontFamily: 'serif',
                            fontSize: '0.95rem',
                            lineHeight: 1.7,
                        }}
                    >
                        <Typography variant="body2" component="span">
                            {highlightContext(
                                item.contextSnippet,
                                item.matchedText,
                            )}
                        </Typography>
                    </Paper>
                </Box>
            )}

            {/* Matched entity section */}
            {item.entityId && item.entityName && (
                <Box>
                    <Typography
                        variant="subtitle2"
                        gutterBottom
                    >
                        Matched Entity
                    </Typography>
                    <Stack
                        direction="row"
                        spacing={1}
                        alignItems="center"
                    >
                        <Typography variant="body1">
                            {item.entityName}
                        </Typography>
                        {item.entityType && (
                            <Chip
                                label={formatEntityType(
                                    item.entityType,
                                )}
                                size="small"
                                color={
                                    entityTypeColors[item.entityType]
                                }
                            />
                        )}
                        {item.similarity !== undefined && (
                            <Chip
                                label={`${Math.round(item.similarity * 100)}% match`}
                                size="small"
                                variant="outlined"
                            />
                        )}
                    </Stack>
                </Box>
            )}

            {/* Resolution status */}
            {!isPending && (
                <Alert severity="info" variant="outlined">
                    This item has been resolved:{' '}
                    <strong>{item.resolution}</strong>
                </Alert>
            )}

            {/* Action buttons */}
            {isPending && (
                <Box>
                    <Typography
                        variant="subtitle2"
                        gutterBottom
                    >
                        Actions
                    </Typography>
                    <Stack direction="row" spacing={1}>
                        {item.entityId && item.entityName && (
                            <Button
                                variant="contained"
                                color="success"
                                disabled={isResolving}
                                startIcon={<Check />}
                                onClick={() =>
                                    onResolve(item.id, 'accepted')
                                }
                            >
                                Link to {item.entityName}
                            </Button>
                        )}
                        <Button
                            variant="outlined"
                            startIcon={<AddCircle />}
                            onClick={onToggleNewEntity}
                        >
                            New Entity
                        </Button>
                        <Button
                            variant="outlined"
                            color="error"
                            disabled={isResolving}
                            startIcon={<Close />}
                            onClick={() =>
                                onResolve(item.id, 'dismissed')
                            }
                        >
                            Dismiss
                        </Button>
                    </Stack>
                </Box>
            )}

            {/* Create entity form */}
            {isPending && showNewEntityForm && (
                <Paper variant="outlined" sx={{ p: 2 }}>
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
                                onNewEntityNameChange(e.target.value)
                            }
                            size="small"
                            fullWidth
                        />
                        <FormControl size="small" fullWidth>
                            <InputLabel id="new-entity-type-label">
                                Entity Type
                            </InputLabel>
                            <Select
                                labelId="new-entity-type-label"
                                value={newEntityType}
                                label="Entity Type"
                                onChange={(e) =>
                                    onNewEntityTypeChange(
                                        e.target.value as EntityType,
                                    )
                                }
                            >
                                {ENTITY_TYPES.map((t) => (
                                    <MenuItem key={t} value={t}>
                                        {formatEntityType(t)}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                        <Typography
                            variant="caption"
                            color="text.secondary"
                        >
                            A full description will be generated
                            during the Enrich phase.
                        </Typography>
                        <Button
                            variant="contained"
                            disabled={
                                !newEntityName.trim() ||
                                isResolving
                            }
                            onClick={() =>
                                onResolve(
                                    item.id,
                                    'new_entity',
                                    newEntityName.trim(),
                                    newEntityType,
                                )
                            }
                        >
                            Create & Link
                        </Button>
                    </Stack>
                </Paper>
            )}
        </Stack>
    );
}
