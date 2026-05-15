# Architecture Conventions

## Core Principles

- SOLID
- DRY
- low coupling
- high cohesion
- explicitness and readability

## Forbidden Patterns

- god services
- giant files
- duplicated business logic
- business logic in handlers
- premature abstractions

## Change Scope Conventions

- Prefer minimal scoped changes.
- If existing implementation already satisfies the requirement, do not rewrite it.
