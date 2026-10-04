<h1 align="center" style="border-bottom: none">
    Property-based API testing for your OpenAPI schema, in Go
</h1>

<div align="center">
  <a href="https://github.com/gonnafaraway/go-schemathesis/actions/workflows/ci.yml"><img src="https://github.com/gonnafaraway/go-schemathesis/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://pkg.go.dev/github.com/gonnafaraway/go-schemathesis"><img src="https://img.shields.io/badge/go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26" /></a>
  <img src="https://img.shields.io/badge/OpenAPI-3.0%20%7C%203.1-00ADD8?logo=openapi&logoColor=white" alt="OpenAPI 3.x" />
</div>

<br />

<p align="center">
    <a href="https://github.com/schemathesis/schemathesis">Inspired by <b>Schemathesis</b>, the industry-standard property-based testing tool for OpenAPI APIs</a>
</p>

## 🔍 What is go-schemathesis?

**go-schemathesis** is a single-binary, dependency-light CLI that tests an HTTP API straight from its OpenAPI schema. It loads the spec, generates request cases from the schema itself, executes them concurrently against a live API, and validates every response with a set of built-in checks — reporting failures together with a ready-to-paste `curl` reproducer.

It is a Go port of the [Schemathesis](https://github.com/schemathesis/schemathesis) approach, written with no runtime beyond the standard library, [`kin-openapi`](https://github.com/getkin/kin-openapi) for spec parsing, and [`cobra`](https://github.com/spf13/cobra) for the CLI. No Python, no Docker, no plugin runtime.

Use it as a pre-merge gate in CI, as a nightly sweep against staging, or as a quick way to answer *"does my API actually behave the way my spec says it does?"*

```text
OpenAPI schema ──▶ schema loader ──▶ case generator ──▶ concurrent workers ──▶ checks ──▶ report + curl
   file / HTTP        validate         examples              up to 64            6 checks    exit code
```

## 📖 Table of Contents

- [✨ Key Features](#-key-features)
- [🚀 Quick Start](#-quick-start)
- [🧪 Generation Phases](#-generation-phases)
- [🎯 Generation Modes](#-generation-modes)
- [🛡️ Checks](#-checks)
- [⚙️ CLI Reference](#-cli-reference)
- [🔐 Testing an Authenticated API](#-testing-an-authenticated-api)
- [💾 Testing with Real Data from Your Database](#-testing-with-real-data-from-your-database)
- [🚦 CI Integration](#-ci-integration)
- [🧱 Project Layout](#-project-layout)
- [⚠️ Known Limitations](#-known-limitations)
- [🧑‍💻 Development](#-development)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [⭐️ Stay Updated](#-stay-updated)

## ✨ Key Features

- **Spec-driven test generation.** Cases come from the schema, not from hand-written fixtures — parameters, request bodies, and their boundaries are derived automatically.
- **Three generation phases.** `examples` uses the examples you already wrote in the spec, `coverage` walks declared boundaries (`minLength`, `minimum`, `maxItems`, …), `fuzzing` explores the space randomly from a seeded RNG.
- **Positive and negative data.** Generate schema-valid data, schema-violating data, or both — and let the checks tell you whether your API actually enforces its own contract.
- **Six built-in checks.** Server errors, undocumented status codes, `Content-Type` conformance, response schema conformance, and acceptance/rejection of valid and invalid data.
- **Concurrency.** 1 to 64 workers, or `auto` to match your CPU count.
- **`curl` reproducers.** Every failure is reported with the exact command that reproduces it, correctly shell-quoted.
- **Reproducible runs.** `--seed` pins the RNG so a failure you found today still reproduces tomorrow.
- **Zero-config input.** Credentials and any extra headers go straight through `-H`. No config file, no env-var conventions to learn.
- **Vendored, auditable, tiny.** Two direct dependencies, no cgo, no runtime services.

## 🚀 Quick Start

### Install

With Go 1.26 or newer:

```bash
go install github.com/gonnafaraway/go-schemathesis/cmd/schemathesis@latest
```

Or download a prebuilt binary — Linux, macOS and Windows, amd64 and arm64, published on the [releases page](https://github.com/gonnafaraway/go-schemathesis/releases) alongside a `SHA256SUMS` file to verify against.

Or build from source:

```bash
git clone https://github.com/gonnafaraway/go-schemathesis.git
cd go-schemathesis
make build

./bin/schemathesis --version   # schemathesis version v0.1.0
```

There is no runtime dependency: no Python, no Docker, no services to run.

<details>
<summary>Shell completion</summary>

```bash
# bash
./bin/schemathesis completion bash > /etc/bash_completion.d/schemathesis

# zsh / fish / powershell are also supported
./bin/schemathesis completion zsh
./bin/schemathesis completion fish
./bin/schemathesis completion powershell
```
</details>

### Run

> [!WARNING]
> **This tool attacks the API you point it at.** The `negative` and `fuzzing` phases deliberately send data that violates your schema, and some of it will reach code paths that were never meant to be reachable. Only run it against APIs you own or have permission to test, prefer a staging environment, and confirm the wiring with `--phases examples --mode positive` before widening. See the [threat model](SECURITY.md#threat-model).

Point it at a schema and a running API:

```bash
schemathesis run examples/petstore.yaml --url http://127.0.0.1:8080
```

A fuller invocation against a remote API with authentication and a tuned fuzzing budget:

```bash
schemathesis run https://api.example.com/openapi.json \
  --url https://api.example.com \
  --workers auto \
  --mode all \
  --phases examples,coverage,fuzzing \
  --max-examples 50 \
  --request-timeout 30s \
  -H "Authorization: Bearer $TOKEN"
```

### Output

Every failure gets a numbered block with a copy-pasteable reproducer:

```text
========================================================================
Tests: 214  Passed: 210  Failed: 4  Errors: 0  Duration: 3.418s
========================================================================

[1] Response violates schema
Check: response_schema_conformance
Operation: GET /pets/{petId} (getPetById)
Phase/Mode: fuzzing / positive
Status: 200
Message: expected string, got number
Reproduce: curl -X 'GET' 'http://127.0.0.1:8080/pets/1'
```

Add `-v` for progress logging on stdout.

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Run completed, no check failed |
| `1` | One or more checks failed, **or** a configuration/schema error occurred |
| `2` | _Not used_ |

> [!WARNING]
> Failures and configuration errors share exit code `1`, and transport errors (`Errors` in the summary) do **not** affect the exit code at all. A completely unreachable API can therefore exit `0`. In CI, assert on the `Errors:` field, or check that `Tests:` is non-zero. See [Known Limitations](#-known-limitations).

## 🧪 Generation Phases

Phases are selected with `--phases` (comma-separated, default `examples,coverage,fuzzing`).

| Phase | What it does |
| --- | --- |
| `examples` | Uses `example`, `examples[*].value`, and `schema.example` straight from your spec. If an operation has no examples at all, it falls back to one generated positive case. |
| `coverage` | Emits one case per declared boundary value — `minLength`/`maxLength`, `minimum`/`maximum`, `minItems`/`maxItems`, `enum` members, both booleans — each in a case where everything else is valid. Also covers omitting a required request body. |
| `fuzzing` | Explores the schema randomly with a seeded RNG: `max-examples` cases per operation, per mode. |

Because the spec's own examples are the highest-value input, a well-annotated OpenAPI document directly increases the quality of the run.

## 🎯 Generation Modes

`--mode` controls whether generated data is expected to be accepted (`positive`), rejected (`negative`), or both (`all`, the default).

| `--mode` | `examples` | `coverage` | `fuzzing` |
| --- | --- | --- | --- |
| `positive` | positive | positive | positive |
| `negative` | positive | negative | negative |
| `all` | positive | positive + negative | positive + negative |

Two things worth internalizing:

- The `examples` phase is **always positive** — your documented examples describe valid usage, so they are never mutated.
- In the default `all` mode, `coverage` and `fuzzing` each emit a positive *and* a negative variant, so a run issues roughly `operations × (examples + coverage + 2 × --max-examples)` requests. Lower `--max-examples` or narrow `--phases` for large specs.

## 🛡️ Checks

Selected with `--checks` (comma-separated names, or `all`) and `--exclude-checks`.

| Check | Fails when |
| --- | --- |
| `not_a_server_error` | The response status is `5xx`. |
| `status_code_conformance` | The status code is not documented for that operation — no exact code, no `NXX`/`Nxx` wildcard, no `default`. |
| `content_type_conformance` | A non-empty body has no `Content-Type`, an unparseable one, or a media type absent from the documented responses. Accepts documented `type/*` and `*/*`. |
| `response_schema_conformance` | A JSON body violates the schema documented for the returned status and media type. |
| `negative_data_rejection` | A **negative** case is accepted: schema-violating data should have been refused. |
| `positive_data_acceptance` | A **positive** case is refused: schema-compliant data should have been accepted. |

Skipping a case is deliberate, not a bug: checks that need a documented response, a schema, or a non-empty body stay quiet when the spec provides none.

Note that `positive_data_acceptance` and `negative_data_rejection` both treat `5xx` as an acceptable outcome and leave that verdict to `not_a_server_error`, so the three checks do not report the same failure twice.

## ⚙️ CLI Reference

```text
schemathesis run SCHEMA [flags]
```

`SCHEMA` is a path or an `http(s)` URL to an OpenAPI 3.x document, in JSON or YAML.

| Flag | Default | Description |
| --- | --- | --- |
| `-u`, `--url` | _from `servers[0]`_ | Base URL of the API under test. Required if the schema declares no server. |
| `-w`, `--workers` | `1` | Concurrent workers, `1`–`64`, or `auto` for `runtime.NumCPU()`. |
| `--phases` | `examples,coverage,fuzzing` | Comma-separated phases. |
| `--mode` | `all` | `all`, `positive`, or `negative`. |
| `-c`, `--checks` | `all` | Comma-separated check names. |
| `--exclude-checks` | _empty_ | Comma-separated check names to skip. |
| `--max-examples` | `100` | Fuzzing cases per operation, per mode. |
| `--max-failures` | `0` | Stop after N failures. `0` disables the limit. |
| `--request-timeout` | `10s` | Whole-request HTTP timeout. |
| `-H`, `--header` | _none_ | Extra request header, `Name: Value`. Repeatable. |
| `--seed` | `0` | RNG seed. `0` uses the current time; pin it to reproduce a run. |
| `-v`, `--verbose` | `false` | Debug-level logging on stdout. |

## 🔐 Testing an Authenticated API

Headers passed with `-H` are attached to the client and sent with **every** generated request. There is no config file and no env-var convention — the shell is your config layer.

```bash
TOKEN=$(curl -s -X POST https://api.example.com/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ci@example.com","password":"s3cret"}' | jq -r .access_token)

schemathesis run https://api.example.com/openapi.json \
  --url https://api.example.com \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Environment: staging" \
  --phases examples,coverage \
  --mode positive
```

- **Multiple headers:** repeat `-H`. The name and value are trimmed, and the split is on the first `:`, so values may contain colons.
- **HTTP Basic:** embed credentials in the base URL — `--url https://ci:s3cret@api.example.com`.
- **Cookie auth:** there is no cookie jar, so fetch the cookie yourself and pass it explicitly with `-H "Cookie: session=..."`.

Two caveats specific to auth:

> [!IMPORTANT]
> **Generated headers override your `-H` headers.** Client headers are applied first, then the headers of each generated case (`internal/httpx/client.go`). If your spec declares `Authorization`, `Cookie`, or `Content-Type` as an explicit header *parameter*, the generator will overwrite your real value with generated data. Model authentication as a `securityScheme` instead of a header parameter, and verify what is actually sent before trusting a run.

> [!NOTE]
> **Redirects are not followed**, so an API that authenticates by redirecting to a login page will surface a `3xx` response rather than the real answer. Multi-step flows (log in, then reuse the token within the same run) are not supported — obtain the credential outside the tool, as shown above.

Expired credentials will not raise a loud alarm: `401` and `403` count as *accepted* for `positive_data_acceptance`, and `not_a_server_error` only inspects `5xx`. Document `401`/`403` responses and watch the `Tests:`/`Failed:` counts to notice that a run stopped authenticating.

## 💾 Testing with Real Data from Your Database

For endpoints that address existing entities — `GET /users/{uuid}`, `GET /orders/{id}` — the tool **has no access to your database**. Fuzzing generates a syntactically valid but fabricated identifier, so the request will return `404`.

That `404` usually passes silently: `not_a_server_error` only flags `5xx`, and `404` is treated as an acceptable outcome by `positive_data_acceptance`. The only check that will react is `status_code_conformance`, and only if `404` is missing from the spec.

**The supported way to feed real identifiers in is the `examples` phase.** Values are taken verbatim from the spec:

```yaml
paths:
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
          examples:
            primary:   { value: 3f2504e0-4f89-11d3-9a0c-0305e82c3301 }
            secondary: { value: 9c858901-8a57-4791-81fe-4c455b099bc9 }
      responses:
        '200': { description: OK }
        '404': { description: Not found }
```

```bash
schemathesis run openapi.yaml --url http://127.0.0.1:8080 \
  --phases examples \
  --mode positive
```

`parameter.example`, `parameter.examples[*].value`, and `parameter.schema.example` are all read. Multiple examples are zipped across parameters, so N examples produce N cases. A practical workflow is to generate this document once from your fixtures or seed data, then commit it — the spec doubles as your fixture manifest.

What will not work, by design:

- **No inter-case correlation.** A `POST /users` followed by a `GET /users/{id}` using the created ID is not possible; the generated ID never comes from an earlier response.
- **`--seed` does not help here.** It makes generation reproducible, not related to your data.
- **No setup or teardown hook.** Fixtures must already exist when the run starts.

For richer, less arbitrary values in general, note that only five formats are generated deliberately: `date-time`, `date`, `email`, `uuid`, and `uri`/`url`. Everything else is a random alphanumeric string, and `pattern` constraints are ignored during generation.

## 🚦 CI Integration

Because the exit code does not distinguish failures from errors, assert on the summary line as well:

```yaml
name: API contract tests
on: [pull_request]

jobs:
  schemathesis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }

      - run: go install github.com/gonnafaraway/go-schemathesis/cmd/schemathesis@latest

      - name: Fetch token
        run: echo "TOKEN=$(curl -sf -X POST $BASE_URL/auth/login -d @creds.json | jq -r .access_token)" >> "$GITHUB_ENV"

      - name: Run
        run: |
          schemathesis run openapi.yaml \
            --url "$BASE_URL" \
            -H "Authorization: Bearer $TOKEN" \
            --workers auto \
            --max-examples 50 \
            | tee schemathesis.log
          grep -qE 'Errors: 0' schemathesis.log   # transport errors must not pass silently
```

> [!TIP]
> Pin `--seed` and keep it stable across runs so that a newly reported failure is the only new thing in the log. Use `--exclude-checks` to silence a known-noisy check while you fix it, rather than deleting the whole run.

## 🧱 Project Layout

```text
cmd/schemathesis/   # entrypoint
internal/
  cli/              # Cobra commands + composition root
  config/           # runtime config normalization and defaults
  schema/           # OpenAPI model + loader (file, HTTP URL, JSON, YAML)
  casegen/          # case generation (examples / coverage / fuzzing, positive & negative)
  check/            # response checks + registry
  httpx/            # HTTP client for the API under test
  runner/           # load → generate → execute → check
  report/           # summary, console output, curl rendering
examples/           # sample OpenAPI schema
```

Dependencies point inward toward the pure packages (`schema`, `casegen`, `check`). All I/O lives in the `schema` loader, `httpx`, and `cli`. Code style follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## ⚠️ Known Limitations

Worth reading before you trust a green run.

**Results and exit codes**

- Configuration errors and check failures both exit `1`; there is no distinct error code.
- Transport errors are counted in `Errors:` but do not affect the exit code. An unreachable or DNS-failing API can exit `0`.

**Data and fixtures**

- No database access, no fixtures, no setup/teardown hooks, and no correlation between generated cases.
- Fabricated identifiers produce `404`, which is deliberately not treated as a failure. See [Testing with Real Data](#-testing-with-real-data-from-your-database).
- Only `date-time`, `date`, `email`, `uuid`, and `uri`/`url` receive format-aware values; `pattern` is ignored while generating.

**Spec support**

- OpenAPI 3.0.x and 3.1.x only — Swagger 2.0 is not supported.
- Only `servers[0]` is used as the base URL.
- In 3.1, only the first entry of a `type` union is read (`type: ["string", "null"]` loses `"null"`), `webhooks` are ignored, and `const` is never extracted.
- Response schema validation is a hand-rolled subset. It covers `type`, `enum`, `required`, `properties`, `items`, `minItems`, `maxItems`, `minLength`, `maxLength`, `minimum`, `maximum`, and `nullable`. It does **not** cover `pattern`, `format`, `uniqueItems`, `minProperties`, `maxProperties`, `additionalProperties`, `const`, `exclusiveMinimum`, `exclusiveMaximum`, or `allOf`/`anyOf`/`oneOf`/`not`.
- Non-JSON response bodies are never schema-validated, and an undocumented response media type falls back to an arbitrary documented one.

**Networking**

- No TLS customization: no CA bundle, no client certificates, no way to skip verification.
- No proxy configuration, no cookie jar, and no environment-variable credentials.
- Redirects are never followed.
- Response bodies are read up to 8 MiB and silently truncated beyond that; schema fetches are capped at 32 MiB with a fixed 30s timeout.

**Output**

- Console output only — no JSON, JUnit, or SARIF report, and no flag to write to a file.
- Case ordering and failure ordering are not deterministic even with a fixed `--seed`, because operation, media-type, and header iteration come from Go maps.

## 🧑‍💻 Development

```bash
make build      # -> bin/schemathesis, version stamped from git describe
make install    # -> $GOBIN/schemathesis, version stamped
make test       # go test ./...
make test-race  # go test -race ./...
make lint       # golangci-lint run ./...
make run        # go run ./cmd/schemathesis
make tidy       # go mod tidy
make vendor     # go mod vendor
make clean      # rm -rf bin
```

The version reported by `--version` comes from `git describe --tags --always --dirty` and is injected with `-ldflags -X`, so a build from a tag reports that tag. Override it with `make build VERSION=v1.2.3`.

The test suite includes an end-to-end test that runs the full pipeline against an in-process `httptest` server (`internal/runner/runner_test.go`). Coverage is currently uneven — `internal/runner` and `internal/schema` are solid, while `internal/cli`, `internal/config`, `internal/httpx` and `internal/report` are barely tested. Adding tests there needs no behaviour change and is the easiest way to contribute.

Dependencies are vendored. Run `make tidy && make vendor` and commit the result together with your change — a pull request that touches `go.mod` without refreshing `vendor/` fails CI.

GitHub Actions runs `golangci-lint` and `go test -race` on every pull request, on every push to `main`, and when a release is published. [`.github/workflows/ci.yml`](.github/workflows/ci.yml) pulls the Go toolchain straight from `go.mod`, so the two can never drift apart. Publishing a release additionally cross-compiles binaries for six platforms and attaches them with checksums — see [`.github/workflows/release.yml`](.github/workflows/release.yml).

> [!NOTE]
> `.gitattributes` pins the working tree to LF. Git for Windows otherwise checks files out with CRLF, which makes `gofmt` — and therefore `make lint` — fail locally on every file while CI stays green.

## 🤝 Contributing

Contributions are welcome. The mechanics live in [CONTRIBUTING.md](CONTRIBUTING.md); this is the short version.

- **Report a bug:** use the [bug report template](https://github.com/gonnafaraway/go-schemathesis/issues/new?template=bug_report.yml). The exact command, `schemathesis --version`, the `-v` log and a minimal schema are the most useful attachments. Redact your tokens.
- **Request a feature:** use the [feature request template](https://github.com/gonnafaraway/go-schemathesis/issues/new?template=feature_request.yml). Describe the situation rather than the solution.
- **Send a pull request:** run `make lint` and `make test-race`, keep the change focused, fill in the template, and follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).
- **Improve the docs:** if a check surprised you or a flag was unclear, that is exactly the kind of gap worth filing.

New checks belong in `internal/check` behind the registry in `registry.go`, with a test in `checks_test.go`. [CONTRIBUTING.md](CONTRIBUTING.md#adding-a-check) walks through the steps.

Participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md). Because this tool generates hostile input on purpose, please read the [threat model in SECURITY.md](SECURITY.md#threat-model) before pointing it at anything you do not own — and report vulnerabilities through [GitHub Security Advisories](https://github.com/gonnafaraway/go-schemathesis/security/advisories/new), not the issue tracker.

## 📄 License

Licensed under the Apache License 2.0 © [gonnafaraway](https://github.com/gonnafaraway). See [LICENSE](LICENSE) for the full text. Notable changes are recorded in [CHANGELOG.md](CHANGELOG.md).


## ⭐️ Stay Updated

If this saved you from hand-writing conformance tests, a star goes a long way.

[![Star the Repo](https://img.shields.io/github/stars/gonnafaraway/go-schemathesis?style=social)](https://github.com/gonnafaraway/go-schemathesis)

---

Built for teams whose API contract and API behaviour have quietly drifted apart.
