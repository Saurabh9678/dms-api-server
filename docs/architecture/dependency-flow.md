# Dependency Flow

## Allowed Direction

- handlers/routes -> services
- services -> repository/provider interfaces
- infra implementations -> external systems

## Restricted Direction

- handlers must not contain business logic
- module logic must not depend on infra concrete implementations
- avoid cyclic dependencies across modules

## Cross-Module Dependencies

- Document each allowed cross-module dependency with rationale.

## Update Rules

- Update this file whenever dependency rules change.
