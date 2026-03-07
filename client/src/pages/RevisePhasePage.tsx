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
 *  - Top section: Header explaining the pin/dismiss workflow.
 *  - Left panel (~40%): Analysis findings grouped by detection type with
 *    severity indicators, and new identification mentions after revision.
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
    CircularProgress,
} from '@mui/material';
import {
    Check,
    CheckCircle,
    Close,
} from '@mui/icons-material';
import { useWizardContext } from '../contexts/AnalysisWizardContext';
import { useResolveItem } from '../hooks/useContentAnalysis';
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
    const { phaseItems, items, job } = useWizardContext();
    const resolveItem = useResolveItem(Number(campaignId));

    // -- Local state -------------------------------------------------------

    const [selectedItemId, setSelectedItemId] = useState<
        number | null
    >(null);

    // -- Derived data ------------------------------------------------------

    /** Analysis findings (this phase's items). */
    const analysisItems = phaseItems;

    /** New identification mentions from all items (pending only). */
    const newMentions = useMemo(
        () =>
            items.filter(
                (i) =>
                    IDENTIFY_TYPES.includes(i.detectionType) &&
                    i.resolution === 'pending',
            ),
        [items],
    );

    /** Group all analysis items by detection type. */
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

    /** Find the selected item across all analysis and mention items. */
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

    // -- Handlers ----------------------------------------------------------

    /**
     * After resolving an item (pin or dismiss), auto-advance to the next
     * pending item.
     */
    const advanceAfterResolve = (resolvedItemId: number) => {
        const allItems = [...analysisItems, ...newMentions];
        const pendingItems = allItems.filter(
            (i) =>
                i.resolution === 'pending' &&
                i.id !== resolvedItemId,
        );
        if (pendingItems.length > 0) {
            setSelectedItemId(pendingItems[0].id);
        } else {
            setSelectedItemId(null);
        }
    };

    const handlePin = (itemId: number) => {
        resolveItem.mutate(
            {
                itemId,
                req: { resolution: 'pinned' },
            },
            {
                onSuccess: () => advanceAfterResolve(itemId),
                onError: (err: Error) => {
                    console.error(
                        'Failed to pin item:',
                        err.message,
                    );
                },
            },
        );
    };

    const handleDismiss = (itemId: number) => {
        resolveItem.mutate(
            {
                itemId,
                req: { resolution: 'dismissed' as const },
            },
            {
                onSuccess: () => advanceAfterResolve(itemId),
                onError: (err: Error) => {
                    console.error(
                        'Failed to dismiss item:',
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
            {/* Revision workflow header */}
            <Paper variant="outlined" sx={{ p: 2 }}>
                <Typography variant="h6">Revise Findings</Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                    Review structural suggestions from analysis. Pin items to
                    keep as reminders on the edit screen, or dismiss.
                </Typography>
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
                                {analysisItems.length} findings
                            </Typography>
                        </Box>
                        <Divider />

                        {/* Analysis findings grouped by detection type */}
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
                                                    <Stack
                                                        direction="row"
                                                        alignItems="center"
                                                        spacing={1}
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
                                                        {item.resolution ===
                                                            'pinned' && (
                                                            <Chip
                                                                label="Pinned"
                                                                size="small"
                                                                color="success"
                                                                variant="outlined"
                                                            />
                                                        )}
                                                        {item.resolution ===
                                                            'dismissed' && (
                                                            <Chip
                                                                label="Dismissed"
                                                                size="small"
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
                                                    <Tooltip title="Pin">
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
                                                                handlePin(
                                                                    item.id,
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
                                                                handleDismiss(
                                                                    item.id,
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

                            {/* Empty state */}
                            {groupedAnalysisItems.length === 0 && (
                                <Box
                                    sx={{
                                        p: 3,
                                        textAlign: 'center',
                                    }}
                                >
                                    {job?.status === 'completed' || job?.status === 'failed' ? (
                                        <Typography
                                            variant="body2"
                                            color="text.secondary"
                                        >
                                            No analysis findings in this
                                            phase.
                                        </Typography>
                                    ) : (
                                        <>
                                            <CircularProgress size={24} sx={{ mb: 1 }} />
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                            >
                                                Analysis in progress...
                                            </Typography>
                                        </>
                                    )}
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
                                                                                handlePin(
                                                                                    item.id,
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
                                                                                handleDismiss(
                                                                                    item.id,
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
                                onPin={handlePin}
                                onDismiss={handleDismiss}
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
// Suggested content renderer
// ---------------------------------------------------------------------------

const SEVERITY_ALERT_MAP: Record<string, 'error' | 'warning' | 'info' | 'success'> = {
    critical: 'error',
    warning: 'warning',
    info: 'info',
};

/**
 * Render suggestedContent as readable prose instead of raw JSON.
 *
 * Two shapes:
 *  - analysis_report: { report: string }
 *  - all others: { category, description, severity, suggestion, lineReference }
 */
function SuggestedContentView({
    content,
}: {
    content: Record<string, unknown>;
}) {
    // Shape 1: analysis_report with a markdown-ish "report" field
    if (typeof content.report === 'string') {
        return (
            <Typography
                variant="body2"
                sx={{ whiteSpace: 'pre-wrap', lineHeight: 1.7 }}
            >
                {content.report}
            </Typography>
        );
    }

    // Shape 2: structured suggestion
    const severity = String(content.severity ?? 'info');
    const alertColor = SEVERITY_ALERT_MAP[severity] ?? 'info';
    const description = content.description ? String(content.description) : '';
    const suggestion = content.suggestion ? String(content.suggestion) : '';
    const category = content.category ? String(content.category).replace(/_/g, ' ') : '';

    return (
        <Stack spacing={1.5}>
            {description && (
                <Typography variant="body2">
                    {description}
                </Typography>
            )}
            {suggestion && (
                <Alert severity={alertColor} variant="outlined">
                    {suggestion}
                </Alert>
            )}
            {category && (
                <Typography variant="caption" color="text.secondary">
                    Category: {category}
                </Typography>
            )}
        </Stack>
    );
}

// ---------------------------------------------------------------------------
// Detail panel sub-component
// ---------------------------------------------------------------------------

interface DetailPanelProps {
    item: ContentAnalysisItem;
    isResolving: boolean;
    onPin: (itemId: number) => void;
    onDismiss: (itemId: number) => void;
}

function DetailPanel({ item, isResolving, onPin, onDismiss }: DetailPanelProps) {
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
                            bgcolor: 'action.hover',
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
                        {item.suggestedContent.report
                            ? 'Analysis Report'
                            : 'Suggestion'}
                    </Typography>
                    <Paper
                        variant="outlined"
                        sx={{
                            p: 2,
                            bgcolor: 'action.hover',
                        }}
                    >
                        <SuggestedContentView content={item.suggestedContent} />
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
                            disabled={isResolving}
                            startIcon={<CheckCircle />}
                            onClick={() => onPin(item.id)}
                        >
                            Pin
                        </Button>
                        <Button
                            variant="outlined"
                            color="error"
                            disabled={isResolving}
                            startIcon={<Close />}
                            onClick={() => onDismiss(item.id)}
                        >
                            Dismiss
                        </Button>
                    </Stack>
                </Box>
            )}
        </Stack>
    );
}
