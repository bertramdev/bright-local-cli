# Next Session Handoff

## Priority

- Add delete commands only after defining a stronger confirmation flow.
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
- Current commands include read-only operations plus confirmed create/update operations for clients and locations.
- Latest feature commit: `c5f6fda`
- Remote: public `tomleelong/bright-local-cli`, branch `main`
- Verification baseline: `go test ./...`, `go vet ./...`, and `git diff --check`

## Completed write support checks

- Verify each write endpoint and request schema against the current BrightLocal Management API documentation.
- Add unit tests proving confirmation is required and requests are not sent when confirmation is missing.
- Run the full verification baseline and review the staged diff.
