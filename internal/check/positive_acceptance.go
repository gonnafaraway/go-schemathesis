package check

import (
	"fmt"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
)

type positiveDataAcceptance struct{}

func (c *positiveDataAcceptance) Name() Name { return PositiveDataAcceptance }

func (c *positiveDataAcceptance) Check(ctx Context) *Failure {
	if ctx.Case.Mode != casegen.ModePositive {
		return nil
	}
	if isAcceptedStatus(ctx.Response.StatusCode) {
		return nil
	}
	return &Failure{
		Check: c.Name(),
		Title: "API rejected schema-compliant request",
		Message: fmt.Sprintf(
			"valid data should have been accepted; got %d",
			ctx.Response.StatusCode,
		),
		Case:       ctx.Case,
		StatusCode: ctx.Response.StatusCode,
	}
}

func isAcceptedStatus(code int) bool {
	if code >= 200 && code < 300 {
		return true
	}
	switch code {
	case 401, 403, 404, 409, 429:
		return true
	default:
		return code >= 500
	}
}
