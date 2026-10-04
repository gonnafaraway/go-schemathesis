package check

import "fmt"

// Checker evaluates one response check.
type Checker interface {
	Name() Name
	Check(ctx Context) *Failure
}

// Registry holds the enabled checkers for a run.
type Registry struct {
	checkers []Checker
}

// NewRegistry builds a registry from check names.
func NewRegistry(names []Name) (*Registry, error) {
	all := map[Name]Checker{
		NotAServerError:           &notAServerError{},
		StatusCodeConformance:     &statusCodeConformance{},
		ContentTypeConformance:    &contentTypeConformance{},
		ResponseSchemaConformance: &responseSchemaConformance{},
		NegativeDataRejection:     &negativeDataRejection{},
		PositiveDataAcceptance:    &positiveDataAcceptance{},
	}

	out := make([]Checker, 0, len(names))
	for _, name := range names {
		checker, ok := all[name]
		if !ok {
			return nil, fmt.Errorf("unknown check %q", name)
		}
		out = append(out, checker)
	}
	return &Registry{checkers: out}, nil
}

// Evaluate runs all enabled checks for one case/response pair.
func (r *Registry) Evaluate(ctx Context) Result {
	failures := make([]Failure, 0)
	for _, checker := range r.checkers {
		if failure := checker.Check(ctx); failure != nil {
			failures = append(failures, *failure)
		}
	}
	return Result{
		Case:     ctx.Case,
		Passed:   len(failures) == 0,
		Failures: failures,
	}
}
