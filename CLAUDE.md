# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go/Gin + PostgreSQL scaffold template** providing a production-ready project structure with:
- **sqlc** for type-safe SQL code generation
- **pgx/v5** for PostgreSQL connectivity
- **Redis** for caching (Cache-Aside pattern)
- **Gin** for HTTP framework
- **golang-migrate** for database migrations

## Key Constraints

- **No soft delete** - use hard deletes only
- **No multi-tenancy** - single tenant architecture
- **Offset pagination** - not cursor-based
- **Cache-Aside** - for single entity by ID lookups only

## Commands

```bash
# Generate sqlc code from SQL queries
make sqlc

# Run database migrations
make migrate-up
make migrate-down

# Create a new migration
make migrate-new name=add_xxx
```

## Architecture

### Layer Structure

```
/cmd/api/main.go           # Entry point, dependency assembly
/internal
  /api/http               # Gin handlers and middleware
  /domain                 # Business errors (AppError with codes)
  /repo                   # Repository interfaces + implementations
  /service                # Business logic, error mapping, caching
  /db                     # PostgreSQL pool initialization
  /cache                  # Redis client and cache implementations
/internal/gen/sqlc        # Generated code (do not edit)
/sql                      # SQL queries for sqlc
/migrations              # Database migrations
```

### Key Patterns

1. **Repository**: Split into CRUD (`UserRepo`) and Query (`UserQueryRepo`) - query repo handles List+Count with dynamic filtering
2. **Service**: Transaction boundary, DB error mapping (`pgx.ErrNoRows` → 404, SQLSTATE `23505` → 409), Cache-Aside for GetByID
3. **HTTP**: Handlers set errors via `c.Error()`, middleware converts to JSON response
4. **Dynamic Filtering**: Use `sqlc.narg()` for optional query parameters
5. **Pagination**: Offset-based with `Limit` and `Offset`, always run List + Count together

### Error Handling

- Service returns `domain.AppError` with codes: `INVALID_ARGUMENT`, `NOT_FOUND`, `CONFLICT`, `INTERNAL`
- HTTP middleware catches errors and returns JSON with HTTP status code mapping
- DB errors mapped: `pgx.ErrNoRows` → 404, unique violation (`23505`) → 409

### Caching Strategy

- Cache by ID only (Cache-Aside pattern)
- Get: check cache → miss → DB → set cache
- Write: after successful write, set or delete cache
- Lists not cached by default (too many dimensions, hard invalidation)

### SQL/Query Patterns

- Use `sqlc.narg()` for optional filter conditions
- List queries always include corresponding Count query
- Use parameterized queries for dynamic filtering
- Order by fixed fields (e.g., `created_at DESC, id DESC`) - never accept order field from user input
