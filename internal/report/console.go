package report

import (
	"fmt"
	"io"
	"strings"
)

// Console writes a human-readable run summary.
type Console struct {
	Out io.Writer
}

// WriteSummary prints the final report.
func (c Console) WriteSummary(summary Summary, baseURL string) error {
	_, _ = fmt.Fprintln(c.Out, "")
	_, _ = fmt.Fprintln(c.Out, strings.Repeat("=", 72))
	_, _ = fmt.Fprintf(c.Out, "Tests: %d  Passed: %d  Failed: %d  Errors: %d  Duration: %s\n",
		summary.TotalCases,
		summary.PassedCases,
		summary.FailedCases,
		summary.Errors,
		summary.Duration().Round(1e6),
	)
	_, _ = fmt.Fprintln(c.Out, strings.Repeat("=", 72))

	for i, failure := range summary.Failures {
		_, _ = fmt.Fprintf(c.Out, "\n[%d] %s\n", i+1, failure.Title)
		_, _ = fmt.Fprintf(c.Out, "Check: %s\n", failure.Check)
		_, _ = fmt.Fprintf(c.Out, "Operation: %s %s (%s)\n", failure.Case.Method, failure.Case.Path, failure.Case.OperationID)
		_, _ = fmt.Fprintf(c.Out, "Phase/Mode: %s / %s\n", failure.Case.Phase, failure.Case.Mode)
		_, _ = fmt.Fprintf(c.Out, "Status: %d\n", failure.StatusCode)
		_, _ = fmt.Fprintf(c.Out, "Message: %s\n", failure.Message)
		curl := failure.CurlCommand
		if curl == "" {
			curl = Curl(baseURL, failure.Case)
		}
		_, _ = fmt.Fprintf(c.Out, "Reproduce: %s\n", curl)
	}
	return nil
}
