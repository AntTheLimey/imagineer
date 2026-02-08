// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useCallback, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import {
    Alert,
    Autocomplete,
    Box,
    Button,
    Chip,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    FormControl,
    FormControlLabel,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Switch,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    Visibility as ViewIcon,
    VisibilityOff as HiddenIcon,
    Event as EventIcon,
    ArrowUpward as AscIcon,
    ArrowDownward as DescIcon,
    FilterList as FilterIcon,
} from '@mui/icons-material';
import {
    useTimelineEvents,
    useTimelineEvent,
    useCreateTimelineEvent,
    useUpdateTimelineEvent,
    useDeleteTimelineEvent,
    useEntities,
} from '../hooks';
import type { TimelineEvent, DatePrecision, Entity } from '../types';

/**
 * Available date precision options.
 */
const DATE_PRECISION_OPTIONS: DatePrecision[] = [
    'exact',
    'approximate',
    'month',
    'year',
    'unknown',
];

/**
 * Human-readable labels for date precision.
 */
const DATE_PRECISION_LABELS: Record<DatePrecision, string> = {
    exact: 'Exact Date',
    approximate: 'Approximate',
    month: 'Month Only',
    year: 'Year Only',
    unknown: 'Unknown',
};

/**
 * Color mapping for date precision (for timeline dots).
 */
const DATE_PRECISION_COLORS: Record<DatePrecision, 'primary' | 'secondary' | 'default' | 'warning' | 'info'> = {
    exact: 'primary',
    approximate: 'secondary',
    month: 'info',
    year: 'warning',
    unknown: 'default',
};

/**
 * Form data structure for create/edit dialogs.
 */
interface TimelineFormData {
    description: string;
    eventDate: string;
    eventTime: string;
    datePrecision: DatePrecision;
    entityIds: number[];
    sessionId: number | undefined;
    isPlayerKnown: boolean;
    sourceDocument: string;
}

/**
 * Default form values.
 */
const DEFAULT_FORM_DATA: TimelineFormData = {
    description: '',
    eventDate: '',
    eventTime: '',
    datePrecision: 'exact',
    entityIds: [],
    sessionId: undefined,
    isPlayerKnown: true,
    sourceDocument: '',
};

/**
 * Formats a date string for display based on precision.
 */
function formatEventDate(
    dateString: string | undefined,
    precision: DatePrecision
): string {
    if (!dateString) {
        return 'Unknown Date';
    }

    const date = new Date(dateString);

    switch (precision) {
        case 'exact':
        case 'approximate':
            return date.toLocaleDateString(undefined, {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
            });
        case 'month':
            return date.toLocaleDateString(undefined, {
                year: 'numeric',
                month: 'long',
            });
        case 'year':
            return date.toLocaleDateString(undefined, {
                year: 'numeric',
            });
        case 'unknown':
        default:
            return 'Unknown Date';
    }
}

/**
 * Formats a date for display in the timeline.
 */
function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

export default function Timeline() {
    const { id } = useParams<{ id: string }>();
    const campaignId = id ? Number(id) : undefined;

    // Pagination state
    const [page, setPage] = useState(1);
    const [pageSize] = useState(20);

    // Sort order (true = newest first, false = oldest first)
    const [sortDescending, setSortDescending] = useState(true);

    // Filter state
    const [precisionFilter, setPrecisionFilter] = useState<DatePrecision | ''>('');
    const [playerKnownFilter, setPlayerKnownFilter] = useState<boolean | ''>('');
    const [startDateFilter, setStartDateFilter] = useState('');
    const [endDateFilter, setEndDateFilter] = useState('');
    const [showFilters, setShowFilters] = useState(false);

    // Dialog states
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [viewDialogOpen, setViewDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    // Selected event for view/edit/delete
    const [selectedEventId, setSelectedEventId] = useState<number | null>(null);

    // Form state
    const [formData, setFormData] = useState<TimelineFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<Partial<Record<keyof TimelineFormData, string>>>({});

    // Fetch timeline events
    const {
        data: events,
        isLoading: eventsLoading,
        error: eventsError,
    } = useTimelineEvents({
        campaignId: campaignId ?? 0,
        page,
        pageSize,
        isPlayerKnown: playerKnownFilter === '' ? undefined : playerKnownFilter,
        startDate: startDateFilter || undefined,
        endDate: endDateFilter || undefined,
    });

    // Fetch entities for the entity selector
    const {
        data: entities,
    } = useEntities({
        campaignId: campaignId ?? 0,
        pageSize: 100, // Load a reasonable amount for the selector
    });

    // Fetch single event for view/edit
    const {
        data: selectedEvent,
        isLoading: selectedEventLoading,
    } = useTimelineEvent(
        campaignId ?? 0,
        selectedEventId ?? 0,
        { enabled: !!selectedEventId && (viewDialogOpen || editDialogOpen) }
    );

    // Mutations
    const createEvent = useCreateTimelineEvent();
    const updateEvent = useUpdateTimelineEvent();
    const deleteEvent = useDeleteTimelineEvent();

    // Sort and filter events
    const sortedEvents = useMemo(() => {
        if (!events) return [];

        let filtered = [...events];

        // Apply precision filter
        if (precisionFilter) {
            filtered = filtered.filter((e) => e.datePrecision === precisionFilter);
        }

        // Sort by date
        filtered.sort((a, b) => {
            const dateA = a.eventDate ? new Date(a.eventDate).getTime() : 0;
            const dateB = b.eventDate ? new Date(b.eventDate).getTime() : 0;
            return sortDescending ? dateB - dateA : dateA - dateB;
        });

        return filtered;
    }, [events, precisionFilter, sortDescending]);

    // Create entity lookup map for displaying entity names
    const entityMap = useMemo(() => {
        const map = new Map<number, Entity>();
        if (entities) {
            for (const entity of entities) {
                map.set(entity.id, entity);
            }
        }
        return map;
    }, [entities]);

    // Validate form
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof TimelineFormData, string>> = {};

        if (!formData.description.trim()) {
            errors.description = 'Description is required';
        }

        if (formData.datePrecision !== 'unknown' && !formData.eventDate) {
            errors.eventDate = 'Date is required unless precision is "Unknown"';
        }

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, [formData]);

    // Handle create
    const handleCreate = async () => {
        if (!validateForm() || !campaignId) return;

        try {
            await createEvent.mutateAsync({
                campaignId,
                description: formData.description.trim(),
                eventDate: formData.eventDate || undefined,
                eventTime: formData.eventTime.trim() || undefined,
                datePrecision: formData.datePrecision,
                entityIds: formData.entityIds.length > 0 ? formData.entityIds : undefined,
                sessionId: formData.sessionId ?? undefined,
                isPlayerKnown: formData.isPlayerKnown,
                sourceDocument: formData.sourceDocument.trim() || undefined,
            });
            setCreateDialogOpen(false);
            setFormData(DEFAULT_FORM_DATA);
            setFormErrors({});
        } catch (error) {
            console.error('Failed to create timeline event:', error);
        }
    };

    // Handle update
    const handleUpdate = async () => {
        if (!validateForm() || !campaignId || !selectedEventId) return;

        try {
            await updateEvent.mutateAsync({
                campaignId,
                eventId: selectedEventId,
                input: {
                    description: formData.description.trim(),
                    eventDate: formData.eventDate || undefined,
                    eventTime: formData.eventTime.trim() || undefined,
                    datePrecision: formData.datePrecision,
                    entityIds: formData.entityIds,
                    sessionId: formData.sessionId ?? undefined,
                    isPlayerKnown: formData.isPlayerKnown,
                    sourceDocument: formData.sourceDocument.trim() || undefined,
                },
            });
            setEditDialogOpen(false);
            setSelectedEventId(null);
            setFormData(DEFAULT_FORM_DATA);
            setFormErrors({});
        } catch (error) {
            console.error('Failed to update timeline event:', error);
        }
    };

    // Handle delete
    const handleDelete = async () => {
        if (!campaignId || !selectedEventId) return;

        try {
            await deleteEvent.mutateAsync({
                campaignId,
                eventId: selectedEventId,
            });
            setDeleteDialogOpen(false);
            setSelectedEventId(null);
        } catch (error) {
            console.error('Failed to delete timeline event:', error);
        }
    };

    // Open view dialog
    const openViewDialog = (event: TimelineEvent) => {
        setSelectedEventId(event.id);
        setViewDialogOpen(true);
    };

    // Open edit dialog
    const openEditDialog = (event: TimelineEvent) => {
        setSelectedEventId(event.id);
        setFormData({
            description: event.description,
            eventDate: event.eventDate ?? '',
            eventTime: event.eventTime ?? '',
            datePrecision: event.datePrecision,
            entityIds: event.entityIds ?? [],
            sessionId: event.sessionId ?? undefined,
            isPlayerKnown: event.isPlayerKnown,
            sourceDocument: event.sourceDocument ?? '',
        });
        setFormErrors({});
        setEditDialogOpen(true);
    };

    // Open delete dialog
    const openDeleteDialog = (event: TimelineEvent) => {
        setSelectedEventId(event.id);
        setDeleteDialogOpen(true);
    };

    // Open create dialog
    const openCreateDialog = () => {
        setFormData(DEFAULT_FORM_DATA);
        setFormErrors({});
        setCreateDialogOpen(true);
    };

    // Clear filters
    const clearFilters = () => {
        setPrecisionFilter('');
        setPlayerKnownFilter('');
        setStartDateFilter('');
        setEndDateFilter('');
    };

    // Loading state
    if (eventsLoading && !events) {
        return (
            <Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                    <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                        Timeline
                    </Typography>
                    <Skeleton variant="rectangular" width={130} height={36} />
                </Box>
                <Paper sx={{ p: 2, mb: 2 }}>
                    <Skeleton variant="rectangular" width={200} height={56} />
                </Paper>
                <Box sx={{ pl: 4 }}>
                    {[1, 2, 3, 4, 5].map((i) => (
                        <Box key={i} sx={{ display: 'flex', mb: 3 }}>
                            <Skeleton variant="circular" width={40} height={40} sx={{ mr: 2 }} />
                            <Box sx={{ flex: 1 }}>
                                <Skeleton width={120} height={20} />
                                <Skeleton width="80%" height={24} />
                                <Skeleton width={200} height={20} />
                            </Box>
                        </Box>
                    ))}
                </Box>
            </Box>
        );
    }

    const hasActiveFilters = precisionFilter || playerKnownFilter !== '' || startDateFilter || endDateFilter;

    return (
        <Box>
            {/* Header */}
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                    Timeline
                </Typography>
                <Box sx={{ display: 'flex', gap: 1 }}>
                    <Tooltip title={sortDescending ? 'Showing newest first' : 'Showing oldest first'}>
                        <IconButton
                            onClick={() => setSortDescending(!sortDescending)}
                            color="primary"
                        >
                            {sortDescending ? <DescIcon /> : <AscIcon />}
                        </IconButton>
                    </Tooltip>
                    <Tooltip title="Toggle filters">
                        <IconButton
                            onClick={() => setShowFilters(!showFilters)}
                            color={hasActiveFilters ? 'primary' : 'default'}
                        >
                            <FilterIcon />
                        </IconButton>
                    </Tooltip>
                    <Button
                        variant="contained"
                        startIcon={<AddIcon />}
                        onClick={openCreateDialog}
                    >
                        New Event
                    </Button>
                </Box>
            </Box>

            {/* Error alerts */}
            {eventsError && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to load timeline events. Please try again later.
                </Alert>
            )}

            {createEvent.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to create event. Please try again.
                </Alert>
            )}

            {updateEvent.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to update event. Please try again.
                </Alert>
            )}

            {deleteEvent.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to delete event. Please try again.
                </Alert>
            )}

            {/* Filters */}
            {showFilters && (
                <Paper sx={{ p: 2, mb: 2 }}>
                    <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center' }}>
                        <FormControl sx={{ minWidth: 160 }}>
                            <InputLabel>Date Precision</InputLabel>
                            <Select
                                value={precisionFilter}
                                label="Date Precision"
                                onChange={(e) => setPrecisionFilter(e.target.value as DatePrecision | '')}
                            >
                                <MenuItem value="">All</MenuItem>
                                {DATE_PRECISION_OPTIONS.map((precision) => (
                                    <MenuItem key={precision} value={precision}>
                                        {DATE_PRECISION_LABELS[precision]}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl sx={{ minWidth: 160 }}>
                            <InputLabel>Visibility</InputLabel>
                            <Select
                                value={playerKnownFilter === '' ? '' : playerKnownFilter ? 'known' : 'hidden'}
                                label="Visibility"
                                onChange={(e) => {
                                    const val = e.target.value;
                                    if (val === '') setPlayerKnownFilter('');
                                    else if (val === 'known') setPlayerKnownFilter(true);
                                    else setPlayerKnownFilter(false);
                                }}
                            >
                                <MenuItem value="">All</MenuItem>
                                <MenuItem value="known">Player Known</MenuItem>
                                <MenuItem value="hidden">Hidden from Players</MenuItem>
                            </Select>
                        </FormControl>

                        <TextField
                            label="Start Date"
                            type="date"
                            value={startDateFilter}
                            onChange={(e) => setStartDateFilter(e.target.value)}
                            InputLabelProps={{ shrink: true }}
                            sx={{ width: 160 }}
                        />

                        <TextField
                            label="End Date"
                            type="date"
                            value={endDateFilter}
                            onChange={(e) => setEndDateFilter(e.target.value)}
                            InputLabelProps={{ shrink: true }}
                            sx={{ width: 160 }}
                        />

                        {hasActiveFilters && (
                            <Button variant="outlined" onClick={clearFilters}>
                                Clear Filters
                            </Button>
                        )}
                    </Box>
                </Paper>
            )}

            {/* Timeline Display */}
            {sortedEvents.length === 0 ? (
                <Paper sx={{ p: 4, textAlign: 'center' }}>
                    <EventIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 2 }} />
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                        No timeline events found
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        {hasActiveFilters
                            ? 'No events match your current filters.'
                            : 'Create your first event to start building your campaign timeline.'}
                    </Typography>
                    {hasActiveFilters ? (
                        <Button variant="outlined" onClick={clearFilters}>
                            Clear Filters
                        </Button>
                    ) : (
                        <Button variant="outlined" onClick={openCreateDialog}>
                            Create Event
                        </Button>
                    )}
                </Paper>
            ) : (
                <Paper sx={{ p: 2 }}>
                    {/* Custom Timeline Implementation */}
                    <Box sx={{ position: 'relative' }}>
                        {sortedEvents.map((event, index) => (
                            <Box
                                key={event.id}
                                sx={{
                                    display: 'flex',
                                    mb: index < sortedEvents.length - 1 ? 0 : 0,
                                }}
                            >
                                {/* Date Column */}
                                <Box
                                    sx={{
                                        width: 140,
                                        flexShrink: 0,
                                        textAlign: 'right',
                                        pr: 2,
                                        pt: 1,
                                    }}
                                >
                                    <Typography variant="body2" color="text.secondary">
                                        {formatEventDate(event.eventDate, event.datePrecision)}
                                    </Typography>
                                    {event.eventTime && (
                                        <Typography variant="caption" display="block" color="text.secondary">
                                            {event.eventTime}
                                        </Typography>
                                    )}
                                    {event.datePrecision !== 'exact' && (
                                        <Chip
                                            label={DATE_PRECISION_LABELS[event.datePrecision]}
                                            size="small"
                                            variant="outlined"
                                            sx={{ mt: 0.5 }}
                                        />
                                    )}
                                </Box>

                                {/* Timeline Line and Dot */}
                                <Box
                                    sx={{
                                        display: 'flex',
                                        flexDirection: 'column',
                                        alignItems: 'center',
                                        width: 40,
                                        flexShrink: 0,
                                    }}
                                >
                                    <Tooltip title={event.isPlayerKnown ? 'Player Known' : 'Hidden from Players'}>
                                        <Box
                                            sx={{
                                                width: 32,
                                                height: 32,
                                                borderRadius: '50%',
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                bgcolor: event.isPlayerKnown
                                                    ? `${DATE_PRECISION_COLORS[event.datePrecision]}.main`
                                                    : 'warning.main',
                                                color: 'white',
                                                cursor: 'pointer',
                                                '&:hover': {
                                                    transform: 'scale(1.1)',
                                                },
                                                transition: 'transform 0.2s',
                                            }}
                                            onClick={() => openViewDialog(event)}
                                        >
                                            {event.isPlayerKnown ? (
                                                <ViewIcon fontSize="small" />
                                            ) : (
                                                <HiddenIcon fontSize="small" />
                                            )}
                                        </Box>
                                    </Tooltip>
                                    {index < sortedEvents.length - 1 && (
                                        <Box
                                            sx={{
                                                width: 2,
                                                flexGrow: 1,
                                                minHeight: 40,
                                                bgcolor: 'divider',
                                            }}
                                        />
                                    )}
                                </Box>

                                {/* Content */}
                                <Box sx={{ flex: 1, pb: 3 }}>
                                    <Paper
                                        elevation={1}
                                        sx={{
                                            p: 2,
                                            cursor: 'pointer',
                                            '&:hover': {
                                                bgcolor: 'action.hover',
                                            },
                                            opacity: event.isPlayerKnown ? 1 : 0.7,
                                            borderLeft: event.isPlayerKnown
                                                ? undefined
                                                : '3px solid',
                                            borderLeftColor: event.isPlayerKnown
                                                ? undefined
                                                : 'warning.main',
                                        }}
                                        onClick={() => openViewDialog(event)}
                                    >
                                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                            <Box sx={{ flex: 1 }}>
                                                <Typography variant="body1">
                                                    {event.description}
                                                </Typography>

                                                {/* Linked Entities */}
                                                {(event.entityIds ?? []).length > 0 && (
                                                    <Box sx={{ mt: 1, display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                                                        {(event.entityIds ?? []).slice(0, 5).map((entityId) => {
                                                            const entity = entityMap.get(entityId);
                                                            return (
                                                                <Chip
                                                                    key={entityId}
                                                                    label={entity?.name ?? String(entityId)}
                                                                    size="small"
                                                                    variant="outlined"
                                                                />
                                                            );
                                                        })}
                                                        {(event.entityIds ?? []).length > 5 && (
                                                            <Chip
                                                                label={`+${(event.entityIds ?? []).length - 5} more`}
                                                                size="small"
                                                                variant="outlined"
                                                            />
                                                        )}
                                                    </Box>
                                                )}

                                                {!event.isPlayerKnown && (
                                                    <Chip
                                                        label="Hidden from Players"
                                                        size="small"
                                                        color="warning"
                                                        sx={{ mt: 1 }}
                                                    />
                                                )}
                                            </Box>

                                            {/* Action buttons */}
                                            <Box sx={{ display: 'flex', ml: 1 }}>
                                                <Tooltip title="Edit">
                                                    <IconButton
                                                        size="small"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            openEditDialog(event);
                                                        }}
                                                    >
                                                        <EditIcon fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                                <Tooltip title="Delete">
                                                    <IconButton
                                                        size="small"
                                                        color="error"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            openDeleteDialog(event);
                                                        }}
                                                    >
                                                        <DeleteIcon fontSize="small" />
                                                    </IconButton>
                                                </Tooltip>
                                            </Box>
                                        </Box>
                                    </Paper>
                                </Box>
                            </Box>
                        ))}
                    </Box>

                    {/* Pagination controls */}
                    {events && events.length >= pageSize && (
                        <Box sx={{ display: 'flex', justifyContent: 'center', gap: 2, mt: 2 }}>
                            <Button
                                disabled={page <= 1}
                                onClick={() => setPage((p) => Math.max(1, p - 1))}
                            >
                                Previous
                            </Button>
                            <Typography variant="body2" sx={{ alignSelf: 'center' }}>
                                Page {page}
                            </Typography>
                            <Button
                                disabled={events.length < pageSize}
                                onClick={() => setPage((p) => p + 1)}
                            >
                                Next
                            </Button>
                        </Box>
                    )}
                </Paper>
            )}

            {/* Create Event Dialog */}
            <Dialog
                open={createDialogOpen}
                onClose={() => setCreateDialogOpen(false)}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>Create Timeline Event</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        label="Description"
                        fullWidth
                        required
                        multiline
                        rows={3}
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        error={!!formErrors.description}
                        helperText={formErrors.description ?? 'What happened in this event?'}
                        sx={{ mt: 1 }}
                    />

                    <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                        <TextField
                            label="Event Date"
                            type="date"
                            value={formData.eventDate}
                            onChange={(e) => setFormData({ ...formData, eventDate: e.target.value })}
                            InputLabelProps={{ shrink: true }}
                            fullWidth
                            error={!!formErrors.eventDate}
                            helperText={formErrors.eventDate}
                        />

                        <TextField
                            label="Time (optional)"
                            value={formData.eventTime}
                            onChange={(e) => setFormData({ ...formData, eventTime: e.target.value })}
                            placeholder="e.g., Morning, 3:00 PM"
                            sx={{ minWidth: 180 }}
                        />

                        <FormControl sx={{ minWidth: 160 }}>
                            <InputLabel>Date Precision</InputLabel>
                            <Select
                                value={formData.datePrecision}
                                label="Date Precision"
                                onChange={(e) =>
                                    setFormData({ ...formData, datePrecision: e.target.value as DatePrecision })
                                }
                            >
                                {DATE_PRECISION_OPTIONS.map((precision) => (
                                    <MenuItem key={precision} value={precision}>
                                        {DATE_PRECISION_LABELS[precision]}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Box>

                    <Autocomplete
                        multiple
                        options={entities ?? []}
                        getOptionLabel={(option) => option.name}
                        value={(entities ?? []).filter((e) => formData.entityIds.includes(e.id))}
                        onChange={(_event, newValue) =>
                            setFormData({ ...formData, entityIds: newValue.map((e) => e.id) })
                        }
                        renderTags={(value, getTagProps) =>
                            value.map((option, index) => (
                                <Chip
                                    label={option.name}
                                    size="small"
                                    {...getTagProps({ index })}
                                    key={option.id}
                                />
                            ))
                        }
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Linked Entities"
                                placeholder="Select entities involved in this event"
                                sx={{ mt: 2 }}
                            />
                        )}
                    />

                    <TextField
                        label="Source Document (optional)"
                        fullWidth
                        value={formData.sourceDocument}
                        onChange={(e) => setFormData({ ...formData, sourceDocument: e.target.value })}
                        placeholder="Reference to source material"
                        sx={{ mt: 2 }}
                    />

                    <FormControlLabel
                        control={
                            <Switch
                                checked={formData.isPlayerKnown}
                                onChange={(e) =>
                                    setFormData({ ...formData, isPlayerKnown: e.target.checked })
                                }
                            />
                        }
                        label="Players know about this event"
                        sx={{ mt: 2 }}
                    />
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => setCreateDialogOpen(false)}
                        disabled={createEvent.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleCreate}
                        variant="contained"
                        disabled={!formData.description.trim() || createEvent.isPending}
                    >
                        {createEvent.isPending ? <CircularProgress size={24} /> : 'Create'}
                    </Button>
                </DialogActions>
            </Dialog>

            {/* View Event Dialog */}
            <Dialog
                open={viewDialogOpen}
                onClose={() => {
                    setViewDialogOpen(false);
                    setSelectedEventId(null);
                }}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>Timeline Event Details</DialogTitle>
                <DialogContent>
                    {selectedEventLoading ? (
                        <Box>
                            <Skeleton height={40} />
                            <Skeleton height={100} />
                            <Skeleton height={40} />
                            <Skeleton height={40} />
                        </Box>
                    ) : selectedEvent ? (
                        <Box sx={{ mt: 1 }}>
                            <Box sx={{ display: 'flex', gap: 2, mb: 2, flexWrap: 'wrap' }}>
                                <Chip
                                    label={DATE_PRECISION_LABELS[selectedEvent.datePrecision]}
                                    color={DATE_PRECISION_COLORS[selectedEvent.datePrecision]}
                                />
                                <Chip
                                    icon={selectedEvent.isPlayerKnown ? <ViewIcon /> : <HiddenIcon />}
                                    label={selectedEvent.isPlayerKnown ? 'Player Known' : 'Hidden'}
                                    color={selectedEvent.isPlayerKnown ? 'success' : 'warning'}
                                    variant="outlined"
                                />
                            </Box>

                            <Box sx={{ mb: 2 }}>
                                <Typography variant="subtitle2" color="text.secondary">
                                    Date
                                </Typography>
                                <Typography variant="body1">
                                    {formatEventDate(selectedEvent.eventDate, selectedEvent.datePrecision)}
                                    {selectedEvent.eventTime && ` at ${selectedEvent.eventTime}`}
                                </Typography>
                            </Box>

                            <Box sx={{ mb: 2 }}>
                                <Typography variant="subtitle2" color="text.secondary">
                                    Description
                                </Typography>
                                <Typography variant="body1" sx={{ whiteSpace: 'pre-wrap' }}>
                                    {selectedEvent.description}
                                </Typography>
                            </Box>

                            {(selectedEvent.entityIds ?? []).length > 0 && (
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                                        Linked Entities
                                    </Typography>
                                    <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                                        {(selectedEvent.entityIds ?? []).map((entityId) => {
                                            const entity = entityMap.get(entityId);
                                            return (
                                                <Chip
                                                    key={entityId}
                                                    label={entity?.name ?? String(entityId)}
                                                    size="small"
                                                />
                                            );
                                        })}
                                    </Box>
                                </Box>
                            )}

                            {selectedEvent.sourceDocument && (
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="subtitle2" color="text.secondary">
                                        Source Document
                                    </Typography>
                                    <Typography variant="body2">
                                        {selectedEvent.sourceDocument}
                                    </Typography>
                                </Box>
                            )}

                            <Divider sx={{ my: 2 }} />

                            <Box sx={{ display: 'flex', gap: 4 }}>
                                <Box>
                                    <Typography variant="caption" color="text.secondary">
                                        Created
                                    </Typography>
                                    <Typography variant="body2">
                                        {formatDate(selectedEvent.createdAt)}
                                    </Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary">
                                        Updated
                                    </Typography>
                                    <Typography variant="body2">
                                        {formatDate(selectedEvent.updatedAt)}
                                    </Typography>
                                </Box>
                            </Box>
                        </Box>
                    ) : null}
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setViewDialogOpen(false);
                            setSelectedEventId(null);
                        }}
                    >
                        Close
                    </Button>
                    {selectedEvent && (
                        <Button
                            variant="contained"
                            onClick={() => {
                                setViewDialogOpen(false);
                                openEditDialog(selectedEvent);
                            }}
                        >
                            Edit
                        </Button>
                    )}
                </DialogActions>
            </Dialog>

            {/* Edit Event Dialog */}
            <Dialog
                open={editDialogOpen}
                onClose={() => {
                    setEditDialogOpen(false);
                    setSelectedEventId(null);
                }}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>Edit Timeline Event</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        label="Description"
                        fullWidth
                        required
                        multiline
                        rows={3}
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        error={!!formErrors.description}
                        helperText={formErrors.description}
                        sx={{ mt: 1 }}
                    />

                    <Box sx={{ display: 'flex', gap: 2, mt: 2 }}>
                        <TextField
                            label="Event Date"
                            type="date"
                            value={formData.eventDate}
                            onChange={(e) => setFormData({ ...formData, eventDate: e.target.value })}
                            InputLabelProps={{ shrink: true }}
                            fullWidth
                            error={!!formErrors.eventDate}
                            helperText={formErrors.eventDate}
                        />

                        <TextField
                            label="Time (optional)"
                            value={formData.eventTime}
                            onChange={(e) => setFormData({ ...formData, eventTime: e.target.value })}
                            placeholder="e.g., Morning, 3:00 PM"
                            sx={{ minWidth: 180 }}
                        />

                        <FormControl sx={{ minWidth: 160 }}>
                            <InputLabel>Date Precision</InputLabel>
                            <Select
                                value={formData.datePrecision}
                                label="Date Precision"
                                onChange={(e) =>
                                    setFormData({ ...formData, datePrecision: e.target.value as DatePrecision })
                                }
                            >
                                {DATE_PRECISION_OPTIONS.map((precision) => (
                                    <MenuItem key={precision} value={precision}>
                                        {DATE_PRECISION_LABELS[precision]}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Box>

                    <Autocomplete
                        multiple
                        options={entities ?? []}
                        getOptionLabel={(option) => option.name}
                        value={(entities ?? []).filter((e) => formData.entityIds.includes(e.id))}
                        onChange={(_event, newValue) =>
                            setFormData({ ...formData, entityIds: newValue.map((e) => e.id) })
                        }
                        renderTags={(value, getTagProps) =>
                            value.map((option, index) => (
                                <Chip
                                    label={option.name}
                                    size="small"
                                    {...getTagProps({ index })}
                                    key={option.id}
                                />
                            ))
                        }
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Linked Entities"
                                placeholder="Select entities involved in this event"
                                sx={{ mt: 2 }}
                            />
                        )}
                    />

                    <TextField
                        label="Source Document (optional)"
                        fullWidth
                        value={formData.sourceDocument}
                        onChange={(e) => setFormData({ ...formData, sourceDocument: e.target.value })}
                        placeholder="Reference to source material"
                        sx={{ mt: 2 }}
                    />

                    <FormControlLabel
                        control={
                            <Switch
                                checked={formData.isPlayerKnown}
                                onChange={(e) =>
                                    setFormData({ ...formData, isPlayerKnown: e.target.checked })
                                }
                            />
                        }
                        label="Players know about this event"
                        sx={{ mt: 2 }}
                    />
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setEditDialogOpen(false);
                            setSelectedEventId(null);
                        }}
                        disabled={updateEvent.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleUpdate}
                        variant="contained"
                        disabled={!formData.description.trim() || updateEvent.isPending}
                    >
                        {updateEvent.isPending ? <CircularProgress size={24} /> : 'Save'}
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Delete Confirmation Dialog */}
            <Dialog
                open={deleteDialogOpen}
                onClose={() => {
                    setDeleteDialogOpen(false);
                    setSelectedEventId(null);
                }}
            >
                <DialogTitle>Delete Timeline Event</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Are you sure you want to delete this timeline event? This action cannot be undone.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setDeleteDialogOpen(false);
                            setSelectedEventId(null);
                        }}
                        disabled={deleteEvent.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleDelete}
                        color="error"
                        variant="contained"
                        disabled={deleteEvent.isPending}
                    >
                        {deleteEvent.isPending ? <CircularProgress size={24} /> : 'Delete'}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
}
