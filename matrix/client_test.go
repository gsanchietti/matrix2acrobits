package matrix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_Success(t *testing.T) {
	cfg := Config{
		HomeserverURL: "http://localhost:8008",
		AsUserID:      "@proxy:example.com",
		AsToken:       "valid_token",
	}

	client, err := NewClient(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.cli)
}

func TestNewClient_MissingHomeserverURL(t *testing.T) {
	cfg := Config{
		HomeserverURL: "",
		AsUserID:      "@proxy:example.com",
		AsToken:       "valid_token",
	}

	client, err := NewClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "homeserver url is required", err.Error())
}

func TestNewClient_MissingAsToken(t *testing.T) {
	cfg := Config{
		HomeserverURL: "http://localhost:8008",
		AsUserID:      "@proxy:example.com",
		AsToken:       "",
	}

	client, err := NewClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "application service token (as_token) is required", err.Error())
}

func TestNewClient_MissingAsUserID(t *testing.T) {
	cfg := Config{
		HomeserverURL: "http://localhost:8008",
		AsUserID:      "",
		AsToken:       "valid_token",
	}

	client, err := NewClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "application service user ID (as_user_id) is required", err.Error())
}
