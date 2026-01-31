/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Multi-User Foundation Migration
-- Creates users table and adds ownership to campaigns

-- ============================================
-- Users Table
-- Stores user accounts authenticated via Google OAuth
-- ============================================
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    google_id  TEXT UNIQUE NOT NULL,
    email      TEXT UNIQUE NOT NULL,
    name       TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE users IS 'User accounts authenticated via Google OAuth';
COMMENT ON COLUMN users.google_id IS 'Unique identifier from Google OAuth provider';
COMMENT ON COLUMN users.email IS 'User email address from Google account';
COMMENT ON COLUMN users.name IS 'Display name from Google account';
COMMENT ON COLUMN users.avatar_url IS 'URL to user profile image from Google';

-- ============================================
-- Add owner_id to campaigns
-- Links campaigns to their owning user
-- ============================================
ALTER TABLE campaigns
    ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE SET NULL;

COMMENT ON COLUMN campaigns.owner_id IS 'User who owns this campaign (nullable for legacy data)';

CREATE INDEX idx_campaigns_owner_id ON campaigns(owner_id);

-- ============================================
-- Apply updated_at trigger to users table
-- ============================================
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('003_add_users');
