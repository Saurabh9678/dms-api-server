# Storage Provider

## Interface Ownership

- Source: `internal/providers/storage/provider.go`
- Contract:
  - `Upload` stores bytes and returns the **object key** to persist in the database.
  - `SignedURL` returns a time-limited client access URL for a stored key.

## Responsibility

- Abstract object/file storage from business modules.
- Keep durable references infra-agnostic (object keys only in DB).
- Issue short-lived access URLs at read/response time.

## Implementations

- `internal/infra/storage.LocalProvider` — writes under `STORAGE_BASE_PATH`; `SignedURL` returns the key unchanged (local/dev).
- `internal/infra/storage.GCSProvider` — uploads to GCS; `SignedURL` returns a V4 signed GET URL.

## Selection

- Factory: `internal/infra/storage.NewProvider`
- `STORAGE_PROVIDER=local` (default) or `gcs`
- `GCS_BUCKET_NAME` required when provider is `gcs` (e.g. `dms-dev-assets`)
- `STORAGE_SIGNED_URL_TTL_SECONDS` (default `3600`)

## Auth notes

- GCS uses Application Default Credentials. Local ADC may impersonate a service account.
- Staging/production should use a service account identity (JSON key mount or equivalent), not a developer user login.

## Object key convention (showroom media)

- `{userID}/showroom/{externalShowroomID}/{YYYYMMDDHHmmss}{ext}`
- Example: `42/showroom/SHOP0001/20260812211000.jpg`

## Object key convention (vehicle photos)

- `{userID}/vehicle/{vehicleID}/{YYYYMMDDHHmmss}{ext}`
- Example: `42/vehicle/18/20260814215100.jpg`
- Multiple rows/photos per `label` are allowed.
- Vehicle photo files: `.jpg` / `.jpeg` / `.png`, max **15 MB**.

## Update Checklist

- Update this file when storage contract, providers, key layout, or signing behavior changes.
