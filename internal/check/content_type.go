package check

import (
	"fmt"
	"mime"
	"strconv"
	"strings"

	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

type contentTypeConformance struct{}

func (c *contentTypeConformance) Name() Name { return ContentTypeConformance }

func (c *contentTypeConformance) Check(ctx Context) *Failure {
	respSchema := findResponse(ctx.Operation, ctx.Response.StatusCode)
	if respSchema == nil || len(respSchema.Content) == 0 {
		return nil
	}
	if len(ctx.Response.Body) == 0 {
		return nil
	}

	raw := ctx.Response.Headers.Get("Content-Type")
	if raw == "" {
		return &Failure{
			Check:      c.Name(),
			Title:      "Missing Content-Type header",
			Message:    "response body is present but Content-Type header is missing",
			Case:       ctx.Case,
			StatusCode: ctx.Response.StatusCode,
		}
	}

	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return &Failure{
			Check:      c.Name(),
			Title:      "Malformed media type",
			Message:    fmt.Sprintf("cannot parse Content-Type %q: %v", raw, err),
			Case:       ctx.Case,
			StatusCode: ctx.Response.StatusCode,
		}
	}

	if matchesMediaType(mediaType, respSchema.Content) {
		return nil
	}

	documented := make([]string, 0, len(respSchema.Content))
	for mt := range respSchema.Content {
		documented = append(documented, mt)
	}
	return &Failure{
		Check: c.Name(),
		Title: "Undocumented Content-Type",
		Message: fmt.Sprintf(
			"received %s; documented: %s",
			mediaType,
			strings.Join(documented, ", "),
		),
		Case:       ctx.Case,
		StatusCode: ctx.Response.StatusCode,
	}
}

func findResponse(op schema.Operation, status int) *schema.Response {
	code := strconv.Itoa(status)
	if resp, ok := op.Responses[code]; ok {
		return &resp
	}
	wildcard := string(code[0]) + "XX"
	if resp, ok := op.Responses[wildcard]; ok {
		return &resp
	}
	wildcardLower := string(code[0]) + "xx"
	if resp, ok := op.Responses[wildcardLower]; ok {
		return &resp
	}
	if resp, ok := op.Responses["default"]; ok {
		return &resp
	}
	return nil
}

func matchesMediaType(got string, content map[string]schema.MediaType) bool {
	if _, ok := content[got]; ok {
		return true
	}
	parts := strings.SplitN(got, "/", 2)
	if len(parts) != 2 {
		return false
	}
	if _, ok := content[parts[0]+"/*"]; ok {
		return true
	}
	if _, ok := content["*/*"]; ok {
		return true
	}
	return false
}
