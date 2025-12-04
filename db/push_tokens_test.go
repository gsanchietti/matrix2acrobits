package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabase(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Close()
	assert.NoError(t, err)
}

func TestSavePushToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	selector := "12869E0E6E553673C54F29105A0647204C416A2A:7C3A0D14"
	tokenMsgs := "QVBBOTFiRzlhcVd2bW54bllCWldHOWh4dnRrZ3pUWFNvcGZpdWZ6bWM2dFAzS2J"
	appIDMsgs := "com.cloudsoftphone.app"
	tokenCalls := "Udl99X2JFP1bWwS5gR/wGeLE1hmAB2CMpr1Ej0wxkrY="
	appIDCalls := "com.cloudsoftphone.app.pushkit"

	err = db.SavePushToken(selector, tokenMsgs, appIDMsgs, tokenCalls, appIDCalls)
	assert.NoError(t, err)

	// Verify it was saved
	token, err := db.GetPushToken(selector)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, selector, token.Selector)
	assert.Equal(t, tokenMsgs, token.TokenMsgs)
	assert.Equal(t, appIDMsgs, token.AppIDMsgs)
	assert.Equal(t, tokenCalls, token.TokenCalls)
	assert.Equal(t, appIDCalls, token.AppIDCalls)
}

func TestUpdatePushToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	selector := "12869E0E6E553673C54F29105A0647204C416A2A:7C3A0D14"
	tokenMsgs1 := "token_v1"
	tokenMsgs2 := "token_v2"

	// Save first version
	err = db.SavePushToken(selector, tokenMsgs1, "app1", "", "")
	assert.NoError(t, err)

	// Update with new token
	err = db.SavePushToken(selector, tokenMsgs2, "app1", "", "")
	assert.NoError(t, err)

	// Verify it was updated
	token, err := db.GetPushToken(selector)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, tokenMsgs2, token.TokenMsgs)
}

func TestGetPushTokenNotFound(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	token, err := db.GetPushToken("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, token)
}

func TestDeletePushToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	selector := "12869E0E6E553673C54F29105A0647204C416A2A:7C3A0D14"

	// Save a token
	err = db.SavePushToken(selector, "token123", "app1", "", "")
	assert.NoError(t, err)

	// Delete it
	err = db.DeletePushToken(selector)
	assert.NoError(t, err)

	// Verify it was deleted
	token, err := db.GetPushToken(selector)
	assert.NoError(t, err)
	assert.Nil(t, token)
}

func TestListPushTokens(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	// Save multiple tokens
	selectors := []string{
		"selector1",
		"selector2",
		"selector3",
	}

	for _, sel := range selectors {
		err = db.SavePushToken(sel, "token_"+sel, "app", "", "")
		assert.NoError(t, err)
	}

	// List all
	tokens, err := db.ListPushTokens()
	assert.NoError(t, err)
	assert.Len(t, tokens, 3)

	// Verify timestamps are set
	for _, token := range tokens {
		assert.False(t, token.CreatedAt.IsZero())
		assert.False(t, token.UpdatedAt.IsZero())
		assert.True(t, token.UpdatedAt.After(time.Time{}))
	}
}

func TestListPushTokensEmpty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_push_tokens_*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	db, err := NewDatabase(tmpFile.Name())
	require.NoError(t, err)
	defer db.Close()

	tokens, err := db.ListPushTokens()
	assert.NoError(t, err)
	assert.Len(t, tokens, 0)
}
