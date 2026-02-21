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
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

// setupIntegrationDB loads the database configuration from the project
// config directory, connects to the database, and registers a cleanup
// function to close the connection when the test completes.
func setupIntegrationDB(t *testing.T) *DB {
	t.Helper()

	config, err := LoadConfig("../../config/db/db.json")
	if err != nil {
		t.Fatalf("failed to load database config: %v", err)
	}

	if v := os.Getenv("TEST_DB_HOST"); v != "" {
		config.Nodes[0].Hostname = v
	}
	if v := os.Getenv("TEST_DB_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("TEST_DB_PORT is not a valid integer: %v", err)
		}
		config.Nodes[0].Port = p
	}
	if v := os.Getenv("TEST_DB_USER"); v != "" {
		config.Users[0].Username = v
	}
	if v := os.Getenv("TEST_DB_PASSWORD"); v != "" {
		config.Users[0].Password = v
	}
	if v := os.Getenv("TEST_DB_NAME"); v != "" {
		config.Database = v
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := Connect(ctx, config)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// createTestCampaign inserts a test user, campaign, and entity into the
// database using raw SQL. It registers cleanup functions to delete the
// rows in reverse dependency order (entity, campaign, user) when the
// test completes. The entity is given a rich Lovecraftian description
// suitable for embedding generation.
func createTestCampaign(t *testing.T, db *DB) (campaignID int64, entityID int64) {
	t.Helper()

	ctx := context.Background()

	var userID int64

	// Create a test user (let BIGSERIAL assign the ID).
	err := db.QueryRow(ctx,
		`INSERT INTO users (google_id, email, name)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		fmt.Sprintf("test-google-%d", time.Now().UnixNano()),
		fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		"Integration Test User",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create a test campaign (let BIGSERIAL assign the ID).
	err = db.QueryRow(ctx,
		`INSERT INTO campaigns (name, owner_id, description, genre)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		"Integration Test Campaign",
		userID,
		"The investigators explore forbidden knowledge related to "+
			"Yog-Sothoth and the Necronomicon in the haunted streets "+
			"of Arkham. Ancient rituals threaten to tear the fabric "+
			"of reality as cultists gather beneath Miskatonic University.",
		"lovecraftian",
	).Scan(&campaignID)
	if err != nil {
		t.Fatalf("failed to create test campaign: %v", err)
	}

	// Create a test entity with a meaningful description for embedding.
	err = db.QueryRow(ctx,
		`INSERT INTO entities (campaign_id, entity_type, name, description, source_confidence)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		campaignID,
		"npc",
		"Professor Henry Armitage",
		"Head librarian at Miskatonic University, Professor Armitage "+
			"is one of the few scholars who understands the true danger "+
			"of the Necronomicon and the cosmic horrors it describes. "+
			"He has dedicated his life to protecting forbidden tomes "+
			"from those who would misuse their eldritch power. Armitage "+
			"played a crucial role in thwarting the Dunwich Horror and "+
			"remains vigilant against incursions from the outer darkness.",
		"AUTHORITATIVE",
	).Scan(&entityID)
	if err != nil {
		t.Fatalf("failed to create test entity: %v", err)
	}

	// Clean up in reverse dependency order.
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cleanupCancel()

		_ = db.Exec(cleanupCtx,
			`DELETE FROM entities WHERE id = $1`, entityID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM campaigns WHERE id = $1`, campaignID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM users WHERE id = $1`, userID)
	})

	return campaignID, entityID
}

// waitForChunks polls the specified chunk table until at least one row
// appears with a non-null embedding for the given source ID, or until
// the timeout expires. The tableName parameter is a trusted internal
// value (not user input). Returns the count of matching rows.
func waitForChunks(
	t *testing.T,
	db *DB,
	tableName string,
	sourceID int64,
	timeout time.Duration,
) int {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	query := fmt.Sprintf(
		`SELECT COUNT(*) FROM %s WHERE source_id = $1 AND embedding IS NOT NULL`,
		tableName,
	)

	for time.Now().Before(deadline) {
		var count int
		err := db.QueryRow(ctx, query, sourceID).Scan(&count)
		if err != nil {
			// The chunk table may not exist yet; keep polling.
			time.Sleep(2 * time.Second)
			continue
		}
		if count > 0 {
			return count
		}
		time.Sleep(2 * time.Second)
	}

	t.Fatalf(
		"timed out after %s waiting for chunks in %s for source_id %d",
		timeout, tableName, sourceID,
	)
	return 0
}

// TestIntegration_VectorizationAvailable verifies that the
// pgedge_vectorizer extension is installed in the database. This is a
// prerequisite for all other embedding and search integration tests.
func TestIntegration_VectorizationAvailable(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	if !db.IsVectorizationAvailable(ctx) {
		t.Fatal("pgedge_vectorizer extension is not available; " +
			"embedding tests require this extension to be installed")
	}
}

// TestIntegration_ChunksCreated inserts an entity with a description
// and polls for chunk rows in the entities_description_chunks table.
// It verifies that the vectorizer produces rows with non-null
// embeddings within a reasonable timeframe.
func TestIntegration_ChunksCreated(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	if !db.IsVectorizationAvailable(ctx) {
		t.Skip("pgedge_vectorizer not available; skipping chunk test")
	}

	_, entityID := createTestCampaign(t, db)

	count := waitForChunks(
		t, db, "entities_description_chunks", entityID, 60*time.Second,
	)

	if count == 0 {
		t.Fatal("expected at least one chunk row with a non-null embedding")
	}

	t.Logf("vectorizer created %d chunk(s) for entity %d", count, entityID)
}

// TestIntegration_SearchReturnsResults exercises the full search
// pipeline: insert entity data, wait for embeddings to be generated,
// then call SearchCampaignContent and verify that results are returned
// with positive scores.
func TestIntegration_SearchReturnsResults(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	if !db.IsVectorizationAvailable(ctx) {
		t.Skip("pgedge_vectorizer not available; skipping search test")
	}

	campaignID, entityID := createTestCampaign(t, db)

	// Wait for entity description chunks to be embedded.
	waitForChunks(
		t, db, "entities_description_chunks", entityID, 60*time.Second,
	)

	results, err := db.SearchCampaignContent(
		ctx, campaignID, "forbidden knowledge Necronomicon", 10,
	)
	if err != nil {
		t.Fatalf("SearchCampaignContent returned error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one search result, got none")
	}

	for i, r := range results {
		if r.CombinedScore <= 0 {
			t.Errorf(
				"result[%d] has non-positive combined score: %f",
				i, r.CombinedScore,
			)
		}
		if r.VectorScore <= 0 {
			t.Errorf(
				"result[%d] has non-positive vector score: %f",
				i, r.VectorScore,
			)
		}
		t.Logf(
			"result[%d]: table=%s name=%q vector=%.4f combined=%.4f",
			i, r.SourceTable, r.SourceName,
			r.VectorScore, r.CombinedScore,
		)
	}
}

// TestIntegration_CampaignDescriptionSearch inserts a campaign with a
// rich Lovecraftian description, waits for chunks in the
// campaigns_description_chunks table, searches for content from the
// description, and verifies that results include source_table
// 'campaigns'.
func TestIntegration_CampaignDescriptionSearch(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	if !db.IsVectorizationAvailable(ctx) {
		t.Skip("pgedge_vectorizer not available; skipping campaign search test")
	}

	campaignID, _ := createTestCampaign(t, db)

	// Wait for campaign description chunks to be embedded.
	waitForChunks(
		t, db, "campaigns_description_chunks", campaignID, 60*time.Second,
	)

	results, err := db.SearchCampaignContent(
		ctx, campaignID, "forbidden knowledge Yog-Sothoth", 10,
	)
	if err != nil {
		t.Fatalf("SearchCampaignContent returned error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one search result, got none")
	}

	foundCampaign := false
	for i, r := range results {
		t.Logf(
			"result[%d]: table=%s name=%q vector=%.4f combined=%.4f",
			i, r.SourceTable, r.SourceName,
			r.VectorScore, r.CombinedScore,
		)
		if r.SourceTable == "campaigns" {
			foundCampaign = true
		}
	}

	if !foundCampaign {
		t.Error("expected at least one result with source_table='campaigns'")
	}
}
