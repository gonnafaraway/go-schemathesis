package schema

import "errors"

var (
	// ErrEmptySchema means the schema source produced no document.
	ErrEmptySchema = errors.New("schema is empty")
	// ErrNoOperations means the schema has no runnable operations.
	ErrNoOperations = errors.New("schema has no operations")
	// ErrUnsupportedSchema means the document format is not supported.
	ErrUnsupportedSchema = errors.New("unsupported schema format")
)
