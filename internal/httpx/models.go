package httpx

import (
	"net/http"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
)

// Request is an outbound HTTP call to the API under test.
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// FromCase builds a Request from a generated test case and base URL.
func FromCase(baseURL string, c casegen.Case) Request {
	return Request{
		Method:  c.Method,
		URL:     joinURL(baseURL, c.ResolvedPath(), c.Query),
		Headers: c.Headers,
		Body:    c.Body,
	}
}

// Response is the transport-level response from the API under test.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ElapsedMS  float64
}
