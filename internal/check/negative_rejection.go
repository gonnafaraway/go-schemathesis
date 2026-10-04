package check

import (
	"fmt"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
)

type negativeDataRejection struct{}

func (c *negativeDataRejection) Name() Name { return NegativeDataRejection }

func (c *negativeDataRejection) Check(ctx Context) *Failure {
	if ctx.Case.Mode != casegen.ModeNegative {
		return nil
	}
	if isRejectedStatus(ctx.Response.StatusCode) {
		return nil
	}
	return &Failure{
		Check: c.Name(),
		Title: "API accepted schema-violating request",
		Message: fmt.Sprintf(
			"invalid data should have been rejected; got %d",
			ctx.Response.StatusCode,
		),
		Case:       ctx.Case,
		StatusCode: ctx.Response.StatusCode,
	}
}

func isRejectedStatus(code int) bool {
	switch code {
	case 400, 401, 403, 404, 405, 406, 409, 415, 422, 428, 429:
		return true
	default:
		return code >= 500
	}
}
