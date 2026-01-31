/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"context"
	"fmt"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/google/uuid"
)

// CreateUser creates a new user in the database.
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	query := `
		INSERT INTO users (id, google_id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	err := db.QueryRow(ctx, query,
		user.ID, user.GoogleID, user.Email, user.Name, user.AvatarURL,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their UUID.
func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, google_id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u models.User
	err := db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.AvatarURL,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &u, nil
}

// GetUserByGoogleID retrieves a user by their Google OAuth ID.
func (db *DB) GetUserByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := `
		SELECT id, google_id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE google_id = $1`

	var u models.User
	err := db.QueryRow(ctx, query, googleID).Scan(
		&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.AvatarURL,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Google ID: %w", err)
	}

	return &u, nil
}

// GetUserByEmail retrieves a user by their email address.
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, google_id, email, name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1`

	var u models.User
	err := db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.AvatarURL,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &u, nil
}

// UpdateUser updates an existing user in the database.
func (db *DB) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET google_id = $2, email = $3, name = $4, avatar_url = $5
		WHERE id = $1
		RETURNING updated_at`

	err := db.QueryRow(ctx, query,
		user.ID, user.GoogleID, user.Email, user.Name, user.AvatarURL,
	).Scan(&user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
