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
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// unmarshalStrict decodes YAML with KnownFields
// enabled, rejecting any unrecognised keys.
func unmarshalStrict(data []byte, out any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	return dec.Decode(out)
}

// LoadEntityTypes parses the entity types YAML file.
func LoadEntityTypes(path string) (
	*EntityTypeFile, error,
) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read entity types: %w", err)
	}
	var f EntityTypeFile
	if err := unmarshalStrict(data, &f); err != nil {
		return nil, fmt.Errorf(
			"parse entity types: %w", err)
	}
	return &f, nil
}

// LoadRelationshipTypes parses the relationship
// types YAML file.
func LoadRelationshipTypes(path string) (
	*RelationshipTypeFile, error,
) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read relationship types: %w", err)
	}
	var f RelationshipTypeFile
	if err := unmarshalStrict(data, &f); err != nil {
		return nil, fmt.Errorf(
			"parse relationship types: %w", err)
	}
	return &f, nil
}

// LoadConstraints parses the constraints YAML file.
func LoadConstraints(path string) (
	*ConstraintsFile, error,
) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read constraints: %w", err)
	}
	var f ConstraintsFile
	if err := unmarshalStrict(data, &f); err != nil {
		return nil, fmt.Errorf(
			"parse constraints: %w", err)
	}
	return &f, nil
}

// LoadOntology loads all three YAML files from a
// directory and returns a complete Ontology struct.
func LoadOntology(dir string) (*Ontology, error) {
	et, err := LoadEntityTypes(
		filepath.Join(dir, "entity-types.yaml"))
	if err != nil {
		return nil, err
	}

	rt, err := LoadRelationshipTypes(
		filepath.Join(dir, "relationship-types.yaml"))
	if err != nil {
		return nil, err
	}

	c, err := LoadConstraints(
		filepath.Join(dir, "constraints.yaml"))
	if err != nil {
		return nil, err
	}

	return &Ontology{
		EntityTypes:       et,
		RelationshipTypes: rt,
		Constraints:       c,
	}, nil
}
