package gateway

import "context"

// CaptureStore persists gateway-captured LLM API traffic.
type CaptureStore interface {
	SaveCaptures(ctx context.Context, captures []Capture) error
}
