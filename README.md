# fastly-purge

A CLI tool for creating Fastly API tokens with purge scopes.

## What It Does

Creates an API token for your Fastly account with configurable scopes (read-only, purge-by-URL, purge-all). Useful for CI/CD pipelines or automation that needs limited-permission tokens.

## Building

```bash
go build
```

## Usage

Create a token (prompts for password if not provided):

```bash
./fastly-purge create-token --username you@example.com
```

With environment variables:

```bash
export FASTLY_USERNAME=you@example.com
export FASTLY_PASSWORD=yourpassword
./fastly-purge create-token
```

If your account has 2FA enabled, provide the OTP:

```bash
./fastly-purge create-token --username you@example.com --otp 123456
```

### Flags

- `--username` - Fastly account email (or set `FASTLY_USERNAME`)
- `--password` - Account password (or set `FASTLY_PASSWORD`); prompted securely if omitted
- `--scope` - Token scope, space-delimited. Default: `global:read purge_all purge_select`
  - Options: `global`, `global:read`, `purge_all`, `purge_select`
- `--service` - Restrict token to specific service IDs (repeatable: `--service id1 --service id2`)
- `--expires-at` - Token expiration in ISO 8601 format (e.g., `2025-12-31T23:59:59+00:00`)
- `--otp` - One-time password for 2FA

## Example

Create a read-only token that expires in 30 days:

```bash
./fastly-purge create-token \
  --username you@example.com \
  --scope global:read \
  --expires-at 2026-09-18T23:59:59+00:00
```

Output includes the token ID and access token (only shown once).

## Testing

```bash
go test ./cmd -v
```

## Requirements

- Go 1.26+
- Fastly account with API access
