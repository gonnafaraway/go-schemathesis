# go-schemathesis

Go implementation of [Schemathesis](https://github.com/schemathesis/schemathesis)-style API testing: load an OpenAPI schema, generate request cases, execute them against a live API, and validate responses with built-in checks.

Code style follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## Features

- OpenAPI 3.x schema loading (file or HTTP URL)
- Generation phases: `examples`, `coverage`, `fuzzing`
- Modes: `positive`, `negative`, `all`
- Concurrent workers
- Checks:
  - `not_a_server_error`
  - `status_code_conformance`
  - `content_type_conformance`
  - `response_schema_conformance`
  - `negative_data_rejection`
  - `positive_data_acceptance`
- Minimal `curl` reproducers in the report

## Install

```bash
go install github.com/gonnafaraway/go-schemathesis/cmd/schemathesis@latest
```

Or build locally:

```bash
make build
```

## Usage

```bash
schemathesis run examples/petstore.yaml --url http://127.0.0.1:8080

schemathesis run https://example.com/openapi.json \
  --workers auto \
  --mode all \
  --phases examples,coverage,fuzzing \
  --max-examples 50 \
  -H "Authorization: Bearer TOKEN"
```

### Exit codes

| Code | Meaning |
| --- | --- |
| 0 | All checks passed |
| 1 | One or more checks failed |
| 2 | Runtime / configuration error |

## Project layout

```text
cmd/schemathesis/   # entrypoint
internal/
  cli/              # Cobra commands + composition root
  config/           # runtime config
  schema/           # OpenAPI model + loader
  casegen/          # case generation (examples/coverage/fuzz)
  check/            # response checks
  httpx/            # HTTP client for API under test
  runner/           # load → generate → execute → check
  report/           # summary + curl/console output
examples/           # sample OpenAPI schema
```

Dependencies point inward to pure packages (`schema`, `casegen`, `check`). I/O lives in `schema` loader, `httpx`, and `cli`.

## Development

```bash
make test
make lint
go mod tidy
go mod vendor
```

## License

Apache License 2.0
