# Game Systems

Overview of TTRPG systems supported by Imagineer.

## Call of Cthulhu 7th Edition

### Overview

Lovecraftian horror investigation game using percentile dice.

### Characteristics

| Characteristic | Range | Description |
|---------------|-------|-------------|
| STR | 1-100 | Physical strength |
| CON | 1-100 | Health and stamina |
| SIZ | 1-100 | Physical size |
| DEX | 1-100 | Agility and reflexes |
| APP | 1-100 | Physical appearance |
| INT | 1-100 | Intelligence |
| POW | 1-100 | Willpower and magical aptitude |
| EDU | 1-100 | Education and knowledge |

### Derived Statistics

| Stat | Formula |
|------|---------|
| HP | (CON + SIZ) / 10 |
| MP | POW / 5 |
| SAN | POW |
| Luck | 3d6 × 5 |
| MOV | Based on STR, DEX, SIZ |

### Skills

Skills are percentile-based (0-99). Key categories:

- **Investigation**: Library Use, Spot Hidden, Listen
- **Social**: Persuade, Fast Talk, Intimidate
- **Combat**: Fighting, Firearms
- **Academic**: Archaeology, History, Occult

### Roll Mechanics

| Result | Roll vs Skill |
|--------|---------------|
| Critical | 01 |
| Extreme | ≤ skill/5 |
| Hard | ≤ skill/2 |
| Regular | ≤ skill |
| Failure | > skill |
| Fumble | 96-100 (if skill < 50) |

### Sanity System

- **Starting SAN**: Equal to POW
- **Maximum SAN**: 99 - Cthulhu Mythos skill
- **Insanity**: At 0 SAN or major loss

### Entity Attributes for CoC

```yaml
npc:
  occupation: string
  characteristics:
    str: number (1-100)
    con: number (1-100)
    # etc.
  skills:
    - name: string
      value: number
  sanity: number
```

## GURPS 4th Edition

### Overview

Generic Universal RolePlaying System using 3d6.

### Primary Attributes

| Attribute | Default | Cost |
|-----------|---------|------|
| ST | 10 | 10/level |
| DX | 10 | 20/level |
| IQ | 10 | 20/level |
| HT | 10 | 10/level |

### Secondary Characteristics

| Characteristic | Formula |
|---------------|---------|
| HP | ST |
| Will | IQ |
| Per | IQ |
| FP | HT |
| Basic Speed | (HT + DX) / 4 |
| Basic Move | Speed (drop fractions) |

### Skills

Difficulty levels:

| Difficulty | Cost Progression |
|------------|------------------|
| Easy | 1, 2, 4, 8... |
| Average | 1, 2, 4, 8... (starts at attribute-1) |
| Hard | 1, 2, 4, 8... (starts at attribute-2) |
| Very Hard | 1, 2, 4, 8... (starts at attribute-3) |

### Roll Mechanics

- Roll 3d6 ≤ skill to succeed
- 3-4: Critical success
- 17-18: Critical failure
- Margin of success/failure matters

### Advantages/Disadvantages

Point-buy system for traits:

- **Advantages**: Beneficial traits (cost points)
- **Disadvantages**: Limiting traits (give points)
- **Quirks**: Minor traits (±1 point each)

### Entity Attributes for GURPS

```yaml
npc:
  attributes:
    st: number
    dx: number
    iq: number
    ht: number
  advantages: [string]
  disadvantages: [string]
  skills:
    - name: string
      level: number
  point_total: number
```

## Forged in the Dark

### Overview

Heist/scoundrel games using d6 dice pools.

### Action Ratings

Grouped by attribute:

**Insight**:
- Hunt, Study, Survey, Tinker

**Prowess**:
- Finesse, Prowl, Skirmish, Wreck

**Resolve**:
- Attune, Command, Consort, Sway

### Dice Pool Mechanics

- Roll d6s equal to action rating
- Highest die determines result:
  - 6: Full success
  - 4-5: Partial success (complication)
  - 1-3: Failure (bad outcome)
- Critical: Two or more 6s

### Position & Effect

**Position** (how dangerous):
- Controlled: Safe, low stakes
- Risky: Standard danger
- Desperate: High danger

**Effect** (how impactful):
- Limited: Minor progress
- Standard: Normal progress
- Great: Major progress

### Stress & Trauma

- **Stress**: 0-9 track, avoid consequences
- **Trauma**: Permanent condition at stress-out
- **Harm**: Physical/mental injury levels

### Faction System

- **Tier**: Power level (0-5)
- **Status**: Relationship (-3 to +3)
- **Hold**: Weak or Strong

### Entity Attributes for FitD

```yaml
npc:
  crew_type: string
  playbook: string
  action_ratings:
    hunt: number (0-4)
    study: number (0-4)
    # etc.
  stress: number
  trauma: [string]
  harm:
    level_1: [string]
    level_2: [string]
    level_3: string

faction:
  tier: number (0-5)
  hold: "weak" | "strong"
  status_with_crew: number (-3 to +3)
```

## System-Agnostic Patterns

### Common Entity Attributes

All systems share:

- Name (required)
- Description
- Tags
- GM notes (GM-only)
- Source document
- Source confidence

### Validation Rules

1. Verify attributes match game system schema
2. Check skill values are in valid range
3. Validate derived statistics
4. Confirm required fields present

### Cross-System Considerations

When designing features:

- Don't hardcode system mechanics
- Use schema validation
- Support flexible attributes
- Allow system-specific extensions
