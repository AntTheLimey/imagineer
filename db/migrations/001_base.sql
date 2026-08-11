/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- 001_base.sql — namespaces, shared trigger, users.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE SCHEMA app;
CREATE SCHEMA world;
CREATE SCHEMA campaign;
CREATE SCHEMA ingest;

COMMENT ON SCHEMA app      IS 'Users, conversations, settings';
COMMENT ON SCHEMA world    IS 'The world plane: shared fictional canon';
COMMENT ON SCHEMA campaign IS 'The campaign plane: what a table did and knows';
COMMENT ON SCHEMA ingest   IS 'Provenance plane: sources, mentions, bindings';

CREATE FUNCTION core.touch_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TABLE app.users (
    id           UUID PRIMARY KEY DEFAULT uuidv7(),
    email        TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE app.users IS 'Accounts. Auth columns arrive in the API plan (argon2id — never MD5).';
CREATE TRIGGER touch BEFORE UPDATE ON app.users
    FOR EACH ROW EXECUTE FUNCTION core.touch_updated_at();
