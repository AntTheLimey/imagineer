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
-- Populates game systems and default relationship types

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

-- ============================================
-- Default Relationship Types
-- System-wide defaults (campaign_id = NULL)
-- ============================================

-- Ownership relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'owns', 'owned_by', false, 'Owns', 'Is owned by', 'Entity possesses or controls another entity'),
(NULL, 'owned_by', 'owns', false, 'Is owned by', 'Owns', 'Inverse of owns relationship');

-- Employment relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'employs', 'employed_by', false, 'Employs', 'Is employed by', 'Entity employs another as worker or servant'),
(NULL, 'employed_by', 'employs', false, 'Is employed by', 'Employs', 'Inverse of employs relationship'),
(NULL, 'works_for', 'employs', false, 'Works for', 'Employs', 'Alias for employed_by relationship');

-- Management relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'reports_to', 'manages', false, 'Reports to', 'Manages', 'Entity reports to another in organizational hierarchy'),
(NULL, 'manages', 'reports_to', false, 'Manages', 'Reports to', 'Entity manages or supervises another');

-- Family relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'parent_of', 'child_of', false, 'Parent of', 'Child of', 'Entity is parent of another'),
(NULL, 'child_of', 'parent_of', false, 'Child of', 'Parent of', 'Entity is child of another');

-- Location relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'located_at', 'contains', false, 'Located at', 'Contains', 'Entity is physically located at another location'),
(NULL, 'contains', 'located_at', false, 'Contains', 'Located at', 'Location contains another entity');

-- Membership relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'member_of', 'has_member', false, 'Member of', 'Has member', 'Entity is member of organization or faction'),
(NULL, 'has_member', 'member_of', false, 'Has member', 'Member of', 'Organization has entity as member');

-- Creation relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'created', 'created_by', false, 'Created', 'Created by', 'Entity created or made another entity'),
(NULL, 'created_by', 'created', false, 'Created by', 'Created', 'Entity was created by another');

-- Political relationships (asymmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'rules', 'ruled_by', false, 'Rules', 'Ruled by', 'Entity has political authority over another'),
(NULL, 'ruled_by', 'rules', false, 'Ruled by', 'Rules', 'Entity is under political authority of another');

-- Social relationships (symmetric)
INSERT INTO relationship_types (campaign_id, name, inverse_name, is_symmetric, display_label, inverse_display_label, description) VALUES
(NULL, 'knows', 'knows', true, 'Knows', 'Knows', 'Entity is acquainted with another'),
(NULL, 'friend_of', 'friend_of', true, 'Friend of', 'Friend of', 'Entity has friendly relationship with another'),
(NULL, 'enemy_of', 'enemy_of', true, 'Enemy of', 'Enemy of', 'Entity has hostile relationship with another'),
(NULL, 'allied_with', 'allied_with', true, 'Allied with', 'Allied with', 'Entity has alliance or partnership with another');

-- ============================================
-- Record migration
-- ============================================
INSERT INTO schema_migrations (version) VALUES ('002_seed_data');
