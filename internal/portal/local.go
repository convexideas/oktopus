package portal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LocalClient is a mock portal that stores auth state on disk.
// Accepts any credentials and simulates a portal connection.
// Used for UX flow testing until a real portal exists.
type LocalClient struct {
	stateDir string
}

// NewLocalClient creates a portal client that persists state to the given directory.
func NewLocalClient(stateDir string) *LocalClient {
	return &LocalClient{stateDir: stateDir}
}

func (c *LocalClient) statePath() string {
	return filepath.Join(c.stateDir, "portal-state.json")
}

func (c *LocalClient) loadState() (*AuthState, error) {
	data, err := os.ReadFile(c.statePath())
	if err != nil {
		return nil, ErrOffline
	}
	var state AuthState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, ErrOffline
	}
	if !state.Authenticated {
		return nil, ErrUnauthorized
	}
	if !state.ExpiresAt.IsZero() && time.Now().After(state.ExpiresAt) {
		return nil, ErrUnauthorized
	}
	return &state, nil
}

func (c *LocalClient) saveState(state *AuthState) error {
	os.MkdirAll(c.stateDir, 0o755)
	data, _ := json.MarshalIndent(state, "", "  ")
	return os.WriteFile(c.statePath(), data, 0o600)
}

func (c *LocalClient) Login(ctx context.Context, serverURL, email, password string) (*AuthState, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password required")
	}

	// Mock: accept any credentials, simulate successful auth
	state := &AuthState{
		Authenticated: true,
		UserID:        "user-" + email,
		Email:         email,
		OrgID:         "org-default",
		OrgName:       "Local Mock Org",
		ServerURL:     serverURL,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	if err := c.saveState(state); err != nil {
		return nil, fmt.Errorf("saving auth state: %w", err)
	}
	return state, nil
}

func (c *LocalClient) LoginWithToken(ctx context.Context, serverURL, token string) (*AuthState, error) {
	if token == "" {
		return nil, fmt.Errorf("token required")
	}

	state := &AuthState{
		Authenticated: true,
		UserID:        "user-token",
		Email:         "token-auth@local",
		OrgID:         "org-default",
		OrgName:       "Local Mock Org",
		ServerURL:     serverURL,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}

	if err := c.saveState(state); err != nil {
		return nil, fmt.Errorf("saving auth state: %w", err)
	}
	return state, nil
}

func (c *LocalClient) Logout(ctx context.Context) error {
	os.Remove(c.statePath())
	return nil
}

func (c *LocalClient) WhoAmI(ctx context.Context) (*AuthState, error) {
	return c.loadState()
}

func (c *LocalClient) Status(ctx context.Context) (*SyncStatus, error) {
	if _, err := c.loadState(); err != nil {
		return nil, err
	}
	// Mock: return zeros — nothing to sync in mock mode
	return &SyncStatus{
		SessionsPending: 0,
		MemoryPending:   0,
		RemotePending:   0,
		LastSyncedAt:    time.Now(),
	}, nil
}

func (c *LocalClient) Push(ctx context.Context) error {
	if _, err := c.loadState(); err != nil {
		return err
	}
	// Mock: no-op, pretend success
	return nil
}

func (c *LocalClient) Pull(ctx context.Context) error {
	if _, err := c.loadState(); err != nil {
		return err
	}
	// Mock: no-op, pretend success
	return nil
}
