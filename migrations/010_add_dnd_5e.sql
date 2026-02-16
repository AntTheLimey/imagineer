/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Migration 010: Add D&D 5th Edition (2024 Revision)
-- Adds the 2024 revised D&D 5e game system to the game_systems table.

-- ============================================
-- Game Systems
-- ============================================

-- D&D 5th Edition (2024 Revision)
INSERT INTO game_systems (name, code, attribute_schema, skill_schema, dice_conventions) VALUES (
    'D&D 5th Edition (2024 Revision)',
    'dnd-5e-2024',
    '{
        "ability_scores": {
            "STR": {"name": "Strength", "type": "standard", "range": [1, 30], "default": 10},
            "DEX": {"name": "Dexterity", "type": "standard", "range": [1, 30], "default": 10},
            "CON": {"name": "Constitution", "type": "standard", "range": [1, 30], "default": 10},
            "INT": {"name": "Intelligence", "type": "standard", "range": [1, 30], "default": 10},
            "WIS": {"name": "Wisdom", "type": "standard", "range": [1, 30], "default": 10},
            "CHA": {"name": "Charisma", "type": "standard", "range": [1, 30], "default": 10}
        },
        "derived_attributes": {
            "AC": {"name": "Armor Class", "formula": "10 + DEX modifier"},
            "HP": {"name": "Hit Points", "formula": "hit_die + CON modifier at 1st level"},
            "Initiative": {"name": "Initiative", "formula": "DEX modifier"},
            "Proficiency_Bonus": {"name": "Proficiency Bonus", "formula": "2 + floor((level - 1) / 4)"},
            "Spell_Save_DC": {"name": "Spell Save DC", "formula": "8 + proficiency_bonus + spellcasting_modifier"},
            "Passive_Perception": {"name": "Passive Perception", "formula": "10 + WIS modifier + proficiency (if proficient)"}
        },
        "classes": {
            "Barbarian": {"hit_die": "d12", "primary_ability": "STR", "saving_throws": ["STR", "CON"]},
            "Bard": {"hit_die": "d8", "primary_ability": "CHA", "saving_throws": ["DEX", "CHA"]},
            "Cleric": {"hit_die": "d8", "primary_ability": "WIS", "saving_throws": ["WIS", "CHA"]},
            "Druid": {"hit_die": "d8", "primary_ability": "WIS", "saving_throws": ["INT", "WIS"]},
            "Fighter": {"hit_die": "d10", "primary_ability": "STR", "saving_throws": ["STR", "CON"]},
            "Monk": {"hit_die": "d8", "primary_ability": "DEX", "saving_throws": ["STR", "DEX"]},
            "Paladin": {"hit_die": "d10", "primary_ability": "STR", "saving_throws": ["WIS", "CHA"]},
            "Ranger": {"hit_die": "d10", "primary_ability": "DEX", "saving_throws": ["STR", "DEX"]},
            "Rogue": {"hit_die": "d8", "primary_ability": "DEX", "saving_throws": ["DEX", "INT"]},
            "Sorcerer": {"hit_die": "d6", "primary_ability": "CHA", "saving_throws": ["CON", "CHA"]},
            "Warlock": {"hit_die": "d8", "primary_ability": "CHA", "saving_throws": ["WIS", "CHA"]},
            "Wizard": {"hit_die": "d6", "primary_ability": "INT", "saving_throws": ["INT", "WIS"]}
        }
    }',
    '{
        "categories": ["STR-based", "DEX-based", "INT-based", "WIS-based", "CHA-based"],
        "skills": {
            "Athletics": {"ability": "STR", "base": 0},
            "Acrobatics": {"ability": "DEX", "base": 0},
            "Sleight_of_Hand": {"ability": "DEX", "base": 0},
            "Stealth": {"ability": "DEX", "base": 0},
            "Arcana": {"ability": "INT", "base": 0},
            "History": {"ability": "INT", "base": 0},
            "Investigation": {"ability": "INT", "base": 0},
            "Nature": {"ability": "INT", "base": 0},
            "Religion": {"ability": "INT", "base": 0},
            "Animal_Handling": {"ability": "WIS", "base": 0},
            "Insight": {"ability": "WIS", "base": 0},
            "Medicine": {"ability": "WIS", "base": 0},
            "Perception": {"ability": "WIS", "base": 0},
            "Survival": {"ability": "WIS", "base": 0},
            "Deception": {"ability": "CHA", "base": 0},
            "Intimidation": {"ability": "CHA", "base": 0},
            "Performance": {"ability": "CHA", "base": 0},
            "Persuasion": {"ability": "CHA", "base": 0}
        }
    }',
    '{
        "primary": "d20",
        "damage": ["d4", "d6", "d8", "d10", "d12"],
        "advantage_disadvantage": "Roll 2d20, take highest or lowest",
        "difficulty_classes": {
            "Very_Easy": 5,
            "Easy": 10,
            "Medium": 15,
            "Hard": 20,
            "Very_Hard": 25,
            "Nearly_Impossible": 30
        }
    }'
) ON CONFLICT (code) DO NOTHING;

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('010_add_dnd_5e');
