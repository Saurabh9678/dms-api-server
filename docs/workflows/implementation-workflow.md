# Implementation Workflow

## Pre-Implementation

1. Start at `docs/index.md`, then load only task-relevant docs.
2. Review `docs/knowledge-base.md` and relevant module/provider/api docs.
3. Inspect existing implementation and dependency boundaries.
4. Confirm folder placement and impacted modules.
5. Ask clarification if requirement or placement is ambiguous.
6. Keep review retrieval scoped to task-relevant files only.
7. Separate required changes from optional improvements before implementation.
8. Keep review notes concise and limited to implementation-relevant findings.

## Implementation

1. Apply minimal scoped changes.
2. Extend existing code where possible; avoid unnecessary rewrites.
3. Keep boundary rules intact.

## Post-Implementation

1. Update tests.
2. Update documentation sections impacted by the change.
3. If APIs changed, update `docs/api/<module>.postman_collection.json` and `postman/collections/DMS API/<module>/` in the same task (keep both in sync; follow Postman YAML conventions in `.cursor/rules/api-documentation.mdc`).
4. Run:
   - `gofmt ./...`
   - `go vet ./...`
   - `go test ./...`
   - `make build`
   - `make graphify-update`
