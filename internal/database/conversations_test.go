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
		ctx, campaignID, models.ScopeTypeChapter, 0, 20)
	require.NoError(t, err)
	for _, c := range convs {
		assert.Equal(t, models.ScopeTypeChapter, c.ScopeType)
	}
}

func TestGetConversation_NotFound(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()

	_, err := db.GetConversation(ctx, 999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"failed to get conversation")
}

func TestCreateMessage(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	conv, _, err :=
		db.GetOrCreateConversation(ctx,
			campaignID,
			models.ScopeTypeChapter,
			int64(1),
		)
	require.NoError(t, err)

	msg, err := db.CreateMessage(ctx, conv.ID,
		models.MessageRoleUser,
		"Hello, help me with Session 5",
		nil, nil, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, conv.ID, msg.ConversationID)
	assert.Equal(t, models.MessageRoleUser, msg.Role)
	assert.Equal(t,
		"Hello, help me with Session 5",
		msg.Content)
	assert.False(t, msg.Compacted)

	// Verify trigger bumped updated_at
	got, err := db.GetConversation(ctx, conv.ID)
	require.NoError(t, err)
	assert.True(t,
		got.UpdatedAt.After(conv.UpdatedAt) ||
			got.UpdatedAt.Equal(conv.UpdatedAt))
}

func TestListMessages(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	conv, _, err :=
		db.GetOrCreateConversation(ctx,
			campaignID,
			models.ScopeTypeChapter,
			int64(1),
		)
	require.NoError(t, err)

	_, err = db.CreateMessage(ctx, conv.ID,
		models.MessageRoleUser, "First message",
		nil, nil, nil, nil, nil)
	require.NoError(t, err)
	_, err = db.CreateMessage(ctx, conv.ID,
		models.MessageRoleAssistant, "Response",
		nil, nil, nil, nil, nil)
	require.NoError(t, err)

	msgs, err := db.ListMessages(ctx,
		conv.ID, 50, 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
	assert.Equal(t, "First message",
		msgs[0].Content)
}

func TestAssembleConversationContext(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	conv, _, err :=
		db.GetOrCreateConversation(ctx,
			campaignID,
			models.ScopeTypeChapter,
			int64(1),
		)
	require.NoError(t, err)

	tokens := 100
	_, err = db.CreateMessage(ctx, conv.ID,
		models.MessageRoleUser, "Hello",
		nil, nil, nil, nil, &tokens)
	require.NoError(t, err)

	cc, err := db.AssembleConversationContext(
		ctx, conv.ID)
	require.NoError(t, err)
	assert.Equal(t, conv.ID,
		cc.Conversation.ID)
	assert.Len(t, cc.Messages, 1)
	assert.Equal(t, 100, cc.TokenEstimate)
}

func TestCompactMessages(t *testing.T) {
	db := setupIntegrationDB(t)
	ctx := context.Background()
	campaignID, _ := createTestCampaign(t, db)

	conv, _, err :=
		db.GetOrCreateConversation(ctx,
			campaignID,
			models.ScopeTypeChapter,
			int64(1),
		)
	require.NoError(t, err)

	msg1, err := db.CreateMessage(ctx, conv.ID,
		models.MessageRoleUser, "Old message",
		nil, nil, nil, nil, nil)
	require.NoError(t, err)
	_, err = db.CreateMessage(ctx, conv.ID,
		models.MessageRoleAssistant, "Recent message",
		nil, nil, nil, nil, nil)
	require.NoError(t, err)

	// Compact the first message
	count, err := db.CompactMessages(ctx,
		conv.ID, []int64{msg1.ID},
		"Summary of old exchange", 50)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify summary was updated
	got, err := db.GetConversation(ctx, conv.ID)
	require.NoError(t, err)
	assert.NotNil(t, got.Summary)
	assert.Equal(t, "Summary of old exchange",
		*got.Summary)
	assert.Equal(t, 50, got.SummaryTokens)

	// Verify context excludes compacted
	cc, err := db.AssembleConversationContext(
		ctx, conv.ID)
	require.NoError(t, err)
	assert.Len(t, cc.Messages, 1)
	assert.Equal(t, "Recent message",
		cc.Messages[0].Content)
}
