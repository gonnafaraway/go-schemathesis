package httpx

import "errors"

var (
	// ErrEmptyBaseURL means the target base URL was not configured.
	ErrEmptyBaseURL = errors.New("base URL is required")
	// ErrRequestFailed means the HTTP transport failed before a response.
	ErrRequestFailed = errors.New("request failed")
)
