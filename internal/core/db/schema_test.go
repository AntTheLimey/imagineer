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
