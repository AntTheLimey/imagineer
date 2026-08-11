/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package store implements the Go data-access layer over the core schema
// (worlds, type tree, and related domain tables).
package store

import "github.com/jackc/pgx/v5/pgxpool"

// Store wraps a connection pool and exposes domain-level queries.
type Store struct {
	Pool *pgxpool.Pool
}

// New returns a Store backed by pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{Pool: pool}
}
