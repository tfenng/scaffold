# Backend Architecture Guide

## Overview

This is a production-ready Go/Gin + PostgreSQL scaffold with:
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

---

## Layer Structure

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

---

## Key Patterns

### 1. Repository Layer

Split into two parts:
- **CRUD Repo** (`UserRepo`): Stable operations - GetByID, Create, Update, Delete
- **Query Repo** (`UserQueryRepo`): List + Count with dynamic filtering

```go
// Example: UserRepo interface
type UserRepo interface {
    GetByID(ctx context.Context, id int64) (sqlc.User, error)
    GetByEmail(ctx context.Context, email string) (sqlc.User, error)
    Create(ctx context.Context, uid, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error)
    Update(ctx context.Context, id int64, name string, usedName, company *string, birth *time.Time) (sqlc.User, error)
    Delete(ctx context.Context, id int64) error
}
```

### 2. Service Layer

- Transaction boundary (service controls transaction)
- DB error mapping:
  - `pgx.ErrNoRows` → 404 NOT_FOUND
  - SQLSTATE `23505` (unique violation) → 409 CONFLICT
- Cache-Aside pattern for GetByID:
  - Read: check cache → miss → DB → set cache
  - Write: after successful write, set or delete cache

```go
// Example: Service error handling
if errors.Is(err, pgx.ErrNoRows) {
    return sqlc.User{}, domain.NotFound("user not found")
}
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation {
    return domain.Conflict("uid or name already exists")
}
```

### 3. HTTP Layer

- Handlers set errors via `c.Error(err)` - do not return JSON directly
- Middleware catches errors and converts to JSON response
- Use binding tags for request validation

```go
// Example: Handler pattern
func (h *UserHandler) Get(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    u, err := h.Svc.GetByID(c.Request.Context(), id)
    if err != nil { c.Error(err); return }
    c.JSON(http.StatusOK, u)
}
```

### 4. Domain Layer

Business error codes:
- `INVALID_ARGUMENT` → HTTP 400
- `NOT_FOUND` → HTTP 404
- `CONFLICT` → HTTP 409
- `INTERNAL` → HTTP 500

```go
// Example: Error creation
func Invalid(msg string) *AppError {
    return &AppError{Code: CodeInvalidArgument, Message: msg, HTTPStatus: http.StatusBadRequest}
}
```

### 5. Dynamic Filtering

Use `sqlc.narg()` for optional query parameters in SQL:

```sql
-- name: ListUsers :many
SELECT id, uid, email, name, created_at, updated_at
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE ('%' || sqlc.narg('name_like')::text || '%'))
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;
```

### 6. Pagination

Always implement List + Count together:
- Use offset-based pagination (NOT cursor-based)
- Order by fixed fields: `created_at DESC, id DESC`
- Never accept order field from user input

---

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

---

## Adding a New Entity

1. **Create migration**: `make migrate-new name=create_xxx`
2. **Write SQL queries**: `sql/xxx.sql`
3. **Generate code**: `make sqlc`
4. **Implement Repo**: `internal/repo/xxx_repo.go` and `internal/repo/xxx_query_repo.go`
5. **Implement Service**: `internal/service/xxx_service.go`
6. **Implement Handler**: `internal/api/http/handler_xxx.go`
7. **Register routes**: `cmd/api/main.go`

---

## Caching Strategy

- Cache by ID only (Cache-Aside pattern)
- Lists are NOT cached by default (too many dimensions, hard invalidation)
- If needed, cache only first page with short TTL

```go
// Example: Cache-Aside in service
func (s *UserService) GetByID(ctx context.Context, id int64) (sqlc.User, error) {
    // Check cache first
    if s.UCache != nil {
        if u, ok, err := s.UCache.Get(ctx, id); err == nil && ok {
            return u, nil
        }
    }

    // Cache miss - fetch from DB
    u, err := s.Users.GetByID(ctx, id)
    if err != nil { return sqlc.User{}, err }

    // Set cache
    if s.UCache != nil {
        _ = s.UCache.Set(ctx, u)
    }
    return u, nil
}
```

---

## Error Handling Summary

| Error Type | HTTP Status | Code |
|------------|-------------|------|
| Invalid input | 400 | INVALID_ARGUMENT |
| Not found | 404 | NOT_FOUND |
| Unique constraint violation | 409 | CONFLICT |
| Internal error | 500 | INTERNAL |

---

## Database Conventions

- Use `BIGSERIAL` for primary keys
- Use `TIMESTAMPTZ` for timestamps
- Always include `created_at` and `updated_at` columns
- Use index on `(created_at DESC, id DESC)` for pagination queries
