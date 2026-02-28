package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type UserRepo interface {
	GetByID(ctx context.Context, id int64) (sqlc.User, error)
	GetByEmail(ctx context.Context, email string) (sqlc.User, error)
	Create(ctx context.Context, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error)
	Update(ctx context.Context, id int64, name string, usedName, company *string, birth *time.Time) (sqlc.User, error)
	Delete(ctx context.Context, id int64) error
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
func (r *userRepo) Create(ctx context.Context, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	return r.q(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		Email:    email,
		Name:     name,
		UsedName: toPgtypeText(usedName),
		Company:  toPgtypeText(company),
		Birth:    toPgtypeDate(birth),
	})
}

func (r *userRepo) Update(ctx context.Context, id int64, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	return r.q(ctx).UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       id,
		Name:     name,
		UsedName: toPgtypeText(usedName),
		Company:  toPgtypeText(company),
		Birth:    toPgtypeDate(birth),
	})
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	return r.q(ctx).DeleteUser(ctx, id)
}

func toPgtypeDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}
