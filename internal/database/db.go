/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package database provides PostgreSQL database access for Imagineer.
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/antonypegg/imagineer/internal/crypto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config represents database configuration.
type Config struct {
	Cluster  string       `json:"cluster"`
	Database string       `json:"database"`
	Nodes    []NodeConfig `json:"nodes"`
	Users    []UserConfig `json:"users"`
}

// NodeConfig represents a database node configuration.
type NodeConfig struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

// UserConfig represents a database user configuration.
type UserConfig struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Superuser bool   `json:"superuser"`
}

// DB wraps a pgxpool.Pool with helper methods.
type DB struct {
	Pool      *pgxpool.Pool
	Encryptor *crypto.Encryptor // nil = no encryption
}

// LoadConfig reads the database configuration from a JSON file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Connect establishes a connection pool to the database.
func Connect(ctx context.Context, config *Config) (*DB, error) {
	if len(config.Nodes) == 0 {
		return nil, fmt.Errorf("no database nodes configured")
	}
	if len(config.Users) == 0 {
		return nil, fmt.Errorf("no database users configured")
	}

	node := config.Nodes[0]
	user := config.Users[0]

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user.Username,
		user.Password,
		node.Hostname,
		node.Port,
		config.Database,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// QueryRow executes a query that returns a single row.
func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

// Query executes a query that returns multiple rows.
func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.Pool.Query(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows.
func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := db.Pool.Exec(ctx, sql, args...)
	return err
}
