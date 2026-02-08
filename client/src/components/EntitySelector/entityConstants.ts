// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

/**
 * Entity-related constants shared across components.
 */

import type { EntityType } from '../../types';

/**
 * Maps entity types to display colours for chips.
 */
export const entityTypeColors: Record<EntityType, 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info'> = {
    npc: 'primary',
    location: 'success',
    item: 'warning',
    faction: 'secondary',
    clue: 'info',
    creature: 'error',
    organization: 'secondary',
    event: 'info',
    document: 'default',
    other: 'default',
};

/**
 * Formats an entity type for display.
 */
export function formatEntityType(type: EntityType): string {
    return type.charAt(0).toUpperCase() + type.slice(1);
}
