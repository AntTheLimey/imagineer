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
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL_V2")
	if url == "" {
		t.Skip("TEST_DATABASE_URL_V2 not set")
	}
	pool, err := Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestMigrateAppliesOnceInOrder(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	pool.Exec(ctx, `DROP TABLE IF EXISTS mig_probe; DROP SCHEMA IF EXISTS core CASCADE`)

	dir := fstest.MapFS{
		"001_a.sql": {Data: []byte(`CREATE TABLE mig_probe (n INT);`)},
		"002_b.sql": {Data: []byte(`INSERT INTO mig_probe VALUES (1);`)},
	}
	applied, err := Migrate(ctx, pool, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 || applied[0] != "001_a.sql" {
		t.Fatalf("applied = %v", applied)
	}
	// Second run applies nothing.
	applied, err = Migrate(ctx, pool, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Fatalf("re-run applied = %v", applied)
	}
}
