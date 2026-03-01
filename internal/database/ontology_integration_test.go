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
	"testing"
	"time"
)

// TestIntegration_CheckRequiredRelationships exercises the
// check_required_relationships(campaign_id) database function.
// It creates a campaign with an entity, defines a required
// relationship rule for that entity type, verifies the function
// reports the missing relationship, then satisfies the requirement
// and verifies the violation disappears.
func TestIntegration_CheckRequiredRelationships(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	campaignID, entityID := createTestCampaign(t, db)

	suffix := time.Now().UnixNano()

	// Insert a campaign_entity_types row so the FK on
	// required_relationships is satisfied. The entity created by
	// createTestCampaign has entity_type = "npc".
	var entityTypeID int64
	err := db.QueryRow(ctx,
		`INSERT INTO campaign_entity_types
		     (campaign_id, name, abstract, description)
		 VALUES ($1, $2, false, $3)
		 RETURNING id`,
		campaignID,
		"npc",
		fmt.Sprintf("test entity type %d", suffix),
	).Scan(&entityTypeID)
	if err != nil {
		t.Fatalf("failed to create campaign_entity_type: %v", err)
	}

	// Create a relationship type for this campaign.
	relTypeName := fmt.Sprintf("mentors-%d", suffix)
	var relTypeID int64
	err = db.QueryRow(ctx,
		`INSERT INTO relationship_types
		     (campaign_id, name, description)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		campaignID,
		relTypeName,
		"Test relationship type for required-relationship check",
	).Scan(&relTypeID)
	if err != nil {
		t.Fatalf("failed to create relationship_type: %v", err)
	}

	// Create a required_relationships rule: every "npc" must have
	// a relationship of the type we just created.
	var requiredRelID int64
	err = db.QueryRow(ctx,
		`INSERT INTO required_relationships
		     (campaign_id, entity_type, relationship_type_name)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		campaignID,
		"npc",
		relTypeName,
	).Scan(&requiredRelID)
	if err != nil {
		t.Fatalf("failed to create required_relationship rule: %v", err)
	}

	// Create a second entity to use as a relationship target later.
	var targetEntityID int64
	err = db.QueryRow(ctx,
		`INSERT INTO entities
		     (campaign_id, entity_type, name, description,
		      source_confidence)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		campaignID,
		"npc",
		fmt.Sprintf("Dr. Laban Shrewsbury %d", suffix),
		"A scholar of the Cthulhu Mythos who has traveled "+
			"to distant worlds through gates opened by the "+
			"Necronomicon.",
		"AUTHORITATIVE",
	).Scan(&targetEntityID)
	if err != nil {
		t.Fatalf("failed to create target entity: %v", err)
	}

	// Register cleanup in reverse dependency order.
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cleanupCancel()

		_ = db.Exec(cleanupCtx,
			`DELETE FROM relationships WHERE campaign_id = $1`,
			campaignID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM required_relationships WHERE id = $1`,
			requiredRelID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM relationship_types WHERE id = $1`,
			relTypeID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM entities WHERE id = $1`,
			targetEntityID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM campaign_entity_types WHERE id = $1`,
			entityTypeID)
	})

	// --- Phase 1: The entity should be reported as missing the
	// required relationship. ---

	rows, err := db.Query(ctx,
		`SELECT entity_id, entity_name, entity_type,
		        missing_relationship_type
		 FROM check_required_relationships($1)`,
		campaignID,
	)
	if err != nil {
		t.Fatalf("check_required_relationships query failed: %v", err)
	}
	defer rows.Close()

	type violation struct {
		EntityID       int64
		EntityName     string
		EntityType     string
		MissingRelType string
	}

	var violations []violation
	for rows.Next() {
		var v violation
		if err := rows.Scan(
			&v.EntityID, &v.EntityName,
			&v.EntityType, &v.MissingRelType,
		); err != nil {
			t.Fatalf("failed to scan violation row: %v", err)
		}
		violations = append(violations, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	// Both the original entity and our target entity are NPCs, so
	// both should appear as violations.
	foundOriginal := false
	foundTarget := false
	for _, v := range violations {
		t.Logf("violation: entity_id=%d name=%q type=%q missing=%q",
			v.EntityID, v.EntityName, v.EntityType, v.MissingRelType)
		if v.EntityID == entityID && v.MissingRelType == relTypeName {
			foundOriginal = true
		}
		if v.EntityID == targetEntityID && v.MissingRelType == relTypeName {
			foundTarget = true
		}
	}

	if !foundOriginal {
		t.Errorf("expected entity %d to appear as missing "+
			"relationship type %q", entityID, relTypeName)
	}
	if !foundTarget {
		t.Errorf("expected entity %d to appear as missing "+
			"relationship type %q", targetEntityID, relTypeName)
	}

	// --- Phase 2: Satisfy the required relationship for both
	// entities and verify no violations remain. ---

	// Create a relationship from the original entity to the target.
	var relID1 int64
	err = db.QueryRow(ctx,
		`INSERT INTO relationships
		     (campaign_id, relationship_type_id,
		      source_entity_id, target_entity_id,
		      description, strength, source_confidence)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		campaignID,
		relTypeID,
		entityID,
		targetEntityID,
		"Test mentorship relationship",
		"strong",
		"AUTHORITATIVE",
	).Scan(&relID1)
	if err != nil {
		t.Fatalf("failed to create relationship: %v", err)
	}

	t.Logf("created relationship %d (entity %d -> entity %d)",
		relID1, entityID, targetEntityID)

	// Both entities participate (source and target) in a
	// relationship of this type, so both should now be satisfied.
	rows2, err := db.Query(ctx,
		`SELECT entity_id, entity_name, entity_type,
		        missing_relationship_type
		 FROM check_required_relationships($1)`,
		campaignID,
	)
	if err != nil {
		t.Fatalf("check_required_relationships query failed: %v", err)
	}
	defer rows2.Close()

	var remaining []violation
	for rows2.Next() {
		var v violation
		if err := rows2.Scan(
			&v.EntityID, &v.EntityName,
			&v.EntityType, &v.MissingRelType,
		); err != nil {
			t.Fatalf("failed to scan violation row: %v", err)
		}
		remaining = append(remaining, v)
	}
	if err := rows2.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	if len(remaining) > 0 {
		for _, v := range remaining {
			t.Logf("unexpected remaining violation: "+
				"entity_id=%d name=%q type=%q missing=%q",
				v.EntityID, v.EntityName,
				v.EntityType, v.MissingRelType)
		}
		t.Errorf("expected no violations after creating "+
			"relationship, got %d", len(remaining))
	}
}

// TestIntegration_CheckCardinalityViolations exercises the
// check_cardinality_violations(campaign_id) database function.
// It creates a cardinality constraint with max_source=1, verifies
// a single relationship produces no violations, then adds a second
// relationship from the same source entity and verifies the
// function reports the cardinality violation.
func TestIntegration_CheckCardinalityViolations(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	campaignID, entityID := createTestCampaign(t, db)

	suffix := time.Now().UnixNano()

	// Create a relationship type for this campaign.
	relTypeName := fmt.Sprintf("guards-%d", suffix)
	var relTypeID int64
	err := db.QueryRow(ctx,
		`INSERT INTO relationship_types
		     (campaign_id, name, description)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		campaignID,
		relTypeName,
		"Test relationship type for cardinality check",
	).Scan(&relTypeID)
	if err != nil {
		t.Fatalf("failed to create relationship_type: %v", err)
	}

	// Create a cardinality constraint: max 1 outgoing relationship
	// of this type per entity (max_source=1).
	var constraintID int64
	err = db.QueryRow(ctx,
		`INSERT INTO cardinality_constraints
		     (campaign_id, relationship_type_id, max_source)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		campaignID,
		relTypeID,
		1,
	).Scan(&constraintID)
	if err != nil {
		t.Fatalf("failed to create cardinality_constraint: %v", err)
	}

	// Create two target entities.
	var targetEntityID1, targetEntityID2 int64
	err = db.QueryRow(ctx,
		`INSERT INTO entities
		     (campaign_id, entity_type, name, description,
		      source_confidence)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		campaignID,
		"npc",
		fmt.Sprintf("Randolph Carter %d", suffix),
		"A dreamer who has visited the Dreamlands and spoken "+
			"with the cats of Ulthar.",
		"AUTHORITATIVE",
	).Scan(&targetEntityID1)
	if err != nil {
		t.Fatalf("failed to create target entity 1: %v", err)
	}

	err = db.QueryRow(ctx,
		`INSERT INTO entities
		     (campaign_id, entity_type, name, description,
		      source_confidence)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		campaignID,
		"npc",
		fmt.Sprintf("Harley Warren %d", suffix),
		"An occultist and companion of Randolph Carter who "+
			"descended into a forbidden crypt and never returned.",
		"AUTHORITATIVE",
	).Scan(&targetEntityID2)
	if err != nil {
		t.Fatalf("failed to create target entity 2: %v", err)
	}

	// Register cleanup in reverse dependency order.
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cleanupCancel()

		_ = db.Exec(cleanupCtx,
			`DELETE FROM relationships WHERE campaign_id = $1`,
			campaignID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM cardinality_constraints WHERE id = $1`,
			constraintID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM relationship_types WHERE id = $1`,
			relTypeID)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM entities WHERE id = $1`,
			targetEntityID2)
		_ = db.Exec(cleanupCtx,
			`DELETE FROM entities WHERE id = $1`,
			targetEntityID1)
	})

	// --- Phase 1: Create one relationship from the source entity.
	// This should NOT violate the max_source=1 constraint. ---

	var relID1 int64
	err = db.QueryRow(ctx,
		`INSERT INTO relationships
		     (campaign_id, relationship_type_id,
		      source_entity_id, target_entity_id,
		      description, strength, source_confidence)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		campaignID,
		relTypeID,
		entityID,
		targetEntityID1,
		"First guard relationship",
		"strong",
		"AUTHORITATIVE",
	).Scan(&relID1)
	if err != nil {
		t.Fatalf("failed to create first relationship: %v", err)
	}

	t.Logf("created relationship %d (entity %d -> entity %d)",
		relID1, entityID, targetEntityID1)

	rows, err := db.Query(ctx,
		`SELECT entity_id, entity_name, entity_type,
		        relationship_type, direction,
		        current_count, max_allowed
		 FROM check_cardinality_violations($1)`,
		campaignID,
	)
	if err != nil {
		t.Fatalf("check_cardinality_violations query failed: %v", err)
	}
	defer rows.Close()

	type cardinalityViolation struct {
		EntityID         int64
		EntityName       string
		EntityType       string
		RelationshipType string
		Direction        string
		CurrentCount     int
		MaxAllowed       int
	}

	var violations []cardinalityViolation
	for rows.Next() {
		var v cardinalityViolation
		if err := rows.Scan(
			&v.EntityID, &v.EntityName, &v.EntityType,
			&v.RelationshipType, &v.Direction,
			&v.CurrentCount, &v.MaxAllowed,
		); err != nil {
			t.Fatalf("failed to scan cardinality violation: %v", err)
		}
		violations = append(violations, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			t.Logf("unexpected violation: entity_id=%d name=%q "+
				"type=%q rel_type=%q dir=%s count=%d max=%d",
				v.EntityID, v.EntityName, v.EntityType,
				v.RelationshipType, v.Direction,
				v.CurrentCount, v.MaxAllowed)
		}
		t.Fatalf("expected no cardinality violations with one "+
			"relationship, got %d", len(violations))
	}

	t.Logf("phase 1: no cardinality violations (as expected)")

	// --- Phase 2: Create a second relationship from the same
	// source entity with the same type. This exceeds max_source=1
	// and should trigger a violation. ---

	var relID2 int64
	err = db.QueryRow(ctx,
		`INSERT INTO relationships
		     (campaign_id, relationship_type_id,
		      source_entity_id, target_entity_id,
		      description, strength, source_confidence)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		campaignID,
		relTypeID,
		entityID,
		targetEntityID2,
		"Second guard relationship (violates cardinality)",
		"moderate",
		"AUTHORITATIVE",
	).Scan(&relID2)
	if err != nil {
		t.Fatalf("failed to create second relationship: %v", err)
	}

	t.Logf("created relationship %d (entity %d -> entity %d)",
		relID2, entityID, targetEntityID2)

	rows2, err := db.Query(ctx,
		`SELECT entity_id, entity_name, entity_type,
		        relationship_type, direction,
		        current_count, max_allowed
		 FROM check_cardinality_violations($1)`,
		campaignID,
	)
	if err != nil {
		t.Fatalf("check_cardinality_violations query failed: %v", err)
	}
	defer rows2.Close()

	var violations2 []cardinalityViolation
	for rows2.Next() {
		var v cardinalityViolation
		if err := rows2.Scan(
			&v.EntityID, &v.EntityName, &v.EntityType,
			&v.RelationshipType, &v.Direction,
			&v.CurrentCount, &v.MaxAllowed,
		); err != nil {
			t.Fatalf("failed to scan cardinality violation: %v", err)
		}
		violations2 = append(violations2, v)
	}
	if err := rows2.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	if len(violations2) == 0 {
		t.Fatal("expected at least one cardinality violation " +
			"after creating second relationship, got none")
	}

	foundSourceViolation := false
	for _, v := range violations2 {
		t.Logf("violation: entity_id=%d name=%q type=%q "+
			"rel_type=%q dir=%s count=%d max=%d",
			v.EntityID, v.EntityName, v.EntityType,
			v.RelationshipType, v.Direction,
			v.CurrentCount, v.MaxAllowed)

		if v.EntityID == entityID &&
			v.RelationshipType == relTypeName &&
			v.Direction == "source" &&
			v.CurrentCount == 2 &&
			v.MaxAllowed == 1 {
			foundSourceViolation = true
		}
	}

	if !foundSourceViolation {
		t.Errorf("expected source cardinality violation for "+
			"entity %d with rel_type %q (count=2, max=1)",
			entityID, relTypeName)
	}
}
