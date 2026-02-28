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

// EntityTypeFile represents the parsed
// entity-types.yaml file.
type EntityTypeFile struct {
	Types map[string]EntityTypeDef `yaml:"types"`
}

// EntityTypeDef defines a single entity type.
type EntityTypeDef struct {
	Parent      string   `yaml:"parent"`
	Abstract    bool     `yaml:"abstract"`
	Description string   `yaml:"description"`
	Children    []string `yaml:"children"`
}

// ResolveToConcreteTypes returns all non-abstract
// descendants of the given type name, including the
// type itself if it is concrete.
func (f *EntityTypeFile) ResolveToConcreteTypes(
	name string,
) []string {
	var result []string
	f.collectConcrete(name, &result)
	return result
}

func (f *EntityTypeFile) collectConcrete(
	name string, result *[]string,
) {
	def, ok := f.Types[name]
	if !ok {
		return
	}
	if !def.Abstract {
		*result = append(*result, name)
	}
	for _, child := range def.Children {
		f.collectConcrete(child, result)
	}
}

// RelationshipTypeFile represents the parsed
// relationship-types.yaml file.
type RelationshipTypeFile struct {
	Types map[string]RelationshipTypeDef `yaml:"types"`
}

// RelationshipTypeDef defines a single relationship
// type.
type RelationshipTypeDef struct {
	Inverse             string   `yaml:"inverse"`
	Symmetric           bool     `yaml:"symmetric"`
	DisplayLabel        string   `yaml:"display_label"`
	InverseDisplayLabel string   `yaml:"inverse_display_label"`
	Domain              []string `yaml:"domain"`
	Range               []string `yaml:"range"`
	Genre               []string `yaml:"genre"`
	Description         string   `yaml:"description"`
}

// ConstraintsFile represents the parsed
// constraints.yaml file.
type ConstraintsFile struct {
	DomainRange map[string]DomainRangeDef `yaml:"domain_range"`
	Cardinality map[string]CardinalityDef `yaml:"cardinality"`
	Required    map[string][]string       `yaml:"required"`
}

// DomainRangeDef defines valid source and target
// entity types for a relationship type.
type DomainRangeDef struct {
	Domain []string `yaml:"domain"`
	Range  []string `yaml:"range"`
}

// CardinalityDef defines limits on how many
// relationships of a type an entity can have.
type CardinalityDef struct {
	MaxSource *int `yaml:"max_source"`
	MaxTarget *int `yaml:"max_target"`
}

// Ontology holds all three parsed YAML files
// representing the complete ontology schema.
type Ontology struct {
	EntityTypes       *EntityTypeFile
	RelationshipTypes *RelationshipTypeFile
	Constraints       *ConstraintsFile
}
