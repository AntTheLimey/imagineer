//go:build integration

/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"context"
	"testing"

	"github.com/antonypegg/imagineer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateConversation(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	// First call: creates the conversation.
	conv, wasCreated, err := db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeChapter, int64(1),
	)
	require.NoError(t, err)
	assert.True(t, wasCreated)
	assert.Equal(t, campaignID, conv.CampaignID)
	assert.Equal(t, models.ScopeTypeChapter, conv.ScopeType)
	assert.Equal(t, int64(1), conv.ScopeID)
	assert.NotZero(t, conv.ID)
	assert.NotZero(t, conv.CreatedAt)

	// Second call: returns existing conversation.
	conv2, wasCreated2, err := db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeChapter, int64(1),
	)
	require.NoError(t, err)
	assert.False(t, wasCreated2)
	assert.Equal(t, conv.ID, conv2.ID)
}

func TestGetConversation(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	created, _, err := db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeEntity, int64(42),
	)
	require.NoError(t, err)

	got, err := db.GetConversation(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, models.ScopeTypeEntity, got.ScopeType)
}

func TestListConversations(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	// Create two conversations with different scopes.
	_, _, err := db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeChapter, int64(1),
	)
	require.NoError(t, err)
	_, _, err = db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeEntity, int64(2),
	)
	require.NoError(t, err)

	convs, err := db.ListConversations(
		ctx, campaignID, "", 0, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(convs), 2)
	for _, c := range convs {
		assert.NotZero(t, c.CampaignID)
	}
}

func TestListConversationsFiltered(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	_, _, err := db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeChapter, int64(1),
	)
	require.NoError(t, err)
	_, _, err = db.GetOrCreateConversation(
		ctx, campaignID,
		models.ScopeTypeEntity, int64(2),
	)
	require.NoError(t, err)

	convs, err := db.ListConversations(
		ctx, campaignID, "chapter", 0, 20)
	require.NoError(t, err)
	for _, c := range convs {
		assert.Equal(t, models.ScopeTypeChapter, c.ScopeType)
	}
}
