package casegen

import (
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

func (g *Generator) fuzz(op schema.Operation, mode Mode) ([]Case, error) {
	n := g.opts.MaxExamples
	cases := make([]Case, 0, n)
	for i := 0; i < n; i++ {
		c := g.baseCase(op, PhaseFuzzing, mode)
		c.Description = "fuzz example"
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
