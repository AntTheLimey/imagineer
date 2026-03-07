/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package defaults

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultRegistry(t *testing.T) {
	reg := NewDefaultRegistry(nil)
	p := reg.BuildPipeline(
		[]string{"identify", "revise", "enrich"})
	require.Len(t, p.Stages, 3)

	assert.Equal(t, "identify", p.Stages[0].Name)
	assert.Equal(t, "identification", p.Stages[0].Phase)
	assert.Len(t, p.Stages[0].Agents, 1)
	assert.Equal(t, "identification",
		p.Stages[0].Agents[0].Name())

	assert.Equal(t, "revise", p.Stages[1].Name)
	assert.Equal(t, "analysis", p.Stages[1].Phase)
	assert.Len(t, p.Stages[1].Agents, 2)

	assert.Equal(t, "enrich", p.Stages[2].Name)
	assert.Equal(t, "enrichment", p.Stages[2].Phase)
	assert.Len(t, p.Stages[2].Agents, 2)
}

func TestDefaultRegistry_SinglePhase(t *testing.T) {
	reg := NewDefaultRegistry(nil)
	p := reg.BuildPipeline([]string{"revise"})
	require.Len(t, p.Stages, 1)
	assert.Equal(t, "analysis", p.Stages[0].Phase)
}
