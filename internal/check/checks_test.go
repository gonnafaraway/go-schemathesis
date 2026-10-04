package check

import (
	"net/http"
	"testing"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

func TestNotAServerError(t *testing.T) {
	t.Parallel()

	c := &notAServerError{}
	failure := c.Check(Context{
		Response: Response{StatusCode: 500},
	})
	if failure == nil {
		t.Fatal("expected failure for 500")
	}
}

func TestNegativeDataRejection(t *testing.T) {
	t.Parallel()

	c := &negativeDataRejection{}
	failure := c.Check(Context{
		Case:     casegen.Case{Mode: casegen.ModeNegative},
		Response: Response{StatusCode: 200},
	})
	if failure == nil {
		t.Fatal("expected failure when invalid data is accepted")
	}
}

func TestStatusCodeConformance(t *testing.T) {
	t.Parallel()

	c := &statusCodeConformance{}
	failure := c.Check(Context{
		Operation: schema.Operation{
			Responses: map[string]schema.Response{
				"200": {},
			},
		},
		Response: Response{StatusCode: 201},
	})
	if failure == nil {
		t.Fatal("expected undocumented status failure")
	}
}

func TestContentTypeConformance(t *testing.T) {
	t.Parallel()

	c := &contentTypeConformance{}
	headers := http.Header{}
	headers.Set("Content-Type", "text/plain")
	failure := c.Check(Context{
		Operation: schema.Operation{
			Responses: map[string]schema.Response{
				"200": {
					Content: map[string]schema.MediaType{
						"application/json": {},
					},
				},
			},
		},
		Response: Response{
			StatusCode: 200,
			Headers:    headers,
			Body:       []byte(`{"ok":true}`),
		},
	})
	if failure == nil {
		t.Fatal("expected content-type failure")
	}
}
