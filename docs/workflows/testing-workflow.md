# Testing Workflow

## Scope

- Every change with behavior impact requires test updates.

## Required Execution

Run the following before considering the task complete:

```bash
make verify
```

`make verify` executes:

1. `gofmt ./...`
2. Lint gate (`scripts/verify-lint.sh`: `go vet`; `golangci-lint` when installed)
3. `go test ./...`
4. Coverage gate (`scripts/verify-changed-coverage.sh`: 100% function coverage on packages changed since `origin/main`)
5. `make build`
6. `make graphify-update`

### Coverage gate options

- `VERIFY_BASE_REF=<git-ref>` — change the diff base (default: `origin/main`)
- `VERIFY_COVERAGE_ALL=1 make verify-coverage` — enforce all `internal/modules/*` packages

## Outcome Rules

- Changed production packages must reach **100% function coverage** with tests for **all branches**.
- Validation is incomplete while lint diagnostics remain in changed files.
- If a test fails or coverage is below 100%, resolve before merge.
- If test coverage is missing due to ambiguity, request clarification and document the gap.
