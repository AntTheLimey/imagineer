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
 * EnrichPhasePage -- the third phase of the analysis wizard.
 *
 * Two-pass workflow:
 *  1. Trigger enrichment via the "Start Enrichment" button and monitor
 *     progress while the backend generates items (description updates,
 *     log entries, relationship suggestions, etc.).
 *  2. Review and resolve the enrichment items in a two-column layout
 *     matching the Identify and Revise phase patterns.
 *
 * Layout:
 *  - Top section: Enrichment trigger controls and progress indicator.
 *  - Left panel (~40%): Items grouped by detection type with counts.
 *  - Right panel (~60%): Detail view for the selected item with context,
 *    suggested content, and resolution actions.
 */

import { useState, useMemo } from 'react';

// Stable empty object to avoid allocating a new reference on each render.
const EMPTY_SUGGESTED: Record<string, unknown> = {};
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
    Paper,
    Tooltip,
    CircularProgress,
    LinearProgress,
} from '@mui/material';
import { Check, Close, PlayArrow, Cancel } from '@mui/icons-material';
import { useWizardContext } from '../contexts/AnalysisWizardContext';
import {
    useResolveItem,
    useTriggerEnrichment,
    useEnrichmentStream,
    useCancelEnrichment,
} from '../hooks/useContentAnalysis';
import {
    entityTypeColors,
    formatEntityType,
} from '../components/EntitySelector/entityConstants';
import type { ContentAnalysisItem } from '../api/contentAnalysis';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Display metadata for each detection group in the Enrich phase. */
const ENRICH_GROUPS = [
    { key: 'description_update', label: 'Description Updates', color: '#1565c0' },
    { key: 'log_entry', label: 'Log Entries', color: '#2e7d32' },
    { key: 'relationship_suggestion', label: 'Relationship Suggestions', color: '#6a1b9a' },
    { key: 'new_entity_suggestion', label: 'New Entity Suggestions', color: '#e65100' },
    { key: 'graph_warning', label: 'Graph Warnings', color: '#c62828' },
    { key: 'redundant_edge', label: 'Redundant Edges', color: '#ff6f00' },
    { key: 'invalid_type_pair', label: 'Invalid Type Pairs', color: '#ad1457' },
    { key: 'orphan_warning', label: 'Orphan Warnings', color: '#00838f' },
] as const;

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

export default function EnrichPhasePage() {
    const { campaignId } = useParams<{ campaignId: string }>();
    const { phaseItems, pendingCount, job } = useWizardContext();
    const resolveItem = useResolveItem(Number(campaignId));
    const triggerEnrichment = useTriggerEnrichment(Number(campaignId));
    const cancelEnrichment = useCancelEnrichment(Number(campaignId));

    // -- Local state -------------------------------------------------------

    const [selectedItemId, setSelectedItemId] = useState<
        number | null
    >(null);

    // -- Derived data ------------------------------------------------------

    /** Whether the job is currently enriching. */
    const isEnriching = job?.status === 'enriching';

    /** Whether the job has completed enrichment. */
    const isEnriched = job?.status === 'enriched';

    // Enable polling when enrichment is in progress.  Derived entirely
    // from server state so polling resumes correctly after navigation.
    useEnrichmentStream(
        Number(campaignId),
        job?.id ?? 0,
        isEnriching,
    );

    /** Group items by detection type, preserving the order of ENRICH_GROUPS. */
    const groupedItems = useMemo(() => {
        const grouped = phaseItems.reduce<
            Record<string, ContentAnalysisItem[]>
        >((acc, item) => {
            const key = item.detectionType;
            if (!acc[key]) acc[key] = [];
            acc[key].push(item);
            return acc;
        }, {});

        return ENRICH_GROUPS.filter(
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

    const handleStartEnrichment = () => {
        if (!job) return;
        triggerEnrichment.mutate(job.id, {
            onError: (err: Error) => {
                console.error(
                    'Failed to trigger enrichment:',
                    err.message,
                );
            },
        });
    };

    const handleCancelEnrichment = () => {
        if (!job) return;
        cancelEnrichment.mutate(job.id, {
            onError: (err: Error) => {
                console.error(
                    'Failed to cancel enrichment:',
                    err.message,
                );
            },
        });
    };

    const handleResolve = (
        itemId: number,
        resolution: 'accepted' | 'dismissed',
    ) => {
        resolveItem.mutate(
            {
                itemId,
                req: { resolution },
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
    };

    // -- Render ------------------------------------------------------------

    return (
        <Stack spacing={2} sx={{ height: '100%' }}>
            {/* Enrichment trigger section */}
            <Paper variant="outlined" sx={{ p: 2 }}>
                <Stack spacing={2}>
                    <Stack
                        direction="row"
                        alignItems="center"
                        spacing={1}
                    >
                        <Typography variant="h6">
                            Enrichment
                        </Typography>
                        {isEnriched && (
                            <Chip
                                label="Complete"
                                size="small"
                                color="success"
                                variant="outlined"
                            />
                        )}
                    </Stack>

                    {/* Start / Cancel buttons */}
                    <Box>
                        {!isEnriching && !isEnriched && (
                            <Button
                                variant="contained"
                                startIcon={
                                    triggerEnrichment.isPending ? (
                                        <CircularProgress
                                            size={16}
                                            color="inherit"
                                        />
                                    ) : (
                                        <PlayArrow />
                                    )
                                }
                                disabled={triggerEnrichment.isPending}
                                onClick={handleStartEnrichment}
                            >
                                Start Enrichment
                            </Button>
                        )}
                        {isEnriching && (
                            <Button
                                variant="outlined"
                                color="warning"
                                startIcon={<Cancel />}
                                disabled={cancelEnrichment.isPending}
                                onClick={handleCancelEnrichment}
                            >
                                Cancel Enrichment
                            </Button>
                        )}
                    </Box>

                    {/* Progress indicator */}
                    {isEnriching && (
                        <Box>
                            <Stack
                                direction="row"
                                spacing={1}
                                alignItems="center"
                                sx={{ mb: 1 }}
                            >
                                <CircularProgress size={16} />
                                <Typography
                                    variant="body2"
                                    color="text.secondary"
                                >
                                    Enrichment in progress...
                                </Typography>
                            </Stack>
                            <LinearProgress />
                        </Box>
                    )}

                    {/* Success message */}
                    {triggerEnrichment.data?.message && !isEnriching && (
                        <Alert severity="info" variant="outlined">
                            {triggerEnrichment.data.message}
                        </Alert>
                    )}
                </Stack>
            </Paper>

            {/* Two-column layout for items and detail */}
            <Grid
                container
                spacing={2}
                sx={{ flexGrow: 1, minHeight: 0 }}
            >
                {/* Left panel -- item list */}
                <Grid item xs={12} md={5}>
                    <Paper
                        variant="outlined"
                        sx={{ height: '100%', overflow: 'auto' }}
                    >
                        <Box sx={{ p: 2, pb: 1 }}>
                            <Typography
                                variant="subtitle2"
                                color="text.secondary"
                            >
                                {phaseItems.length} items (
                                {pendingCount} pending)
                            </Typography>
                        </Box>
                        <Divider />
                        <List disablePadding>
                            {groupedItems.map((group) => (
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
                                                selectedItemId ===
                                                item.id
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
                                                                              item.entityType
                                                                          ]
                                                                        : 'default'
                                                                }
                                                                variant="outlined"
                                                            />
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
                                                            disabled={
                                                                resolveItem.isPending
                                                            }
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
                                                            disabled={
                                                                resolveItem.isPending
                                                            }
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
                            ))}
                            {groupedItems.length === 0 && (
                                <Box
                                    sx={{
                                        p: 3,
                                        textAlign: 'center',
                                    }}
                                >
                                    <Typography
                                        variant="body2"
                                        color="text.secondary"
                                    >
                                        {isEnriching
                                            ? 'Waiting for enrichment items...'
                                            : 'No items in this phase.'}
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
                        sx={{
                            height: '100%',
                            overflow: 'auto',
                            p: 3,
                        }}
                    >
                        {selectedItem ? (
                            <DetailPanel
                                item={selectedItem}
                                isResolving={resolveItem.isPending}
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
        </Stack>
    );
}

// ---------------------------------------------------------------------------
// Detail panel sub-component
// ---------------------------------------------------------------------------

interface DetailPanelProps {
    item: ContentAnalysisItem;
    isResolving: boolean;
    onResolve: (
        itemId: number,
        resolution: 'accepted' | 'dismissed',
    ) => void;
}

function DetailPanel({ item, isResolving, onResolve }: DetailPanelProps) {
    const isPending = item.resolution === 'pending';

    /** Look up group metadata for the detection type. */
    const groupInfo = useMemo(
        () => ENRICH_GROUPS.find((g) => g.key === item.detectionType),
        [item.detectionType],
    );

    /** Parse suggestedContent for type-specific rendering. */
    const suggested = item.suggestedContent ?? EMPTY_SUGGESTED;

    return (
        <Stack spacing={3}>
            {/* Header */}
            <Box>
                <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                >
                    <Typography variant="h6">
                        {item.matchedText}
                    </Typography>
                    {item.entityType && (
                        <Chip
                            label={formatEntityType(item.entityType)}
                            size="small"
                            color={entityTypeColors[item.entityType]}
                        />
                    )}
                </Stack>
                {groupInfo && (
                    <Typography
                        variant="caption"
                        color="text.secondary"
                    >
                        {groupInfo.label}
                    </Typography>
                )}
            </Box>

            {/* Context section */}
            {item.contextSnippet && (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
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

            {/* Type-specific suggested content */}
            {renderSuggestedContent(item.detectionType, suggested)}

            {/* Entity info */}
            {item.entityName && (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Entity
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
                    <Typography variant="subtitle2" gutterBottom>
                        Actions
                    </Typography>
                    <Stack direction="row" spacing={1}>
                        <Button
                            variant="contained"
                            color="success"
                            disabled={isResolving}
                            startIcon={<Check />}
                            onClick={() =>
                                onResolve(item.id, 'accepted')
                            }
                        >
                            Accept
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
        </Stack>
    );
}

// ---------------------------------------------------------------------------
// Suggested content renderers
// ---------------------------------------------------------------------------

/**
 * Render type-specific suggested content based on the detection type.
 */
function renderSuggestedContent(
    detectionType: string,
    suggested: Record<string, unknown>,
): React.ReactNode {
    if (!suggested || Object.keys(suggested).length === 0) {
        return null;
    }

    switch (detectionType) {
        case 'description_update':
            return (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Description Update
                    </Typography>
                    {'currentDescription' in suggested && (
                        <Paper
                            variant="outlined"
                            sx={{ p: 2, mb: 1, bgcolor: 'grey.50' }}
                        >
                            <Typography
                                variant="caption"
                                color="text.secondary"
                                gutterBottom
                                component="div"
                            >
                                Current
                            </Typography>
                            <Typography
                                variant="body2"
                                sx={{ whiteSpace: 'pre-wrap' }}
                            >
                                {String(suggested.currentDescription)}
                            </Typography>
                        </Paper>
                    )}
                    {'suggestedDescription' in suggested && (
                        <Paper
                            variant="outlined"
                            sx={{ p: 2, bgcolor: 'success.50' }}
                        >
                            <Typography
                                variant="caption"
                                color="text.secondary"
                                gutterBottom
                                component="div"
                            >
                                Suggested
                            </Typography>
                            <Typography
                                variant="body2"
                                sx={{ whiteSpace: 'pre-wrap' }}
                            >
                                {String(
                                    suggested.suggestedDescription,
                                )}
                            </Typography>
                        </Paper>
                    )}
                </Box>
            );

        case 'relationship_suggestion':
            return (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Relationship Suggestion
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{ p: 2, bgcolor: 'grey.50' }}
                    >
                        {'relationshipType' in suggested && (
                            <Typography variant="body2">
                                <strong>Type:</strong>{' '}
                                {String(suggested.relationshipType)}
                            </Typography>
                        )}
                        {'targetEntity' in suggested && (
                            <Typography variant="body2">
                                <strong>Target:</strong>{' '}
                                {String(suggested.targetEntity)}
                            </Typography>
                        )}
                        {'description' in suggested && (
                            <Typography
                                variant="body2"
                                sx={{ mt: 1 }}
                            >
                                {String(suggested.description)}
                            </Typography>
                        )}
                    </Paper>
                </Box>
            );

        case 'log_entry':
            return (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Log Entry
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{ p: 2, bgcolor: 'grey.50' }}
                    >
                        <Typography
                            variant="body2"
                            sx={{ whiteSpace: 'pre-wrap' }}
                        >
                            {'content' in suggested
                                ? String(suggested.content)
                                : JSON.stringify(suggested, null, 2)}
                        </Typography>
                    </Paper>
                </Box>
            );

        case 'new_entity_suggestion':
            return (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        New Entity Suggestion
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{ p: 2, bgcolor: 'grey.50' }}
                    >
                        {'suggestedName' in suggested && (
                            <Typography variant="body2">
                                <strong>Name:</strong>{' '}
                                {String(suggested.suggestedName)}
                            </Typography>
                        )}
                        {'suggestedType' in suggested && (
                            <Typography variant="body2">
                                <strong>Type:</strong>{' '}
                                {String(suggested.suggestedType)}
                            </Typography>
                        )}
                        {'description' in suggested && (
                            <Typography
                                variant="body2"
                                sx={{ mt: 1 }}
                            >
                                {String(suggested.description)}
                            </Typography>
                        )}
                    </Paper>
                </Box>
            );

        // Graph items: graph_warning, redundant_edge,
        // invalid_type_pair, orphan_warning
        default:
            return (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Details
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{ p: 2, bgcolor: 'grey.50' }}
                    >
                        <Typography
                            variant="body2"
                            sx={{ whiteSpace: 'pre-wrap' }}
                        >
                            {'description' in suggested
                                ? String(suggested.description)
                                : JSON.stringify(suggested, null, 2)}
                        </Typography>
                        {'recommendation' in suggested && (
                            <Typography
                                variant="body2"
                                sx={{ mt: 1 }}
                                color="text.secondary"
                            >
                                <strong>Recommendation:</strong>{' '}
                                {String(suggested.recommendation)}
                            </Typography>
                        )}
                    </Paper>
                </Box>
            );
    }
}
