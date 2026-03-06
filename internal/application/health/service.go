package health

import (
	"context"
	"time"
)

// Pinger is the interface for checking database connectivity.
type Pinger interface {
	Ping(ctx context.Context) error
}

// CheckResult holds the result of a single health check.
type CheckResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
}

// Status is the top-level health response.
type Status struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

// NoopPinger is a Pinger that always succeeds. Useful for tests without a real database.
type NoopPinger struct{}

// Ping always returns nil.
func (NoopPinger) Ping(_ context.Context) error { return nil }

// CheckMongoDB pings MongoDB and returns a CheckResult.
func CheckMongoDB(ctx context.Context, pinger Pinger) CheckResult {
	start := time.Now()
	err := pinger.Ping(ctx)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return CheckResult{Name: "mongodb", Status: "down", LatencyMs: latencyMs}
	}
	return CheckResult{Name: "mongodb", Status: "up", LatencyMs: latencyMs}
}
