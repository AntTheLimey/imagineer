/*-------------------------------------------------------------------------
 *
 * 004_add_other_game_system.sql
 *
 * Adds a generic "Other / Homebrew" game system for campaigns
 * using unsupported or custom rule sets.
 *
 *-------------------------------------------------------------------------
 */

INSERT INTO game_systems (name, code, attribute_schema, skill_schema, dice_conventions)
VALUES (
    'Other / Homebrew',
    'other',
    '{}',
    '{}',
    '{}'
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO schema_migrations (version) VALUES ('004_add_other_game_system');
