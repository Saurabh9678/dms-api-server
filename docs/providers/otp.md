# OTP Provider

## Interface Ownership

- Source: `internal/providers/otp/provider.go`
- Contract: `Send(ctx, phone, message) error`

## Responsibility

- Abstract OTP delivery from business modules.

## Implementations

- Infra implementations may include dummy and external gateway clients.

## Update Checklist

- Update this file when OTP provider contract or implementation behavior changes.
