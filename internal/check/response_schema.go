package check

import (
	"encoding/json"
	"fmt"
	"mime"
	"reflect"

	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

type responseSchemaConformance struct{}

func (c *responseSchemaConformance) Name() Name { return ResponseSchemaConformance }

func (c *responseSchemaConformance) Check(ctx Context) *Failure {
	respSchema := findResponse(ctx.Operation, ctx.Response.StatusCode)
	if respSchema == nil || len(respSchema.Content) == 0 {
		return nil
	}
	if len(ctx.Response.Body) == 0 {
		return nil
	}

	raw := ctx.Response.Headers.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil || mediaType == "" {
		mediaType = "application/json"
	}
	media, ok := respSchema.Content[mediaType]
	if !ok {
		for _, candidate := range respSchema.Content {
			media = candidate
			ok = true
			break
		}
	}
	if !ok {
		return nil
	}

	var payload any
	if err := json.Unmarshal(ctx.Response.Body, &payload); err != nil {
		// non-JSON bodies are skipped for schema validation
		return nil
	}
	if err := validateValue(media.Schema, payload); err != nil {
		return &Failure{
			Check:      c.Name(),
			Title:      "Response violates schema",
			Message:    err.Error(),
			Case:       ctx.Case,
			StatusCode: ctx.Response.StatusCode,
		}
	}
	return nil
}

func validateValue(s schema.JSONSchema, v any) error {
	if v == nil {
		if s.Nullable || s.Type == "null" {
			return nil
		}
		if s.Type == "" {
			return nil
		}
		return fmt.Errorf("expected %s, got null", s.Type)
	}

	if len(s.Enum) > 0 {
		for _, candidate := range s.Enum {
			if reflect.DeepEqual(candidate, v) {
				return nil
			}
		}
		return fmt.Errorf("value %v is not one of the allowed enum values", v)
	}

	switch s.Type {
	case "", "object":
		obj, ok := v.(map[string]any)
		if s.Type == "object" && !ok {
			return fmt.Errorf("expected object, got %T", v)
		}
		if !ok {
			return nil
		}
		for _, name := range s.Required {
			if _, exists := obj[name]; !exists {
				return fmt.Errorf("%q is a required property", name)
			}
		}
		for name, prop := range s.Properties {
			if child, exists := obj[name]; exists {
				if err := validateValue(prop, child); err != nil {
					return fmt.Errorf("%s: %w", name, err)
				}
			}
		}
	case "array":
		arr, ok := v.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", v)
		}
		if s.MinItems != nil && len(arr) < *s.MinItems {
			return fmt.Errorf("array length %d is less than minItems %d", len(arr), *s.MinItems)
		}
		if s.MaxItems != nil && len(arr) > *s.MaxItems {
			return fmt.Errorf("array length %d is greater than maxItems %d", len(arr), *s.MaxItems)
		}
		if s.Items != nil {
			for i, item := range arr {
				if err := validateValue(*s.Items, item); err != nil {
					return fmt.Errorf("[%d]: %w", i, err)
				}
			}
		}
	case "string":
		text, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", v)
		}
		if s.MinLength != nil && len(text) < *s.MinLength {
			return fmt.Errorf("string length %d is less than minLength %d", len(text), *s.MinLength)
		}
		if s.MaxLength != nil && len(text) > *s.MaxLength {
			return fmt.Errorf("string length %d is greater than maxLength %d", len(text), *s.MaxLength)
		}
	case "integer":
		switch n := v.(type) {
		case float64:
			if n != float64(int64(n)) {
				return fmt.Errorf("expected integer, got %v", n)
			}
		case json.Number, int, int64:
		default:
			return fmt.Errorf("expected integer, got %T", v)
		}
	case "number":
		switch v.(type) {
		case float64, json.Number, int, int64:
		default:
			return fmt.Errorf("expected number, got %T", v)
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", v)
		}
	}
	return nil
}
