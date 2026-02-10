<!--
  Imagineer - TTRPG Campaign Intelligence Platform

  Copyright (c) 2025 - 2026
  This software is released under The MIT License
-->

# Relationship Model Refactor: Single-Edge LPG Design

This design document describes the refactoring of the entity
relationship model from a dual-edge pattern to a single-edge
Labeled Property Graph (LPG) pattern. The refactor eliminates
duplicate storage, simplifies queries, reduces enrichment noise,
and aligns the data model with graph database best practices.

## Context and Motivation

The current relationship model stores two rows per logical
relationship: a forward edge and its inverse. For example,
"Viktor located_in Silver Fox Inn" creates two rows:

- Row 1 stores Viktor as the source with `located_in` as the
  type and Silver Fox Inn as the target.
- Row 2 stores Silver Fox Inn as the source with `contains` as
  the type and Viktor as the target.

This dual-edge approach causes four problems:

- Deduplication logic is required in queries to avoid displaying
  both directions of the same relationship.
- Two rows per relationship double both the storage footprint and
  the write cost.
- The LLM enrichment pipeline suggests inverse relationships that
  already exist, creating duplicate review items in the triage
  interface.
- Deleting or updating one row without the other leaves the data
  in an inconsistent state.

Graph design best practice (per Neo4j, GraphAware, and the
broader LPG community) recommends storing one edge per
relationship. The inverse verb belongs as metadata on the
relationship type, not as a separate row. A single-edge model
allows traversal in both directions at equal speed.

## Design Changes

This section describes each change required to move to the
single-edge model.

### Single-Edge Storage

The system stores one row per relationship. The
`relationship_types` table defines how to read the relationship
from either direction.

When the system creates "Viktor is located_in Silver Fox Inn",
the `relationships` table receives a single row:

- `source_entity_id` references Viktor.
- `target_entity_id` references Silver Fox Inn.
- `relationship_type_id` references the `located_in` type.

When viewing Silver Fox Inn's relationships, the system looks up
the relationship type, finds `inverse_name = "contains"`, and
displays "contains Viktor" without needing a second row.

### Numeric Foreign Key for Relationship Type

The current `relationships.relationship_type` column stores type
names as `TEXT` with no referential integrity. The refactor
replaces this column with
`relationship_type_id BIGINT NOT NULL REFERENCES
relationship_types(id)`.

This change provides several benefits:

- The database enforces referential integrity through a proper
  foreign key constraint.
- Integer comparisons replace text matching in joins.
- Renaming a type requires updating only one row in the
  `relationship_types` table.

### Template Pattern for Campaign-Scoped Types

The current design uses `campaign_id = NULL` for system defaults
and requires
`(rt.campaign_id = r.campaign_id OR rt.campaign_id IS NULL)` in
every query. The refactor introduces a
`relationship_type_templates` table for system defaults. When the
application creates a campaign, the system copies all templates
into `relationship_types` with the new campaign's ID.

This approach provides the following benefits:

- Each campaign has its own isolated set of types with no shared
  rows.
- The `campaign_id` column is always `NOT NULL`, eliminating null
  handling.
- The foreign key from `relationships.relationship_type_id` to
  `relationship_types.id` requires no fallback logic.
- Adding new template types does not affect existing campaigns.
- Campaigns can rename, delete, or add custom types freely.
- Queries no longer require
  `OR campaign_id IS NULL` fallback clauses.

### Deduplicate relationship_types Rows

The current seed data stores both directions as separate rows.
For example, both `owns` (with `inverse_name = "owned_by"`) and
`owned_by` (with `inverse_name = "owns"`) exist as distinct
rows. With single-edge storage, each relationship pair needs only
one canonical row that stores the forward name and the inverse
name.

The current 26 rows (13 pairs) collapse to 13 canonical rows:

- `owns` / `owned_by` (2 rows become 1 row with
  `name = "owns"`, `inverse_name = "owned_by"`).
- `employs` / `employed_by` (2 rows become 1 row; `works_for`
  becomes a separate type if needed).
- `reports_to` / `manages` becomes 1 row.
- `parent_of` / `child_of` becomes 1 row.
- `located_at` / `contains` becomes 1 row.
- `member_of` / `has_member` becomes 1 row.
- `created` / `created_by` becomes 1 row.
- `rules` / `ruled_by` becomes 1 row.
- `knows` (symmetric) becomes 1 row.
- `friend_of` (symmetric) becomes 1 row.
- `enemy_of` (symmetric) becomes 1 row.
- `allied_with` (symmetric) becomes 1 row.
- `headquartered_at` / `headquarters_of` (from migration 006)
  becomes 1 row.

### Remove the bidirectional Column

The refactor removes the `bidirectional` boolean column from the
`relationships` table. Directionality is entirely driven by the
`relationship_types` table:

- When `is_symmetric = true`, the system uses the same verb in
  both directions (for example, "knows").
- When `is_symmetric = false`, the system uses `inverse_name`
  when reading from the target side.

Every relationship type has an inverse verb. Even one-way-feeling
relationships like "worships" receive an inverse such as
"worshipped_by". The system always traverses relationships from
both sides; only the displayed verbs differ.

### Database Trigger: Prevent Inverse Relationships

A `BEFORE INSERT OR UPDATE` trigger on `relationships` prevents
creating an inverse of an existing relationship. The trigger
raises an exception rather than silently skipping the insert, and
the application layer catches the exception and handles the
situation gracefully (for example, filtering enrichment
suggestions that would create inverse duplicates).

The following function implements the trigger logic:

```sql
CREATE OR REPLACE FUNCTION prevent_inverse_relationship()
RETURNS TRIGGER AS $$
BEGIN
    -- For symmetric types: block the same type with
    -- swapped entities
    IF EXISTS (
        SELECT 1 FROM relationships r
        JOIN relationship_types rt
            ON rt.id = NEW.relationship_type_id
        WHERE rt.is_symmetric = true
        AND r.campaign_id = NEW.campaign_id
        AND r.source_entity_id = NEW.target_entity_id
        AND r.target_entity_id = NEW.source_entity_id
        AND r.relationship_type_id
            = NEW.relationship_type_id
        AND (TG_OP = 'INSERT' OR r.id != NEW.id)
    ) THEN
        RAISE EXCEPTION
            'Symmetric inverse relationship already exists';
    END IF;

    -- For asymmetric types: block if the inverse type
    -- exists with swapped entities. Find the inverse
    -- type: if NEW uses "located_at", find "contains".
    -- The inverse is the type whose name matches the
    -- current type's inverse_name.
    IF EXISTS (
        SELECT 1 FROM relationships r
        JOIN relationship_types rt_new
            ON rt_new.id = NEW.relationship_type_id
        JOIN relationship_types rt_inv
            ON rt_inv.campaign_id = rt_new.campaign_id
            AND rt_inv.name = rt_new.inverse_name
        WHERE rt_new.is_symmetric = false
        AND r.campaign_id = NEW.campaign_id
        AND r.source_entity_id = NEW.target_entity_id
        AND r.target_entity_id = NEW.source_entity_id
        AND r.relationship_type_id = rt_inv.id
        AND (TG_OP = 'INSERT' OR r.id != NEW.id)
    ) THEN
        RAISE EXCEPTION
            'Inverse relationship already exists';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

The following statement attaches the trigger to the
`relationships` table:

```sql
CREATE TRIGGER trg_prevent_inverse_relationship
    BEFORE INSERT OR UPDATE ON relationships
    FOR EACH ROW
    EXECUTE FUNCTION prevent_inverse_relationship();
```

No new indexes are needed. The existing unique constraint on
`relationships` creates an index on
`(campaign_id, source_entity_id, target_entity_id,
relationship_type_id)` which covers the swapped lookup. The
`relationship_types` unique constraint covers the type name
lookup.

### Display View: entity_relationships_view

A regular PostgreSQL view presents relationships from both
perspectives, eliminating application-side flipping logic.

The following view definition provides forward and inverse
projections:

```sql
CREATE VIEW entity_relationships_view AS
-- Forward: entity is source, use display_label
SELECT
    r.id,
    r.campaign_id,
    r.source_entity_id AS from_entity_id,
    r.target_entity_id AS to_entity_id,
    rt.name AS relationship_type,
    rt.display_label,
    r.tone,
    r.description,
    r.strength,
    r.created_at,
    r.updated_at,
    se.name AS from_entity_name,
    se.entity_type AS from_entity_type,
    te.name AS to_entity_name,
    te.entity_type AS to_entity_type,
    'forward' AS direction
FROM relationships r
JOIN relationship_types rt
    ON rt.id = r.relationship_type_id
JOIN entities se ON se.id = r.source_entity_id
JOIN entities te ON te.id = r.target_entity_id

UNION ALL

-- Inverse: entity is target, flip source/target,
-- use inverse labels
SELECT
    r.id,
    r.campaign_id,
    r.target_entity_id AS from_entity_id,
    r.source_entity_id AS to_entity_id,
    rt.inverse_name AS relationship_type,
    rt.inverse_display_label AS display_label,
    r.tone,
    r.description,
    r.strength,
    r.created_at,
    r.updated_at,
    te.name AS from_entity_name,
    te.entity_type AS from_entity_type,
    se.name AS to_entity_name,
    se.entity_type AS to_entity_type,
    'inverse' AS direction
FROM relationships r
JOIN relationship_types rt
    ON rt.id = r.relationship_type_id
JOIN entities se ON se.id = r.source_entity_id
JOIN entities te ON te.id = r.target_entity_id
WHERE rt.is_symmetric = false;
```

For symmetric types, the view includes only the forward direction
to avoid duplicating "knows" entries. For asymmetric types, the
view includes both forward and inverse projections.

Querying all relationships for a specific entity becomes a single
statement:

```sql
SELECT * FROM entity_relationships_view
WHERE from_entity_id = $1
ORDER BY relationship_type, to_entity_name;
```

A materialized view is unnecessary because the `relationships`
table for a TTRPG campaign contains hundreds to low thousands of
rows, making the join sub-millisecond with existing indexes.

## New Schema

This section defines the new and modified table structures.

### relationship_type_templates Table

The `relationship_type_templates` table stores system-default
relationship types that the system copies to each new campaign.

```sql
CREATE TABLE relationship_type_templates (
    id                    BIGSERIAL PRIMARY KEY,
    name                  TEXT NOT NULL UNIQUE,
    inverse_name          TEXT NOT NULL,
    is_symmetric          BOOLEAN NOT NULL
                              DEFAULT false,
    display_label         TEXT NOT NULL,
    inverse_display_label TEXT NOT NULL,
    description           TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT tpl_symmetric_inverse_match
        CHECK (
            NOT is_symmetric
            OR name = inverse_name
        )
);

COMMENT ON TABLE relationship_type_templates IS
    'System-default relationship types copied to '
    'each new campaign';
```

### Modified relationship_types Table

The `relationship_types` table now requires a non-null
`campaign_id` and stores one canonical row per relationship pair.

```sql
CREATE TABLE relationship_types (
    id                    BIGSERIAL PRIMARY KEY,
    campaign_id           BIGINT NOT NULL
                          REFERENCES campaigns(id)
                              ON DELETE CASCADE,
    name                  TEXT NOT NULL,
    inverse_name          TEXT NOT NULL,
    is_symmetric          BOOLEAN NOT NULL
                              DEFAULT false,
    display_label         TEXT NOT NULL,
    inverse_display_label TEXT NOT NULL,
    description           TEXT,
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT unique_relationship_type_per_campaign
        UNIQUE (campaign_id, name),
    CONSTRAINT symmetric_inverse_match
        CHECK (
            NOT is_symmetric
            OR name = inverse_name
        )
);

COMMENT ON TABLE relationship_types IS
    'Campaign-scoped relationship types, seeded '
    'from templates on campaign creation';
```

### Modified relationships Table

The `relationships` table replaces the `relationship_type` text
column with `relationship_type_id` and removes the
`bidirectional` column.

```sql
CREATE TABLE relationships (
    id                   BIGSERIAL PRIMARY KEY,
    campaign_id          BIGINT NOT NULL
                         REFERENCES campaigns(id)
                             ON DELETE CASCADE,
    source_entity_id     BIGINT NOT NULL
                         REFERENCES entities(id)
                             ON DELETE CASCADE,
    target_entity_id     BIGINT NOT NULL
                         REFERENCES entities(id)
                             ON DELETE CASCADE,
    relationship_type_id BIGINT NOT NULL
                         REFERENCES
                             relationship_types(id),
    tone                 TEXT CHECK (tone IN (
                             'friendly', 'hostile',
                             'neutral', 'romantic',
                             'professional', 'fearful',
                             'respectful', 'unknown')),
    description          TEXT,
    strength             INT CHECK (
                             strength >= 1
                             AND strength <= 10),
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT no_self_relationship
        CHECK (
            source_entity_id != target_entity_id
        ),
    CONSTRAINT uq_relationships_source_target_type
        UNIQUE (
            campaign_id,
            source_entity_id,
            target_entity_id,
            relationship_type_id
        )
);

COMMENT ON COLUMN relationships.relationship_type_id IS
    'FK to relationship_types; the canonical '
    'forward direction';
```

## Migration Strategy

The migration must handle existing data without loss. The
following steps describe the migration sequence in the order the
database must execute them.

### Step 1: Create relationship_type_templates

The migration creates the templates table and populates the
table from the existing `NULL`-campaign rows in
`relationship_types`. The migration keeps only the forward
direction of each pair (for example, the migration keeps `owns`
and drops `owned_by`).

### Step 2: Seed Campaign-Scoped Types

For each existing campaign, the migration copies the templates
into `relationship_types` with the campaign's ID.

### Step 3: Add the relationship_type_id Column

The migration adds the new column to `relationships` as nullable
initially:

```sql
ALTER TABLE relationships
    ADD COLUMN relationship_type_id BIGINT;
```

### Step 4: Populate relationship_type_id

The migration maps existing text types to the new numeric IDs.
For inverse types (such as "owned_by"), the migration maps to the
canonical forward type (such as "owns") and swaps the source and
target entities.

The following statement handles forward types with a direct
mapping:

```sql
UPDATE relationships r
SET relationship_type_id = rt.id
FROM relationship_types rt
WHERE rt.campaign_id = r.campaign_id
AND rt.name = r.relationship_type;
```

The following statement handles inverse types by mapping to the
forward type and swapping entities:

```sql
UPDATE relationships r
SET relationship_type_id = rt.id,
    source_entity_id = r.target_entity_id,
    target_entity_id = r.source_entity_id
FROM relationship_types rt
WHERE rt.campaign_id = r.campaign_id
AND rt.inverse_name = r.relationship_type
AND rt.name != rt.inverse_name
AND r.relationship_type_id IS NULL;
```

### Step 5: Deduplicate Rows

After swapping, some rows may now be duplicates (a forward row
and a former inverse row that point the same way). The migration
keeps the older row:

```sql
DELETE FROM relationships
WHERE id NOT IN (
    SELECT MIN(id)
    FROM relationships
    GROUP BY campaign_id, source_entity_id,
             target_entity_id, relationship_type_id
);
```

### Step 6: Enforce NOT NULL and Add the Foreign Key

The migration makes `relationship_type_id` non-nullable and adds
the foreign key constraint:

```sql
ALTER TABLE relationships
    ALTER COLUMN relationship_type_id SET NOT NULL;

ALTER TABLE relationships
    ADD CONSTRAINT fk_relationships_type
    FOREIGN KEY (relationship_type_id)
    REFERENCES relationship_types(id);
```

### Step 7: Drop Old Columns

The migration removes the replaced columns:

```sql
ALTER TABLE relationships
    DROP COLUMN relationship_type,
    DROP COLUMN bidirectional;
```

### Step 8: Update the Unique Constraint

The migration replaces the old unique constraint with one that
references `relationship_type_id`:

```sql
ALTER TABLE relationships
    DROP CONSTRAINT IF EXISTS
        uq_relationships_source_target_type;

ALTER TABLE relationships
    ADD CONSTRAINT uq_relationships_source_target_type
    UNIQUE (
        campaign_id,
        source_entity_id,
        target_entity_id,
        relationship_type_id
    );
```

### Step 9: Remove NULL-Campaign Relationship Types

The migration deletes the global relationship types and enforces
the `NOT NULL` constraint on `campaign_id`:

```sql
DELETE FROM relationship_types
WHERE campaign_id IS NULL;

ALTER TABLE relationship_types
    ALTER COLUMN campaign_id SET NOT NULL;
```

### Step 10: Remove Inverse-Only Type Rows

The migration deletes `relationship_types` rows that served
purely as inverse entries (for example, "owned_by" when "owns"
already exists with `inverse_name = "owned_by"`):

```sql
DELETE FROM relationship_types rt
WHERE EXISTS (
    SELECT 1 FROM relationship_types rt2
    WHERE rt2.campaign_id = rt.campaign_id
    AND rt2.inverse_name = rt.name
    AND rt2.name != rt.name
    AND rt2.id < rt.id
);
```

### Step 11: Create the Trigger and View

The migration creates the `prevent_inverse_relationship` trigger
function and the `entity_relationships_view` as defined in the
Design Changes section.

### Step 12: Hook Campaign Creation

The migration requires a corresponding change to the Go
application code. The `CreateCampaignWithOwner` database method
must copy templates when creating a campaign.

The following Go code shows the required addition:

```go
func (db *DB) CreateCampaignWithOwner(
    ctx context.Context,
    req models.CreateCampaignRequest,
    ownerID int64,
) (*models.Campaign, error) {
    // ... existing campaign INSERT ...

    // Copy relationship type templates for the
    // new campaign
    _, err = db.Exec(ctx, `
        INSERT INTO relationship_types
            (campaign_id, name, inverse_name,
             is_symmetric, display_label,
             inverse_display_label, description)
        SELECT $1, name, inverse_name,
               is_symmetric, display_label,
               inverse_display_label, description
        FROM relationship_type_templates
    `, campaign.ID)
    if err != nil {
        return nil, fmt.Errorf(
            "seed relationship types: %w", err)
    }

    return &c, nil
}
```

## Application Layer Changes

This section describes the required changes to the Go backend
and the React frontend.

### Go Backend

The backend requires changes across several packages.

The `models.go` file must update the `Relationship` struct by
replacing `RelationshipType string` with
`RelationshipTypeID int64`, removing `Bidirectional bool`, and
adding `RelationshipTypeName string` and `DisplayLabel string`
for joined display data.

The `models.go` file must also update the
`CreateRelationshipRequest` struct by replacing
`RelationshipType string` with `RelationshipTypeID int64` and
removing `Bidirectional *bool`.

The `database/relationships.go` file must update the
`CreateRelationship` function to use `relationship_type_id` in
the `INSERT` statement, update the `ON CONFLICT` clause for the
new unique constraint, and catch trigger exceptions for inverse
conflicts gracefully.

The `database/relationships.go` file must update the
`GetEntityRelationships` function to replace the current query
and deduplication logic with a simple
`SELECT` from `entity_relationships_view WHERE from_entity_id
= $1`.

The `database/relationships.go` file must update the
`ListRelationshipsByCampaign` function to join
`relationship_types` for display information.

The `api/handlers.go` file must update the
`CreateRelationship` handler to accept
`relationship_type_id` instead of `relationship_type` text and
to remove the inverse row creation logic.

The `api/content_analysis_handler.go` file must remove the code
that creates inverse relationships, catch the trigger exception
to silently skip duplicate inverse suggestions, and update
`RunContentEnrichment` to look up `relationship_type_id` by
name.

The `enrichment/prompts.go` file must update the LLM prompt to
suggest only canonical forward relationship types (not inverses
like "owned_by").

### React Frontend

The frontend requires changes to types, API calls, and several
components.

The `types/index.ts` file must update the `Relationship`
interface by replacing `relationshipType: string` with
`relationshipTypeId: number`, removing `bidirectional`, and
adding `displayLabel: string` and
`direction: 'forward' | 'inverse'`.

The `api/relationships.ts` file must update create and update
payloads to send `relationship_type_id`.

The `RelationshipEditor` component must use
`relationship_type_id` for lookups instead of text matching.

The `AddRelationshipDialog` component must swap source and target
before submitting when the direction is "incoming", so the stored
row always uses the canonical forward direction.

The `EntityView.tsx` component must display relationships using
`displayLabel` from the view, with no client-side flipping
needed.

The `AnalysisTriagePage.tsx` component must update the
relationship suggestion display to use type names from the joined
data.

## Verification Plan

This section defines the verification steps for the refactor.

### Automated Tests

The following commands must pass after the refactor:

1. Run `go build ./...` to verify compilation.
2. Run `go test ./...` to verify all Go tests pass.
3. Run `cd client && npx tsc --noEmit` to verify TypeScript
   compiles without errors.
4. Run `cd client && npx vitest run` to verify all frontend
   tests pass.
5. Run `make test-all` to verify the full suite passes.

### Manual Testing

The following scenarios require manual verification:

- Create a new campaign and verify that the system seeds the
  relationship types from templates.
- Create a relationship and verify that the database stores a
  single row.
- View an entity and verify that both forward and inverse
  relationships display correctly.
- Attempt to create an inverse relationship via the API and
  verify that the trigger blocks the insert.
- Run the enrichment pipeline and verify that no inverse
  suggestions appear.
- Accept a relationship suggestion and verify that the system
  creates a single row with no inverse.
- View an entity that had old dual-edge data and verify that the
  migration preserved the data correctly.
- Delete a relationship type and verify appropriate cascade or
  error handling.

## References

The following resources informed this design:

- [GraphAware: Modelling Bidirectional Relationships in Neo4j](https://graphaware.com/blog/neo4j-bidirectional-relationships/)
- [Neo4j: RDF vs Property Graphs](https://neo4j.com/blog/knowledge-graph/rdf-vs-property-graphs-knowledge-graphs/)
- [Neo4j Community: Single Relationship vs Two for Efficiency](https://community.neo4j.com/t/creating-single-relationship-instead-of-two-in-order-to-drive-efficiency/59925)
