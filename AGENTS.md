# Engineering Constitution

This file defines mandatory engineering governance for this repository. These rules apply to every task unless the user explicitly overrides them.

## 1) Before Starting Any Task

The agent must complete all checks below before implementing:

1. Understand existing architecture and current implementation.
2. Check `docs/knowledge-base.md` and relevant docs sections first.
3. Identify impacted modules, providers, APIs, and database areas.
4. Verify dependency direction and module boundaries.
5. Verify target folder placement before creating files.
6. Search for existing implementation to avoid duplicate logic.
7. Prefer extending existing code over rewriting stable code.
8. Plan minimal scoped changes that satisfy only the requirement.

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

## 5) Knowledge Base Rule

`docs/knowledge-base.md` is required project memory and must stay current.

Before implementing:

- consult knowledge base sections relevant to the task

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
