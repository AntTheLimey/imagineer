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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOrphanedEntity_Structure tests the OrphanedEntity struct fields.
func TestOrphanedEntity_Structure(t *testing.T) {
	id := uuid.New()
	orphan := OrphanedEntity{
		ID:         id,
		Name:       "Test Entity",
		EntityType: models.EntityTypeNPC,
	}

	assert.Equal(t, id, orphan.ID)
	assert.Equal(t, "Test Entity", orphan.Name)
	assert.Equal(t, models.EntityTypeNPC, orphan.EntityType)
}

// TestDuplicateNamePair_Structure tests the DuplicateNamePair struct fields.
func TestDuplicateNamePair_Structure(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	pair := DuplicateNamePair{
		EntityID1:  id1,
		Name1:      "John Smith",
		EntityID2:  id2,
		Name2:      "Jon Smith",
		Similarity: 0.85,
	}

	assert.Equal(t, id1, pair.EntityID1)
	assert.Equal(t, "John Smith", pair.Name1)
	assert.Equal(t, id2, pair.EntityID2)
	assert.Equal(t, "Jon Smith", pair.Name2)
	assert.InDelta(t, 0.85, pair.Similarity, 0.001)
}

// TestTimelineConflict_Structure tests the TimelineConflict struct fields.
func TestTimelineConflict_Structure(t *testing.T) {
	entityID := uuid.New()
	eventID1 := uuid.New()
	eventID2 := uuid.New()
	eventDate := time.Date(1925, 3, 15, 0, 0, 0, 0, time.UTC)

	conflict := TimelineConflict{
		EntityID:   entityID,
		EntityName: "Professor Armitage",
		EventDate:  eventDate,
		EventCount: 2,
		EventIDs:   []uuid.UUID{eventID1, eventID2},
	}

	assert.Equal(t, entityID, conflict.EntityID)
	assert.Equal(t, "Professor Armitage", conflict.EntityName)
	assert.Equal(t, eventDate, conflict.EventDate)
	assert.Equal(t, 2, conflict.EventCount)
	assert.Len(t, conflict.EventIDs, 2)
}

// TestInvalidReference_Structure tests the InvalidReference struct fields.
func TestInvalidReference_Structure(t *testing.T) {
	relID := uuid.New()
	missingID := uuid.New()

	ref := InvalidReference{
		RelationshipID:  relID,
		MissingEntityID: missingID,
		ReferenceType:   "target",
	}

	assert.Equal(t, relID, ref.RelationshipID)
	assert.Equal(t, missingID, ref.MissingEntityID)
	assert.Equal(t, "target", ref.ReferenceType)
}

// TestEmptySession_Structure tests the EmptySession struct fields.
func TestEmptySession_Structure(t *testing.T) {
	id := uuid.New()

	session := EmptySession{
		ID:            id,
		SessionNumber: 5,
	}

	assert.Equal(t, id, session.ID)
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

	for _, et := range entityTypes {
		t.Run(string(et), func(t *testing.T) {
			orphan := OrphanedEntity{
				ID:         uuid.New(),
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := DuplicateNamePair{
				EntityID1:  uuid.New(),
				Name1:      "Name1",
				EntityID2:  uuid.New(),
				Name2:      "Name2",
				Similarity: tt.similarity,
			}
			assert.InDelta(t, tt.similarity, pair.Similarity, 0.001)
		})
	}
}

// TestTimelineConflict_MultipleEvents tests conflicts with multiple event IDs.
func TestTimelineConflict_MultipleEvents(t *testing.T) {
	eventIDs := make([]uuid.UUID, 5)
	for i := range eventIDs {
		eventIDs[i] = uuid.New()
	}

	conflict := TimelineConflict{
		EntityID:   uuid.New(),
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := InvalidReference{
				RelationshipID:  uuid.New(),
				MissingEntityID: uuid.New(),
				ReferenceType:   tt.refType,
			}
			assert.Equal(t, tt.refType, ref.ReferenceType)
		})
	}
}

// TestEmptySession_ZeroSessionNumber tests EmptySession with session number 0.
func TestEmptySession_ZeroSessionNumber(t *testing.T) {
	session := EmptySession{
		ID:            uuid.New(),
		SessionNumber: 0,
	}

	assert.Equal(t, 0, session.SessionNumber)
}
