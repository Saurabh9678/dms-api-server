# Migration Rules

## Governance

- Every schema change requires a migration.
- Do not modify old applied migrations; create a new migration instead.
- Keep migrations backward-safe for deployment process.

## Required Checks

- Validate SQL up/down behavior.
- Verify migration ordering and naming.
- Ensure related docs are updated.

## Update Checklist

- Update this file when migration process rules evolve.
