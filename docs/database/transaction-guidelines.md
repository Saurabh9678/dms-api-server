# Transaction Guidelines

## Principles

- Use transactions for multi-step state changes requiring atomicity.
- Keep transaction boundaries explicit and minimal.
- Do not mix unrelated operations in a single transaction.

## Failure Handling

- Roll back on error.
- Return domain-relevant errors for upstream mapping.

## Update Checklist

- Update this file when transaction handling conventions change.
