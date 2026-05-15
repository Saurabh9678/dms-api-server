# Release Workflow

## Pre-Release Checks

1. Validate required tests and build pass.
2. Ensure docs and API contracts are updated.
3. Ensure graph artifacts are refreshed when structural changes exist.

## Required Validation Commands

- `gofmt ./...`
- `go vet ./...`
- `go test ./...`
- `make build`
- `make graphify-update` (when structural changes are included)

## Release Notes Checklist

- Include behavior changes, API changes, and migration notes.
