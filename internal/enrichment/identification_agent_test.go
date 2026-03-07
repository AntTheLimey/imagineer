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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// IdentificationAgent tests
// ---------------------------------------------------------------------------

func TestIdentificationAgent_Name(t *testing.T) {
	agent := NewIdentificationAgent(nil)
	assert.Equal(t, "identification", agent.Name())
}

func TestIdentificationAgent_DependsOn(t *testing.T) {
	agent := NewIdentificationAgent(nil)
	assert.Nil(t, agent.DependsOn())
}

func TestIdentificationAgent_RunEmptyContent(t *testing.T) {
	agent := NewIdentificationAgent(nil)

	items, err := agent.Run(
		context.Background(),
		nil,
		PipelineInput{Content: ""},
	)

	require.NoError(t, err)
	assert.Empty(t, items)
}
