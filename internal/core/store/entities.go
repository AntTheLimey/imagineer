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

	"github.com/antonypegg/imagineer/internal/core/worldtime"
	"github.com/google/uuid"
)

// Entity is a row of world.entities.
type Entity struct {
	ID          uuid.UUID
	WorldID     uuid.UUID
	TypeID      uuid.UUID
	Name        string
	Description string
	CanonStatus string
}

// Relation is a row of world.entity_relations joined to its predicate type.
// The stored direction is source -> target; inverses are presentation only.
type Relation struct {
	ID          uuid.UUID
	SourceID    uuid.UUID
	TargetID    uuid.UUID
	PredicateID uuid.UUID
	Predicate   string
}

// CreateEntity inserts an entity of the given type into a world.
func (s *Store) CreateEntity(ctx context.Context, worldID, typeID uuid.UUID,
	name, description string) (Entity, error) {
	var e Entity
	err := s.Pool.QueryRow(ctx, `INSERT INTO world.entities
        (world_id, type_id, name, description) VALUES ($1,$2,$3,$4)
        RETURNING id, world_id, type_id, name, coalesce(description,''), canon_status`,
		worldID, typeID, name, description).
		Scan(&e.ID, &e.WorldID, &e.TypeID, &e.Name, &e.Description, &e.CanonStatus)
	return e, err
}

// CreateRelation records that source stands in the predicate relation to
// target for the half-open hour interval validity. A relation is an entity, so
// both the world.entities row (typed by the predicate) and the
// world.entity_relations subclass row are written in one transaction: either
// the edge exists in full or not at all.
//
// An empty validity (Lo == Hi) is rejected by the validity_not_empty
// constraint; that is deliberate, since an empty multirange overlaps nothing
// and would slip past the non-overlap trigger.
func (s *Store) CreateRelation(ctx context.Context, worldID, predicateID,
	sourceID, targetID uuid.UUID, validity worldtime.Range) (uuid.UUID, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var id uuid.UUID
	if err := tx.QueryRow(ctx, `INSERT INTO world.entities (world_id, type_id, name)
        VALUES ($1,$2,'(relation)') RETURNING id`, worldID, predicateID).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO world.entity_relations
        (id, source_id, target_id, predicate_id, validity)
        VALUES ($1,$2,$3,$4, nummultirange(numrange($5,$6)))`,
		id, sourceID, targetID, predicateID, validity.Lo, validity.Hi); err != nil {
		return uuid.Nil, err
	}
	return id, tx.Commit(ctx)
}

// RelationsAt returns every relation touching entityID — as either source or
// target — whose validity contains atHours. This is the time-scrubber query.
//
// atHours is cast to numeric in SQL so that PostgreSQL resolves the parameter
// type explicitly; pgx marshals the int64 into that numeric without help.
func (s *Store) RelationsAt(ctx context.Context, entityID uuid.UUID,
	atHours int64) ([]Relation, error) {
	rows, err := s.Pool.Query(ctx, `
        SELECT r.id, r.source_id, r.target_id, r.predicate_id, t.name
        FROM world.entity_relations r
        JOIN world.types t ON t.id = r.predicate_id
        WHERE (r.source_id = $1 OR r.target_id = $1)
          AND r.validity @> $2::numeric`, entityID, atHours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Relation
	for rows.Next() {
		var r Relation
		if err := rows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.PredicateID, &r.Predicate); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
