/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

import { useState } from 'react';
import {
    Alert,
    AlertTitle,
    Box,
    Button,
    Chip,
    Divider,
    IconButton,
    LinearProgress,
    Link,
    List,
    ListItem,
    ListItemSecondaryAction,
    ListItemText,
    Paper,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Add as AddIcon,
    Check as CheckIcon,
    Close as CloseIcon,
    Delete as DeleteIcon,
    Psychology as PsychologyIcon,
    Star as StarIcon,
    StarBorder as StarBorderIcon,
} from '@mui/icons-material';
import { Link as RouterLink } from 'react-router-dom';
import type { ChapterEntity, Entity } from '../../types';
import { EntitySelector, entityTypeColors } from '../EntitySelector';

/**
 * Represents an AI-detected entity suggestion from similarity search.
 */
export interface EntitySuggestion {
    /** The matched entity. */
    entity: Entity;
    /** The similarity score (0-1). */
    similarity: number;
    /** The text in the chapter content that matched. */
    text: string;
}

/**
 * Props for the ChapterEntityPanel component.
 */
export interface ChapterEntityPanelProps {
    /** The current campaign ID. */
    campaignId: number;
    /** Entities currently linked to this chapter. */
    linkedEntities: ChapterEntity[];
    /** AI-detected entity suggestions from similarity search. */
    suggestions: EntitySuggestion[];
    /** Entity IDs pending to be linked (before chapter is saved). */
    pendingLinkIds: number[];
    /** Whether the embedding API key is configured. */
    isEmbeddingConfigured: boolean;
    /** Whether AI analysis is currently in progress. */
    isAnalyzing: boolean;
    /** Callback when a suggestion is accepted. */
    onAcceptSuggestion: (entityId: number) => void;
    /** Callback when a suggestion is rejected. */
    onRejectSuggestion: (entityId: number) => void;
    /** Callback when an entity is manually linked. */
    onLinkEntity: (entity: Entity) => void;
    /** Callback when an entity link is removed. */
    onUnlinkEntity: (entityId: number) => void;
    /** Callback when an entity is set as featured. */
    onSetFeatured: (entityId: number) => void;
    /** Callback to create a new entity. */
    onCreateNewEntity: () => void;
    /** Callback to trigger AI analysis of chapter content. */
    onAnalyze: () => void;
}

/**
 * Formats a similarity score as a percentage.
 */
function formatSimilarity(score: number): string {
    return `${Math.round(score * 100)}%`;
}

/**
 * Formats an entity type for display.
 */
function formatEntityType(type: string): string {
    return type.charAt(0).toUpperCase() + type.slice(1);
}

/**
 * Formats a mention type for display.
 */
function formatMentionType(type: string): string {
    return type.charAt(0).toUpperCase() + type.slice(1);
}

/**
 * The ChapterEntityPanel displays and manages entity associations for a chapter.
 *
 * This component appears as the right-side panel in the chapter editor and
 * provides three main sections:
 *
 * 1. AI Suggestions - Entities detected via similarity search with accept/reject
 *    actions. This section is disabled with a warning if the embedding API key
 *    is not configured.
 *
 * 2. Linked Entities - Currently associated entities with remove and set-featured
 *    actions.
 *
 * 3. Add Entity - Manual entity search and creation capabilities.
 *
 * @param props - The component props.
 * @returns A React element containing the entity panel.
 */
export default function ChapterEntityPanel({
    campaignId,
    linkedEntities,
    suggestions,
    pendingLinkIds,
    isEmbeddingConfigured,
    isAnalyzing,
    onAcceptSuggestion,
    onRejectSuggestion,
    onLinkEntity,
    onUnlinkEntity,
    onSetFeatured,
    onCreateNewEntity,
    onAnalyze,
}: ChapterEntityPanelProps) {
    const [selectedEntity, setSelectedEntity] = useState<Entity | null>(null);

    // Get IDs of all linked and pending entities to exclude from search
    const linkedEntityIds = linkedEntities.map((ce) => ce.entityId);
    const excludeIds = [...linkedEntityIds, ...pendingLinkIds];

    // Handle entity selection from the autocomplete
    const handleEntitySelect = (entity: Entity | null) => {
        if (entity) {
            onLinkEntity(entity);
            setSelectedEntity(null);
        } else {
            setSelectedEntity(entity);
        }
    };

    return (
        <Paper
            elevation={0}
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                borderLeft: 1,
                borderColor: 'divider',
            }}
        >
            {/* AI Suggestions Section */}
            <Box sx={{ p: 2 }}>
                <Box
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        mb: 1,
                    }}
                >
                    <Typography variant="subtitle1" fontWeight="medium">
                        AI Suggestions
                    </Typography>
                    <Tooltip title="Analyze chapter content">
                        <span>
                            <IconButton
                                size="small"
                                onClick={onAnalyze}
                                disabled={!isEmbeddingConfigured || isAnalyzing}
                            >
                                <PsychologyIcon />
                            </IconButton>
                        </span>
                    </Tooltip>
                </Box>

                {isAnalyzing && <LinearProgress sx={{ mb: 1 }} />}

                {!isEmbeddingConfigured && (
                    <Alert severity="warning" sx={{ mb: 1 }}>
                        <AlertTitle>Embedding API Not Configured</AlertTitle>
                        AI entity detection requires an embedding service.
                        Configure Ollama (local) or a cloud provider.
                        <Link
                            component={RouterLink}
                            to="/settings"
                            sx={{ display: 'block', mt: 0.5 }}
                        >
                            Configure in Settings
                        </Link>
                    </Alert>
                )}

                {isEmbeddingConfigured && suggestions.length === 0 && !isAnalyzing && (
                    <Typography variant="body2" color="text.secondary">
                        No suggestions. Click the analyze button to detect entities.
                    </Typography>
                )}

                {suggestions.length > 0 && (
                    <List dense disablePadding>
                        {suggestions.map((suggestion) => (
                            <ListItem
                                key={suggestion.entity.id}
                                sx={{
                                    bgcolor: 'action.hover',
                                    borderRadius: 1,
                                    mb: 0.5,
                                }}
                            >
                                <ListItemText
                                    primary={
                                        <Box
                                            sx={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 1,
                                            }}
                                        >
                                            <Typography variant="body2">
                                                {suggestion.entity.name}
                                            </Typography>
                                            <Chip
                                                label={formatEntityType(suggestion.entity.entityType)}
                                                size="small"
                                                color={entityTypeColors[suggestion.entity.entityType]}
                                                variant="outlined"
                                                sx={{ height: 20, fontSize: '0.7rem' }}
                                            />
                                        </Box>
                                    }
                                    secondary={
                                        <Typography
                                            variant="caption"
                                            color="text.secondary"
                                            component="span"
                                        >
                                            {formatSimilarity(suggestion.similarity)} match
                                            {suggestion.text && ` - "${suggestion.text}"`}
                                        </Typography>
                                    }
                                />
                                <ListItemSecondaryAction>
                                    <Tooltip title="Accept suggestion">
                                        <IconButton
                                            size="small"
                                            color="success"
                                            onClick={() => onAcceptSuggestion(suggestion.entity.id)}
                                        >
                                            <CheckIcon fontSize="small" />
                                        </IconButton>
                                    </Tooltip>
                                    <Tooltip title="Reject suggestion">
                                        <IconButton
                                            size="small"
                                            color="error"
                                            onClick={() => onRejectSuggestion(suggestion.entity.id)}
                                        >
                                            <CloseIcon fontSize="small" />
                                        </IconButton>
                                    </Tooltip>
                                </ListItemSecondaryAction>
                            </ListItem>
                        ))}
                    </List>
                )}
            </Box>

            <Divider />

            {/* Linked Entities Section */}
            <Box sx={{ p: 2, flexGrow: 1, overflow: 'auto' }}>
                <Typography variant="subtitle1" fontWeight="medium" sx={{ mb: 1 }}>
                    Linked Entities ({linkedEntities.length})
                </Typography>

                {linkedEntities.length === 0 ? (
                    <Typography variant="body2" color="text.secondary">
                        No entities linked to this chapter.
                    </Typography>
                ) : (
                    <List dense disablePadding>
                        {linkedEntities.map((chapterEntity) => {
                            const entity = chapterEntity.entity;
                            if (!entity) return null;

                            const isFeatured = chapterEntity.mentionType === 'featured';

                            return (
                                <ListItem
                                    key={chapterEntity.id}
                                    sx={{
                                        bgcolor: isFeatured ? 'action.selected' : 'transparent',
                                        borderRadius: 1,
                                        mb: 0.5,
                                    }}
                                >
                                    <ListItemText
                                        primary={
                                            <Box
                                                sx={{
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: 1,
                                                }}
                                            >
                                                <Typography variant="body2">
                                                    {entity.name}
                                                </Typography>
                                                <Chip
                                                    label={formatEntityType(entity.entityType)}
                                                    size="small"
                                                    color={entityTypeColors[entity.entityType]}
                                                    variant="outlined"
                                                    sx={{ height: 20, fontSize: '0.7rem' }}
                                                />
                                            </Box>
                                        }
                                        secondary={
                                            <Chip
                                                label={formatMentionType(chapterEntity.mentionType)}
                                                size="small"
                                                variant="filled"
                                                sx={{
                                                    height: 18,
                                                    fontSize: '0.65rem',
                                                    mt: 0.5,
                                                }}
                                            />
                                        }
                                    />
                                    <ListItemSecondaryAction>
                                        <Tooltip
                                            title={isFeatured ? 'Featured entity' : 'Set as featured'}
                                        >
                                            <IconButton
                                                size="small"
                                                color={isFeatured ? 'warning' : 'default'}
                                                onClick={() => onSetFeatured(chapterEntity.entityId)}
                                            >
                                                {isFeatured ? (
                                                    <StarIcon fontSize="small" />
                                                ) : (
                                                    <StarBorderIcon fontSize="small" />
                                                )}
                                            </IconButton>
                                        </Tooltip>
                                        <Tooltip title="Remove link">
                                            <IconButton
                                                size="small"
                                                color="error"
                                                onClick={() => onUnlinkEntity(chapterEntity.entityId)}
                                            >
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </Tooltip>
                                    </ListItemSecondaryAction>
                                </ListItem>
                            );
                        })}
                    </List>
                )}
            </Box>

            <Divider />

            {/* Add Entity Section */}
            <Box sx={{ p: 2 }}>
                <Typography variant="subtitle1" fontWeight="medium" sx={{ mb: 1 }}>
                    Add Entity
                </Typography>

                <EntitySelector
                    campaignId={campaignId}
                    value={selectedEntity}
                    onChange={handleEntitySelect}
                    excludeIds={excludeIds}
                    label=""
                    placeholder="Search entities..."
                />

                <Button
                    variant="outlined"
                    startIcon={<AddIcon />}
                    fullWidth
                    sx={{ mt: 1 }}
                    onClick={onCreateNewEntity}
                >
                    Create New Entity
                </Button>
            </Box>
        </Paper>
    );
}
