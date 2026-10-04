package check

import (
	"net/http"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

// Name identifies a response check.
type Name string

const (
	NotAServerError           Name = "not_a_server_error"
	StatusCodeConformance     Name = "status_code_conformance"
	ContentTypeConformance    Name = "content_type_conformance"
	ResponseSchemaConformance Name = "response_schema_conformance"
	NegativeDataRejection     Name = "negative_data_rejection"
	PositiveDataAcceptance    Name = "positive_data_acceptance"
)

// AllNames returns the default enabled check set.
func AllNames() []Name {
	return []Name{
		NotAServerError,
		StatusCodeConformance,
		ContentTypeConformance,
		ResponseSchemaConformance,
		NegativeDataRejection,
		PositiveDataAcceptance,
	}
}

// Response is the observed HTTP response for check evaluation.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ElapsedMS  float64
}

// Context carries data required to evaluate checks for one case.
type Context struct {
	Case      casegen.Case
	Response  Response
	Operation schema.Operation
}

// Failure is a single failed check verdict.
type Failure struct {
	Check       Name
	Title       string
	Message     string
	Case        casegen.Case
	StatusCode  int
	CurlCommand string
}

// Result aggregates check outcomes for one executed case.
type Result struct {
	Case     casegen.Case
	Passed   bool
	Failures []Failure
}
