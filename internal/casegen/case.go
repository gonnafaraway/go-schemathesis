package casegen

// Phase is a Schemathesis-style generation phase.
type Phase string

const (
	PhaseExamples Phase = "examples"
	PhaseCoverage Phase = "coverage"
	PhaseFuzzing  Phase = "fuzzing"
)

// Mode controls whether generated data should be valid, invalid, or both.
type Mode string

const (
	ModePositive Mode = "positive"
	ModeNegative Mode = "negative"
	ModeAll      Mode = "all"
)

// Case is a concrete HTTP request ready to execute against the API under test.
type Case struct {
	OperationID string
	Method      string
	Path        string
	PathParams  map[string]string
	Query       map[string]string
	Headers     map[string]string
	Body        []byte
	MediaType   string
	Phase       Phase
	Mode        Mode
	Description string
}

// Clone returns a shallow copy with independent maps and body.
func (c Case) Clone() Case {
	out := c
	out.PathParams = copyStringMap(c.PathParams)
	out.Query = copyStringMap(c.Query)
	out.Headers = copyStringMap(c.Headers)
	if c.Body != nil {
		out.Body = append([]byte(nil), c.Body...)
	}
	return out
}

// ResolvedPath substitutes path parameters into the path template.
func (c Case) ResolvedPath() string {
	path := c.Path
	for name, value := range c.PathParams {
		path = replacePathParam(path, name, value)
	}
	return path
}

func replacePathParam(path, name, value string) string {
	braced := "{" + name + "}"
	for {
		idx := indexOf(path, braced)
		if idx < 0 {
			return path
		}
		path = path[:idx] + value + path[idx+len(braced):]
	}
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func copyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
