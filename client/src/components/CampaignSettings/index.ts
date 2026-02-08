// -------------------------------------------------------------------------
//
// Imagineer - TTRPG Campaign Intelligence Platform
//
// Copyright (c) 2025 - 2026
// This software is released under The MIT License
//
// -------------------------------------------------------------------------

import type { Campaign } from '../../types';
import type { CampaignSettingsData } from './CampaignSettings';

export { default } from './CampaignSettings';
export type { CampaignSettingsData } from './CampaignSettings';

/**
 * Available genre options for campaigns.
 */
export const GENRE_OPTIONS = [
    'Fantasy',
    'Sci-Fi',
    'Horror',
    'Mystery',
    'Cyberpunk',
    'Steampunk',
    'Post-Apocalyptic',
    'Superhero',
    'Historical',
    'Urban Fantasy',
    'Space Opera',
    'Dark Fantasy',
    'Pulp Adventure',
    'Western',
    'Mythological',
    'Modern Day',
    'Noir',
    'Wuxia',
    'Military',
    'Comedy',
] as const;

/**
 * Convert a campaign object to form data structure.
 */
export function campaignToFormData(campaign: Campaign): CampaignSettingsData {
    return {
        name: campaign.name,
        description: campaign.description ?? '',
        gameSystemId: String(campaign.systemId),
        genre: (campaign.settings?.genre as string) ?? '',
        imageStylePrompt: (campaign.settings?.imageStylePrompt as string) ?? '',
    };
}
