package check

import (
	"fmt"
)

type notAServerError struct{}

func (c *notAServerError) Name() Name { return NotAServerError }

func (c *notAServerError) Check(ctx Context) *Failure {
	if ctx.Response.StatusCode < 500 {
		return nil
	}
	return &Failure{
		Check:      c.Name(),
		Title:      "Server error",
		Message:    fmt.Sprintf("received server error status %d", ctx.Response.StatusCode),
		Case:       ctx.Case,
		StatusCode: ctx.Response.StatusCode,
	}
}
