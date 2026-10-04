package casegen

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

// ValueGenerator produces values from JSON Schema fragments.
type ValueGenerator struct {
	rng *rand.Rand
}

// NewValueGenerator constructs a seeded value generator.
func NewValueGenerator(seed int64) *ValueGenerator {
	return &ValueGenerator{rng: rand.New(rand.NewSource(seed))}
}

// Positive generates a schema-compliant value.
func (g *ValueGenerator) Positive(s schema.JSONSchema) any {
	if s.Const != nil {
		return s.Const
	}
	if len(s.Enum) > 0 {
		return s.Enum[g.rng.Intn(len(s.Enum))]
	}
	if s.Example != nil {
		return s.Example
	}
	if s.Default != nil {
		return s.Default
	}

	switch s.Type {
	case "object", "":
		if len(s.Properties) > 0 || s.Type == "object" {
			return g.positiveObject(s)
		}
		fallthrough
	case "string":
		return g.positiveString(s)
	case "integer":
		return g.positiveInteger(s)
	case "number":
		return g.positiveNumber(s)
	case "boolean":
		return g.rng.Intn(2) == 0
	case "array":
		return g.positiveArray(s)
	case "null":
		return nil
	default:
		return g.positiveString(s)
	}
}

// Negative generates a deliberately invalid value for the schema.
func (g *ValueGenerator) Negative(s schema.JSONSchema) any {
	if len(s.Enum) > 0 {
		return "__invalid_enum__"
	}
	switch s.Type {
	case "string":
		if s.MinLength != nil && *s.MinLength > 0 {
			return ""
		}
		if s.MaxLength != nil {
			return strings.Repeat("x", *s.MaxLength+1)
		}
		return 42
	case "integer", "number":
		if s.Minimum != nil {
			return *s.Minimum - 1
		}
		if s.Maximum != nil {
			return *s.Maximum + 1
		}
		return "not-a-number"
	case "boolean":
		return "not-bool"
	case "array":
		return map[string]any{"unexpected": true}
	case "object":
		return "not-an-object"
	default:
		return nil
	}
}

// BoundaryValues returns deterministic coverage candidates for a schema.
func (g *ValueGenerator) BoundaryValues(s schema.JSONSchema) []any {
	out := make([]any, 0, 8)
	if len(s.Enum) > 0 {
		out = append(out, s.Enum...)
		return out
	}
	switch s.Type {
	case "string":
		if s.MinLength != nil {
			n := *s.MinLength
			out = append(out, strings.Repeat("a", n))
			if n > 0 {
				out = append(out, strings.Repeat("a", n-1))
			}
			out = append(out, strings.Repeat("a", n+1))
		}
		if s.MaxLength != nil {
			n := *s.MaxLength
			out = append(out, strings.Repeat("b", n))
			out = append(out, strings.Repeat("b", n+1))
			if n > 0 {
				out = append(out, strings.Repeat("b", n-1))
			}
		}
		if s.MinLength == nil && s.MaxLength == nil {
			out = append(out, "", "a", "coverage")
		}
	case "integer", "number":
		if s.Minimum != nil {
			out = append(out, *s.Minimum, *s.Minimum-1, *s.Minimum+1)
		}
		if s.Maximum != nil {
			out = append(out, *s.Maximum, *s.Maximum+1, *s.Maximum-1)
		}
		if s.Minimum == nil && s.Maximum == nil {
			out = append(out, 0, 1, -1)
		}
	case "boolean":
		out = append(out, true, false)
	case "array":
		if s.MinItems != nil {
			out = append(out, g.arrayOfSize(s, *s.MinItems))
			if *s.MinItems > 0 {
				out = append(out, g.arrayOfSize(s, *s.MinItems-1))
			}
		}
		if s.MaxItems != nil {
			out = append(out, g.arrayOfSize(s, *s.MaxItems))
			out = append(out, g.arrayOfSize(s, *s.MaxItems+1))
		}
	default:
		out = append(out, g.Positive(s))
	}
	return uniqueValues(out)
}

func (g *ValueGenerator) positiveObject(s schema.JSONSchema) map[string]any {
	obj := map[string]any{}
	required := map[string]struct{}{}
	for _, name := range s.Required {
		required[name] = struct{}{}
	}
	for name, prop := range s.Properties {
		_, need := required[name]
		if need || g.rng.Intn(3) != 0 {
			obj[name] = g.Positive(prop)
		}
	}
	for name := range required {
		if _, ok := obj[name]; !ok {
			if prop, exists := s.Properties[name]; exists {
				obj[name] = g.Positive(prop)
			} else {
				obj[name] = "value"
			}
		}
	}
	return obj
}

func (g *ValueGenerator) positiveArray(s schema.JSONSchema) []any {
	size := 1
	if s.MinItems != nil {
		size = *s.MinItems
	}
	if s.MaxItems != nil && size > *s.MaxItems {
		size = *s.MaxItems
	}
	if size == 0 {
		size = 1
	}
	return g.arrayOfSize(s, size)
}

func (g *ValueGenerator) arrayOfSize(s schema.JSONSchema, size int) []any {
	if size < 0 {
		size = 0
	}
	out := make([]any, 0, size)
	for i := 0; i < size; i++ {
		if s.Items != nil {
			out = append(out, g.Positive(*s.Items))
			continue
		}
		out = append(out, i)
	}
	return out
}

func (g *ValueGenerator) positiveString(s schema.JSONSchema) string {
	switch s.Format {
	case "date-time":
		return time.Now().UTC().Format(time.RFC3339)
	case "date":
		return time.Now().UTC().Format("2006-01-02")
	case "email":
		return fmt.Sprintf("user%d@example.com", g.rng.Intn(10000))
	case "uuid":
		return fmt.Sprintf("00000000-0000-4000-8000-%012d", g.rng.Intn(1_000_000_000_000))
	case "uri", "url":
		return "https://example.com/resource"
	}

	minLen := 1
	if s.MinLength != nil {
		minLen = *s.MinLength
	}
	maxLen := minLen + 8
	if s.MaxLength != nil {
		maxLen = *s.MaxLength
		if maxLen < minLen {
			maxLen = minLen
		}
	}
	n := minLen
	if maxLen > minLen {
		n = minLen + g.rng.Intn(maxLen-minLen+1)
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		b.WriteByte(alphabet[g.rng.Intn(len(alphabet))])
	}
	return b.String()
}

func (g *ValueGenerator) positiveInteger(s schema.JSONSchema) int64 {
	min := int64(0)
	max := int64(100)
	if s.Minimum != nil {
		min = int64(math.Ceil(*s.Minimum))
	}
	if s.Maximum != nil {
		max = int64(math.Floor(*s.Maximum))
	}
	if max < min {
		max = min
	}
	if max == min {
		return min
	}
	return min + g.rng.Int63n(max-min+1)
}

func (g *ValueGenerator) positiveNumber(s schema.JSONSchema) float64 {
	min := 0.0
	max := 100.0
	if s.Minimum != nil {
		min = *s.Minimum
	}
	if s.Maximum != nil {
		max = *s.Maximum
	}
	if max < min {
		max = min
	}
	if max == min {
		return min
	}
	return min + g.rng.Float64()*(max-min)
}

// EncodeJSON serializes a value as JSON bytes.
func EncodeJSON(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v)
}

// Stringify converts a generated value to a request parameter string.
func Stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(b)
	}
}

func uniqueValues(values []any) []any {
	seen := map[string]struct{}{}
	out := make([]any, 0, len(values))
	for _, v := range values {
		key := fmt.Sprintf("%T:%v", v, v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}
