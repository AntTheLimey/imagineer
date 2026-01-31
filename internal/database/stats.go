/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"context"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
)

// GetDashboardStats retrieves statistics for the dashboard.
func (db *DB) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Count campaigns
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM campaigns").Scan(&stats.TotalCampaigns)
	if err != nil {
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Count entities
	stats.TotalEntities, err = db.CountEntities(ctx)
	if err != nil {
		return nil, err
	}

	// Count relationships
	stats.TotalRelationships, err = db.CountRelationships(ctx)
	if err != nil {
		return nil, err
	}

	// Count sessions
	stats.TotalSessions, err = db.CountSessions(ctx)
	if err != nil {
		return nil, err
	}

	// Get entities by type
	stats.EntitiesByType, err = db.CountEntitiesByType(ctx)
	if err != nil {
		return nil, err
	}

	// Get recent campaigns
	stats.RecentCampaigns, err = db.GetRecentCampaigns(ctx, 5)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetFrontendDashboardStats retrieves statistics in the format expected by the frontend.
func (db *DB) GetFrontendDashboardStats(ctx context.Context) (*models.FrontendDashboardStats, error) {
	stats := &models.FrontendDashboardStats{}

	// Count campaigns
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM campaigns").Scan(&stats.CampaignCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Count NPCs
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM entities WHERE entity_type = 'npc'").Scan(&stats.NPCCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count NPCs: %w", err)
	}

	// Count locations
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM entities WHERE entity_type = 'location'").Scan(&stats.LocationCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count locations: %w", err)
	}

	// Count items
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM entities WHERE entity_type = 'item'").Scan(&stats.ItemCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	// Count factions
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM entities WHERE entity_type = 'faction'").Scan(&stats.FactionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count factions: %w", err)
	}

	// Count timeline events
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM timeline_events").Scan(&stats.TimelineEventCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count timeline events: %w", err)
	}

	// Count total entities
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM entities").Scan(&stats.TotalEntityCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count total entities: %w", err)
	}

	return stats, nil
}
