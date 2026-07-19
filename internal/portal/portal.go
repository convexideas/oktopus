// Package portal defines the interface between the CLI and a remote Oktopus portal.
// The portal is the source of truth for teams/orgs. The CLI works offline (local SQLite)
// and syncs when connected — same model as git.
//
// Until the portal exists, a mock client satisfies the interface and returns ErrOffline.
package portal

import (
	"context"
	"errors"
	"time"
)

// ErrOffline is returned when no portal connection is available.
var ErrOffline = errors.New("portal: not connected (working offline)")

// ErrUnauthorized is returned when the auth token is invalid or expired.
var ErrUnauthorized = errors.New("portal: unauthorized — run 'ok login'")

// AuthState represents the current portal authentication state.
type AuthState struct {
	Authenticated bool
	UserID        string
	Email         string
	OrgID         string
	OrgName       string
	ServerURL     string
	ExpiresAt     time.Time
}

// SyncStatus reports what's pending between local and remote.
type SyncStatus struct {
	SessionsPending  int // local sessions not yet pushed
	MemoryPending    int // local episodes not yet pushed
	RemotePending    int // remote items not yet pulled
	LastSyncedAt     time.Time
}

// Client is the interface the CLI uses to talk to a remote portal.
// Implementations: MockClient (offline), HTTPClient (future).
type Client interface {
	// Auth
	Login(ctx context.Context, serverURL, email, password string) (*AuthState, error)
	LoginWithToken(ctx context.Context, serverURL, token string) (*AuthState, error)
	Logout(ctx context.Context) error
	WhoAmI(ctx context.Context) (*AuthState, error)

	// Sync
	Status(ctx context.Context) (*SyncStatus, error)
	Push(ctx context.Context) error
	Pull(ctx context.Context) error

	// Remote operations (when connected)
	// These mirror the local store interfaces but talk to the portal API.
	// The CLI decides: use local store OR portal client based on auth state + flags.
}
