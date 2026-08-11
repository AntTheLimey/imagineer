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

// seedTypedWorld returns (worldID, npcTypeID).
func seedTypedWorld(t *testing.T, pool *pgxpool.Pool) (string, string) {
	t.Helper()
	wid := seedWorld(t, pool)
	var tid string
	if err := pool.QueryRow(context.Background(), `INSERT INTO world.types
        (world_id, kind, name, display_label) VALUES ($1,'entity','npc','NPC')
        RETURNING id`, wid).Scan(&tid); err != nil {
		t.Fatal(err)
	}
	return wid, tid
}

// seedRelWorld returns (wid, npcType, predType, aliceID, bobID).
func seedRelWorld(t *testing.T, pool *pgxpool.Pool) (wid, npc, pred, a, b string) {
	t.Helper()
	ctx := context.Background()
	wid, npc = seedTypedWorld(t, pool)
	if err := pool.QueryRow(ctx, `INSERT INTO world.types (world_id, kind, name, display_label)
        VALUES ($1,'relation','ally_of','Ally of') RETURNING id`, wid).Scan(&pred); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
        VALUES ($1,$2,'Alice') RETURNING id`, wid, npc).Scan(&a); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
        VALUES ($1,$2,'Bob') RETURNING id`, wid, npc).Scan(&b); err != nil {
		t.Fatal(err)
	}
	return
}

// addRelation creates the entity row + subclass row in one tx; returns relation id.
// lo/hi are hours; the interval is [lo, hi).
func addRelation(t *testing.T, pool *pgxpool.Pool, wid, pred, src, tgt string, lo, hi int64) (string, error) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var relType, id string
	if err := tx.QueryRow(ctx,
		`SELECT id FROM world.types WHERE id = $1 AND kind = 'relation'`, pred).Scan(&relType); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
        VALUES ($1,$2,'(relation)') RETURNING id`, wid, relType).Scan(&id); err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `INSERT INTO world.entity_relations
        (id, source_id, target_id, predicate_id, validity)
        VALUES ($1,$2,$3,$4, nummultirange(numrange($5,$6)))`,
		id, src, tgt, pred, lo, hi)
	if err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

func TestRelationRecurrenceAndOverlap(t *testing.T) {
	pool := migratedPool(t)
	wid, _, pred, a, b := seedRelWorld(t, pool)

	// First interval: hours [0, 100).
	if _, err := addRelation(t, pool, wid, pred, a, b, 0, 100); err != nil {
		t.Fatal(err)
	}
	// Recurrence across a disjoint interval [200, 300) is ALLOWED.
	if _, err := addRelation(t, pool, wid, pred, a, b, 200, 300); err != nil {
		t.Fatalf("disjoint recurrence rejected: %v", err)
	}
	// Overlapping duplicate [50, 150) is REJECTED.
	if _, err := addRelation(t, pool, wid, pred, a, b, 50, 150); err == nil {
		t.Fatal("overlapping duplicate allowed")
	}
}

func TestNoSelfRelation(t *testing.T) {
	pool := migratedPool(t)
	wid, _, pred, a, _ := seedRelWorld(t, pool)
	if _, err := addRelation(t, pool, wid, pred, a, a, 0, 10); err == nil {
		t.Fatal("self-relation allowed")
	}
}

func TestRelationOfRelation(t *testing.T) {
	// A relation is an entity: it can be the target of another relation.
	pool := migratedPool(t)
	wid, _, pred, a, b := seedRelWorld(t, pool)
	rel, err := addRelation(t, pool, wid, pred, a, b, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := addRelation(t, pool, wid, pred, a, rel, 0, 10); err != nil {
		t.Fatalf("relation-of-relation rejected: %v", err)
	}
}

// Two overlapping same-edge rows inserted inside ONE transaction must still
// fail: the advisory lock only serializes across transactions, so the check
// itself has to catch the earlier same-txn row (which is visible to it).
func TestOverlapWithinSingleTransaction(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	wid, _, pred, a, b := seedRelWorld(t, pool)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	insert := func(lo, hi int64) error {
		var id string
		if err := tx.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
            VALUES ($1,$2,'(relation)') RETURNING id`, wid, pred).Scan(&id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO world.entity_relations
            (id, source_id, target_id, predicate_id, validity)
            VALUES ($1,$2,$3,$4, nummultirange(numrange($5,$6)))`, id, a, b, pred, lo, hi)
		return err
	}
	if err := insert(0, 100); err != nil {
		t.Fatal(err)
	}
	if err := insert(50, 150); err == nil {
		t.Fatal("overlapping same-transaction insert allowed")
	}
}

// The unbounded default covers the whole world line, so any second row on the
// same edge overlaps it.
func TestRelationDefaultValidityIsUnbounded(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	wid, _, pred, a, b := seedRelWorld(t, pool)

	var id string
	if err := pool.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
        VALUES ($1,$2,'(relation)') RETURNING id`, wid, pred).Scan(&id); err != nil {
		t.Fatal(err)
	}
	var validity string
	if err := pool.QueryRow(ctx, `INSERT INTO world.entity_relations
        (id, source_id, target_id, predicate_id) VALUES ($1,$2,$3,$4)
        RETURNING validity::text`, id, a, b, pred).Scan(&validity); err != nil {
		t.Fatal(err)
	}
	if validity != "{(,)}" {
		t.Fatalf("default validity = %q, want {(,)}", validity)
	}
	if _, err := addRelation(t, pool, wid, pred, a, b, 900, 1000); err == nil {
		t.Fatal("row overlapping the unbounded default allowed")
	}
}

// Widening an existing interval into a sibling's span must be rejected too.
func TestRelationUpdateOverlapRejected(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	wid, _, pred, a, b := seedRelWorld(t, pool)

	first, err := addRelation(t, pool, wid, pred, a, b, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := addRelation(t, pool, wid, pred, a, b, 200, 300); err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `UPDATE world.entity_relations
        SET validity = nummultirange(numrange(0,250)) WHERE id = $1`, first)
	if err == nil {
		t.Fatal("update widening into a sibling interval allowed")
	}
}

func TestEntities(t *testing.T) {
	pool := migratedPool(t)
	ctx := context.Background()
	wid, tid := seedTypedWorld(t, pool)

	var eid, status string
	err := pool.QueryRow(ctx, `INSERT INTO world.entities
        (world_id, type_id, name, description)
        VALUES ($1,$2,'Nell','Thief from Southwark')
        RETURNING id, canon_status`, wid, tid).Scan(&eid, &status)
	if err != nil {
		t.Fatal(err)
	}
	if status != "draft" {
		t.Fatalf("default canon_status = %q, want draft", status)
	}
	// Bad status rejected.
	if _, err := pool.Exec(ctx,
		`UPDATE world.entities SET canon_status = 'confirmed' WHERE id = $1`, eid); err == nil {
		t.Fatal("invalid canon_status allowed")
	}
}
