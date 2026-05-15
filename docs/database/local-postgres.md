# Local Postgres (Docker)

## Purpose

- Run a local Postgres instance for development and migrations.

## Default Local Configuration

- Container: `dms-postgres`
- Image: `postgres:16-alpine`
- Port: `5432`
- Database: `dms`
- User: `postgres`
- Password: `postgres`
- Connection URL: `postgres://postgres:postgres@localhost:5432/dms?sslmode=disable`

## Commands

- Start: `make docker-postgres-up`
- Stop: `make docker-postgres-down`
- Logs: `make docker-postgres-logs`

## Environment

- Copy `.env.example` to `.env` if `.env` does not exist.
- Ensure `DB_URL` matches the local container URL.
