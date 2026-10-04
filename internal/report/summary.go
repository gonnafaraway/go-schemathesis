package report

import (
	"time"

	"github.com/gonnafaraway/go-schemathesis/internal/check"
)

// Summary is the aggregate outcome of a run.
type Summary struct {
	BaseURL     string
	StartedAt   time.Time
	FinishedAt  time.Time
	TotalCases  int
	PassedCases int
	FailedCases int
	Errors      int
	Failures    []check.Failure
}

// Duration returns wall-clock duration of the run.
func (s Summary) Duration() time.Duration {
	if s.FinishedAt.IsZero() || s.StartedAt.IsZero() {
		return 0
	}
	return s.FinishedAt.Sub(s.StartedAt)
}

// HasFailures reports whether any check failed.
func (s Summary) HasFailures() bool {
	return s.FailedCases > 0 || len(s.Failures) > 0
}
