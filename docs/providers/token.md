# Token Provider

## Interface Ownership

- Source: `internal/providers/token/provider.go`
- Contract: issue/rotate token pairs and hash refresh tokens.

## Responsibility

- Abstract token issuance, rotation, and refresh token hashing.

## Security Notes

- Token semantics and expiry must remain consistent with auth service expectations.

## Update Checklist

- Update this file for provider contract, token semantics, or security-related changes.
