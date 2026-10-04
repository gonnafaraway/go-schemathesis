# Contributing to go-schemathesis

Contributions are welcome, from bug reports to whole features. This document covers the mechanics; the project's goals and limitations are described in the [README](README.md).

By participating you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md). Security problems go through [SECURITY.md](SECURITY.md), not the issue tracker.

## Getting Set Up

You will need Go 1.26 or newer — the version in `go.mod` is authoritative. Everything is vendored, so no dependency download is required for a first build.

```bash
git clone https://github.com/gonnafaraway/go-schemathesis.git
cd go-schemathesis

make build   # -> bin/schemathesis
make test    # go test ./...
make lint    # golangci-lint run ./...
```

If you touch dependencies:

```bash
make tidy && make vendor   # commit go.mod, go.sum and vendor/ together
```

> [!IMPORTANT]
> Because `vendor/` is committed, **updating a dependency requires re-running `make vendor`**. A pull request that changes `go.mod` without refreshing `vendor/modules.txt` will fail CI with an `inconsistent vendoring` error. This applies to automated Dependabot pull requests too — run `make vendor` and push the result.

## Code Style

The [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) is followed throughout. Beyond what `golangci-lint` enforces:

- Dependencies point inward. `schema`, `casegen` and `check` stay pure; all I/O lives in the `schema` loader, `httpx` and `cli`.
- Errors are wrapped with `%w` and carry context.
- Exported identifiers get a doc comment starting with the identifier's name. The `revive` `exported` rule is disabled, so this is on you.
- Line endings are LF everywhere, enforced by `.gitattributes`. Git for Windows otherwise checks files out with CRLF and `gofmt` fails locally while CI stays green.

Run `make lint` before opening a pull request. It runs the same `golangci-lint` configuration as CI, with the same pinned toolchain.

## Tests

```bash
make test                                        # everything
go test -race ./...                              # the runner is concurrent; please use -race
go test -run TestRunAgainstMockAPI ./internal/runner -v
```

Coverage is currently uneven and we know it: `internal/runner` and `internal/schema` are well exercised, while `internal/cli`, `internal/config`, `internal/httpx` and `internal/report` are close to untested. **Pull requests that add tests to those packages are especially welcome** — no behaviour change required.

`internal/runner/runner_test.go` runs the whole pipeline against an in-process `httptest` server. Extend that rather than reaching for a live API in tests.

## Adding a Check

Checks live in `internal/check` and are wired through the registry:

1. Add the name as a constant in `types.go`.
2. Implement `Check(ctx Context) *Failure` in its own file. Return `nil` to pass.
3. Register the implementation in `registry.go`.
4. Add it to `AllNames()` in `types.go` so `all` picks it up.
5. Add a test in `checks_test.go`.

A check must be deterministic, must not perform I/O, and must document what it deliberately skips — silently skipping is a design decision, not a bug.

## Adding a Generation Phase or Mode

Phases live in `internal/casegen` as a file per phase, selected in `generator.go`. Modes are resolved by `modesFor`, which decides which modes run for which phase — keep that table in mind, because `examples` is intentionally always positive.

## Pull Requests

- Branch from `main`, one topic per pull request.
- Fill in the template. If behaviour changes, say what a user will observe differently.
- Add or update tests, and update the README if you changed flags, checks, phases or output.
- CI must be green. Lint and tests run on every pull request, on every push to `main`, and on published releases.
- Conventional-ish commit messages (`feat:`, `fix:`, `docs:`, `ci:`, `refactor:`, `test:`, `chore:`) keep the changelog readable.

## Reporting Bugs

Use the [bug report template](https://github.com/gonnafaraway/go-schemathesis/issues/new?template=bug_report.yml). The most useful attachments are the exact command, the `schemathesis --version` output, the `-v` log, and a minimal schema that reproduces the problem. Please redact any tokens.
