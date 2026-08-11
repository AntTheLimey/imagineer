/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migratedPool: drop-and-rebuild the full schema from db/migrations.
func migratedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testPool(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
        DROP SCHEMA IF EXISTS app, world, campaign, ingest, core CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Migrate(ctx, pool, os.DirFS("../../../db/migrations")); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestBaseSchemas(t *testing.T) {
	pool := migratedPool(t)
	var n int
	err := pool.QueryRow(context.Background(), `
        SELECT count(*) FROM information_schema.schemata
        WHERE schema_name IN ('app','world','campaign','ingest')`).Scan(&n)
	if err != nil || n != 4 {
		t.Fatalf("schemas found = %d, err = %v", n, err)
	}
}

func TestOnePrimaryCalendarPerWorld(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	var uid, wid string
	pool.QueryRow(ctx, `INSERT INTO app.users (email, display_name)
        VALUES ('gm@example.com','Ant') RETURNING id`).Scan(&uid)
	pool.QueryRow(ctx, `INSERT INTO world.worlds (owner_id, name)
        VALUES ($1,'Canticle') RETURNING id`, uid).Scan(&wid)
	if _, err := pool.Exec(ctx, `INSERT INTO world.calendars
        (world_id, is_primary, epoch_label) VALUES ($1, true, 'Rata Die')`, wid); err != nil {
		t.Fatal(err)
	}
	_, err := pool.Exec(ctx, `INSERT INTO world.calendars
        (world_id, is_primary, epoch_label) VALUES ($1, true, 'Other')`, wid)
	if err == nil {
		t.Fatal("second primary calendar was allowed")
	}
}

// seedWorld inserts a user+world and returns the world id.
func seedWorld(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	var uid, wid string
	pool.QueryRow(ctx, `INSERT INTO app.users (email, display_name)
        VALUES (concat(uuidv7()::text,'@x.io'),'GM') RETURNING id`).Scan(&uid)
	if err := pool.QueryRow(ctx, `INSERT INTO world.worlds (owner_id, name)
        VALUES ($1,'W') RETURNING id`, uid).Scan(&wid); err != nil {
		t.Fatal(err)
	}
	return wid
}

func TestTypeTree(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	wid := seedWorld(t, pool)

	var root, child string
	// Root type: parent_id NULL must be allowed (the v3 correction).
	if err := pool.QueryRow(ctx, `INSERT INTO world.types
        (world_id, kind, name, display_label) VALUES ($1,'entity','thing','Thing')
        RETURNING id`, wid).Scan(&root); err != nil {
		t.Fatal(err)
	}
	pool.QueryRow(ctx, `INSERT INTO world.types
        (world_id, parent_id, kind, name, display_label)
        VALUES ($1,$2,'entity','npc','NPC') RETURNING id`, wid, root).Scan(&child)

	// Case-insensitive sibling name collision rejected.
	if _, err := pool.Exec(ctx, `INSERT INTO world.types
        (world_id, parent_id, kind, name, display_label)
        VALUES ($1,$2,'entity','NPC','Dup')`, wid, root); err == nil {
		t.Fatal("duplicate sibling name allowed")
	}
	// Duplicate root name (parent_id NULL) rejected: NULLS NOT DISTINCT
	// on the expression index must treat two NULL parents as colliding.
	if _, err := pool.Exec(ctx, `INSERT INTO world.types
        (world_id, kind, name, display_label)
        VALUES ($1,'entity','thing','Thing 2')`, wid); err == nil {
		t.Fatal("duplicate root name allowed")
	}
	// Cycle rejected: root cannot become child of its descendant.
	if _, err := pool.Exec(ctx,
		`UPDATE world.types SET parent_id = $1 WHERE id = $2`, child, root); err == nil {
		t.Fatal("cycle allowed")
	}
}
