package schema

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadPetstoreExample(t *testing.T) {
	t.Parallel()

	loader := NewLoader()
	source := filepath.Join("..", "..", "examples", "petstore.yaml")
	spec, err := loader.Load(context.Background(), source)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}
	if len(spec.Operations) != 3 {
		t.Fatalf("expected 3 operations, got %d", len(spec.Operations))
	}
	if spec.Title != "Sample Petstore" {
		t.Fatalf("unexpected title %q", spec.Title)
	}
}
