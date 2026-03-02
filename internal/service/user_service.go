package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"

	"github.com/tfenng/scaffold/internal/cache"
	"github.com/tfenng/scaffold/internal/domain"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
	"github.com/tfenng/scaffold/internal/repo"
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

func (s *UserService) Create(ctx context.Context, uid string, email *string, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	uid = strings.TrimSpace(uid)
	name = strings.TrimSpace(name)
	if uid == "" || name == "" {
		return sqlc.User{}, domain.Invalid("uid and name are required")
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return sqlc.User{}, domain.Invalid("email must be a valid email address")
	}

	var out sqlc.User
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		u, err := s.Users.Create(ctx, uid, normalizedEmail, name, usedName, company, birth)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation {
				return domain.Conflict("uid, name or email already exists")
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

func (s *UserService) List(ctx context.Context, f repo.UserListFilter) (repo.Page[sqlc.User], error) {
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

func normalizeEmail(email *string) (*string, error) {
	if email == nil {
		return nil, nil
	}

	v := strings.TrimSpace(*email)
	if v == "" {
		return nil, nil
	}

	if _, err := mail.ParseAddress(v); err != nil {
		return nil, err
	}
	return &v, nil
}
