# Next Session Handoff

## Priority

- Add BrightLocal Management API write commands, starting with create/update operations.
- Keep write operations separate from the existing read-only command groups.

## Required safety behavior

- Require an explicit `--confirm` flag for every mutating command.
- Prompt interactively with the target resource and operation before sending a write.
- In non-interactive environments, refuse writes unless `--confirm` is present.
- Require stronger confirmation for deletes (for example, `--confirm DELETE`).
- Add `--dry-run` to show the request without sending it.
- Do not add legacy `apidocs.brightlocal.com` endpoints unless explicitly requested.

## Current state

- Project: `bright-local-cli`
- API base: `https://api.brightlocal.com`
- Authentication: `x-api-key` header from `BRIGHTLOCAL_API_KEY` or `--api-key`
- Current commands are GET-only and intentionally perform no mutations.
- Latest feature commit: `00772d5`
- Remote: private `tomleelong/bright-local-cli`, branch `main`
- Verification baseline: `go test ./...`, `go vet ./...`, and `git diff --check`

## Before committing write support

- Verify each write endpoint and request schema against the current BrightLocal Management API documentation.
- Add unit tests proving confirmation is required and requests are not sent when confirmation is missing.
- Run the full verification baseline and review the staged diff.
