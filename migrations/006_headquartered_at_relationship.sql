/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 006: Add headquartered_at relationship type
--
-- This migration adds a system-default relationship type pair for
-- indicating that an entity (typically an organisation or faction) has
-- its primary base of operations at a specific location.

-- ============================================
-- Headquarters relationship (asymmetric)
-- ============================================
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'headquartered_at', 'headquarters_of', false, 'Headquartered at', 'Headquarters of', 'Entity has its primary base of operations at a location'),
(NULL, 'headquarters_of', 'headquartered_at', false, 'Headquarters of', 'Headquartered at', 'Location serves as the primary base of operations for an entity');

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('006_headquartered_at_relationship');
