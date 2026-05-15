# Debugging Workflow

## Process

1. Reproduce issue with clear inputs and expected output.
2. Identify impacted layer/module without broad refactors.
3. Fix with minimal scoped change.
4. Add or update tests to prevent regression.
5. Update docs if behavior or contracts change.

## Clarification Rule

- If expected behavior is unclear, ask clarification before implementing a fix.

## Validation

- Run required checks:
  - `gofmt ./...`
  - `go vet ./...`
  - `go test ./...`
  - `make build`
