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

	"github.com/google/uuid"
)

// World is a row of world.worlds.
type World struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	IsTemplate bool
	ForkedFrom *uuid.UUID
}

// TypeNode is a row of world.types.
type TypeNode struct {
	ID           uuid.UUID
	WorldID      uuid.UUID
	ParentID     *uuid.UUID
	Kind         string
	Name         string
	DisplayLabel string
}

// CreateWorld inserts a new world owned by ownerID.
func (s *Store) CreateWorld(ctx context.Context, ownerID uuid.UUID, name string) (World, error) {
	var w World
	err := s.Pool.QueryRow(ctx, `INSERT INTO world.worlds (owner_id, name)
        VALUES ($1,$2) RETURNING id, owner_id, name, is_template`, ownerID, name).
		Scan(&w.ID, &w.OwnerID, &w.Name, &w.IsTemplate)
	return w, err
}

// CreateType inserts a node into a world's type tree. parent is nil for a
// root node.
func (s *Store) CreateType(ctx context.Context, worldID uuid.UUID, parent *uuid.UUID,
	kind, name, label string) (TypeNode, error) {
	var n TypeNode
	err := s.Pool.QueryRow(ctx, `INSERT INTO world.types
        (world_id, parent_id, kind, name, display_label)
        VALUES ($1,$2,$3,$4,$5)
        RETURNING id, world_id, parent_id, kind, name, display_label`,
		worldID, parent, kind, name, label).
		Scan(&n.ID, &n.WorldID, &n.ParentID, &n.Kind, &n.Name, &n.DisplayLabel)
	return n, err
}

// TypeAncestors returns the ancestor chain of typeID, nearest first.
func (s *Store) TypeAncestors(ctx context.Context, typeID uuid.UUID) ([]TypeNode, error) {
	rows, err := s.Pool.Query(ctx, `
        WITH RECURSIVE up AS (
            SELECT t.*, 0 AS depth FROM world.types t WHERE t.id = $1
            UNION ALL
            SELECT p.*, up.depth + 1 FROM world.types p
            JOIN up ON p.id = up.parent_id
        )
        SELECT id, world_id, parent_id, kind, name, display_label
        FROM up WHERE depth > 0 ORDER BY depth`, typeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TypeNode
	for rows.Next() {
		var n TypeNode
		if err := rows.Scan(&n.ID, &n.WorldID, &n.ParentID, &n.Kind, &n.Name, &n.DisplayLabel); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
