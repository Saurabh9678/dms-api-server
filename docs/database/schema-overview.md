# Schema Overview

## Purpose

- Summarize major tables, ownership by module, and key relationships.

## Module Ownership

- `auth` related tables
- `user` related tables
- `vehicle` related tables
- `showroom` related tables
- `customer` related tables

## Auth Schema Notes

- `user_otps.request_id` is an 8-character unique, non-null identifier used by OTP verification APIs.

## Update Checklist

- Update this file whenever schema structure or ownership changes.
