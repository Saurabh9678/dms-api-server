# Engineering Constitution

This file defines mandatory engineering governance for this repository. These rules apply to every task unless the user explicitly overrides them.

## 1) Before Starting Any Task

The agent must complete all checks below before implementing:

1. Understand existing architecture and current implementation.
2. Check `docs/index.md` first, then open only task-relevant docs.
3. Check `docs/knowledge-base.md` and relevant docs sections first.
4. Use module docs (`docs/modules/*.md`) to trace endpoint/function flow before changing behavior.
5. Use schema docs (`docs/database/tables/*.md`) to validate table/column/constraint assumptions before DB-related changes.
6. Identify impacted modules, providers, APIs, and database areas.
7. Verify dependency direction and module boundaries.
8. Verify target folder placement before creating files.
9. Search for existing implementation to avoid duplicate logic.
10. Prefer extending existing code over rewriting stable code.
11. Plan minimal scoped changes that satisfy only the requirement.

Mandatory execution gate:

- For every new task, provide an implementation plan first.
- Resolve all requirement/behavior ambiguities before coding.
- Start implementation only after explicit user go-ahead.
- If go-ahead is not explicit, stay in planning/clarification mode.

Mandatory clarification rule:

- When requirements are ambiguous, clarification is mandatory before implementation.
- Never assume behavior, naming, placement, or architecture intent.
- Never invent architecture decisions without confirmation.

## 2) Implementation Rules

Required principles:

- SOLID
- DRY
- modular architecture
- low coupling
- high cohesion
- explicit, readable code over clever code

Forbidden patterns:

- giant files
- god services
- business logic in handlers/controllers
- module code importing infra implementation directly
- duplicated business logic
- premature abstractions

Additional constraints:

- Respect current folder/module architecture.
- Keep interfaces and ownership explicit.
- If existing implementation already satisfies the requirement, do not rewrite it.

## 3) Testing Rules

After every task, the agent must:

1. Create or update tests for changed behavior.
2. Run validation commands.
3. Ensure build passes.

Mandatory validation commands:

```bash
gofmt ./...
go vet ./...
go test ./...
make build
make graphify-update
```

## 4) Documentation Rules

Documentation updates are mandatory for any behavioral, structural, or contract change.

- Architecture change -> update `docs/architecture/`
- API change -> update `docs/api/`
- Module change -> update `docs/modules/`
- Provider change -> update `docs/providers/`
- Database/schema/migration behavior change -> update `docs/database/`
- Workflow/process change -> update `docs/workflows/`
- Module documentation must include endpoint flow details: route entry, middleware chain, handler, service logic path, and response branches.
- Schema documentation must be maintained in `docs/database/tables/` with one file per table and explicit column/constraint details.

## 5) Knowledge Base Rule

`docs/knowledge-base.md` is required project memory and must stay current.

Before implementing:

- consult knowledge base sections relevant to the task
- consult module flow docs and per-table schema docs for quick flow tracing

After implementing:

- update knowledge base with new decisions and caveats

Knowledge base must track:

- architecture decisions
- module responsibilities
- provider responsibilities
- dependency rules
- conventions
- migration notes
- important implementation details
- known caveats
- important workflows
- endpoint flow tracing references
- per-table schema references and caveats

## 6) API Documentation Rule

Whenever any API endpoint is created, modified, or removed, update `docs/api/<module>.postman_collection.json`.

API documentation must be written as Postman-importable JSON (Collection v2.1) and include:

- endpoint path
- method
- auth requirements
- request payload
- response payload
- error responses
- example request
- example response

## 7) Graphify Rule

After structural code changes:

1. Run `make graphify-update`.
2. Ensure graph artifacts remain updated.
3. Use graph output for architecture consistency checks when relevant.

## 8) Minimal Change Rule

The agent must:

- prefer the smallest safe implementation
- avoid unnecessary refactors
- avoid rewriting stable code
- avoid touching unrelated files

Explicit rule:

- If an existing implementation already satisfies the requirement, do not rewrite it.

## 9) Clarification Rule

If any of the following is unclear, stop and ask clarification questions before implementing:

- architecture intent
- naming
- folder placement
- requirement details
- competing implementation approaches

Never silently decide ambiguous details.

## 10) Dependency Direction Rule

Maintain directional boundaries:

- Handlers -> Services -> Repositories/Providers interfaces
- Infra implementations satisfy interfaces; domain/module logic should not depend on infra concrete types
- Module boundaries must remain explicit; cross-module coupling requires clear justification and docs updates

## 11) API v1 Device Context Rule

- Every `/api/v1/*` endpoint must require `X-Platform` and `X-Device-Id` headers.
- `X-Platform` must be one of: `web`, `ios_mobile`, `android_mobile`, `desktop`.
- `X-Device-Id` must be non-empty for all platforms, including web.
- Missing/invalid device-context headers must return error code `INVALID_DEVICE_CONTEXT` with message `invalid request`.
- Ensure new `/api/v1/*` routes inherit device-context enforcement automatically.

## 12) Scope Discipline Rule

- Do not refactor unrelated modules while implementing a task.
- Do not move existing modules unless explicitly requested.
- Keep PR scope narrow and verifiable.

## 13) Review Efficiency Rule

- Keep review narration concise and implementation-focused.
- Report only findings that impact implementation decisions or safety.
- Clearly separate required changes from optional improvements.
- Retrieve only task-relevant files; avoid scanning unrelated modules/docs.
- Prefer concise summaries and targeted retrieval over broad repository sweeps.
