package report

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
)

// Curl builds a minimal curl reproducer for a failing case.
func Curl(baseURL string, c casegen.Case) string {
	fullURL := joinURL(baseURL, c.ResolvedPath(), c.Query)
	parts := []string{"curl", "-X", shellQuote(c.Method), shellQuote(fullURL)}
	for k, v := range c.Headers {
		parts = append(parts, "-H", shellQuote(k+": "+v))
	}
	if len(c.Body) > 0 {
		parts = append(parts, "-d", shellQuote(string(c.Body)))
	}
	return strings.Join(parts, " ")
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

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`") {
		return s
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", `'\''`))
}
