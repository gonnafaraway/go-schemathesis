# Security Policy

## Reporting a Vulnerability

Please **do not** open a public issue for a security problem.

Report it privately through [GitHub Security Advisories](https://github.com/gonnafaraway/go-schemathesis/security/advisories/new). Use the "Report a vulnerability" button, or the `Security` tab → `Advisories` → `New draft advisory`.

Include the version or commit, the command you ran, the schema fragment that triggers it, and the observed behaviour. You can expect an acknowledgement within a few days and a status update as the fix progresses.

## Supported Versions

| Version | Supported |
| --- | --- |
| `0.1.x` | ✅ |
| `< 0.1.0` | ❌ |

This project is pre-1.0. Security fixes are applied to the latest minor release only.

## Threat Model

go-schemathesis **generates hostile input on purpose**. Understanding this is the most important part of using it safely.

### It will attack the API you point it at

The `negative` and `fuzzing` phases deliberately send requests that violate your schema: wrong types, missing required fields, out-of-bounds numbers, unexpected media types. Some of those payloads will reach code paths that were never meant to be reachable.

- Only run it against APIs you own or have explicit permission to test.
- Start with `--phases examples --mode positive` to confirm the wiring, then widen.
- Prefer a dedicated staging environment. Never point the fuzzer at production.

### Treat schemas from untrusted sources as code

The schema is the input that drives everything, and external references are resolved:

- **External `$ref`s are followed.** A schema hosted somewhere you do not control can cause go-schemathesis to make outbound requests to arbitrary hosts, and to read local files via `file://` references. Only load schemas you trust.
- **Schema values are echoed into the report.** Parameter headers and request bodies are rendered into the `curl` reproducer. A secret pasted into a schema `example` or `default` will be printed to stdout and captured in CI logs.

### Credentials

Headers passed with `-H` are attached to the client and are **not** included in the `curl` reproducers, so a token supplied that way does not leak into the report.

The exception is anything the schema itself declares as a header parameter: those values are generated from the spec and do appear in the output. Model authentication as an OpenAPI `securityScheme`, not as a header parameter with a real credential in its example.

### What the tool deliberately does not do

- No TLS verification is ever skipped — there is no `--insecure` flag.
- Redirects are not followed, which limits credential-forwarding to redirect targets.
- Response bodies are read through an 8 MiB cap, and a fetched schema through a 32 MiB cap.
- The generated `curl` commands are shell-quoted for POSIX shells. They are safe to paste, but they are rendered for `sh`, not `cmd.exe` or PowerShell.

## Non-Issues

Reports about the following are not vulnerabilities, and will be closed as such:

- A generated request causing an error, panic, or unexpected response in the **API under test**. That is the tool working as designed — report it as an issue against your API, or use `--exclude-checks` while you fix it.
- The exit code not distinguishing configuration errors from check failures.
- Documented limitations listed in the [README](README.md#️-known-limitations).
