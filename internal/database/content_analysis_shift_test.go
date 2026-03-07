//go:build integration

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
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestAnalysisJob inserts a minimal content analysis job for a
// campaign and returns its ID. The caller is responsible for ensuring
// the campaign exists and cleaning up afterward.
func createTestAnalysisJob(t *testing.T, db *DB, campaignID int64) int64 {
	t.Helper()

	ctx := context.Background()

	job := &models.ContentAnalysisJob{
		CampaignID:  campaignID,
		SourceTable: "chapters",
		SourceID:    1,
		SourceField: "overview",
		Status:      "pending",
	}

	created, err := db.CreateAnalysisJob(ctx, job)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Exec(context.Background(),
			`DELETE FROM content_analysis_jobs WHERE id = $1`, created.ID)
	})

	return created.ID
}

// createTestAnalysisItemAtPosition inserts a content analysis item
// with the given position range and returns its ID.
func createTestAnalysisItemAtPosition(t *testing.T, db *DB, jobID int64, posStart, posEnd int) int64 {
	t.Helper()

	ctx := context.Background()

	items := []models.ContentAnalysisItem{
		{
			JobID:         jobID,
			DetectionType: "untagged_mention",
			MatchedText:   "Test Entity",
			PositionStart: &posStart,
			PositionEnd:   &posEnd,
			Resolution:    "pending",
			Phase:         "identification",
		},
	}

	err := db.CreateAnalysisItems(ctx, items)
	require.NoError(t, err)

	// Retrieve the ID of the item we just inserted.
	var itemID int64
	err = db.QueryRow(ctx,
		`SELECT id FROM content_analysis_items
		 WHERE job_id = $1 AND position_start = $2 AND position_end = $3
		 ORDER BY id DESC LIMIT 1`,
		jobID, posStart, posEnd,
	).Scan(&itemID)
	require.NoError(t, err)

	return itemID
}

// TestShiftItemPositions creates three items at positions 10-20, 30-40,
// and 50-60, then shifts by +5 after position 25. Item 1 should be
// unchanged; items 2 and 3 should be shifted to 35-45 and 55-65.
func TestShiftItemPositions(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	campaignID, _ := createTestCampaign(t, db)
	jobID := createTestAnalysisJob(t, db, campaignID)

	id1 := createTestAnalysisItemAtPosition(t, db, jobID, 10, 20)
	id2 := createTestAnalysisItemAtPosition(t, db, jobID, 30, 40)
	id3 := createTestAnalysisItemAtPosition(t, db, jobID, 50, 60)

	// Shift positions by +5 for items after position 25.
	err := db.ShiftItemPositions(ctx, jobID, 25, 5)
	require.NoError(t, err)

	// Fetch all items and verify positions.
	items, err := db.GetAnalysisItemsByIDs(ctx, []int64{id1, id2, id3})
	require.NoError(t, err)
	require.Len(t, items, 3)

	// Build a map for easy lookup by ID.
	byID := make(map[int64]models.ContentAnalysisItem)
	for _, item := range items {
		byID[item.ID] = item
	}

	// Item 1: position_start=10 <= 25, should be unchanged.
	assert.Equal(t, 10, *byID[id1].PositionStart)
	assert.Equal(t, 20, *byID[id1].PositionEnd)

	// Item 2: position_start=30 > 25, should be shifted to 35-45.
	assert.Equal(t, 35, *byID[id2].PositionStart)
	assert.Equal(t, 45, *byID[id2].PositionEnd)

	// Item 3: position_start=50 > 25, should be shifted to 55-65.
	assert.Equal(t, 55, *byID[id3].PositionStart)
	assert.Equal(t, 65, *byID[id3].PositionEnd)
}

// TestShiftItemPositions_NegativeDelta verifies that a negative delta
// decreases positions for items after the threshold.
func TestShiftItemPositions_NegativeDelta(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	campaignID, _ := createTestCampaign(t, db)
	jobID := createTestAnalysisJob(t, db, campaignID)

	id1 := createTestAnalysisItemAtPosition(t, db, jobID, 10, 20)
	id2 := createTestAnalysisItemAtPosition(t, db, jobID, 30, 40)
	id3 := createTestAnalysisItemAtPosition(t, db, jobID, 50, 60)

	// Shift positions by -10 for items after position 20.
	err := db.ShiftItemPositions(ctx, jobID, 20, -10)
	require.NoError(t, err)

	items, err := db.GetAnalysisItemsByIDs(ctx, []int64{id1, id2, id3})
	require.NoError(t, err)
	require.Len(t, items, 3)

	byID := make(map[int64]models.ContentAnalysisItem)
	for _, item := range items {
		byID[item.ID] = item
	}

	// Item 1: position_start=10 <= 20, should be unchanged.
	assert.Equal(t, 10, *byID[id1].PositionStart)
	assert.Equal(t, 20, *byID[id1].PositionEnd)

	// Item 2: position_start=30 > 20, shifted by -10 to 20-30.
	assert.Equal(t, 20, *byID[id2].PositionStart)
	assert.Equal(t, 30, *byID[id2].PositionEnd)

	// Item 3: position_start=50 > 20, shifted by -10 to 40-50.
	assert.Equal(t, 40, *byID[id3].PositionStart)
	assert.Equal(t, 50, *byID[id3].PositionEnd)
}

// TestGetAnalysisItemsByIDs creates two items and verifies they can be
// fetched by their IDs.
func TestGetAnalysisItemsByIDs(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	campaignID, _ := createTestCampaign(t, db)
	jobID := createTestAnalysisJob(t, db, campaignID)

	id1 := createTestAnalysisItemAtPosition(t, db, jobID, 5, 15)
	id2 := createTestAnalysisItemAtPosition(t, db, jobID, 25, 35)

	items, err := db.GetAnalysisItemsByIDs(ctx, []int64{id1, id2})
	require.NoError(t, err)
	require.Len(t, items, 2)

	// Verify ordering by position_start ascending.
	assert.Equal(t, id1, items[0].ID)
	assert.Equal(t, id2, items[1].ID)
}

// TestGetAnalysisItemsByIDs_Empty verifies that an empty input slice
// returns an empty result without error.
func TestGetAnalysisItemsByIDs_Empty(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	items, err := db.GetAnalysisItemsByIDs(ctx, []int64{})
	require.NoError(t, err)
	assert.Empty(t, items)
}
