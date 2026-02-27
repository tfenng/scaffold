好，那我按你约束（**Gin + Postgres、无软删/无多租户、offset 分页、需要缓存**）给一套可以直接照搬的工程化模板。核心点：

* **Repo：**sqlc + narg 做动态过滤，offset 分页（List + Count）
* **Service：**统一错误码体系 + DB 错误映射（比如 unique 冲突 → 409）+ Cache-Aside（按 ID 缓存）
* **HTTP：**Gin 中间件统一把业务错误 → HTTP JSON
* **迁移：**给出 golang-migrate 和 Atlas 的整合方式（你二选一）

---

## 1) 目录结构（可直接用）

```text
/cmd/api/main.go

/internal
  /api/http
    middleware_error.go
    handler_user.go
  /domain
    errors.go
  /repo
    tx.go
    user_repo.go
    user_query_repo.go
  /service
    user_service.go
  /db
    postgres.go
  /cache
    redis.go
    user_cache.go

/internal/gen/sqlc         # sqlc 生成

/sql
  user.sql

/migrations
  000001_init.up.sql
  000001_init.down.sql

/sqlc.yaml
/Makefile
```

---

## 2) sqlc：SQL（动态过滤 + offset 分页）

> 这里用 `sqlc.narg()` 实现可选条件，几乎够用。复杂动态排序建议固定白名单字段做“分支 SQL”，不要把字段名当参数传进去。

`/sql/user.sql`

```sql
-- name: GetUserByID :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING id, email, name, created_at, updated_at;

-- name: ListUsers :many
SELECT id, email, name, created_at, updated_at
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE ('%' || sqlc.narg('name_like')::text || '%'))
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(1)
FROM users
WHERE (sqlc.narg('email')::text IS NULL OR email = sqlc.narg('email')::text)
  AND (sqlc.narg('name_like')::text IS NULL OR name ILIKE ('%' || sqlc.narg('name_like')::text || '%'));
```

---

## 3) sqlc.yaml（pgx 版本）

`/sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "migrations"
    queries: "sql"
    gen:
      go:
        package: "sqlc"
        out: "internal/gen/sqlc"
        sql_package: "pgx/v5"
```

---

## 4) DB：pgxpool 初始化

`/internal/db/postgres.go`

```go
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
```

---

## 5) Repo：事务 + CRUD + 动态查询/分页

### 5.1 事务管理（Service 控事务边界）

`/internal/repo/tx.go`

```go
package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type txKey struct{}

type PgxTxManager struct{ Pool *pgxpool.Pool }

func (m PgxTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func TxFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
```

### 5.2 UserRepo（稳定 CRUD）

`/internal/repo/user_repo.go`

```go
package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type UserRepo interface {
	GetByID(ctx context.Context, id int64) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	Create(ctx context.Context, email, name string) (sqlc.User, error)
}

type userRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) UserRepo { return &userRepo{pool: pool} }

func (r *userRepo) q(ctx context.Context) *sqlc.Queries {
	if tx, ok := TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (sqlc.User, error) {
	return r.q(ctx).GetUserByID(ctx, id)
}
func (r *userRepo) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q(ctx).GetUserByEmail(ctx, email)
}
func (r *userRepo) Create(ctx context.Context, email, name string) (sqlc.User, error) {
	return r.q(ctx).CreateUser(ctx, sqlc.CreateUserParams{Email: email, Name: name})
}
```

### 5.3 UserQueryRepo（offset 分页 + 动态过滤）

`/internal/repo/user_query_repo.go`

```go
package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type Page[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	TotalPages int32 `json:"total_pages"`
}

type UserListFilter struct {
	Email    *string
	NameLike *string
	Page     int32
	PageSize int32
}

type UserQueryRepo interface {
	List(ctx context.Context, f UserListFilter) (Page[sqlc.User], error)
}

type userQueryRepo struct{ pool *pgxpool.Pool }

func NewUserQueryRepo(pool *pgxpool.Pool) UserQueryRepo { return &userQueryRepo{pool: pool} }

func (r *userQueryRepo) q(ctx context.Context) *sqlc.Queries {
	if tx, ok := TxFrom(ctx); ok {
		return sqlc.New(tx)
	}
	return sqlc.New(r.pool)
}

func (r *userQueryRepo) List(ctx context.Context, f UserListFilter) (Page[sqlc.User], error) {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 200 {
		f.PageSize = 20
	}
	limit := f.PageSize
	offset := (f.Page - 1) * f.PageSize

	total, err := r.q(ctx).CountUsers(ctx, sqlc.CountUsersParams{
		Email:    f.Email,
		NameLike: f.NameLike,
	})
	if err != nil {
		return Page[sqlc.User]{}, err
	}

	items, err := r.q(ctx).ListUsers(ctx, sqlc.ListUsersParams{
		Limit:    limit,
		Offset:   offset,
		Email:    f.Email,
		NameLike: f.NameLike,
	})
	if err != nil {
		return Page[sqlc.User]{}, err
	}

	totalPages := int32((total + int64(limit) - 1) / int64(limit))
	return Page[sqlc.User]{Items: items, Total: total, Page: f.Page, PageSize: limit, TotalPages: totalPages}, nil
}
```

---

## 6) Domain：错误码体系（业务错误 → HTTP）

`/internal/domain/errors.go`

```go
package domain

import "net/http"

type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeInternal        Code = "INTERNAL"
)

type AppError struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Cause      error  `json:"-"`
}

func (e *AppError) Error() string { return string(e.Code) + ": " + e.Message }

func Invalid(msg string) *AppError  { return &AppError{Code: CodeInvalidArgument, Message: msg, HTTPStatus: http.StatusBadRequest} }
func NotFound(msg string) *AppError { return &AppError{Code: CodeNotFound, Message: msg, HTTPStatus: http.StatusNotFound} }
func Conflict(msg string) *AppError { return &AppError{Code: CodeConflict, Message: msg, HTTPStatus: http.StatusConflict} }
func Internal(err error) *AppError  { return &AppError{Code: CodeInternal, Message: "internal error", HTTPStatus: http.StatusInternalServerError, Cause: err} }
```

---

## 7) Cache：Redis + UserCache（按 ID 缓存）

`/internal/cache/redis.go`

```go
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  600 * time.Millisecond,
		WriteTimeout: 600 * time.Millisecond,
	})
}

func Ping(ctx context.Context, rdb *redis.Client) error { return rdb.Ping(ctx).Err() }
```

`/internal/cache/user_cache.go`

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type UserCache struct {
	Rdb *redis.Client
	TTL time.Duration
}

func NewUserCache(rdb *redis.Client) *UserCache {
	return &UserCache{Rdb: rdb, TTL: 5 * time.Minute}
}

func (c *UserCache) key(id int64) string { return fmt.Sprintf("user:v1:id:%d", id) }

func (c *UserCache) Get(ctx context.Context, id int64) (sqlc.User, bool, error) {
	val, err := c.Rdb.Get(ctx, c.key(id)).Result()
	if err == redis.Nil {
		return sqlc.User{}, false, nil
	}
	if err != nil {
		return sqlc.User{}, false, err
	}
	var u sqlc.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return sqlc.User{}, false, err
	}
	return u, true, nil
}

func (c *UserCache) Set(ctx context.Context, u sqlc.User) error {
	b, _ := json.Marshal(u)
	return c.Rdb.Set(ctx, c.key(u.ID), b, c.TTL).Err()
}

func (c *UserCache) Del(ctx context.Context, id int64) error {
	return c.Rdb.Del(ctx, c.key(id)).Err()
}
```

---

## 8) Service：错误映射 + 缓存策略（Cache-Aside）

关键点：

* `pgx.ErrNoRows` → 404
* Postgres 唯一约束冲突：SQLSTATE `23505` → 409（Conflict）
* 缓存：**读缓存 miss → DB → Set**；写成功后 **Set 或 Del**

`/internal/service/user_service.go`

```go
package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgconn"

	"github.com/tfenng/scaffold/internal/cache"
	"github.com/tfenng/scaffold/internal/domain"
	"github.com/tfenng/scaffold/internal/repo"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type UserService struct {
	Tx     repo.TxManager
	Users  repo.UserRepo
	Query  repo.UserQueryRepo
	UCache *cache.UserCache
}

// Postgres SQLSTATE
const (
	sqlStateUniqueViolation = "23505"
)

func (s *UserService) GetByID(ctx context.Context, id int64) (sqlc.User, error) {
	if id <= 0 {
		return sqlc.User{}, domain.Invalid("id must be positive")
	}

	if s.UCache != nil {
		if u, ok, err := s.UCache.Get(ctx, id); err == nil && ok {
			return u, nil
		}
		// 缓存错误：不阻断主流程（建议打日志/指标）
	}

	u, err := s.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, domain.NotFound("user not found")
		}
		return sqlc.User{}, domain.Internal(err)
	}

	if s.UCache != nil {
		_ = s.UCache.Set(ctx, u)
	}
	return u, nil
}

func (s *UserService) Create(ctx context.Context, email, name string) (sqlc.User, error) {
	if email == "" || name == "" {
		return sqlc.User{}, domain.Invalid("email and name are required")
	}

	var out sqlc.User
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.Users.Create(ctx, email, name)
		if err != nil {
			// 映射唯一约束冲突
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation {
				return domain.Conflict("email already exists")
			}
			return domain.Internal(err)
		}
		out = u
		return nil
	})
	if err != nil {
		// 事务内部我们直接 return domain.AppError（或 Internal）
		var ae *domain.AppError
		if errors.As(err, &ae) {
			return sqlc.User{}, ae
		}
		return sqlc.User{}, domain.Internal(err)
	}

	if s.UCache != nil {
		_ = s.UCache.Set(ctx, out)
	}
	return out, nil
}

func (s *UserService) List(ctx context.Context, f repo.UserListFilter) (repo.Page[sqlc.User], error) {
	// 列表默认不做缓存（参数维度多、失效难），你后续要做“第一页短 TTL”也很好加
	out, err := s.Query.List(ctx, f)
	if err != nil {
		return repo.Page[sqlc.User]{}, domain.Internal(err)
	}
	return out, nil
}
```

---

## 9) HTTP：Gin 错误中间件（统一 JSON）

Handler 只 `c.Error(err)`，中间件统一输出。

`/internal/api/http/middleware_error.go`

```go
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tfenng/scaffold/internal/domain"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err

		var ae *domain.AppError
		if errors.As(err, &ae) {
			c.JSON(ae.HTTPStatus, gin.H{"code": ae.Code, "message": ae.Message})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"code": domain.CodeInternal, "message": "internal error"})
	}
}
```

---

## 10) Handler：分页/动态过滤（Gin bind）

`/internal/api/http/handler_user.go`

```go
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

type UserHandler struct{ Svc *service.UserService }

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	u, err := h.Svc.GetByID(c.Request.Context(), id)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusOK, u)
}

type createUserReq struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(/*domain.Invalid*/ err) // 也可以把 bind 错误映射成 domain.Invalid
		return
	}
	u, err := h.Svc.Create(c.Request.Context(), req.Email, req.Name)
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusCreated, u)
}

type listUsersQuery struct {
	Email    *string `form:"email"`
	NameLike *string `form:"name_like"`
	Page     int32   `form:"page"`
	PageSize int32   `form:"page_size"`
}

func (h *UserHandler) List(c *gin.Context) {
	var q listUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil { c.Error(err); return }

	out, err := h.Svc.List(c.Request.Context(), repo.UserListFilter{
		Email: q.Email, NameLike: q.NameLike, Page: q.Page, PageSize: q.PageSize,
	})
	if err != nil { c.Error(err); return }
	c.JSON(http.StatusOK, out)
}
```

> 小建议：把 bind 错误统一映射成 `domain.Invalid("...")`，避免直接把 gin 的错误结构暴露出去。

---

## 11) main.go：依赖组装

`/cmd/api/main.go`

```go
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/tfenng/scaffold/internal/api/http"
	"github.com/tfenng/scaffold/internal/cache"
	"github.com/tfenng/scaffold/internal/db"
	"github.com/tfenng/scaffold/internal/repo"
	"github.com/tfenng/scaffold/internal/service"
)

func main() {
	ctx := context.Background()

	dsn := "postgres://user:pass@localhost:5432/app?sslmode=disable"
	pool, err := db.NewPostgres(ctx, dsn)
	if err != nil { log.Fatal(err) }
	defer pool.Close()

	rdb := cache.NewRedis("127.0.0.1:6379")
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Println("redis unavailable, continue without cache:", err)
		// 你也可以在这里决定直接退出
	}

	txMgr := repo.PgxTxManager{Pool: pool}
	userRepo := repo.NewUserRepo(pool)
	userQueryRepo := repo.NewUserQueryRepo(pool)
	userCache := cache.NewUserCache(rdb)

	userSvc := &service.UserService{
		Tx: txMgr, Users: userRepo, Query: userQueryRepo, UCache: userCache,
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(http.ErrorMiddleware())

	h := &http.UserHandler{Svc: userSvc}
	r.GET("/users/:id", h.Get)
	r.POST("/users", h.Create)
	r.GET("/users", h.List)

	log.Fatal(r.Run(":8080"))
}
```

---

## 12) 迁移：golang-migrate（推荐先用这个）

`/migrations/000001_init.up.sql`

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_created_at_id ON users (created_at DESC, id DESC);
```

`/migrations/000001_init.down.sql`

```sql
DROP TABLE IF EXISTS users;
```

`/Makefile`

```makefile
DB_URL=postgres://user:pass@localhost:5432/app?sslmode=disable

sqlc:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-new:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: sqlc migrate-up migrate-down migrate-new
```

> 推荐工作流：
>
> 1. `make migrate-new name=add_xxx`
> 2. 写 up/down
> 3. `make migrate-up`
> 4. `make sqlc`（保证生成代码跟 schema/queries 匹配）

---

## 13) Atlas（可选替代 migrate）

如果你更喜欢 schema 驱动 + 自动 diff，可以用 Atlas；最终仍然输出 migration 文件并 apply。

典型命令（示意，放到 Makefile 里）：

```makefile
atlas-diff:
	atlas migrate diff $(name) \
	  --dir "file://migrations" \
	  --to "file://schema.hcl" \
	  --dev-url "docker://postgres/16/dev?search_path=public"

atlas-apply:
	atlas migrate apply --dir "file://migrations" --url "$(DB_URL)"
```

---

# 你直接拿去用的关键点总结

1. **分页/动态查询（Repo）**：`List + Count`，sqlc `narg` 管可选条件；分页参数统一在 QueryRepo 里校正。
2. **错误码体系（Service）**：`domain.AppError{Code, Message, HTTPStatus}`；DB 错误映射（`pgx.ErrNoRows` → 404，`23505` → 409）。
3. **迁移整合**：推荐先 migrate（简单稳），Atlas 作为进阶选项。
4. **缓存策略**：按 ID 缓存（Cache-Aside），列表默认不缓存（要缓存就只缓存第一页 + 短 TTL）。

