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
