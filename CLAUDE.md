# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`common` is a shared Go library module (`module common`, Go 1.27) used by the other sumni-finance internal services. It is not an application: there is no `main` package. Root package `common` holds cross-cutting primitives; subpackages are `db/`, `http/`, `log/`, and `testutils/`.

Key dependencies: `pgx/v5` (Postgres), `golang-migrate/migrate/v4` (migrations), `labstack/echo/v5` (HTTP), `cenkalti/backoff/v5` (retries), `ThreeDotsLabs/humanslog` (human-readable slog output), `testify` (tests).

## Commands

```bash
go build ./...                          # compile everything
go vet ./...
go test ./...                           # all tests
go test ./ -run TestName                # single test in the root package
go test ./db/... -run TestName -v       # single test in a subpackage
```

Tests under `db/` may need a running Postgres; check `testutils/` for how the test database is provisioned.

## Architecture

- **`error.go` – `common.Error`**: the single error type carrying an HTTP status (`HttpErrorCode`), a client-safe `PublicError`, a machine-readable `ErrorSlug`, an optional `InternalError` (never for clients) and `Details`. Build with the `New*Error(slug, format, args...)` constructors (NotFound, InvalidInput, Unauthorized, Forbidden, Expired, Conflict); attach context with `WithDetails` / `WithInternalError`, which return copies. The `http/` layer is expected to translate these into responses, so return `common.Error` from domain code rather than raw HTTP codes.
- **`enum.go` – `Enum[T Enumerable]`**: string-backed enum wrapper. `T` implements `Values() []string` and is the source of truth for valid values. Validation happens in `UnmarshalText` (used by JSON and `Scan`), and an invalid value yields a `common.Error` (`invalid-enum-value`, 400). An empty string is accepted as the zero value (`IsZero`). Concrete enums are defined as `struct{ common.Enum[T] }` and built with `MustEnum`.
- **`uuid.go` – `UUID`**: `[16]byte` type implementing text marshalling plus `driver.Valuer`/`sql.Scanner`, so it works directly with pgx/database/sql and JSON. `NewUUIDv7()` is the generator for new IDs.

`db/`, `http/`, `log/` and `testutils/` were not yet reviewed when this file was generated; extend this section with their conventions (migrations, transactions, middleware, logger setup) when working in them.
