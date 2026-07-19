package portal

import "context"

// MockClient satisfies Client for offline/local-only mode.
// All operations return ErrOffline — the CLI falls back to local stores.
type MockClient struct{}

func (m *MockClient) Login(ctx context.Context, serverURL, email, password string) (*AuthState, error) {
	return nil, ErrOffline
}

func (m *MockClient) LoginWithToken(ctx context.Context, serverURL, token string) (*AuthState, error) {
	return nil, ErrOffline
}

func (m *MockClient) Logout(ctx context.Context) error {
	return ErrOffline
}

func (m *MockClient) WhoAmI(ctx context.Context) (*AuthState, error) {
	return nil, ErrOffline
}

func (m *MockClient) Status(ctx context.Context) (*SyncStatus, error) {
	return nil, ErrOffline
}

func (m *MockClient) Push(ctx context.Context) error {
	return ErrOffline
}

func (m *MockClient) Pull(ctx context.Context) error {
	return ErrOffline
}
