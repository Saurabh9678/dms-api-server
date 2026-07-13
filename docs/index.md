# Documentation Index

Primary documentation entrypoint. Start here, then open only task-relevant docs.

## Architecture

- `docs/architecture/overview.md`
- `docs/architecture/folder-structure.md`
- `docs/architecture/dependency-flow.md`
- `docs/architecture/conventions.md`

## Modules

- `docs/modules/auth.md`
- `docs/modules/user.md`
- `docs/modules/showroom.md`
- `docs/modules/vehicle.md`
- `docs/modules/customer.md`

## Database

- `docs/database/schema-overview.md`
- `docs/database/migration-rules.md`
- `docs/database/transaction-guidelines.md`
- `docs/database/local-postgres.md`
- `docs/database/tables/`

## Providers

- `docs/providers/otp.md`
- `docs/providers/token.md`
- `docs/providers/email.md`
- `docs/providers/storage.md`
- `docs/providers/payment.md`

## Workflows

- `docs/workflows/implementation-workflow.md`
- `docs/workflows/testing-workflow.md`
- `docs/workflows/debugging-workflow.md`
- `docs/workflows/release-workflow.md`

## API Collections

Importable JSON (per module):

- `docs/api/auth.postman_collection.json`
- `docs/api/user.postman_collection.json`
- `docs/api/vehicle.postman_collection.json`
- `docs/api/dashboard.postman_collection.json`

Postman workspace YAML collection (cloud-synced; keep in sync with JSON collections above):

- `postman/collections/DMS API/`

Postman environments:

- `postman/environments/local.environment.yaml` (`base_url: http://localhost:8080`)
- `postman/environments/staging.environment.yaml` (`base_url: https://stag-api.infiniour.com`)

## Knowledge Base

- `docs/knowledge-base.md`
