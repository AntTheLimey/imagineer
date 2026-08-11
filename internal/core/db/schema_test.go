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
