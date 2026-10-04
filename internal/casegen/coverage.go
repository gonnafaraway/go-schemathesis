package casegen

import (
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

func (g *Generator) coverage(op schema.Operation, mode Mode) ([]Case, error) {
	cases := make([]Case, 0)

	for _, p := range op.Parameters {
		boundaries := g.values.BoundaryValues(p.Schema)
		for _, boundary := range boundaries {
			valid := mode == ModePositive
			if mode == ModeNegative {
				// keep only values that are likely invalid for negative mode
				if isLikelyValidBoundary(p.Schema, boundary) {
					continue
				}
			}
			if mode == ModePositive && !isLikelyValidBoundary(p.Schema, boundary) {
				continue
			}
			_ = valid

			c := g.baseCase(op, PhaseCoverage, mode)
			c.Description = "coverage for parameter " + p.Name
			target := p.Name
			value := boundary
			if err := g.fillParameters(&c, op, ModePositive, func(param schema.Parameter) any {
				if param.Name == target {
					return value
				}
				return g.values.Positive(param.Schema)
			}); err != nil {
				return nil, err
			}
			if err := g.fillBody(&c, op, ModePositive, nil); err != nil {
				return nil, err
			}
			cases = append(cases, c)
		}
	}

	if op.RequestBody != nil {
		_, media := pickMediaType(op.RequestBody.Content)
		boundaries := g.values.BoundaryValues(media.Schema)
		for _, boundary := range boundaries {
			if mode == ModePositive && !isLikelyValidBoundary(media.Schema, boundary) {
				continue
			}
			if mode == ModeNegative && isLikelyValidBoundary(media.Schema, boundary) {
				continue
			}
			c := g.baseCase(op, PhaseCoverage, mode)
			c.Description = "coverage for request body"
			if err := g.fillParameters(&c, op, ModePositive, nil); err != nil {
				return nil, err
			}
			if err := g.fillBody(&c, op, mode, boundary); err != nil {
				return nil, err
			}
			cases = append(cases, c)
		}

		if mode == ModeNegative && op.RequestBody.Required {
			c := g.baseCase(op, PhaseCoverage, mode)
			c.Description = "missing required request body"
			if err := g.fillParameters(&c, op, ModePositive, nil); err != nil {
				return nil, err
			}
			cases = append(cases, c)
		}
	}

	if len(cases) == 0 {
		c := g.baseCase(op, PhaseCoverage, mode)
		c.Description = "coverage baseline"
		if err := g.fillParameters(&c, op, mode, nil); err != nil {
			return nil, err
		}
		if err := g.fillBody(&c, op, mode, nil); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func isLikelyValidBoundary(s schema.JSONSchema, v any) bool {
	switch s.Type {
	case "string":
		text, ok := v.(string)
		if !ok {
			return false
		}
		if s.MinLength != nil && len(text) < *s.MinLength {
			return false
		}
		if s.MaxLength != nil && len(text) > *s.MaxLength {
			return false
		}
		return true
	case "integer", "number":
		var num float64
		switch t := v.(type) {
		case int:
			num = float64(t)
		case int64:
			num = float64(t)
		case float64:
			num = t
		default:
			return false
		}
		if s.Minimum != nil && num < *s.Minimum {
			return false
		}
		if s.Maximum != nil && num > *s.Maximum {
			return false
		}
		return true
	case "boolean":
		_, ok := v.(bool)
		return ok
	default:
		return true
	}
}
