package schema

// Spec is a loaded OpenAPI document reduced to runnable operations.
type Spec struct {
	Title      string
	Version    string
	BaseURL    string
	Operations []Operation
}

// Operation describes a single HTTP operation from the schema.
type Operation struct {
	ID          string
	Method      string
	Path        string
	Summary     string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]Response
}

// Parameter is a path, query, header, or cookie parameter.
type Parameter struct {
	Name     string
	In       string
	Required bool
	Schema   JSONSchema
	Example  any
	Examples []any
}

// RequestBody describes an operation request body.
type RequestBody struct {
	Required bool
	Content  map[string]MediaType
}

// MediaType pairs a content type with its schema and examples.
type MediaType struct {
	Schema   JSONSchema
	Example  any
	Examples []any
}

// Response describes a documented HTTP response.
type Response struct {
	Description string
	Headers     map[string]Header
	Content     map[string]MediaType
}

// Header is a documented response header.
type Header struct {
	Required bool
	Schema   JSONSchema
}

// JSONSchema is a simplified JSON Schema used for generation and validation.
type JSONSchema struct {
	Type                 string
	Format               string
	Enum                 []any
	Const                any
	Default              any
	Example              any
	Minimum              *float64
	Maximum              *float64
	ExclusiveMinimum     *float64
	ExclusiveMaximum     *float64
	MinLength            *int
	MaxLength            *int
	Pattern              string
	MinItems             *int
	MaxItems             *int
	UniqueItems          bool
	MinProperties        *int
	MaxProperties        *int
	Required             []string
	Properties           map[string]JSONSchema
	AdditionalProperties *JSONSchema
	Items                *JSONSchema
	Nullable             bool
	Raw                  map[string]any
}
