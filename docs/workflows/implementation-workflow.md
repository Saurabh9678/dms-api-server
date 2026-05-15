# Implementation Workflow

## Pre-Implementation

1. Review `docs/knowledge-base.md` and relevant module/provider/api docs.
2. Inspect existing implementation and dependency boundaries.
3. Confirm folder placement and impacted modules.
4. Ask clarification if requirement or placement is ambiguous.

## Implementation

1. Apply minimal scoped changes.
2. Extend existing code where possible; avoid unnecessary rewrites.
3. Keep boundary rules intact.

## Post-Implementation

1. Update tests.
2. Update documentation sections impacted by the change.
3. If APIs changed, update `docs/api/<module>.postman_collection.json` (Postman importable).
4. Run:
   - `gofmt ./...`
   - `go vet ./...`
   - `go test ./...`
   - `make build`
   - `make graphify-update`
