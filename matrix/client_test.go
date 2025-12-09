package matrix

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nethesis/matrix2acrobits/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maunium.net/go/mautrix/id"
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

// TestSetPusher_Success tests successful pusher registration
func TestSetPusher_Success(t *testing.T) {
	// Mock server that expects a POST request and validates the body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "_matrix/client/v3/pushers/set")

		// Verify query parameter for user impersonation
		assert.Equal(t, "@alice:example.com", r.URL.Query().Get("user_id"))

		// Read and verify the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var pusherReq models.SetPusherRequest
		err = json.Unmarshal(body, &pusherReq)
		require.NoError(t, err)

		assert.Equal(t, "com.acrobits.softphone", pusherReq.AppID)
		assert.Equal(t, "test-device-token-12345", pusherReq.Pushkey)
		assert.Equal(t, "http", *pusherReq.Kind)
		assert.NotNil(t, pusherReq.Data)
		assert.Equal(t, "https://proxy.example.com/_matrix/push/v1/notify", pusherReq.Data.URL)

		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	// Create client with mock server
	cfg := Config{
		HomeserverURL: server.URL,
		AsUserID:      "@proxy:example.com",
		AsToken:       "test_token",
	}
	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Create pusher request
	httpKind := "http"
	pusherReq := &models.SetPusherRequest{
		AppDisplayName:    "Acrobits",
		AppID:             "com.acrobits.softphone",
		Append:            false,
		DeviceDisplayName: "iPhone",
		Kind:              &httpKind,
		Lang:              "en",
		Pushkey:           "test-device-token-12345",
		Data: &models.PusherData{
			Format: "event_id_only",
			URL:    "https://proxy.example.com/_matrix/push/v1/notify",
		},
	}

	// Call SetPusher
	err = client.SetPusher(context.Background(), id.UserID("@alice:example.com"), pusherReq)
	assert.NoError(t, err)
}

// TestSetPusher_InvalidRequest tests SetPusher with invalid request

// TestSetPusher_ServerError tests SetPusher when server returns error
func TestSetPusher_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errcode":"M_UNKNOWN","error":"Internal server error"}`))
	}))
	defer server.Close()

	cfg := Config{
		HomeserverURL: server.URL,
		AsUserID:      "@proxy:example.com",
		AsToken:       "test_token",
	}
	client, err := NewClient(cfg)
	require.NoError(t, err)

	httpKind := "http"
	pusherReq := &models.SetPusherRequest{
		AppID:   "com.acrobits.softphone",
		Kind:    &httpKind,
		Pushkey: "test-token",
		Data: &models.PusherData{
			URL: "https://proxy.example.com/_matrix/push/v1/notify",
		},
	}

	err = client.SetPusher(context.Background(), id.UserID("@alice:example.com"), pusherReq)
	assert.Error(t, err)
}

// TestSetPusher_MissingURL tests SetPusher with missing gateway URL
func TestSetPusher_MissingURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	cfg := Config{
		HomeserverURL: server.URL,
		AsUserID:      "@proxy:example.com",
		AsToken:       "test_token",
	}
	client, err := NewClient(cfg)
	require.NoError(t, err)

	httpKind := "http"
	pusherReq := &models.SetPusherRequest{
		AppID:   "com.acrobits.softphone",
		Kind:    &httpKind,
		Pushkey: "test-token",
		Data: &models.PusherData{
			URL: "", // Missing URL
		},
	}

	// Should still succeed as validation is done by server
	err = client.SetPusher(context.Background(), id.UserID("@alice:example.com"), pusherReq)
	assert.NoError(t, err)
}
