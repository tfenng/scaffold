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
