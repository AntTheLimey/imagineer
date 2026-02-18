// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * PlayEntitySidebar - A collapsible left-edge sidebar displaying entities
 * grouped by type for the Play mode layout.
 *
 * In its collapsed state the sidebar renders a narrow icon strip showing
 * one badge-decorated icon per entity type that has members. Expanding the
 * sidebar reveals a grouped entity list where each type header is followed
 * by clickable entity rows.
 *
 * When `sceneEntityIds` is non-empty, only entities whose IDs appear in
 * that array are shown; otherwise all entities are displayed.
 */

import React, { useMemo } from 'react';
import {
    Badge,
    Box,
    IconButton,
    List,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Tooltip,
    Typography,
} from '@mui/material';
import {
    Person as NpcIcon,
    Place as LocationIcon,
    Inventory as ItemIcon,
    Groups as FactionIcon,
    Search as ClueIcon,
    Pets as CreatureIcon,
    Business as OrganizationIcon,
    Event as EventIcon,
    Description as DocumentIcon,
    Category as OtherIcon,
    ChevronRight as ExpandIcon,
    ChevronLeft as CollapseIcon,
} from '@mui/icons-material';
import type { Entity, EntityType } from '../../types';
import { useDraftIndicators } from '../../hooks/useDraftIndicators';
import { DraftIndicator } from '../DraftIndicator';

/**
 * Maps each entity type to its display icon and human-readable label.
 */
const ENTITY_TYPE_CONFIG: Record<EntityType, { icon: React.ReactElement; label: string }> = {
    npc: { icon: <NpcIcon fontSize="small" />, label: 'NPCs' },
    location: { icon: <LocationIcon fontSize="small" />, label: 'Locations' },
    item: { icon: <ItemIcon fontSize="small" />, label: 'Items' },
    faction: { icon: <FactionIcon fontSize="small" />, label: 'Factions' },
    clue: { icon: <ClueIcon fontSize="small" />, label: 'Clues' },
    creature: { icon: <CreatureIcon fontSize="small" />, label: 'Creatures' },
    organization: { icon: <OrganizationIcon fontSize="small" />, label: 'Organizations' },
    event: { icon: <EventIcon fontSize="small" />, label: 'Events' },
    document: { icon: <DocumentIcon fontSize="small" />, label: 'Documents' },
    other: { icon: <OtherIcon fontSize="small" />, label: 'Other' },
};

/**
 * Canonical display order for entity type groups.
 */
const TYPE_ORDER: EntityType[] = [
    'npc', 'location', 'faction', 'item', 'event',
    'clue', 'creature', 'organization', 'document', 'other',
];

/**
 * Props for the PlayEntitySidebar component.
 */
export interface PlayEntitySidebarProps {
    /** The campaign ID for fetching draft indicators. */
    campaignId: number;
    /** Entity IDs associated with the active scene; empty means show all. */
    sceneEntityIds: number[];
    /** The full set of entities available in the campaign. */
    allEntities: Entity[];
    /** Callback fired when an entity row is clicked. */
    onEntitySelect: (entity: Entity) => void;
    /** Whether the sidebar is in its narrow collapsed state. */
    collapsed: boolean;
    /** Callback to toggle between collapsed and expanded states. */
    onToggle: () => void;
}

/**
 * A collapsible left-edge sidebar that displays campaign entities grouped
 * by type. In its collapsed state it shows a vertical icon strip with
 * badge counts; when expanded it shows a full grouped entity list.
 *
 * @param props - The component props.
 * @returns A React element containing the sidebar.
 *
 * @example
 * ```tsx
 * <PlayEntitySidebar
 *     campaignId={42}
 *     sceneEntityIds={[10, 20, 30]}
 *     allEntities={entities}
 *     onEntitySelect={(e) => openEntityDetail(e)}
 *     collapsed={isSidebarCollapsed}
 *     onToggle={() => setIsSidebarCollapsed((v) => !v)}
 * />
 * ```
 */
export function PlayEntitySidebar(props: PlayEntitySidebarProps) {
    const {
        campaignId,
        sceneEntityIds,
        allEntities,
        onEntitySelect,
        collapsed,
        onToggle,
    } = props;

    const { hasDraft } = useDraftIndicators(campaignId, 'entities');
    /**
     * Filter entities to the active scene when scene entity IDs are
     * provided; otherwise show the full campaign entity set.
     */
    const filteredEntities = useMemo(() => {
        if (sceneEntityIds.length > 0) {
            const idSet = new Set(sceneEntityIds);
            return allEntities.filter(e => idSet.has(e.id));
        }
        return allEntities;
    }, [allEntities, sceneEntityIds]);

    /**
     * Group the filtered entities by their entity type for display in
     * categorised sections.
     */
    const groupedEntities = useMemo(() => {
        const groups: Partial<Record<EntityType, Entity[]>> = {};
        for (const entity of filteredEntities) {
            if (!groups[entity.entityType]) groups[entity.entityType] = [];
            groups[entity.entityType]!.push(entity);
        }
        return groups;
    }, [filteredEntities]);

    return (
        <Box
            sx={{
                width: collapsed ? 40 : 220,
                height: '100%',
                overflow: 'auto',
                borderRight: 1,
                borderColor: 'divider',
                transition: 'width 0.2s ease',
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            {collapsed ? (
                /* ------- Collapsed state: icon strip ------- */
                <Box
                    sx={{
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        pt: 0.5,
                    }}
                >
                    <Tooltip title="Expand entity sidebar" placement="right">
                        <IconButton
                            size="small"
                            onClick={onToggle}
                            aria-label="expand entity sidebar"
                        >
                            <ExpandIcon fontSize="small" />
                        </IconButton>
                    </Tooltip>

                    {TYPE_ORDER.map((type) => {
                        const entities = groupedEntities[type];
                        if (!entities || entities.length === 0) return null;

                        const config = ENTITY_TYPE_CONFIG[type];
                        return (
                            <Tooltip
                                key={type}
                                title={config.label}
                                placement="right"
                            >
                                <IconButton
                                    size="small"
                                    onClick={onToggle}
                                    aria-label={`${config.label} (${entities.length})`}
                                    sx={{ my: 0.25 }}
                                >
                                    <Badge
                                        badgeContent={entities.length}
                                        color="primary"
                                        max={99}
                                    >
                                        {config.icon}
                                    </Badge>
                                </IconButton>
                            </Tooltip>
                        );
                    })}
                </Box>
            ) : (
                /* ------- Expanded state: grouped entity list ------- */
                <Box sx={{ display: 'flex', flexDirection: 'column' }}>
                    <Box
                        sx={{
                            display: 'flex',
                            justifyContent: 'flex-end',
                            pt: 0.5,
                            pr: 0.5,
                        }}
                    >
                        <Tooltip title="Collapse entity sidebar" placement="right">
                            <IconButton
                                size="small"
                                onClick={onToggle}
                                aria-label="collapse entity sidebar"
                            >
                                <CollapseIcon fontSize="small" />
                            </IconButton>
                        </Tooltip>
                    </Box>

                    {TYPE_ORDER.map((type) => {
                        const entities = groupedEntities[type];
                        if (!entities || entities.length === 0) return null;

                        const config = ENTITY_TYPE_CONFIG[type];
                        return (
                            <Box key={type}>
                                <Box
                                    sx={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 0.5,
                                        px: 1,
                                        pt: 1,
                                    }}
                                >
                                    {config.icon}
                                    <Typography variant="overline">
                                        {config.label}
                                    </Typography>
                                </Box>

                                <List dense disablePadding>
                                    {entities.map((entity) => (
                                        <ListItemButton
                                            key={entity.id}
                                            onClick={() => onEntitySelect(entity)}
                                            sx={{ pl: 2 }}
                                        >
                                            <ListItemIcon sx={{ minWidth: 28 }}>
                                                {config.icon}
                                            </ListItemIcon>
                                            <ListItemText
                                                primary={entity.name}
                                                primaryTypographyProps={{
                                                    variant: 'body2',
                                                    noWrap: true,
                                                }}
                                            />
                                            <DraftIndicator hasDraft={hasDraft(entity.id)} />
                                        </ListItemButton>
                                    ))}
                                </List>
                            </Box>
                        );
                    })}
                </Box>
            )}
        </Box>
    );
}

export default PlayEntitySidebar;
