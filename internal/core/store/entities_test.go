/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package store

import (
	"context"
	"testing"

	"github.com/antonypegg/imagineer/internal/core/worldtime"
)

func TestCreateRelationAndQueryAt(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	uid := seedUser(t, s)
	w, _ := s.CreateWorld(ctx, uid, "W")
	npc, _ := s.CreateType(ctx, w.ID, nil, "entity", "npc", "NPC")
	ally, _ := s.CreateType(ctx, w.ID, nil, "relation", "ally_of", "Ally of")
	a, _ := s.CreateEntity(ctx, w.ID, npc.ID, "Alice", "")
	b, _ := s.CreateEntity(ctx, w.ID, npc.ID, "Bob", "")

	_, err := s.CreateRelation(ctx, w.ID, ally.ID, a.ID, b.ID,
		worldtime.Range{Lo: 100, Hi: 200})
	if err != nil {
		t.Fatal(err)
	}
	in, err := s.RelationsAt(ctx, a.ID, 150)
	if err != nil || len(in) != 1 || in[0].Predicate != "ally_of" {
		t.Fatalf("at 150: %+v err=%v", in, err)
	}
	out, err := s.RelationsAt(ctx, a.ID, 250)
	if err != nil || len(out) != 0 {
		t.Fatalf("at 250: %+v err=%v", out, err)
	}

	// The target's side of the edge answers too: the scrubber asks about an
	// entity, not about a direction.
	tgt, err := s.RelationsAt(ctx, b.ID, 150)
	if err != nil || len(tgt) != 1 || tgt[0].Predicate != "ally_of" {
		t.Fatalf("target at 150: %+v err=%v", tgt, err)
	}
	if tgt[0].SourceID != a.ID || tgt[0].TargetID != b.ID {
		t.Fatalf("stored direction lost: %+v", tgt[0])
	}
	// Half-open: the low bound is inside, the high bound is outside.
	lo, err := s.RelationsAt(ctx, b.ID, 100)
	if err != nil || len(lo) != 1 {
		t.Fatalf("at 100: %+v err=%v", lo, err)
	}
	hi, err := s.RelationsAt(ctx, b.ID, 200)
	if err != nil || len(hi) != 0 {
		t.Fatalf("at 200: %+v err=%v", hi, err)
	}
}

// An empty validity is rejected by the schema, and CreateRelation must not
// leave the relation's entity row behind when the subclass insert fails.
func TestCreateRelationEmptyValidityRollsBack(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	uid := seedUser(t, s)
	w, _ := s.CreateWorld(ctx, uid, "W")
	npc, _ := s.CreateType(ctx, w.ID, nil, "entity", "npc", "NPC")
	ally, _ := s.CreateType(ctx, w.ID, nil, "relation", "ally_of", "Ally of")
	a, _ := s.CreateEntity(ctx, w.ID, npc.ID, "Alice", "")
	b, _ := s.CreateEntity(ctx, w.ID, npc.ID, "Bob", "")

	if _, err := s.CreateRelation(ctx, w.ID, ally.ID, a.ID, b.ID,
		worldtime.Range{Lo: 100, Hi: 100}); err == nil {
		t.Fatal("empty validity accepted")
	}
	var n int
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM world.entities
        WHERE world_id = $1 AND type_id = $2`, w.ID, ally.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("orphaned relation entity rows = %d", n)
	}
}
