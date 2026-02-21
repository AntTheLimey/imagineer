/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

-- Seed Data Migration
-- Populates game systems and default relationship type templates

-- ============================================
-- Game Systems
-- ============================================

-- Call of Cthulhu 7th Edition
INSERT INTO game_systems (name, code, attribute_schema, skill_schema, dice_conventions) VALUES (
    'Call of Cthulhu 7th Edition',
    'coc-7e',
    '{
        "characteristics": {
            "STR": {"name": "Strength", "type": "percentile", "range": [15, 90], "default": 50},
            "CON": {"name": "Constitution", "type": "percentile", "range": [15, 90], "default": 50},
            "SIZ": {"name": "Size", "type": "percentile", "range": [40, 90], "default": 65},
            "DEX": {"name": "Dexterity", "type": "percentile", "range": [15, 90], "default": 50},
            "APP": {"name": "Appearance", "type": "percentile", "range": [15, 90], "default": 50},
            "INT": {"name": "Intelligence", "type": "percentile", "range": [40, 90], "default": 65},
            "POW": {"name": "Power", "type": "percentile", "range": [15, 90], "default": 50},
            "EDU": {"name": "Education", "type": "percentile", "range": [40, 90], "default": 65}
        },
        "derived": {
            "HP": {"formula": "(CON + SIZ) / 10", "round": "down"},
            "MP": {"formula": "POW / 5", "round": "down"},
            "SAN": {"formula": "POW", "max_formula": "99 - Cthulhu_Mythos"},
            "Luck": {"formula": "3d6 * 5"}
        }
    }',
    '{
        "categories": ["Combat", "Communication", "Investigation", "Knowledge", "Physical", "Technical", "Unusual"],
        "sample_skills": {
            "Spot_Hidden": {"base": 25, "category": "Investigation"},
            "Library_Use": {"base": 20, "category": "Investigation"},
            "Listen": {"base": 20, "category": "Investigation"},
            "Psychology": {"base": 10, "category": "Investigation"},
            "Cthulhu_Mythos": {"base": 0, "category": "Unusual", "special": true},
            "Occult": {"base": 5, "category": "Knowledge"},
            "Fighting_Brawl": {"base": 25, "category": "Combat"},
            "Dodge": {"base": "DEX/2", "category": "Combat"}
        }
    }',
    '{
        "primary": "d100",
        "success_levels": [
            {"name": "Critical Success", "threshold": "01"},
            {"name": "Extreme Success", "threshold": "skill/5"},
            {"name": "Hard Success", "threshold": "skill/2"},
            {"name": "Regular Success", "threshold": "skill"},
            {"name": "Failure", "threshold": ">skill"},
            {"name": "Fumble", "threshold": "96-100"}
        ],
        "sanity_loss_format": "XdY/ZdW"
    }'
);

-- GURPS 4th Edition
INSERT INTO game_systems (name, code, attribute_schema, skill_schema, dice_conventions) VALUES (
    'GURPS 4th Edition',
    'gurps-4e',
    '{
        "primary_attributes": {
            "ST": {"name": "Strength", "base": 10, "cost_per_level": 10},
            "DX": {"name": "Dexterity", "base": 10, "cost_per_level": 20},
            "IQ": {"name": "Intelligence", "base": 10, "cost_per_level": 20},
            "HT": {"name": "Health", "base": 10, "cost_per_level": 10}
        },
        "secondary_characteristics": {
            "HP": {"base_formula": "ST", "cost_per_level": 2},
            "Will": {"base_formula": "IQ", "cost_per_level": 5},
            "Per": {"base_formula": "IQ", "cost_per_level": 5},
            "FP": {"base_formula": "HT", "cost_per_level": 3},
            "Basic_Speed": {"formula": "(HT + DX) / 4", "cost_per_quarter": 5},
            "Basic_Move": {"formula": "floor(Basic_Speed)", "cost_per_level": 5}
        },
        "point_costs": {
            "starting_points": {"heroic": 150, "cinematic": 250, "superheroic": 500},
            "disadvantage_limit": -75
        }
    }',
    '{
        "difficulty_levels": [
            {"code": "E", "name": "Easy", "base": "Attribute"},
            {"code": "A", "name": "Average", "base": "Attribute - 1"},
            {"code": "H", "name": "Hard", "base": "Attribute - 2"},
            {"code": "VH", "name": "Very Hard", "base": "Attribute - 3"}
        ],
        "cost_progression": [1, 2, 4, 8]
    }',
    '{
        "primary": "3d6",
        "success_roll": "3d6 <= effective skill",
        "margin_of_success": "skill - roll",
        "critical_success": ["3-4 always", "5 if skill >= 15", "6 if skill >= 16"],
        "critical_failure": ["18 always", "17 if skill <= 15", "10+ over skill"]
    }'
);

-- Forged in the Dark
INSERT INTO game_systems (name, code, attribute_schema, skill_schema, dice_conventions) VALUES (
    'Forged in the Dark',
    'fitd',
    '{
        "attributes": {
            "Insight": {"description": "Mental acuity, observation, technical skill"},
            "Prowess": {"description": "Physical capability, speed, strength"},
            "Resolve": {"description": "Determination, social skill, willpower"}
        },
        "action_ratings": {
            "Insight": ["Hunt", "Study", "Survey", "Tinker"],
            "Prowess": ["Finesse", "Prowl", "Skirmish", "Wreck"],
            "Resolve": ["Attune", "Command", "Consort", "Sway"]
        },
        "stress": {"max": 9},
        "trauma": {"max": 4}
    }',
    '{
        "actions": {
            "Hunt": "Track, ambush, attack from range",
            "Study": "Scrutinize details, research, analyze",
            "Survey": "Observe, search, understand circumstances",
            "Tinker": "Craft, repair, disable devices",
            "Finesse": "Employ dexterity, subtle maneuvering",
            "Prowl": "Traverse skillfully, quietly, quickly",
            "Skirmish": "Entangle a target in close combat",
            "Wreck": "Unleash savage force, demolish",
            "Attune": "Open mind to arcane power, spirits",
            "Command": "Compel obedience with authority",
            "Consort": "Socialize with friends, contacts",
            "Sway": "Influence with guile, charm, deception"
        }
    }',
    '{
        "primary": "d6 dice pool",
        "resolution": {
            "0_dice": "Roll 2d6, take lowest",
            "1+_dice": "Roll Xd6, take highest"
        },
        "results": {
            "6": "Full Success",
            "4-5": "Partial Success (success with consequence)",
            "1-3": "Bad Outcome",
            "critical": "Two or more 6s = Critical (enhanced effect)"
        },
        "position_levels": ["Controlled", "Risky", "Desperate"],
        "effect_levels": ["Zero", "Limited", "Standard", "Great", "Extreme"]
    }'
);

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
);

-- ============================================
-- Relationship Type Templates
-- System-wide defaults copied to each new campaign
-- ============================================

-- Asymmetric relationships (forward direction only)
INSERT INTO relationship_type_templates (name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
('owns', 'owned_by', false, 'Owns', 'Is owned by', 'Entity possesses or controls another entity'),
('employs', 'employed_by', false, 'Employs', 'Is employed by', 'Entity employs another as worker or servant'),
('works_for', 'employs', false, 'Works for', 'Employs', 'Alias for employed_by relationship'),
('reports_to', 'manages', false, 'Reports to', 'Manages', 'Entity reports to another in organizational hierarchy'),
('parent_of', 'child_of', false, 'Parent of', 'Child of', 'Entity is parent of another'),
('located_at', 'contains', false, 'Located at', 'Contains', 'Entity is physically located at another location'),
('member_of', 'has_member', false, 'Member of', 'Has member', 'Entity is member of organization or faction'),
('created', 'created_by', false, 'Created', 'Created by', 'Entity created or made another entity'),
('rules', 'ruled_by', false, 'Rules', 'Ruled by', 'Entity has political authority over another'),
('headquartered_at', 'headquarters_of', false, 'Headquartered at', 'Headquarters of', 'Entity has its primary base of operations at a location');

-- Symmetric relationships
INSERT INTO relationship_type_templates (name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
('knows', 'knows', true, 'Knows', 'Knows', 'Entity is acquainted with another'),
('friend_of', 'friend_of', true, 'Friend of', 'Friend of', 'Entity has friendly relationship with another'),
('enemy_of', 'enemy_of', true, 'Enemy of', 'Enemy of', 'Entity has hostile relationship with another'),
('allied_with', 'allied_with', true, 'Allied with', 'Allied with', 'Entity has alliance or partnership with another');

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('002_seed_data');
