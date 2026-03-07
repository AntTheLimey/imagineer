/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package enrichment

import (
	"context"
	"testing"

	"github.com/antonypegg/imagineer/internal/llm"
	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Minimal mock agent for registry tests
// ---------------------------------------------------------------------------

// mockAgent is a minimal PipelineAgent implementation used exclusively
// by the PhaseRegistry tests.
type mockAgent struct {
	name string
}

func (m *mockAgent) Name() string        { return m.name }
func (m *mockAgent) DependsOn() []string { return nil }
func (m *mockAgent) Run(
	_ context.Context,
	_ llm.Provider,
	_ PipelineInput,
) ([]models.ContentAnalysisItem, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// PhaseRegistry tests
// ---------------------------------------------------------------------------

func TestNewPhaseRegistry(t *testing.T) {
	reg := NewPhaseRegistry(nil)
	require.NotNil(t, reg)
}

func TestRegister(t *testing.T) {
	reg := NewPhaseRegistry(nil)
	agent := &mockAgent{name: "detector"}

	reg.Register("analysis", "analysis_phase", agent)

	p := reg.BuildPipeline([]string{"analysis"})
	require.NotNil(t, p)
	require.Len(t, p.Stages, 1)
	assert.Equal(t, "analysis", p.Stages[0].Name)
	assert.Equal(t, "analysis_phase", p.Stages[0].Phase)
	assert.Len(t, p.Stages[0].Agents, 1)
}

func TestBuildPipeline_MultiplePhases(t *testing.T) {
	reg := NewPhaseRegistry(nil)
	reg.Register("alpha", "phase_a", &mockAgent{name: "a1"})
	reg.Register("beta", "phase_b", &mockAgent{name: "b1"})

	p := reg.BuildPipeline([]string{"alpha", "beta"})
	require.NotNil(t, p)
	require.Len(t, p.Stages, 2)
	assert.Equal(t, "alpha", p.Stages[0].Name)
	assert.Equal(t, "beta", p.Stages[1].Name)
}

func TestBuildPipeline_RepeatedPhase(t *testing.T) {
	reg := NewPhaseRegistry(nil)
	reg.Register("x", "x_phase", &mockAgent{name: "x-agent"})

	p := reg.BuildPipeline([]string{"x", "x"})
	require.NotNil(t, p)
	require.Len(t, p.Stages, 2)
	assert.Equal(t, "x", p.Stages[0].Name)
	assert.Equal(t, "x", p.Stages[1].Name)
}

func TestBuildPipeline_UnknownPhaseSkipped(t *testing.T) {
	reg := NewPhaseRegistry(nil)
	reg.Register("known", "known_phase", &mockAgent{name: "k-agent"})

	p := reg.BuildPipeline([]string{"known", "unknown"})
	require.NotNil(t, p)
	require.Len(t, p.Stages, 1)
	assert.Equal(t, "known", p.Stages[0].Name)
}

func TestBuildPipeline_EmptyPhases(t *testing.T) {
	reg := NewPhaseRegistry(nil)

	p := reg.BuildPipeline(nil)
	require.NotNil(t, p)
	assert.Len(t, p.Stages, 0)
}
