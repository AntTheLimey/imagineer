/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ontology

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEntityTypes(t *testing.T) {
	et, err := LoadEntityTypes(
		"../../schemas/ontology/entity-types.yaml",
	)
	require.NoError(t, err)
	require.NotNil(t, et)

	// Check abstract parents exist.
	assert.True(t, et.Types["agent"].Abstract)
	assert.True(t, et.Types["person"].Abstract)
	assert.True(t, et.Types["character"].Abstract)
	assert.True(t, et.Types["place"].Abstract)
	assert.True(t, et.Types["artifact"].Abstract)
	assert.True(t, et.Types["narrative"].Abstract)

	// Check concrete types exist.
	assert.False(t, et.Types["npc"].Abstract)
	assert.Equal(t, "character", et.Types["npc"].Parent)

	assert.False(t, et.Types["pc"].Abstract)
	assert.Equal(t, "character", et.Types["pc"].Parent)

	assert.False(t, et.Types["location"].Abstract)
	assert.Equal(t, "place", et.Types["location"].Parent)

	// Check sub-types.
	assert.Equal(t, "creature",
		et.Types["undead"].Parent)
	assert.Equal(t, "item",
		et.Types["weapon"].Parent)
	assert.Equal(t, "document",
		et.Types["spell"].Parent)
	assert.Equal(t, "event",
		et.Types["battle"].Parent)
}

func TestLoadRelationshipTypes(t *testing.T) {
	rt, err := LoadRelationshipTypes(
		"../../schemas/ontology/relationship-types.yaml",
	)
	require.NoError(t, err)
	require.NotNil(t, rt)

	// Check an existing type.
	owns := rt.Types["owns"]
	assert.Equal(t, "owned_by", owns.Inverse)
	assert.False(t, owns.Symmetric)

	// Check a symmetric type.
	knows := rt.Types["knows"]
	assert.True(t, knows.Symmetric)

	// Check display labels.
	assert.Equal(t, "Employs",
		rt.Types["employs"].DisplayLabel)
	assert.Equal(t, "Employed by",
		rt.Types["employs"].InverseDisplayLabel)

	// Check backward-compatible templates exist.
	backwardCompat := []string{
		"owns", "employs", "works_for", "reports_to",
		"parent_of", "located_at", "member_of",
		"created", "rules", "headquartered_at",
		"knows", "friend_of", "enemy_of", "allied_with",
	}
	for _, name := range backwardCompat {
		_, ok := rt.Types[name]
		assert.True(t, ok,
			"missing backward-compatible type: %s", name)
	}

	// Check total count is approximately 80+.
	assert.Greater(t, len(rt.Types), 70)
}

func TestLoadConstraints(t *testing.T) {
	c, err := LoadConstraints(
		"../../schemas/ontology/constraints.yaml",
	)
	require.NoError(t, err)
	require.NotNil(t, c)

	// Check domain/range for a known type.
	plays := c.DomainRange["plays"]
	assert.Contains(t, plays.Domain, "player")
	assert.Contains(t, plays.Range, "pc")

	// Check constraints for spatial types.
	locatedAt := c.DomainRange["located_at"]
	assert.Contains(t, locatedAt.Domain, "agent")
	assert.Contains(t, locatedAt.Range, "place")

	// Check required relationships.
	assert.Contains(t, c.Required["npc"], "located_at")
	assert.Contains(t, c.Required["pc"], "played_by")
	assert.Contains(t, c.Required["pc"], "located_at")
	assert.Contains(t, c.Required["faction"],
		"headquartered_at")

	// Cardinality defaults are empty.
	assert.Empty(t, c.Cardinality)
}

func TestResolveConcreteTypes(t *testing.T) {
	et, err := LoadEntityTypes(
		"../../schemas/ontology/entity-types.yaml",
	)
	require.NoError(t, err)

	// "agent" should resolve to npc, pc, creature,
	// faction, organization and all their sub-types.
	concrete := et.ResolveToConcreteTypes("agent")
	assert.Contains(t, concrete, "npc")
	assert.Contains(t, concrete, "pc")
	assert.Contains(t, concrete, "creature")
	assert.Contains(t, concrete, "undead")
	assert.Contains(t, concrete, "faction")
	assert.Contains(t, concrete, "organization")
	assert.Contains(t, concrete, "government")
	assert.NotContains(t, concrete, "agent")
	assert.NotContains(t, concrete, "character")

	// "artifact" should resolve to item, document and
	// all their sub-types.
	artConcrete := et.ResolveToConcreteTypes("artifact")
	assert.Contains(t, artConcrete, "item")
	assert.Contains(t, artConcrete, "weapon")
	assert.Contains(t, artConcrete, "document")
	assert.Contains(t, artConcrete, "spell")
	assert.NotContains(t, artConcrete, "artifact")

	// A concrete type with no children resolves to
	// just itself.
	locConcrete := et.ResolveToConcreteTypes("location")
	assert.Equal(t, []string{"location"}, locConcrete)

	// An unknown type resolves to empty.
	unknown := et.ResolveToConcreteTypes("nonexistent")
	assert.Empty(t, unknown)
}

func TestLoadOntology(t *testing.T) {
	ont, err := LoadOntology(
		"../../schemas/ontology",
	)
	require.NoError(t, err)
	require.NotNil(t, ont)
	require.NotNil(t, ont.EntityTypes)
	require.NotNil(t, ont.RelationshipTypes)
	require.NotNil(t, ont.Constraints)

	// Sanity check across all three files.
	assert.Greater(t, len(ont.EntityTypes.Types), 30)
	assert.Greater(t, len(ont.RelationshipTypes.Types), 70)
	assert.Greater(t, len(ont.Constraints.DomainRange), 50)
}
