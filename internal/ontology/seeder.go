/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package ontology

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimal database interface needed by
// seeder functions. Both *pgxpool.Pool and pgx.Tx
// satisfy this interface.
type DBTX interface {
	Exec(ctx context.Context, sql string,
		arguments ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string,
		args ...interface{}) pgx.Row
}

// SeedCampaignEntityTypes inserts entity type rows
// into campaign_entity_types for the given campaign.
func SeedCampaignEntityTypes(
	ctx context.Context,
	db DBTX,
	campaignID int64,
	et *EntityTypeFile,
) error {
	if len(et.Types) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`INSERT INTO campaign_entity_types
		(campaign_id, name, parent_name, abstract, description)
		VALUES `)

	args := []interface{}{}
	i := 0
	for name, def := range et.Types {
		if i > 0 {
			sb.WriteString(", ")
		}
		base := i * 5
		sb.WriteString(fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5,
		))

		var parentName *string
		if def.Parent != "" {
			p := def.Parent
			parentName = &p
		}

		args = append(args,
			campaignID, name, parentName,
			def.Abstract, def.Description,
		)
		i++
	}

	_, err := db.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf(
			"seed campaign entity types: %w", err)
	}
	return nil
}

// SeedCampaignRelationshipTypes inserts relationship
// type rows into the campaign-scoped relationship_types
// table. This replaces the INSERT...SELECT from
// relationship_type_templates.
func SeedCampaignRelationshipTypes(
	ctx context.Context,
	db DBTX,
	campaignID int64,
	rt *RelationshipTypeFile,
) error {
	if len(rt.Types) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`INSERT INTO relationship_types
		(campaign_id, name, inverse_name, is_symmetric,
		 display_label, inverse_display_label, description)
		VALUES `)

	args := []interface{}{}
	i := 0
	for name, def := range rt.Types {
		if i > 0 {
			sb.WriteString(", ")
		}
		base := i * 7
		sb.WriteString(fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4,
			base+5, base+6, base+7,
		))

		args = append(args,
			campaignID, name, def.Inverse,
			def.Symmetric, def.DisplayLabel,
			def.InverseDisplayLabel, def.Description,
		)
		i++
	}

	_, err := db.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf(
			"seed campaign relationship types: %w", err)
	}
	return nil
}

// SeedCampaignConstraints populates the
// relationship_type_constraints table with
// domain/range constraints. Abstract parent types are
// resolved to concrete types using the entity type
// hierarchy.
func SeedCampaignConstraints(
	ctx context.Context,
	db DBTX,
	campaignID int64,
	constraints *ConstraintsFile,
	entityTypes *EntityTypeFile,
) error {
	if len(constraints.DomainRange) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`INSERT INTO relationship_type_constraints
		(relationship_type_id, source_entity_type,
		 target_entity_type)
		VALUES `)

	args := []interface{}{}
	idx := 0

	for relTypeName, dr := range constraints.DomainRange {
		// Resolve domain types.
		domainTypes := resolveTypes(
			dr.Domain, entityTypes)
		// Resolve range types.
		rangeTypes := resolveTypes(
			dr.Range, entityTypes)

		for _, srcType := range domainTypes {
			for _, tgtType := range rangeTypes {
				if idx > 0 {
					sb.WriteString(", ")
				}
				base := idx * 4
				// Use a subquery to look up the
				// relationship_type_id by name and
				// campaign_id.
				sb.WriteString(fmt.Sprintf(
					`((SELECT id FROM relationship_types
					   WHERE campaign_id = $%d
					     AND name = $%d),
					  $%d, $%d)`,
					base+1, base+2, base+3, base+4,
				))
				args = append(args,
					campaignID, relTypeName,
					srcType, tgtType,
				)
				idx++
			}
		}
	}

	if idx == 0 {
		return nil
	}

	sb.WriteString(` ON CONFLICT
		(relationship_type_id, source_entity_type,
		 target_entity_type) DO NOTHING`)

	_, err := db.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf(
			"seed campaign constraints: %w", err)
	}
	return nil
}

// resolveTypes expands abstract types to concrete
// types. The special value "any" expands to all
// concrete types.
func resolveTypes(
	types []string,
	entityTypes *EntityTypeFile,
) []string {
	var result []string
	seen := map[string]bool{}

	for _, t := range types {
		if t == "any" {
			// Expand "any" to all concrete types.
			for name, def := range entityTypes.Types {
				if !def.Abstract && !seen[name] {
					result = append(result, name)
					seen[name] = true
				}
			}
			continue
		}

		concrete := entityTypes.ResolveToConcreteTypes(t)
		for _, c := range concrete {
			if !seen[c] {
				result = append(result, c)
				seen[c] = true
			}
		}
	}

	return result
}

// SeedCampaignRequiredRelationships populates the
// required_relationships table.
func SeedCampaignRequiredRelationships(
	ctx context.Context,
	db DBTX,
	campaignID int64,
	constraints *ConstraintsFile,
) error {
	if len(constraints.Required) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(`INSERT INTO required_relationships
		(campaign_id, entity_type,
		 relationship_type_name)
		VALUES `)

	args := []interface{}{}
	idx := 0

	for entityType, relTypes := range constraints.Required {
		for _, relType := range relTypes {
			if idx > 0 {
				sb.WriteString(", ")
			}
			base := idx * 3
			sb.WriteString(fmt.Sprintf(
				"($%d, $%d, $%d)",
				base+1, base+2, base+3,
			))
			args = append(args,
				campaignID, entityType, relType,
			)
			idx++
		}
	}

	_, err := db.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf(
			"seed required relationships: %w", err)
	}
	return nil
}

// SeedDefaultEra creates a single era named
// "Present Day" with sequence 1 and scale "now" for
// every new campaign.
func SeedDefaultEra(
	ctx context.Context,
	db DBTX,
	campaignID int64,
) error {
	_, err := db.Exec(ctx, `
		INSERT INTO eras
			(campaign_id, sequence, name, scale,
			 description)
		VALUES ($1, 1, 'Present Day', 'now',
			'The current time in the campaign')`,
		campaignID,
	)
	if err != nil {
		return fmt.Errorf(
			"seed default era: %w", err)
	}
	return nil
}
