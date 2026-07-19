package gateway

import (
	"sync"
	"time"
)

// Capture is a single captured LLM API request/response pair.
type Capture struct {
	ID        string        `json:"id"`
	SessionID string        `json:"session_id"`
	Timestamp time.Time     `json:"timestamp"`
	Provider  string        `json:"provider"`
	Model     string        `json:"model"`
	Endpoint  string        `json:"endpoint"`
	TokensIn  int           `json:"tokens_in"`
	TokensOut int           `json:"tokens_out"`
	Cost      float64       `json:"cost"`
	Duration  time.Duration `json:"duration"`
	Request   []byte        `json:"request"`
	Response  []byte        `json:"response"`
}

// Meter accumulates token usage and cost for a session.
type Meter struct {
	mu        sync.Mutex
	sessionID string
	tokensIn  int
	tokensOut int
	cost      float64
	captures  []Capture
}

// NewMeter creates a meter for the given session.
func NewMeter(sessionID string) *Meter {
	return &Meter{sessionID: sessionID}
}

// Record adds a capture and updates running totals.
func (m *Meter) Record(c Capture) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensIn += c.TokensIn
	m.tokensOut += c.TokensOut
	m.cost += c.Cost
	m.captures = append(m.captures, c)
}

// Totals returns current accumulated usage.
func (m *Meter) Totals() (tokensIn, tokensOut int, cost float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tokensIn, m.tokensOut, m.cost
}

// Flush returns all captures and resets.
func (m *Meter) Flush() []Capture {
	m.mu.Lock()
	defer m.mu.Unlock()
	caps := m.captures
	m.captures = nil
	m.tokensIn = 0
	m.tokensOut = 0
	m.cost = 0
	return caps
}

// Count returns how many captures have been recorded.
func (m *Meter) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.captures)
}
