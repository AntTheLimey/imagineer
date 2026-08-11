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
	"os"
	"testing"

	coredb "github.com/antonypegg/imagineer/internal/core/db"
	"github.com/google/uuid"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL_V2")
	if url == "" {
		t.Skip("TEST_DATABASE_URL_V2 not set")
	}
	pool, err := coredb.Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	// Ensure the schema exists regardless of package run order: Migrate is
	// idempotent. (test-core also runs -p 1 so db tests' DROP SCHEMA can't
	// race this package.)
	if _, err := coredb.Migrate(context.Background(), pool,
		os.DirFS("../../../db/migrations")); err != nil {
		t.Fatal(err)
	}
	return New(pool)
}

// seedUser inserts a user and returns its id.
func seedUser(t *testing.T, s *Store) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := s.Pool.QueryRow(context.Background(), `INSERT INTO app.users
        (email, display_name)
        VALUES (concat(uuidv7()::text, '@x.io'), 'GM') RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestTypeAncestors(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	uid := seedUser(t, s) // helper: INSERT app.users, return uuid.UUID
	w, err := s.CreateWorld(ctx, uid, "Canticle")
	if err != nil {
		t.Fatal(err)
	}
	root, _ := s.CreateType(ctx, w.ID, nil, "entity", "thing", "Thing")
	person, _ := s.CreateType(ctx, w.ID, &root.ID, "entity", "person", "Person")
	npc, err := s.CreateType(ctx, w.ID, &person.ID, "entity", "npc", "NPC")
	if err != nil {
		t.Fatal(err)
	}
	anc, err := s.TypeAncestors(ctx, npc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(anc) != 2 || anc[0].Name != "person" || anc[1].Name != "thing" {
		t.Fatalf("ancestors = %+v", anc)
	}
}
