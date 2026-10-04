## What this changes

<!-- One or two sentences. What will a user observe differently after this merges? -->

## Why

<!-- Link the issue: Fixes #123 -->

## Type of change

- [ ] Bug fix
- [ ] New check
- [ ] New generation phase or mode
- [ ] CLI / configuration
- [ ] Output or reporting
- [ ] Documentation
- [ ] Refactor or cleanup
- [ ] Build / CI

## Checklist

- [ ] `make lint` passes
- [ ] `make test` passes, including `go test -race ./...` if the change touches `internal/runner` or `internal/httpx`
- [ ] New behaviour is covered by a test
- [ ] README updated if flags, checks, phases, exit codes or output changed
- [ ] If `go.mod` changed, `make vendor` was re-run and `vendor/` is committed with it
- [ ] No tokens, credentials, internal hostnames or customer data in the diff

## Notes for the reviewer

<!-- Anything non-obvious: a deliberate trade-off, a known gap, a follow-up you are not including. -->
