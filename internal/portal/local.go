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
// All portals share one file (portals.json) keyed by portal name.
type LocalClient struct {
	filePath   string
	portalName string
}

// NewLocalClient creates a portal client for a named portal slot.
func NewLocalClient(stateDir, portalName string) *LocalClient {
	return &LocalClient{
		filePath:   filepath.Join(stateDir, "portals.json"),
		portalName: portalName,
	}
}

func (c *LocalClient) loadAll() (map[string]*AuthState, error) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return make(map[string]*AuthState), nil
	}
	var states map[string]*AuthState
	if err := json.Unmarshal(data, &states); err != nil {
		return make(map[string]*AuthState), nil
	}
	return states, nil
}

func (c *LocalClient) saveAll(states map[string]*AuthState) error {
	os.MkdirAll(filepath.Dir(c.filePath), 0o755)
	data, _ := json.MarshalIndent(states, "", "  ")
	return os.WriteFile(c.filePath, data, 0o600)
}

func (c *LocalClient) loadState() (*AuthState, error) {
	states, _ := c.loadAll()
	state, ok := states[c.portalName]
	if !ok || state == nil || !state.Authenticated {
		return nil, ErrOffline
	}
	if !state.ExpiresAt.IsZero() && time.Now().After(state.ExpiresAt) {
		return nil, ErrUnauthorized
	}
	return state, nil
}

func (c *LocalClient) Login(ctx context.Context, serverURL, email, password string) (*AuthState, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password required")
	}

	state := &AuthState{
		Authenticated: true,
		UserID:        "user-" + email,
		Email:         email,
		OrgID:         "org-" + c.portalName,
		OrgName:       c.portalName,
		ServerURL:     serverURL,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	states, _ := c.loadAll()
	states[c.portalName] = state
	if err := c.saveAll(states); err != nil {
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
		Email:         "token-auth@" + c.portalName,
		OrgID:         "org-" + c.portalName,
		OrgName:       c.portalName,
		ServerURL:     serverURL,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
	}

	states, _ := c.loadAll()
	states[c.portalName] = state
	if err := c.saveAll(states); err != nil {
		return nil, fmt.Errorf("saving auth state: %w", err)
	}
	return state, nil
}

func (c *LocalClient) Logout(ctx context.Context) error {
	states, _ := c.loadAll()
	delete(states, c.portalName)
	if len(states) == 0 {
		os.Remove(c.filePath)
		return nil
	}
	return c.saveAll(states)
}

func (c *LocalClient) WhoAmI(ctx context.Context) (*AuthState, error) {
	return c.loadState()
}

func (c *LocalClient) Status(ctx context.Context) (*SyncStatus, error) {
	if _, err := c.loadState(); err != nil {
		return nil, err
	}
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
	return nil
}

func (c *LocalClient) Pull(ctx context.Context) error {
	if _, err := c.loadState(); err != nil {
		return err
	}
	return nil
}
