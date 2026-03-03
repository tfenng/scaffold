# Refactor Backlog (P1-P3)

## P1

### 1. Frontend/Backend API contract drift
- Problem:
  - API docs use snake_case and lowercase field examples (`id`, `name`), while actual Go JSON output is PascalCase (`ID`, `Name`, `Uid`).
  - Create API docs do not include required `uid`.
- Risk:
  - Client integration errors and unclear external contract.
- Suggested refactor:
  - Add explicit JSON tags on response DTOs (or introduce API response DTO layer).
  - Align `web/api.md`, frontend types, and backend response shape to one canonical contract.

### 2. Email validation vs DB constraints mismatch
- Problem:
  - DB requires `email NOT NULL UNIQUE`.
  - Frontend schema currently allows optional/empty email.
- Risk:
  - User-facing 409/500-like failures during create due to avoidable invalid payloads.
- Suggested refactor:
  - Make frontend create schema require non-empty valid email.
  - Keep backend-side explicit validation for email format and required semantics.

### 3. Redis degradation strategy incomplete
- Problem:
  - App logs Redis unavailable but still wires `UserCache`, so cache calls continue and may add latency.
- Risk:
  - Increased request tail latency when Redis is down.
- Suggested refactor:
  - If ping fails, set `userCache = nil` and run in no-cache mode.
  - Add metrics/logging for cache hit/miss/error and circuit-breaker style fallback if needed.

## P2

### 4. Update conflict message inaccurate
- Problem:
  - Update path maps unique violation to `"email already exists"` even though update does not change email and DB also enforces unique `name`.
- Risk:
  - Misleading API diagnostics and harder troubleshooting.
- Suggested refactor:
  - Return neutral message like `"unique constraint violation"` or parse violated constraint name for precise messaging.

### 5. 204 response should not include body
- Problem:
  - Delete handler returns `c.JSON(http.StatusNoContent, nil)`.
- Risk:
  - Non-standard behavior for strict HTTP clients/proxies.
- Suggested refactor:
  - Use `c.Status(http.StatusNoContent)`.

## P3

### 6. Migration history has overlapping schema changes
- Problem:
  - `000001_init.up.sql` already contains `used_name/company/birth`, while `000002` adds same columns with `IF NOT EXISTS`.
- Risk:
  - Confusing migration intent and onboarding cost.
- Decision:
  - Do not modify historical migrations now.
  - Document migration policy and overlap rationale in README.
  - Plan a baseline migration set for new environments in a future milestone.
- Why:
  - Historical migration edits can create environment drift and rollback risk.
  - Current overlap is noisy but functionally safe due to `IF NOT EXISTS`.
- Follow-up task:
  - Introduce a baseline migration chain for fresh installs (e.g. `migrations_baseline/` or `000001_baseline`), while keeping existing chain immutable for already-deployed environments.

### 7. Frontend type hygiene
- Problem:
  - `users/page.tsx` uses `any` for list items and imports unused `userListQuerySchema`.
- Risk:
  - Type safety erosion and hidden runtime shape bugs.
- Suggested refactor:
  - Type API responses via `Page<User>`.
  - Remove unused imports and enforce lint rule for `no-explicit-any`.
