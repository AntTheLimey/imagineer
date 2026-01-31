/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account authenticated via Google OAuth.
type User struct {
	ID        uuid.UUID `json:"id"`
	GoogleID  string    `json:"googleId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
