package schema

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Loader loads OpenAPI documents from a local file or HTTP(S) URL.
type Loader struct {
	httpClient *http.Client
}

// NewLoader constructs a schema loader.
func NewLoader() *Loader {
	return &Loader{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Load reads and parses an OpenAPI schema into domain types.
func (l *Loader) Load(ctx context.Context, source string) (*Spec, error) {
	data, err := l.readSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrEmptySchema
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedSchema, err)
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate schema: %w", err)
	}

	spec, err := mapDocument(doc, source)
	if err != nil {
		return nil, err
	}
	if len(spec.Operations) == 0 {
		return nil, ErrNoOperations
	}
	return spec, nil
}

func (l *Loader) readSource(ctx context.Context, source string) ([]byte, error) {
	if isURL(source) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return nil, fmt.Errorf("build schema request: %w", err)
		}
		resp, err := l.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch schema: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("fetch schema: unexpected status %d", resp.StatusCode)
		}
		return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	}
	return os.ReadFile(source)
}

func isURL(source string) bool {
	parsed, err := url.Parse(source)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func mapDocument(doc *openapi3.T, source string) (*Spec, error) {
	baseURL := ""
	if len(doc.Servers) > 0 && doc.Servers[0].URL != "" {
		baseURL = doc.Servers[0].URL
		if !isURL(baseURL) && isURL(source) {
			baseURL = resolveAgainst(source, baseURL)
		}
	}

	title := ""
	version := ""
	if doc.Info != nil {
		title = doc.Info.Title
		version = doc.Info.Version
	}

	ops := make([]Operation, 0)
	if doc.Paths != nil {
		for path, item := range doc.Paths.Map() {
			if item == nil {
				continue
			}
			for method, op := range item.Operations() {
				if op == nil {
					continue
				}
				mapped, err := mapOperation(method, path, op, item.Parameters)
				if err != nil {
					return nil, err
				}
				ops = append(ops, mapped)
			}
		}
	}

	return &Spec{
		Title:      title,
		Version:    version,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Operations: ops,
	}, nil
}

func mapOperation(
	method, path string,
	op *openapi3.Operation,
	pathParams openapi3.Parameters,
) (Operation, error) {
	params := make([]Parameter, 0, len(pathParams)+len(op.Parameters))
	for _, ref := range pathParams {
		if ref == nil || ref.Value == nil {
			continue
		}
		params = append(params, mapParameter(ref.Value))
	}
	for _, ref := range op.Parameters {
		if ref == nil || ref.Value == nil {
			continue
		}
		params = append(params, mapParameter(ref.Value))
	}

	var body *RequestBody
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		body = mapRequestBody(op.RequestBody.Value)
	}

	responses := map[string]Response{}
	if op.Responses != nil {
		for code, ref := range op.Responses.Map() {
			if ref == nil || ref.Value == nil {
				continue
			}
			responses[code] = mapResponse(ref.Value)
		}
	}

	id := op.OperationID
	if id == "" {
		id = strings.ToUpper(method) + " " + path
	}

	return Operation{
		ID:          id,
		Method:      strings.ToUpper(method),
		Path:        path,
		Summary:     op.Summary,
		Parameters:  params,
		RequestBody: body,
		Responses:   responses,
	}, nil
}

func mapParameter(p *openapi3.Parameter) Parameter {
	out := Parameter{
		Name:     p.Name,
		In:       p.In,
		Required: p.Required,
		Example:  p.Example,
	}
	if p.Schema != nil && p.Schema.Value != nil {
		out.Schema = mapSchema(p.Schema.Value)
	}
	if len(p.Examples) > 0 {
		out.Examples = make([]any, 0, len(p.Examples))
		for _, ex := range p.Examples {
			if ex != nil {
				out.Examples = append(out.Examples, ex.Value)
			}
		}
	}
	return out
}

func mapRequestBody(body *openapi3.RequestBody) *RequestBody {
	content := map[string]MediaType{}
	for mt, media := range body.Content {
		if media == nil {
			continue
		}
		content[mt] = mapMediaType(media)
	}
	return &RequestBody{
		Required: body.Required,
		Content:  content,
	}
}

func mapResponse(resp *openapi3.Response) Response {
	content := map[string]MediaType{}
	for mt, media := range resp.Content {
		if media == nil {
			continue
		}
		content[mt] = mapMediaType(media)
	}
	headers := map[string]Header{}
	for name, ref := range resp.Headers {
		if ref == nil || ref.Value == nil {
			continue
		}
		h := Header{Required: ref.Value.Required}
		if ref.Value.Schema != nil && ref.Value.Schema.Value != nil {
			h.Schema = mapSchema(ref.Value.Schema.Value)
		}
		headers[name] = h
	}
	desc := ""
	if resp.Description != nil {
		desc = *resp.Description
	}
	return Response{
		Description: desc,
		Headers:     headers,
		Content:     content,
	}
}

func mapMediaType(media *openapi3.MediaType) MediaType {
	out := MediaType{Example: media.Example}
	if media.Schema != nil && media.Schema.Value != nil {
		out.Schema = mapSchema(media.Schema.Value)
	}
	if len(media.Examples) > 0 {
		out.Examples = make([]any, 0, len(media.Examples))
		for _, ex := range media.Examples {
			if ex != nil {
				out.Examples = append(out.Examples, ex.Value)
			}
		}
	}
	return out
}

func mapSchema(s *openapi3.Schema) JSONSchema {
	out := JSONSchema{
		Format:      s.Format,
		Enum:        append([]any(nil), s.Enum...),
		Default:     s.Default,
		Example:     s.Example,
		Pattern:     s.Pattern,
		UniqueItems: s.UniqueItems,
		Required:    append([]string(nil), s.Required...),
		Nullable:    s.Nullable,
		Raw:         map[string]any{},
	}

	if s.Type != nil {
		types := s.Type.Slice()
		if len(types) > 0 {
			out.Type = types[0]
		}
	}
	if s.Min != nil {
		out.Minimum = s.Min
	}
	if s.Max != nil {
		out.Maximum = s.Max
	}
	if s.ExclusiveMin.IsSet() {
		switch {
		case s.ExclusiveMin.Value != nil:
			v := *s.ExclusiveMin.Value
			out.ExclusiveMinimum = &v
		case s.ExclusiveMin.IsTrue() && s.Min != nil:
			v := *s.Min
			out.ExclusiveMinimum = &v
		}
	}
	if s.ExclusiveMax.IsSet() {
		switch {
		case s.ExclusiveMax.Value != nil:
			v := *s.ExclusiveMax.Value
			out.ExclusiveMaximum = &v
		case s.ExclusiveMax.IsTrue() && s.Max != nil:
			v := *s.Max
			out.ExclusiveMaximum = &v
		}
	}
	if s.MinLength != 0 {
		v := int(s.MinLength)
		out.MinLength = &v
	}
	if s.MaxLength != nil {
		v := int(*s.MaxLength)
		out.MaxLength = &v
	}
	if s.MinItems != 0 {
		v := int(s.MinItems)
		out.MinItems = &v
	}
	if s.MaxItems != nil {
		v := int(*s.MaxItems)
		out.MaxItems = &v
	}
	if s.MinProps != 0 {
		v := int(s.MinProps)
		out.MinProperties = &v
	}
	if s.MaxProps != nil {
		v := int(*s.MaxProps)
		out.MaxProperties = &v
	}
	if len(s.Properties) > 0 {
		out.Properties = make(map[string]JSONSchema, len(s.Properties))
		for name, ref := range s.Properties {
			if ref != nil && ref.Value != nil {
				out.Properties[name] = mapSchema(ref.Value)
			}
		}
	}
	if s.Items != nil && s.Items.Value != nil {
		items := mapSchema(s.Items.Value)
		out.Items = &items
	}
	if s.AdditionalProperties.Schema != nil && s.AdditionalProperties.Schema.Value != nil {
		add := mapSchema(s.AdditionalProperties.Schema.Value)
		out.AdditionalProperties = &add
	}
	return out
}

func resolveAgainst(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return baseURL.ResolveReference(refURL).String()
}
