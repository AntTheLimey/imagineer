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
 * RevisePhasePage -- the second phase of the analysis wizard.
 *
 * Layout:
 *  - Top section: Revision workflow (generate, diff view, edit, apply).
 *  - Left panel (~40%): Analysis findings grouped by detection type with
 *    severity indicators + new identification mentions after revision.
 *  - Right panel (~60%): Detail view for the selected finding or mention.
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
    Paper,
    Tooltip,
    TextField,
    CircularProgress,
} from '@mui/material';
import { Check, Close, Edit, PlayArrow } from '@mui/icons-material';
import { useWizardContext } from '../contexts/AnalysisWizardContext';
import {
    useResolveItem,
    useGenerateRevision,
    useApplyRevision,
} from '../hooks/useContentAnalysis';
import type { ContentAnalysisItem } from '../api/contentAnalysis';

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Display metadata for each detection group in the Revise phase. */
const REVISE_GROUPS = [
    {
        key: 'analysis_report',
        label: 'Analysis Reports',
        color: '#1565c0',
        severity: 'info' as const,
    },
    {
        key: 'content_suggestion',
        label: 'Content Suggestions',
        color: '#2e7d32',
        severity: 'low' as const,
    },
    {
        key: 'mechanics_warning',
        label: 'Mechanics Warnings',
        color: '#e65100',
        severity: 'medium' as const,
    },
    {
        key: 'investigation_gap',
        label: 'Investigation Gaps',
        color: '#6a1b9a',
        severity: 'medium' as const,
    },
    {
        key: 'pacing_note',
        label: 'Pacing Notes',
        color: '#00838f',
        severity: 'low' as const,
    },
    {
        key: 'canon_contradiction',
        label: 'Canon Contradictions',
        color: '#c62828',
        severity: 'high' as const,
    },
    {
        key: 'temporal_inconsistency',
        label: 'Temporal Inconsistencies',
        color: '#ff6f00',
        severity: 'high' as const,
    },
    {
        key: 'character_inconsistency',
        label: 'Character Inconsistencies',
        color: '#ad1457',
        severity: 'high' as const,
    },
] as const;

/** Display metadata for identification-phase groups (new mentions). */
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

/** Detection types from the identification phase (new mentions). */
const IDENTIFY_TYPES = IDENTIFY_GROUPS.map((g) => g.key) as string[];

/** Map severity levels to MUI Chip color props. */
const SEVERITY_CHIP_COLOR: Record<
    string,
    'error' | 'warning' | 'success' | 'info'
> = {
    high: 'error',
    medium: 'warning',
    low: 'success',
    info: 'info',
};

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

export default function RevisePhasePage() {
    const { campaignId } = useParams<{ campaignId: string }>();
    const { phaseItems, pendingCount, job, items } = useWizardContext();
    const resolveItem = useResolveItem(Number(campaignId));
    const generateRevision = useGenerateRevision(Number(campaignId));
    const applyRevision = useApplyRevision(Number(campaignId));

    // -- Local state -------------------------------------------------------

    const [selectedItemId, setSelectedItemId] = useState<
        number | null
    >(null);
    const [revisionContent, setRevisionContent] = useState<
        string | null
    >(null);
    const [isEditing, setIsEditing] = useState(false);
    const [revisionCount, setRevisionCount] = useState(0);

    // -- Derived data ------------------------------------------------------

    /** Analysis findings (this phase's items). */
    const analysisItems = phaseItems;

    /** New identification mentions from all items. */
    const newMentions = useMemo(
        () => items.filter((i) => IDENTIFY_TYPES.includes(i.detectionType)),
        [items],
    );

    /** Group analysis items by detection type. */
    const groupedAnalysisItems = useMemo(() => {
        const grouped = analysisItems.reduce<
            Record<string, ContentAnalysisItem[]>
        >((acc, item) => {
            const key = item.detectionType;
            if (!acc[key]) acc[key] = [];
            acc[key].push(item);
            return acc;
        }, {});

        return REVISE_GROUPS.filter(
            (g) => grouped[g.key] && grouped[g.key].length > 0,
        ).map((g) => ({
            ...g,
            items: grouped[g.key],
        }));
    }, [analysisItems]);

    /** Group new mention items by detection type. */
    const groupedNewMentions = useMemo(() => {
        const grouped = newMentions.reduce<
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
    }, [newMentions]);

    /** Find the selected item across both analysis and mention items. */
    const selectedItem = useMemo(() => {
        if (selectedItemId === null) return null;
        const fromAnalysis = analysisItems.find(
            (i) => i.id === selectedItemId,
        );
        if (fromAnalysis) return fromAnalysis;
        return (
            newMentions.find((i) => i.id === selectedItemId) ?? null
        );
    }, [analysisItems, newMentions, selectedItemId]);

    /** Whether any acknowledged findings exist. */
    const hasAcknowledgedFindings = useMemo(
        () => analysisItems.some((i) => i.resolution === 'acknowledged'),
        [analysisItems],
    );

    // -- Handlers ----------------------------------------------------------

    const handleResolve = (
        itemId: number,
        resolution: 'acknowledged' | 'accepted' | 'dismissed',
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
                    const allItems = [...analysisItems, ...newMentions];
                    const pendingItems = allItems.filter(
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

    const handleGenerateRevision = () => {
        if (!job) return;
        generateRevision.mutate(job.id, {
            onSuccess: (data) => {
                setRevisionContent(data.revisedContent);
                setIsEditing(false);
            },
            onError: (err: Error) => {
                console.error(
                    'Failed to generate revision:',
                    err.message,
                );
            },
        });
    };

    const handleApplyRevision = () => {
        if (!job || !revisionContent) return;
        applyRevision.mutate(
            {
                jobId: job.id,
                req: { revisedContent: revisionContent },
            },
            {
                onSuccess: () => {
                    setRevisionCount((prev) => prev + 1);
                    setRevisionContent(null);
                    setIsEditing(false);
                },
                onError: (err: Error) => {
                    console.error(
                        'Failed to apply revision:',
                        err.message,
                    );
                },
            },
        );
    };

    // -- Render ------------------------------------------------------------

    return (
        <Stack spacing={2} sx={{ height: '100%' }}>
            {/* Revision workflow section */}
            <Paper variant="outlined" sx={{ p: 2 }}>
                <Stack spacing={2}>
                    {/* Header with iteration counter */}
                    <Stack
                        direction="row"
                        alignItems="center"
                        spacing={1}
                    >
                        <Typography variant="h6">
                            Revision Workflow
                        </Typography>
                        {revisionCount > 0 && (
                            <Chip
                                label={`Iteration ${revisionCount}`}
                                size="small"
                                color="primary"
                                variant="outlined"
                            />
                        )}
                    </Stack>

                    {/* Generate button */}
                    <Box>
                        <Button
                            variant="contained"
                            startIcon={
                                generateRevision.isPending ? (
                                    <CircularProgress
                                        size={16}
                                        color="inherit"
                                    />
                                ) : (
                                    <PlayArrow />
                                )
                            }
                            disabled={
                                generateRevision.isPending ||
                                !hasAcknowledgedFindings
                            }
                            onClick={handleGenerateRevision}
                        >
                            Generate Revision
                        </Button>
                    </Box>

                    {/* Summary */}
                    {generateRevision.data?.summary && (
                        <Alert severity="info" variant="outlined">
                            {generateRevision.data.summary}
                        </Alert>
                    )}

                    {/* Diff view */}
                    {revisionContent !== null &&
                        generateRevision.data && (
                            <Box>
                                <Stack
                                    direction="row"
                                    spacing={1}
                                    sx={{ mb: 1 }}
                                >
                                    <Button
                                        variant="outlined"
                                        size="small"
                                        startIcon={<Edit />}
                                        onClick={() =>
                                            setIsEditing(
                                                (prev) => !prev,
                                            )
                                        }
                                    >
                                        {isEditing
                                            ? 'Stop Editing'
                                            : 'Edit Revision'}
                                    </Button>
                                    <Button
                                        variant="contained"
                                        size="small"
                                        color="success"
                                        disabled={
                                            applyRevision.isPending
                                        }
                                        onClick={handleApplyRevision}
                                    >
                                        Apply Revision
                                    </Button>
                                </Stack>
                                <Grid container spacing={2}>
                                    <Grid item xs={12} md={6}>
                                        <Paper
                                            variant="outlined"
                                            sx={{
                                                p: 2,
                                                bgcolor: 'grey.50',
                                                maxHeight: 300,
                                                overflow: 'auto',
                                            }}
                                        >
                                            <Typography
                                                variant="subtitle2"
                                                gutterBottom
                                                color="text.secondary"
                                            >
                                                Original
                                            </Typography>
                                            <Typography
                                                variant="body2"
                                                sx={{
                                                    whiteSpace:
                                                        'pre-wrap',
                                                }}
                                            >
                                                {
                                                    generateRevision
                                                        .data
                                                        .originalContent
                                                }
                                            </Typography>
                                        </Paper>
                                    </Grid>
                                    <Grid item xs={12} md={6}>
                                        <Paper
                                            variant="outlined"
                                            sx={{
                                                p: 2,
                                                bgcolor: 'grey.50',
                                                maxHeight: 300,
                                                overflow: 'auto',
                                            }}
                                        >
                                            <Typography
                                                variant="subtitle2"
                                                gutterBottom
                                                color="text.secondary"
                                            >
                                                Revised
                                            </Typography>
                                            {isEditing ? (
                                                <TextField
                                                    multiline
                                                    fullWidth
                                                    value={
                                                        revisionContent
                                                    }
                                                    onChange={(e) =>
                                                        setRevisionContent(
                                                            e.target
                                                                .value,
                                                        )
                                                    }
                                                    variant="outlined"
                                                    size="small"
                                                    minRows={6}
                                                    inputProps={{
                                                        'aria-label':
                                                            'Edit revised content',
                                                    }}
                                                />
                                            ) : (
                                                <Typography
                                                    variant="body2"
                                                    sx={{
                                                        whiteSpace:
                                                            'pre-wrap',
                                                    }}
                                                >
                                                    {revisionContent}
                                                </Typography>
                                            )}
                                        </Paper>
                                    </Grid>
                                </Grid>
                            </Box>
                        )}
                </Stack>
            </Paper>

            {/* Two-column layout for findings and detail */}
            <Grid
                container
                spacing={2}
                sx={{ flexGrow: 1, minHeight: 0 }}
            >
                {/* Left panel -- findings list */}
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
                                {analysisItems.length} findings (
                                {pendingCount} pending)
                            </Typography>
                        </Box>
                        <Divider />

                        {/* Section 1: Analysis findings */}
                        <List disablePadding>
                            {groupedAnalysisItems.map((group) => (
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
                                            label={group.severity}
                                            size="small"
                                            color={
                                                SEVERITY_CHIP_COLOR[
                                                    group.severity
                                                ]
                                            }
                                        />
                                        <Chip
                                            label={
                                                group.items.length
                                            }
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
                                                handleSelectItem(
                                                    item,
                                                )
                                            }
                                            sx={{ pl: 4 }}
                                        >
                                            <ListItemText
                                                primary={
                                                    <Typography
                                                        variant="body2"
                                                        fontWeight="bold"
                                                        noWrap
                                                    >
                                                        {
                                                            item.matchedText
                                                        }
                                                    </Typography>
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
                                                    <Tooltip title="Acknowledge">
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
                                                                    'acknowledged',
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
                            {groupedAnalysisItems.length === 0 && (
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
                                        No analysis findings in this
                                        phase.
                                    </Typography>
                                </Box>
                            )}
                        </List>

                        {/* Section 2: New mentions after revision */}
                        {groupedNewMentions.length > 0 && (
                            <>
                                <Divider />
                                <Box sx={{ px: 2, py: 1 }}>
                                    <Typography
                                        variant="subtitle2"
                                        color="text.secondary"
                                    >
                                        New Mentions (after revision)
                                    </Typography>
                                </Box>
                                <Divider />
                                <List disablePadding>
                                    {groupedNewMentions.map(
                                        (group) => (
                                            <Box key={group.key}>
                                                {/* Group header */}
                                                <Box
                                                    sx={{
                                                        display:
                                                            'flex',
                                                        alignItems:
                                                            'center',
                                                        gap: 1,
                                                        px: 2,
                                                        py: 1,
                                                        bgcolor:
                                                            'action.hover',
                                                    }}
                                                >
                                                    <Box
                                                        sx={{
                                                            width: 10,
                                                            height: 10,
                                                            borderRadius:
                                                                '50%',
                                                            bgcolor:
                                                                group.color,
                                                            flexShrink: 0,
                                                        }}
                                                    />
                                                    <Typography
                                                        variant="subtitle2"
                                                        sx={{
                                                            flexGrow: 1,
                                                        }}
                                                    >
                                                        {group.label}
                                                    </Typography>
                                                    <Chip
                                                        label={
                                                            group
                                                                .items
                                                                .length
                                                        }
                                                        size="small"
                                                    />
                                                </Box>
                                                {/* Items */}
                                                {group.items.map(
                                                    (item) => (
                                                        <ListItemButton
                                                            key={
                                                                item.id
                                                            }
                                                            selected={
                                                                selectedItemId ===
                                                                item.id
                                                            }
                                                            onClick={() =>
                                                                handleSelectItem(
                                                                    item,
                                                                )
                                                            }
                                                            sx={{
                                                                pl: 4,
                                                            }}
                                                        >
                                                            <ListItemText
                                                                primary={
                                                                    <Typography
                                                                        variant="body2"
                                                                        fontWeight="bold"
                                                                        noWrap
                                                                    >
                                                                        {
                                                                            item.matchedText
                                                                        }
                                                                    </Typography>
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
                                                            {item.resolution ===
                                                                'pending' && (
                                                                <Stack
                                                                    direction="row"
                                                                    spacing={
                                                                        0.5
                                                                    }
                                                                    sx={{
                                                                        ml: 1,
                                                                    }}
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
                                                    ),
                                                )}
                                            </Box>
                                        ),
                                    )}
                                </List>
                            </>
                        )}
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
        resolution: 'acknowledged' | 'accepted' | 'dismissed',
    ) => void;
}

function DetailPanel({ item, isResolving, onResolve }: DetailPanelProps) {
    const isPending = item.resolution === 'pending';

    /** Look up severity from REVISE_GROUPS for analysis items. */
    const severityInfo = REVISE_GROUPS.find(
        (g) => g.key === item.detectionType,
    );

    return (
        <Stack spacing={3}>
            {/* Header with matched text and severity */}
            <Box>
                <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                >
                    <Typography variant="h6">
                        {item.matchedText}
                    </Typography>
                    {severityInfo && (
                        <Chip
                            label={severityInfo.severity}
                            size="small"
                            color={
                                SEVERITY_CHIP_COLOR[
                                    severityInfo.severity
                                ]
                            }
                        />
                    )}
                </Stack>
                <Typography
                    variant="caption"
                    color="text.secondary"
                >
                    {item.detectionType.replace(/_/g, ' ')}
                </Typography>
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

            {/* Suggested content */}
            {item.suggestedContent && (
                <Box>
                    <Typography variant="subtitle2" gutterBottom>
                        Suggested Content
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{
                            p: 2,
                            bgcolor: 'grey.50',
                        }}
                    >
                        <Typography
                            variant="body2"
                            sx={{ whiteSpace: 'pre-wrap' }}
                        >
                            {typeof item.suggestedContent === 'object'
                                ? JSON.stringify(
                                      item.suggestedContent,
                                      null,
                                      2,
                                  )
                                : String(item.suggestedContent)}
                        </Typography>
                    </Paper>
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
                                onResolve(item.id, 'acknowledged')
                            }
                        >
                            Acknowledge
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
