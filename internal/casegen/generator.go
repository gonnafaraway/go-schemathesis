package casegen

import (
	"fmt"
	"strings"

	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

// Options controls case generation for a run.
type Options struct {
	Phases      []Phase
	Mode        Mode
	MaxExamples int
	Seed        int64
}

// Generator builds Schemathesis-style test cases from a schema.
type Generator struct {
	values *ValueGenerator
	opts   Options
}

// New constructs a case generator.
func New(opts Options) *Generator {
	if opts.MaxExamples <= 0 {
		opts.MaxExamples = 100
	}
	return &Generator{
		values: NewValueGenerator(opts.Seed),
		opts:   opts,
	}
}

// Generate produces cases for all operations and enabled phases.
func (g *Generator) Generate(spec *schema.Spec) ([]Case, error) {
	if spec == nil {
		return nil, schema.ErrEmptySchema
	}
	cases := make([]Case, 0)
	for _, op := range spec.Operations {
		for _, phase := range g.opts.Phases {
			phaseCases, err := g.generateOperation(op, phase)
			if err != nil {
				return nil, err
			}
			cases = append(cases, phaseCases...)
		}
	}
	return cases, nil
}

func (g *Generator) generateOperation(op schema.Operation, phase Phase) ([]Case, error) {
	modes := modesFor(g.opts.Mode, phase)
	out := make([]Case, 0)
	for _, mode := range modes {
		var generated []Case
		var err error
		switch phase {
		case PhaseExamples:
			generated, err = g.examples(op, mode)
		case PhaseCoverage:
			generated, err = g.coverage(op, mode)
		case PhaseFuzzing:
			generated, err = g.fuzz(op, mode)
		default:
			return nil, fmt.Errorf("unsupported phase %q", phase)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, generated...)
	}
	return out, nil
}

func modesFor(mode Mode, phase Phase) []Mode {
	if phase == PhaseExamples {
		return []Mode{ModePositive}
	}
	switch mode {
	case ModePositive:
		return []Mode{ModePositive}
	case ModeNegative:
		return []Mode{ModeNegative}
	default:
		return []Mode{ModePositive, ModeNegative}
	}
}

func (g *Generator) baseCase(op schema.Operation, phase Phase, mode Mode) Case {
	return Case{
		OperationID: op.ID,
		Method:      op.Method,
		Path:        op.Path,
		PathParams:  map[string]string{},
		Query:       map[string]string{},
		Headers:     map[string]string{},
		Phase:       phase,
		Mode:        mode,
	}
}

func (g *Generator) fillParameters(c *Case, op schema.Operation, mode Mode, mutate func(schema.Parameter) any) error {
	for _, p := range op.Parameters {
		var value any
		if mutate != nil {
			value = mutate(p)
		} else if mode == ModeNegative && p.Required {
			// omit required parameter as a negative strategy occasionally
			if g.values.rng.Intn(2) == 0 {
				continue
			}
			value = g.values.Negative(p.Schema)
		} else {
			value = g.values.Positive(p.Schema)
		}
		if value == nil {
			continue
		}
		text := Stringify(value)
		switch strings.ToLower(p.In) {
		case "path":
			c.PathParams[p.Name] = text
		case "query":
			c.Query[p.Name] = text
		case "header":
			c.Headers[p.Name] = text
		case "cookie":
			c.Headers["Cookie"] = mergeCookie(c.Headers["Cookie"], p.Name, text)
		}
	}

	for _, p := range op.Parameters {
		if strings.ToLower(p.In) == "path" {
			if _, ok := c.PathParams[p.Name]; !ok {
				c.PathParams[p.Name] = Stringify(g.values.Positive(p.Schema))
			}
		}
	}
	return nil
}

func (g *Generator) fillBody(c *Case, op schema.Operation, mode Mode, override any) error {
	if op.RequestBody == nil || len(op.RequestBody.Content) == 0 {
		return nil
	}
	mediaType, media := pickMediaType(op.RequestBody.Content)
	var payload any
	switch {
	case override != nil:
		payload = override
	case mode == ModeNegative:
		payload = g.values.Negative(media.Schema)
	default:
		payload = g.values.Positive(media.Schema)
	}
	body, err := EncodeJSON(payload)
	if err != nil {
		return err
	}
	c.Body = body
	c.MediaType = mediaType
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	c.Headers["Content-Type"] = mediaType
	return nil
}

func pickMediaType(content map[string]schema.MediaType) (string, schema.MediaType) {
	if media, ok := content["application/json"]; ok {
		return "application/json", media
	}
	for mt, media := range content {
		return mt, media
	}
	return "application/json", schema.MediaType{}
}

func mergeCookie(existing, name, value string) string {
	part := name + "=" + value
	if existing == "" {
		return part
	}
	return existing + "; " + part
}
