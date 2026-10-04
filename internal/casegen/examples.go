package casegen

import (
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

func (g *Generator) examples(op schema.Operation, mode Mode) ([]Case, error) {
	_ = mode
	cases := make([]Case, 0)

	exampleSets := collectExampleSets(op)
	if len(exampleSets) == 0 {
		c := g.baseCase(op, PhaseExamples, ModePositive)
		c.Description = "default positive example"
		if err := g.fillParameters(&c, op, ModePositive, nil); err != nil {
			return nil, err
		}
		if err := g.fillBody(&c, op, ModePositive, nil); err != nil {
			return nil, err
		}
		return []Case{c}, nil
	}

	for i, set := range exampleSets {
		c := g.baseCase(op, PhaseExamples, ModePositive)
		c.Description = "schema example"
		if err := g.fillParameters(&c, op, ModePositive, func(p schema.Parameter) any {
			if v, ok := set.params[p.Name]; ok {
				return v
			}
			return g.values.Positive(p.Schema)
		}); err != nil {
			return nil, err
		}
		if err := g.fillBody(&c, op, ModePositive, set.body); err != nil {
			return nil, err
		}
		if i == 0 && c.Description == "" {
			c.Description = "schema example"
		}
		cases = append(cases, c)
	}
	return cases, nil
}

type exampleSet struct {
	params map[string]any
	body   any
}

func collectExampleSets(op schema.Operation) []exampleSet {
	sets := make([]exampleSet, 0)

	paramExamples := map[string][]any{}
	for _, p := range op.Parameters {
		values := make([]any, 0)
		if p.Example != nil {
			values = append(values, p.Example)
		}
		values = append(values, p.Examples...)
		if p.Schema.Example != nil {
			values = append(values, p.Schema.Example)
		}
		if len(values) > 0 {
			paramExamples[p.Name] = values
		}
	}

	bodyExamples := make([]any, 0)
	if op.RequestBody != nil {
		for _, media := range op.RequestBody.Content {
			if media.Example != nil {
				bodyExamples = append(bodyExamples, media.Example)
			}
			bodyExamples = append(bodyExamples, media.Examples...)
			if media.Schema.Example != nil {
				bodyExamples = append(bodyExamples, media.Schema.Example)
			}
		}
	}

	max := 0
	for _, values := range paramExamples {
		if len(values) > max {
			max = len(values)
		}
	}
	if len(bodyExamples) > max {
		max = len(bodyExamples)
	}
	if max == 0 {
		return sets
	}

	for i := 0; i < max; i++ {
		set := exampleSet{params: map[string]any{}}
		for name, values := range paramExamples {
			set.params[name] = values[i%len(values)]
		}
		if len(bodyExamples) > 0 {
			set.body = bodyExamples[i%len(bodyExamples)]
		}
		sets = append(sets, set)
	}
	return sets
}
