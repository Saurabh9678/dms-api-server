# Testing Workflow

## Scope

- Every change with behavior impact requires test updates.

## Required Execution

Run the following before considering the task complete:

1. `gofmt ./...`
2. `go vet ./...`
3. `go test ./...`
4. `make build`

## Outcome Rules

- If a test fails, resolve or clarify before merge.
- If test coverage is missing due to ambiguity, request clarification and document the gap.
