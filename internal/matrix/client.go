package matrix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"maunium.net/go/mautrix"
)

const defaultSessionTTL = 10 * time.Minute

// Config configures the Matrix client wrapper.
type Config struct {
	HomeserverURL string
	CacheTTL      time.Duration
	HTTPClient    *http.Client
}

// MatrixClient manages Matrix sessions keyed by username so we can cache access tokens.
type MatrixClient struct {
	homeserver string
	cacheTTL   time.Duration
	httpClient *http.Client

	mu       sync.RWMutex
	sessions map[string]*session
}

type session struct {
	client    *mautrix.Client
	expiresAt time.Time
}

// NewClient creates a MatrixClient that can log in using username/password and cache the session per user.
func NewClient(cfg Config) (*MatrixClient, error) {
	if strings.TrimSpace(cfg.HomeserverURL) == "" {
		return nil, errors.New("homeserver url is required")
	}

	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &MatrixClient{
		homeserver: cfg.HomeserverURL,
		cacheTTL:   ttl,
		httpClient: httpClient,
		sessions:   make(map[string]*session),
	}, nil
}

// EnsureSession returns a logged-in mautrix client for the provided credentials.
// If the cached session is expired or missing it performs a new login.
func (mc *MatrixClient) EnsureSession(ctx context.Context, username, password string) (*mautrix.Client, error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("username and password are required")
	}

	key := normalizeKey(username)
	if sess := mc.loadSession(key); sess != nil {
		return sess.client, nil
	}

	return mc.createSession(ctx, key, username, password)
}

// Invalidate removes any cached session for the provided username.
func (mc *MatrixClient) Invalidate(username string) {
	key := normalizeKey(username)
	mc.mu.Lock()
	delete(mc.sessions, key)
	mc.mu.Unlock()
}

func (mc *MatrixClient) loadSession(key string) *session {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	sess, ok := mc.sessions[key]
	if !ok || sess == nil || sess.isExpired() {
		return nil
	}
	return sess
}

func (mc *MatrixClient) createSession(ctx context.Context, key, username, password string) (*mautrix.Client, error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if sess := mc.sessions[key]; sess != nil && !sess.isExpired() {
		return sess.client, nil
	}

	client, err := mautrix.NewClient(mc.homeserver, "", "")
	if err != nil {
		return nil, fmt.Errorf("create matrix client: %w", err)
	}
	client.Client = mc.httpClient

	req := &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			Type: mautrix.IdentifierTypeUser,
			User: username,
		},
		Password:           password,
		StoreCredentials:   true,
		StoreHomeserverURL: false,
	}

	if _, err = client.Login(ctx, req); err != nil {
		return nil, fmt.Errorf("matrix login failed: %w", err)
	}

	mc.sessions[key] = &session{
		client:    client,
		expiresAt: time.Now().Add(mc.cacheTTL),
	}

	return client, nil
}

func (s *session) isExpired() bool {
	return time.Now().After(s.expiresAt)
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
