// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useState, useCallback, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
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
    FormControl,
    IconButton,
    InputLabel,
    MenuItem,
    Paper,
    Select,
    Skeleton,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TablePagination,
    TableRow,
    TextField,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Delete as DeleteIcon,
    Edit as EditIcon,
    Visibility as ViewIcon,
} from '@mui/icons-material';
import { EntityPreviewPanel } from '../components/EntityPreviewPanel';
import {
    useEntities,
    useEntity,
    useCreateEntity,
    useUpdateEntity,
    useDeleteEntity,
    useSimilarEntities,
    useCampaignOwnership,
} from '../hooks';
import { sanitizeHtml, stripHtml } from '../utils';
import type { Entity, EntityType, SourceConfidence } from '../types';

/**
 * All available entity types.
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
 * Human-readable labels for entity types.
 */
const ENTITY_TYPE_LABELS: Record<EntityType, string> = {
    npc: 'NPC',
    location: 'Location',
    item: 'Item',
    faction: 'Faction',
    clue: 'Clue',
    creature: 'Creature',
    organization: 'Organization',
    event: 'Event',
    document: 'Document',
    other: 'Other',
};

/**
 * Source confidence options.
 */
const SOURCE_CONFIDENCE_OPTIONS: SourceConfidence[] = [
    'DRAFT',
    'AUTHORITATIVE',
    'SUPERSEDED',
];

/**
 * Form data structure for create/edit dialogs.
 */
interface EntityFormData {
    name: string;
    entityType: EntityType;
    description: string;
    tags: string[];
    attributes: string;
    gmNotes: string;
    sourceConfidence: SourceConfidence;
}

/**
 * Default form values.
 */
const DEFAULT_FORM_DATA: EntityFormData = {
    name: '',
    entityType: 'npc',
    description: '',
    tags: [],
    attributes: '{}',
    gmNotes: '',
    sourceConfidence: 'DRAFT',
};

/**
 * Validates JSON string.
 */
function isValidJson(str: string): boolean {
    try {
        JSON.parse(str);
        return true;
    } catch {
        return false;
    }
}

/**
 * Formats a date string for display.
 */
function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    });
}

/**
 * Render the Entities management UI for a campaign, providing listing, filtering, and CRUD interactions for campaign entities.
 *
 * Creation and editing navigate to the full-screen entity editor routes; viewing and deletion are handled via in-page dialogs. Access to GM-only fields is gated by campaign ownership.
 *
 * @returns The rendered Entities page as a JSX element
 */
/**
 * Valid entity types for URL parameter validation.
 */
const VALID_ENTITY_TYPES: Set<string> = new Set(ENTITY_TYPES);

/**
 * Parse entity type from URL search params.
 */
function parseTypeParam(typeParam: string | null): EntityType | 'all' {
    if (!typeParam) return 'all';
    if (VALID_ENTITY_TYPES.has(typeParam)) {
        return typeParam as EntityType;
    }
    return 'all';
}

export default function Entities() {
    const { id: campaignId } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    // Check if current user is the campaign owner (GM)
    const { isOwner: isGM } = useCampaignOwnership(campaignId ?? '');

    // Pagination state
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);

    // Get type filter from URL query params
    const typeFromUrl = parseTypeParam(searchParams.get('type'));
    const [typeFilter, setTypeFilter] = useState<EntityType | 'all'>(typeFromUrl);

    // Sync typeFilter with URL changes
    useEffect(() => {
        const newType = parseTypeParam(searchParams.get('type'));
        if (newType !== typeFilter) {
            setTypeFilter(newType);
            setPage(0);
            setSelectedEntity(null);
        }
    }, [searchParams]); // eslint-disable-line react-hooks/exhaustive-deps

    // Selected entity for preview
    const [selectedEntity, setSelectedEntity] = useState<Entity | null>(null);

    // Dialog states
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [viewDialogOpen, setViewDialogOpen] = useState(false);
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

    // Selected entity for view/edit/delete dialogs
    const [selectedEntityId, setSelectedEntityId] = useState<string | null>(null);

    // Form state
    const [formData, setFormData] = useState<EntityFormData>(DEFAULT_FORM_DATA);
    const [formErrors, setFormErrors] = useState<Partial<Record<keyof EntityFormData, string>>>({});

    // Fetch entities
    const {
        data: entities,
        isLoading: entitiesLoading,
        error: entitiesError,
    } = useEntities({
        campaignId: campaignId ?? '',
        page: page + 1,
        pageSize: rowsPerPage,
        entityType: typeFilter === 'all' ? undefined : typeFilter,
    });

    // Fetch single entity for view/edit
    const {
        data: dialogEntity,
        isLoading: dialogEntityLoading,
    } = useEntity(
        campaignId ?? '',
        selectedEntityId ?? '',
        { enabled: !!selectedEntityId && (viewDialogOpen || editDialogOpen) }
    );

    // Similar entities search for duplicate detection
    const {
        data: similarEntities,
    } = useSimilarEntities(
        campaignId ?? '',
        formData.name,
        { enabled: createDialogOpen && formData.name.length >= 2 }
    );

    // Mutations
    const createEntity = useCreateEntity();
    const updateEntity = useUpdateEntity();
    const deleteEntity = useDeleteEntity();

    // Validate form
    const validateForm = useCallback((): boolean => {
        const errors: Partial<Record<keyof EntityFormData, string>> = {};

        if (!formData.name.trim()) {
            errors.name = 'Name is required';
        }

        if (!formData.entityType) {
            errors.entityType = 'Type is required';
        }

        if (!isValidJson(formData.attributes)) {
            errors.attributes = 'Invalid JSON format';
        }

        setFormErrors(errors);
        return Object.keys(errors).length === 0;
    }, [formData]);

    // Handle create
    const handleCreate = async () => {
        if (!validateForm() || !campaignId) return;

        try {
            await createEntity.mutateAsync({
                campaignId,
                name: formData.name.trim(),
                entityType: formData.entityType,
                description: formData.description.trim() || undefined,
                tags: formData.tags.length > 0 ? formData.tags : undefined,
                attributes: JSON.parse(formData.attributes),
                gmNotes: formData.gmNotes.trim() || undefined,
                sourceConfidence: formData.sourceConfidence,
            });
            setCreateDialogOpen(false);
            setFormData(DEFAULT_FORM_DATA);
            setFormErrors({});
        } catch (error) {
            console.error('Failed to create entity:', error);
        }
    };

    // Handle update
    const handleUpdate = async () => {
        if (!validateForm() || !campaignId || !selectedEntityId) return;

        try {
            await updateEntity.mutateAsync({
                campaignId,
                entityId: selectedEntityId,
                input: {
                    name: formData.name.trim(),
                    description: formData.description.trim() || undefined,
                    tags: formData.tags,
                    attributes: JSON.parse(formData.attributes),
                    gmNotes: formData.gmNotes.trim() || undefined,
                    sourceConfidence: formData.sourceConfidence,
                },
            });
            setEditDialogOpen(false);
            setSelectedEntityId(null);
            setFormData(DEFAULT_FORM_DATA);
            setFormErrors({});
        } catch (error) {
            console.error('Failed to update entity:', error);
        }
    };

    // Handle delete
    const handleDelete = async () => {
        if (!campaignId || !selectedEntityId) return;

        try {
            await deleteEntity.mutateAsync({
                campaignId,
                entityId: selectedEntityId,
            });
            setDeleteDialogOpen(false);
            setSelectedEntityId(null);
            // Clear preview if deleted entity was selected
            if (selectedEntity?.id === selectedEntityId) {
                setSelectedEntity(null);
            }
        } catch (error) {
            console.error('Failed to delete entity:', error);
        }
    };

    // Open view dialog
    const openViewDialog = (entity: Entity) => {
        setSelectedEntityId(entity.id);
        setViewDialogOpen(true);
    };

    // Open delete dialog
    const openDeleteDialog = (entity: Entity) => {
        setSelectedEntityId(entity.id);
        setDeleteDialogOpen(true);
    };

    // Handle preview panel delete
    const handlePreviewDelete = () => {
        if (selectedEntity) {
            openDeleteDialog(selectedEntity);
        }
    };

    // Handle preview panel edit
    const handlePreviewEdit = () => {
        if (selectedEntity && campaignId) {
            navigate(`/campaigns/${campaignId}/entities/${selectedEntity.id}/edit`);
        }
    };

    // Navigate to full-screen entity editor for creating
    const openCreateDialog = () => {
        if (!campaignId) {
            return;
        }
        // Navigate to full-screen entity editor
        navigate(`/campaigns/${campaignId}/entities/new`);
    };

    // Navigate to full-screen entity editor for editing
    const navigateToEditor = (entityId: string) => {
        if (!campaignId) {
            return;
        }
        navigate(`/campaigns/${campaignId}/entities/${entityId}/edit`);
    };

    // Handle page change
    const handleChangePage = (_event: unknown, newPage: number) => {
        setPage(newPage);
    };

    // Handle rows per page change
    const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    // Handle row click - select for preview
    const handleRowClick = (entity: Entity) => {
        setSelectedEntity(entity);
    };

    // Loading state
    if (entitiesLoading && !entities) {
        return (
            <Box sx={{ display: 'flex', gap: 3, height: '100%' }}>
                {/* Main content area */}
                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                        <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                            Entities
                        </Typography>
                        <Skeleton variant="rectangular" width={130} height={36} />
                    </Box>
                    <TableContainer component={Paper}>
                        <Table>
                            <TableHead>
                                <TableRow>
                                    <TableCell><Skeleton width={80} /></TableCell>
                                    <TableCell><Skeleton width={60} /></TableCell>
                                    <TableCell><Skeleton width={200} /></TableCell>
                                    <TableCell><Skeleton width={100} /></TableCell>
                                    <TableCell><Skeleton width={80} /></TableCell>
                                    <TableCell><Skeleton width={100} /></TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {[1, 2, 3, 4, 5].map((i) => (
                                    <TableRow key={i}>
                                        <TableCell><Skeleton /></TableCell>
                                        <TableCell><Skeleton width={60} /></TableCell>
                                        <TableCell><Skeleton /></TableCell>
                                        <TableCell><Skeleton /></TableCell>
                                        <TableCell><Skeleton width={80} /></TableCell>
                                        <TableCell><Skeleton width={100} /></TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </Box>

                {/* Preview panel */}
                <Paper
                    sx={{
                        width: 320,
                        flexShrink: 0,
                        p: 2,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }}
                >
                    <Typography variant="body1" color="text.secondary">
                        Select an entity to preview
                    </Typography>
                </Paper>
            </Box>
        );
    }

    const entityList = entities ?? [];

    // Main content - the entity table
    const mainContent = (
        <Box>
            {/* Header */}
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                <Typography variant="h4" sx={{ fontFamily: 'Cinzel' }}>
                    Entities
                </Typography>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={openCreateDialog}
                >
                    New Entity
                </Button>
            </Box>

            {/* Error alerts */}
            {entitiesError && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to load entities. Please try again later.
                </Alert>
            )}

            {createEntity.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to create entity. Please try again.
                </Alert>
            )}

            {updateEntity.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to update entity. Please try again.
                </Alert>
            )}

            {deleteEntity.error && (
                <Alert severity="error" sx={{ mb: 2 }}>
                    Failed to delete entity. Please try again.
                </Alert>
            )}

            {/* Entity Table */}
            {entityList.length === 0 ? (
                <Paper sx={{ p: 4, textAlign: 'center' }}>
                    <Typography variant="h6" color="text.secondary" gutterBottom>
                        No entities found
                    </Typography>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                        {typeFilter !== 'all'
                            ? `No ${ENTITY_TYPE_LABELS[typeFilter]} entities in this campaign.`
                            : 'Create your first entity to get started.'}
                    </Typography>
                    <Button variant="outlined" onClick={openCreateDialog}>
                        Create Entity
                    </Button>
                </Paper>
            ) : (
                <Paper>
                    <TableContainer>
                        <Table>
                            <TableHead>
                                <TableRow>
                                    <TableCell>Name</TableCell>
                                    <TableCell>Type</TableCell>
                                    <TableCell>Description</TableCell>
                                    <TableCell>Tags</TableCell>
                                    <TableCell>Created</TableCell>
                                    <TableCell align="right">Actions</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {entityList.map((entity) => (
                                    <TableRow
                                        key={entity.id}
                                        hover
                                        selected={selectedEntity?.id === entity.id}
                                        sx={{
                                            cursor: 'pointer',
                                            '&.Mui-selected': {
                                                backgroundColor: 'action.selected',
                                            },
                                            '&.Mui-selected:hover': {
                                                backgroundColor: 'action.selected',
                                            },
                                        }}
                                        onClick={() => handleRowClick(entity)}
                                    >
                                        <TableCell>
                                            <Typography variant="body2" fontWeight="medium">
                                                {entity.name}
                                            </Typography>
                                        </TableCell>
                                        <TableCell>
                                            <Chip
                                                label={ENTITY_TYPE_LABELS[entity.entityType]}
                                                size="small"
                                                variant="outlined"
                                            />
                                        </TableCell>
                                        <TableCell>
                                            <Typography
                                                variant="body2"
                                                color="text.secondary"
                                                sx={{
                                                    maxWidth: 300,
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    whiteSpace: 'nowrap',
                                                }}
                                            >
                                                {entity.description ? stripHtml(entity.description) : '-'}
                                            </Typography>
                                        </TableCell>
                                        <TableCell>
                                            <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                                                {entity.tags?.slice(0, 3).map((tag) => (
                                                    <Chip key={tag} label={tag} size="small" />
                                                ))}
                                                {entity.tags && entity.tags.length > 3 && (
                                                    <Chip
                                                        label={`+${entity.tags.length - 3}`}
                                                        size="small"
                                                        variant="outlined"
                                                    />
                                                )}
                                            </Box>
                                        </TableCell>
                                        <TableCell>
                                            <Typography variant="body2" color="text.secondary">
                                                {formatDate(entity.createdAt)}
                                            </Typography>
                                        </TableCell>
                                        <TableCell align="right">
                                            <Tooltip title="View">
                                                <IconButton
                                                    size="small"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        openViewDialog(entity);
                                                    }}
                                                >
                                                    <ViewIcon fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                            <Tooltip title="Edit">
                                                <IconButton
                                                    size="small"
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        navigateToEditor(entity.id);
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
                                                        openDeleteDialog(entity);
                                                    }}
                                                >
                                                    <DeleteIcon fontSize="small" />
                                                </IconButton>
                                            </Tooltip>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                    <TablePagination
                        component="div"
                        count={-1}
                        rowsPerPage={rowsPerPage}
                        page={page}
                        onPageChange={handleChangePage}
                        onRowsPerPageChange={handleChangeRowsPerPage}
                        rowsPerPageOptions={[5, 10, 25, 50]}
                        labelDisplayedRows={({ from, to }) => `${from}-${to}`}
                    />
                </Paper>
            )}
        </Box>
    );

    // Preview panel content
    const previewPanel = campaignId ? (
        <EntityPreviewPanel
            entity={selectedEntity}
            campaignId={campaignId}
            onEdit={handlePreviewEdit}
            onDelete={handlePreviewDelete}
        />
    ) : null;

    return (
        <>
            <Box sx={{ display: 'flex', gap: 3, height: '100%' }}>
                {/* Main content area */}
                <Box sx={{ flex: 1, minWidth: 0 }}>
                    {mainContent}
                </Box>

                {/* Preview panel */}
                <Paper
                    sx={{
                        width: 320,
                        flexShrink: 0,
                        p: 2,
                        overflow: 'auto',
                    }}
                >
                    {previewPanel}
                </Paper>
            </Box>

            {/* Create Entity Dialog */}
            <Dialog
                open={createDialogOpen}
                onClose={() => setCreateDialogOpen(false)}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>Create New Entity</DialogTitle>
                <DialogContent>
                    {/* Similar entities warning */}
                    {similarEntities && similarEntities.length > 0 && (
                        <Alert severity="warning" sx={{ mb: 2, mt: 1 }}>
                            Similar entities found: {similarEntities.map((e) => e.name).join(', ')}.
                            Please verify this is not a duplicate.
                        </Alert>
                    )}

                    <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                        <TextField
                            autoFocus
                            label="Name"
                            fullWidth
                            required
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            error={!!formErrors.name}
                            helperText={formErrors.name}
                        />
                        <FormControl sx={{ minWidth: 200 }} error={!!formErrors.entityType}>
                            <InputLabel required>Type</InputLabel>
                            <Select
                                value={formData.entityType}
                                label="Type"
                                onChange={(e) =>
                                    setFormData({ ...formData, entityType: e.target.value as EntityType })
                                }
                            >
                                {ENTITY_TYPES.map((type) => (
                                    <MenuItem key={type} value={type}>
                                        {ENTITY_TYPE_LABELS[type]}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Box>

                    <TextField
                        label="Description"
                        fullWidth
                        multiline
                        rows={3}
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        sx={{ mt: 2 }}
                    />

                    <Autocomplete
                        multiple
                        freeSolo
                        autoSelect
                        options={[]}
                        value={formData.tags}
                        onChange={(_event, newValue) =>
                            setFormData({ ...formData, tags: newValue as string[] })
                        }
                        onBlur={(event) => {
                            const inputValue = (event.target as HTMLInputElement).value?.trim();
                            if (inputValue && !formData.tags.includes(inputValue)) {
                                setFormData({ ...formData, tags: [...formData.tags, inputValue] });
                            }
                        }}
                        renderTags={(value, getTagProps) =>
                            value.map((option, index) => (
                                <Chip
                                    label={option}
                                    size="small"
                                    {...getTagProps({ index })}
                                    key={option}
                                />
                            ))
                        }
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Tags"
                                placeholder="Type and press Enter to add tags"
                                sx={{ mt: 2 }}
                            />
                        )}
                    />

                    <TextField
                        label="Attributes (JSON)"
                        fullWidth
                        multiline
                        rows={4}
                        value={formData.attributes}
                        onChange={(e) => setFormData({ ...formData, attributes: e.target.value })}
                        error={!!formErrors.attributes}
                        helperText={formErrors.attributes ?? 'Game system specific attributes in JSON format'}
                        sx={{ mt: 2 }}
                        InputProps={{
                            sx: { fontFamily: 'monospace', fontSize: '0.875rem' },
                        }}
                    />

                    {isGM && (
                        <TextField
                            label="GM Notes"
                            fullWidth
                            multiline
                            rows={3}
                            value={formData.gmNotes}
                            onChange={(e) => setFormData({ ...formData, gmNotes: e.target.value })}
                            sx={{ mt: 2 }}
                            helperText="Private notes visible only to the Game Master"
                        />
                    )}

                    <FormControl fullWidth sx={{ mt: 2 }}>
                        <InputLabel>Source Confidence</InputLabel>
                        <Select
                            value={formData.sourceConfidence}
                            label="Source Confidence"
                            onChange={(e) =>
                                setFormData({
                                    ...formData,
                                    sourceConfidence: e.target.value as SourceConfidence,
                                })
                            }
                        >
                            {SOURCE_CONFIDENCE_OPTIONS.map((opt) => (
                                <MenuItem key={opt} value={opt}>
                                    {opt}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => setCreateDialogOpen(false)}
                        disabled={createEntity.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleCreate}
                        variant="contained"
                        disabled={!formData.name.trim() || createEntity.isPending}
                    >
                        {createEntity.isPending ? <CircularProgress size={24} /> : 'Create'}
                    </Button>
                </DialogActions>
            </Dialog>

            {/* View Entity Dialog */}
            <Dialog
                open={viewDialogOpen}
                onClose={() => {
                    setViewDialogOpen(false);
                    setSelectedEntityId(null);
                }}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>
                    {dialogEntityLoading ? (
                        <Skeleton width={200} />
                    ) : (
                        dialogEntity?.name ?? 'Entity Details'
                    )}
                </DialogTitle>
                <DialogContent>
                    {dialogEntityLoading ? (
                        <Box>
                            <Skeleton height={40} />
                            <Skeleton height={100} />
                            <Skeleton height={40} />
                            <Skeleton height={40} />
                        </Box>
                    ) : dialogEntity ? (
                        <Box sx={{ mt: 1 }}>
                            <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                                <Chip
                                    label={ENTITY_TYPE_LABELS[dialogEntity.entityType]}
                                    color="primary"
                                />
                                <Chip
                                    label={dialogEntity.sourceConfidence}
                                    variant="outlined"
                                    color={
                                        dialogEntity.sourceConfidence === 'AUTHORITATIVE'
                                            ? 'success'
                                            : dialogEntity.sourceConfidence === 'DRAFT'
                                                ? 'warning'
                                                : 'default'
                                    }
                                />
                            </Box>

                            {dialogEntity.description && (
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="subtitle2" color="text.secondary">
                                        Description
                                    </Typography>
                                    <Box
                                        component="div"
                                        sx={{
                                            '& p': { mt: 0, mb: 1 },
                                            '& p:last-child': { mb: 0 },
                                        }}
                                        dangerouslySetInnerHTML={{
                                            __html: sanitizeHtml(dialogEntity.description),
                                        }}
                                    />
                                </Box>
                            )}

                            {dialogEntity.tags && dialogEntity.tags.length > 0 && (
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                                        Tags
                                    </Typography>
                                    <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap' }}>
                                        {dialogEntity.tags.map((tag) => (
                                            <Chip key={tag} label={tag} size="small" />
                                        ))}
                                    </Box>
                                </Box>
                            )}

                            {dialogEntity.attributes &&
                                Object.keys(dialogEntity.attributes).length > 0 && (
                                    <Box sx={{ mb: 2 }}>
                                        <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                                            Attributes
                                        </Typography>
                                        <Paper variant="outlined" sx={{ p: 2 }}>
                                            <pre style={{ margin: 0, fontSize: '0.875rem', overflow: 'auto' }}>
                                                {JSON.stringify(dialogEntity.attributes, null, 2)}
                                            </pre>
                                        </Paper>
                                    </Box>
                                )}

                            {isGM && dialogEntity.gmNotes && (
                                <Box sx={{ mb: 2 }}>
                                    <Typography variant="subtitle2" color="text.secondary">
                                        GM Notes
                                    </Typography>
                                    <Paper
                                        variant="outlined"
                                        sx={{ p: 2, bgcolor: 'warning.light', color: 'warning.contrastText' }}
                                    >
                                        <Typography variant="body2">
                                            {dialogEntity.gmNotes}
                                        </Typography>
                                    </Paper>
                                </Box>
                            )}

                            <Box sx={{ display: 'flex', gap: 4, mt: 3 }}>
                                <Box>
                                    <Typography variant="caption" color="text.secondary">
                                        Created
                                    </Typography>
                                    <Typography variant="body2">
                                        {formatDate(dialogEntity.createdAt)}
                                    </Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary">
                                        Updated
                                    </Typography>
                                    <Typography variant="body2">
                                        {formatDate(dialogEntity.updatedAt)}
                                    </Typography>
                                </Box>
                                <Box>
                                    <Typography variant="caption" color="text.secondary">
                                        Version
                                    </Typography>
                                    <Typography variant="body2">
                                        {dialogEntity.version}
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
                            setSelectedEntityId(null);
                        }}
                    >
                        Close
                    </Button>
                    {dialogEntity && (
                        <Button
                            variant="contained"
                            onClick={() => {
                                setViewDialogOpen(false);
                                navigateToEditor(dialogEntity.id);
                            }}
                        >
                            Edit
                        </Button>
                    )}
                </DialogActions>
            </Dialog>

            {/* Edit Entity Dialog */}
            <Dialog
                open={editDialogOpen}
                onClose={() => {
                    setEditDialogOpen(false);
                    setSelectedEntityId(null);
                }}
                maxWidth="md"
                fullWidth
            >
                <DialogTitle>Edit Entity</DialogTitle>
                <DialogContent>
                    <Box sx={{ display: 'flex', gap: 2, mt: 1 }}>
                        <TextField
                            autoFocus
                            label="Name"
                            fullWidth
                            required
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            error={!!formErrors.name}
                            helperText={formErrors.name}
                        />
                        <FormControl sx={{ minWidth: 200 }} disabled>
                            <InputLabel>Type</InputLabel>
                            <Select value={formData.entityType} label="Type">
                                <MenuItem value={formData.entityType}>
                                    {ENTITY_TYPE_LABELS[formData.entityType]}
                                </MenuItem>
                            </Select>
                        </FormControl>
                    </Box>

                    <TextField
                        label="Description"
                        fullWidth
                        multiline
                        rows={3}
                        value={formData.description}
                        onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                        sx={{ mt: 2 }}
                    />

                    <Autocomplete
                        multiple
                        freeSolo
                        autoSelect
                        options={[]}
                        value={formData.tags}
                        onChange={(_event, newValue) =>
                            setFormData({ ...formData, tags: newValue as string[] })
                        }
                        onBlur={(event) => {
                            const inputValue = (event.target as HTMLInputElement).value?.trim();
                            if (inputValue && !formData.tags.includes(inputValue)) {
                                setFormData({ ...formData, tags: [...formData.tags, inputValue] });
                            }
                        }}
                        renderTags={(value, getTagProps) =>
                            value.map((option, index) => (
                                <Chip
                                    label={option}
                                    size="small"
                                    {...getTagProps({ index })}
                                    key={option}
                                />
                            ))
                        }
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Tags"
                                placeholder="Type and press Enter to add tags"
                                sx={{ mt: 2 }}
                            />
                        )}
                    />

                    <TextField
                        label="Attributes (JSON)"
                        fullWidth
                        multiline
                        rows={4}
                        value={formData.attributes}
                        onChange={(e) => setFormData({ ...formData, attributes: e.target.value })}
                        error={!!formErrors.attributes}
                        helperText={formErrors.attributes ?? 'Game system specific attributes in JSON format'}
                        sx={{ mt: 2 }}
                        InputProps={{
                            sx: { fontFamily: 'monospace', fontSize: '0.875rem' },
                        }}
                    />

                    {isGM && (
                        <TextField
                            label="GM Notes"
                            fullWidth
                            multiline
                            rows={3}
                            value={formData.gmNotes}
                            onChange={(e) => setFormData({ ...formData, gmNotes: e.target.value })}
                            sx={{ mt: 2 }}
                            helperText="Private notes visible only to the Game Master"
                        />
                    )}

                    <FormControl fullWidth sx={{ mt: 2 }}>
                        <InputLabel>Source Confidence</InputLabel>
                        <Select
                            value={formData.sourceConfidence}
                            label="Source Confidence"
                            onChange={(e) =>
                                setFormData({
                                    ...formData,
                                    sourceConfidence: e.target.value as SourceConfidence,
                                })
                            }
                        >
                            {SOURCE_CONFIDENCE_OPTIONS.map((opt) => (
                                <MenuItem key={opt} value={opt}>
                                    {opt}
                                </MenuItem>
                            ))}
                        </Select>
                    </FormControl>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setEditDialogOpen(false);
                            setSelectedEntityId(null);
                        }}
                        disabled={updateEntity.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleUpdate}
                        variant="contained"
                        disabled={!formData.name.trim() || updateEntity.isPending}
                    >
                        {updateEntity.isPending ? <CircularProgress size={24} /> : 'Save'}
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Delete Confirmation Dialog */}
            <Dialog
                open={deleteDialogOpen}
                onClose={() => {
                    setDeleteDialogOpen(false);
                    setSelectedEntityId(null);
                }}
            >
                <DialogTitle>Delete Entity</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        Are you sure you want to delete this entity? This action cannot be undone.
                        Any relationships connected to this entity will also be removed.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button
                        onClick={() => {
                            setDeleteDialogOpen(false);
                            setSelectedEntityId(null);
                        }}
                        disabled={deleteEntity.isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={handleDelete}
                        color="error"
                        variant="contained"
                        disabled={deleteEntity.isPending}
                    >
                        {deleteEntity.isPending ? <CircularProgress size={24} /> : 'Delete'}
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
}
