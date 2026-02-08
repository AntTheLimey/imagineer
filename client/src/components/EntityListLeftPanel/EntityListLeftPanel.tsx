// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import { useMemo } from 'react';
import {
    Box,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Skeleton,
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
    List as AllIcon,
} from '@mui/icons-material';
import { useCampaign, useEntities } from '../../hooks';
import type { EntityType } from '../../types';

/**
 * Props for the EntityListLeftPanel component.
 */
export interface EntityListLeftPanelProps {
    /** The campaign ID to display entity filters for. */
    campaignId: number;
    /** The currently selected entity type filter, or 'all' for no filter. */
    selectedType: EntityType | 'all';
    /** Callback fired when a type filter is selected. */
    onTypeChange: (type: EntityType | 'all') => void;
}

/**
 * Display configuration for each entity type.
 */
interface EntityTypeConfig {
    label: string;
    icon: React.ReactNode;
}

/**
 * Entity type display configuration with labels and icons.
 */
const ENTITY_TYPE_CONFIG: Record<EntityType | 'all', EntityTypeConfig> = {
    all: { label: 'All Entities', icon: <AllIcon /> },
    npc: { label: 'NPCs', icon: <NpcIcon /> },
    location: { label: 'Locations', icon: <LocationIcon /> },
    faction: { label: 'Factions', icon: <FactionIcon /> },
    item: { label: 'Items', icon: <ItemIcon /> },
    event: { label: 'Events', icon: <EventIcon /> },
    clue: { label: 'Clues', icon: <ClueIcon /> },
    creature: { label: 'Creatures', icon: <CreatureIcon /> },
    organization: { label: 'Organizations', icon: <OrganizationIcon /> },
    document: { label: 'Documents', icon: <DocumentIcon /> },
    other: { label: 'Other', icon: <OtherIcon /> },
};

/**
 * Order in which entity types are displayed in the list.
 */
const TYPE_ORDER: Array<EntityType | 'all'> = [
    'all',
    'npc',
    'location',
    'faction',
    'item',
    'event',
    'clue',
    'creature',
    'organization',
    'document',
    'other',
];

/**
 * Left panel component for the Entities page.
 *
 * Displays the campaign name and a list of entity type filters with counts.
 * The user can select a type to filter the entity table, or select "All Entities"
 * to show entities of all types.
 *
 * @param props - The component props.
 * @returns A React element containing the left panel navigation.
 */
export default function EntityListLeftPanel({
    campaignId,
    selectedType,
    onTypeChange,
}: EntityListLeftPanelProps) {
    // Fetch campaign details for the header
    const { data: campaign, isLoading: campaignLoading } = useCampaign(campaignId);

    // Fetch all entities (with large page size) to calculate counts.
    // NOTE: This approach has a limitation for campaigns with >1000 entities.
    // A dedicated count endpoint would be more robust for very large campaigns.
    const { data: entities, isLoading: entitiesLoading } = useEntities({
        campaignId,
        pageSize: 1000,
    });

    // Calculate entity counts by type
    const typeCounts = useMemo(() => {
        const counts: Record<EntityType | 'all', number> = {
            all: 0,
            npc: 0,
            location: 0,
            item: 0,
            faction: 0,
            clue: 0,
            creature: 0,
            organization: 0,
            event: 0,
            document: 0,
            other: 0,
        };

        if (!entities) return counts;

        for (const entity of entities) {
            counts[entity.entityType]++;
            counts.all++;
        }

        return counts;
    }, [entities]);

    return (
        <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
            {/* Campaign name header */}
            <Box sx={{ mb: 2 }}>
                {campaignLoading ? (
                    <Skeleton variant="text" width="80%" height={32} />
                ) : (
                    <Typography
                        variant="h6"
                        sx={{
                            fontFamily: 'Cinzel',
                            fontWeight: 600,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        {campaign?.name ?? 'Campaign'}
                    </Typography>
                )}
            </Box>

            {/* Entity type filter list */}
            <List sx={{ flexGrow: 1, overflow: 'auto' }}>
                {TYPE_ORDER.map((type) => {
                    const config = ENTITY_TYPE_CONFIG[type];
                    const count = typeCounts[type];
                    const isSelected = selectedType === type;

                    return (
                        <ListItem key={type} disablePadding>
                            <ListItemButton
                                selected={isSelected}
                                onClick={() => onTypeChange(type)}
                                sx={{
                                    borderRadius: 1,
                                    mb: 0.5,
                                    '&.Mui-selected': {
                                        backgroundColor: 'primary.main',
                                        color: 'primary.contrastText',
                                        '&:hover': {
                                            backgroundColor: 'primary.dark',
                                        },
                                        '& .MuiListItemIcon-root': {
                                            color: 'primary.contrastText',
                                        },
                                    },
                                }}
                            >
                                <ListItemIcon
                                    sx={{
                                        minWidth: 36,
                                        color: isSelected ? 'inherit' : 'text.secondary',
                                    }}
                                >
                                    {config.icon}
                                </ListItemIcon>
                                <ListItemText
                                    primary={config.label}
                                    primaryTypographyProps={{
                                        variant: 'body2',
                                        fontWeight: isSelected ? 600 : 400,
                                    }}
                                />
                                {entitiesLoading ? (
                                    <Skeleton variant="circular" width={24} height={24} />
                                ) : (
                                    <Typography
                                        variant="body2"
                                        sx={{
                                            color: isSelected ? 'inherit' : 'text.secondary',
                                            fontWeight: 500,
                                            minWidth: 24,
                                            textAlign: 'right',
                                        }}
                                    >
                                        {count}
                                    </Typography>
                                )}
                            </ListItemButton>
                        </ListItem>
                    );
                })}
            </List>
        </Box>
    );
}
