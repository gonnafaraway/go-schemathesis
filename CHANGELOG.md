# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Note that while the project is pre-1.0, `internal/` is the entire public surface: the module ships a CLI and no importable Go API, so breaking changes are limited to CLI behaviour.

## [Unreleased]

Nothing yet.

## [0.1.0] - 2026-10-04

First release.

### Added

- `schemathesis run SCHEMA` — property-based API testing driven by an OpenAPI 3.x document, loaded from a file path or an `http(s)` URL, in JSON or YAML.
- Three generation phases: `examples` (values taken verbatim from the spec), `coverage` (declared boundaries such as `minLength`, `minimum`, `maxItems`), and `fuzzing` (seeded random exploration).
- Three generation modes: `positive`, `negative`, `all`.
- Six checks: `not_a_server_error`, `status_code_conformance`, `content_type_conformance`, `response_schema_conformance`, `negative_data_rejection`, `positive_data_acceptance`.
- Concurrent execution, from 1 to 64 workers, or `auto` to match the CPU count.
- Console report with a numbered failure block and a shell-quoted `curl` reproducer for each failure.
- Authentication through repeatable `-H` headers, which are attached to every generated request.
- Reproducible runs via `--seed`.
- Shell completion for bash, zsh, fish and PowerShell.

[Unreleased]: https://github.com/gonnafaraway/go-schemathesis/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/gonnafaraway/go-schemathesis/releases/tag/v0.1.0
