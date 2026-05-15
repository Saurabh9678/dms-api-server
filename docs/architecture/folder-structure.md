# Folder Structure

## Objective

Define ownership and placement rules for repository folders.

## Top-Level Placement Guide

- `cmd/` for entry points
- `internal/modules/` for business modules
- `internal/providers/` for provider contracts
- `internal/infra/` for concrete infrastructure implementations
- `pkg/` for shared reusable package utilities
- `tests/` for integration/smoke/contract coverage

## Placement Rules

- Put code in the narrowest folder matching ownership.
- Do not place business logic in transport or infra layers.
- Validate placement before creating files.
