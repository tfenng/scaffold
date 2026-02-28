package service

import (
	"context"
	"errors"
	"time"

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

func (s *UserService) Create(ctx context.Context, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	if email == "" || name == "" {
		return sqlc.User{}, domain.Invalid("email and name are required")
	}

	var out sqlc.User
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.Users.Create(ctx, email, name, usedName, company, birth)
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
	// 列表默认不做缓存（参数维度多、失效难），你后续要做"第一页短 TTL"也很好加
	out, err := s.Query.List(ctx, f)
	if err != nil {
		return repo.Page[sqlc.User]{}, domain.Internal(err)
	}
	return out, nil
}

func (s *UserService) Update(ctx context.Context, id int64, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	if id <= 0 {
		return sqlc.User{}, domain.Invalid("id must be positive")
	}
	if name == "" {
		return sqlc.User{}, domain.Invalid("name is required")
	}

	var out sqlc.User
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.Users.Update(ctx, id, name, usedName, company, birth)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.NotFound("user not found")
			}
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

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.Invalid("id must be positive")
	}

	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		err := s.Users.Delete(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.NotFound("user not found")
			}
			return domain.Internal(err)
		}
		return nil
	})
	if err != nil {
		var ae *domain.AppError
		if errors.As(err, &ae) {
			return ae
		}
		return domain.Internal(err)
	}

	if s.UCache != nil {
		_ = s.UCache.Del(ctx, id)
	}
	return nil
}
