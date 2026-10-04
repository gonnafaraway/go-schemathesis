package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client executes HTTP requests against the API under test.
type Client interface {
	Do(ctx context.Context, req Request) (Response, error)
}

type httpClient struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient constructs an HTTP target client.
func NewClient(baseURL string, timeout time.Duration, headers map[string]string) (Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, ErrEmptyBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: invalid base URL %q", ErrEmptyBaseURL, baseURL)
	}

	copied := make(map[string]string, len(headers))
	for k, v := range headers {
		copied[k] = v
	}

	return &httpClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		headers: copied,
	}, nil
}

func (c *httpClient) Do(ctx context.Context, req Request) (Response, error) {
	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return Response{}, fmt.Errorf("%w: build request: %v", ErrRequestFailed, err)
	}

	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	started := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	elapsed := float64(time.Since(started).Microseconds()) / 1000.0
	if err != nil {
		return Response{}, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return Response{}, fmt.Errorf("%w: read body: %v", ErrRequestFailed, err)
	}

	return Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       payload,
		ElapsedMS:  elapsed,
	}, nil
}

func joinURL(baseURL, path string, query map[string]string) string {
	base := strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := base + path
	if len(query) == 0 {
		return u
	}
	values := url.Values{}
	for k, v := range query {
		values.Set(k, v)
	}
	return u + "?" + values.Encode()
}
