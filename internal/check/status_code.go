package check

import (
	"fmt"
	"strconv"
	"strings"
)

type statusCodeConformance struct{}

func (c *statusCodeConformance) Name() Name { return StatusCodeConformance }

func (c *statusCodeConformance) Check(ctx Context) *Failure {
	if len(ctx.Operation.Responses) == 0 {
		return nil
	}
	if _, ok := ctx.Operation.Responses["default"]; ok {
		return nil
	}
	code := strconv.Itoa(ctx.Response.StatusCode)
	if _, ok := ctx.Operation.Responses[code]; ok {
		return nil
	}
	wildcard := string(code[0]) + "XX"
	if _, ok := ctx.Operation.Responses[wildcard]; ok {
		return nil
	}
	wildcardLower := string(code[0]) + "xx"
	if _, ok := ctx.Operation.Responses[wildcardLower]; ok {
		return nil
	}

	documented := make([]string, 0, len(ctx.Operation.Responses))
	for k := range ctx.Operation.Responses {
		documented = append(documented, k)
	}
	return &Failure{
		Check: c.Name(),
		Title: "Undocumented HTTP status code",
		Message: fmt.Sprintf(
			"received %d; documented: %s",
			ctx.Response.StatusCode,
			strings.Join(documented, ", "),
		),
		Case:       ctx.Case,
		StatusCode: ctx.Response.StatusCode,
	}
}
