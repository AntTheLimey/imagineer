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
	"testing"
	"time"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestOrphanedEntity_Structure tests the OrphanedEntity struct fields.
func TestOrphanedEntity_Structure(t *testing.T) {
	orphan := OrphanedEntity{
		ID:         1,
		Name:       "Test Entity",
		EntityType: models.EntityTypeNPC,
	}

	assert.Equal(t, int64(1), orphan.ID)
	assert.Equal(t, "Test Entity", orphan.Name)
	assert.Equal(t, models.EntityTypeNPC, orphan.EntityType)
}

// TestDuplicateNamePair_Structure tests the DuplicateNamePair struct fields.
func TestDuplicateNamePair_Structure(t *testing.T) {
	pair := DuplicateNamePair{
		EntityID1:  1,
		Name1:      "John Smith",
		EntityID2:  2,
		Name2:      "Jon Smith",
		Similarity: 0.85,
	}

	assert.Equal(t, int64(1), pair.EntityID1)
	assert.Equal(t, "John Smith", pair.Name1)
	assert.Equal(t, int64(2), pair.EntityID2)
	assert.Equal(t, "Jon Smith", pair.Name2)
	assert.InDelta(t, 0.85, pair.Similarity, 0.001)
}

// TestTimelineConflict_Structure tests the TimelineConflict struct fields.
func TestTimelineConflict_Structure(t *testing.T) {
	eventDate := time.Date(1925, 3, 15, 0, 0, 0, 0, time.UTC)

	conflict := TimelineConflict{
		EntityID:   1,
		EntityName: "Professor Armitage",
		EventDate:  eventDate,
		EventCount: 2,
		EventIDs:   []int64{10, 11},
	}

	assert.Equal(t, int64(1), conflict.EntityID)
	assert.Equal(t, "Professor Armitage", conflict.EntityName)
	assert.Equal(t, eventDate, conflict.EventDate)
	assert.Equal(t, 2, conflict.EventCount)
	assert.Len(t, conflict.EventIDs, 2)
}

// TestInvalidReference_Structure tests the InvalidReference struct fields.
func TestInvalidReference_Structure(t *testing.T) {
	ref := InvalidReference{
		RelationshipID:  1,
		MissingEntityID: 2,
		ReferenceType:   "target",
	}

	assert.Equal(t, int64(1), ref.RelationshipID)
	assert.Equal(t, int64(2), ref.MissingEntityID)
	assert.Equal(t, "target", ref.ReferenceType)
}

// TestEmptySession_Structure tests the EmptySession struct fields.
func TestEmptySession_Structure(t *testing.T) {
	session := EmptySession{
		ID:            1,
		SessionNumber: 5,
	}

	assert.Equal(t, int64(1), session.ID)
	assert.Equal(t, 5, session.SessionNumber)
}

// TestOrphanedEntity_AllEntityTypes tests OrphanedEntity with different entity types.
func TestOrphanedEntity_AllEntityTypes(t *testing.T) {
	entityTypes := []models.EntityType{
		models.EntityTypeNPC,
		models.EntityTypeLocation,
		models.EntityTypeItem,
		models.EntityTypeFaction,
		models.EntityTypeClue,
		models.EntityTypeCreature,
		models.EntityTypeOrganization,
		models.EntityTypeEvent,
		models.EntityTypeDocument,
		models.EntityTypeOther,
	}

	for i, et := range entityTypes {
		t.Run(string(et), func(t *testing.T) {
			orphan := OrphanedEntity{
				ID:         int64(i + 1),
				Name:       "Test",
				EntityType: et,
			}
			assert.Equal(t, et, orphan.EntityType)
		})
	}
}

// TestDuplicateNamePair_SimilarityRange tests similarity values at boundaries.
func TestDuplicateNamePair_SimilarityRange(t *testing.T) {
	tests := []struct {
		name       string
		similarity float64
	}{
		{"exact match", 1.0},
		{"high similarity", 0.9},
		{"threshold similarity", 0.7},
		{"low similarity", 0.3},
		{"no similarity", 0.0},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := DuplicateNamePair{
				EntityID1:  int64(i*2 + 1),
				Name1:      "Name1",
				EntityID2:  int64(i*2 + 2),
				Name2:      "Name2",
				Similarity: tt.similarity,
			}
			assert.InDelta(t, tt.similarity, pair.Similarity, 0.001)
		})
	}
}

// TestTimelineConflict_MultipleEvents tests conflicts with multiple event IDs.
func TestTimelineConflict_MultipleEvents(t *testing.T) {
	eventIDs := []int64{10, 11, 12, 13, 14}

	conflict := TimelineConflict{
		EntityID:   1,
		EntityName: "Test Entity",
		EventDate:  time.Now(),
		EventCount: 5,
		EventIDs:   eventIDs,
	}

	assert.Equal(t, 5, conflict.EventCount)
	assert.Len(t, conflict.EventIDs, 5)
}

// TestInvalidReference_ReferenceTypes tests both source and target reference types.
func TestInvalidReference_ReferenceTypes(t *testing.T) {
	tests := []struct {
		name    string
		refType string
	}{
		{"source reference", "source"},
		{"target reference", "target"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := InvalidReference{
				RelationshipID:  int64(i + 1),
				MissingEntityID: int64(i + 100),
				ReferenceType:   tt.refType,
			}
			assert.Equal(t, tt.refType, ref.ReferenceType)
		})
	}
}

// TestEmptySession_ZeroSessionNumber tests EmptySession with session number 0.
func TestEmptySession_ZeroSessionNumber(t *testing.T) {
	session := EmptySession{
		ID:            1,
		SessionNumber: 0,
	}

	assert.Equal(t, 0, session.SessionNumber)
}
