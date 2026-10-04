package casegen

import (
	"testing"

	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

func TestPositiveStringRespectsBounds(t *testing.T) {
	t.Parallel()

	minLen := 3
	maxLen := 5
	g := NewValueGenerator(42)
	got := g.Positive(schema.JSONSchema{
		Type:      "string",
		MinLength: &minLen,
		MaxLength: &maxLen,
	})
	text, ok := got.(string)
	if !ok {
		t.Fatalf("expected string, got %T", got)
	}
	if len(text) < minLen || len(text) > maxLen {
		t.Fatalf("length %d outside [%d, %d]", len(text), minLen, maxLen)
	}
}

func TestNegativeStringBreaksType(t *testing.T) {
	t.Parallel()

	g := NewValueGenerator(7)
	got := g.Negative(schema.JSONSchema{Type: "boolean"})
	if _, ok := got.(bool); ok {
		t.Fatalf("expected non-bool negative value, got %v", got)
	}
}

func TestBoundaryValuesIncludeEdges(t *testing.T) {
	t.Parallel()

	minLen := 2
	g := NewValueGenerator(1)
	values := g.BoundaryValues(schema.JSONSchema{
		Type:      "string",
		MinLength: &minLen,
	})
	if len(values) == 0 {
		t.Fatal("expected boundary values")
	}
}
