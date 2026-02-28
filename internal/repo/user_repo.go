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
	GetByUID(ctx context.Context, uid string) (sqlc.User, error)
	Create(ctx context.Context, uid, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error)
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
	row, err := r.q(ctx).GetUserByID(ctx, id)
	if err != nil {
		return sqlc.User{}, err
	}
	return toUser(row)
}
func (r *userRepo) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	row, err := r.q(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc.User{}, err
	}
	return toUserFromEmail(row)
}
func (r *userRepo) GetByUID(ctx context.Context, uid string) (sqlc.User, error) {
	row, err := r.q(ctx).GetUserByUID(ctx, uid)
	if err != nil {
		return sqlc.User{}, err
	}
	return toUserFromUID(row)
}
func (r *userRepo) Create(ctx context.Context, uid, email, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	row, err := r.q(ctx).CreateUser(ctx, sqlc.CreateUserParams{
		Uid:      uid,
		Email:    email,
		Name:     name,
		UsedName: toPgtypeText(usedName),
		Company:  toPgtypeText(company),
		Birth:    toPgtypeDate(birth),
	})
	if err != nil {
		return sqlc.User{}, err
	}
	return toUserFromCreate(row)
}

func (r *userRepo) Update(ctx context.Context, id int64, name string, usedName, company *string, birth *time.Time) (sqlc.User, error) {
	row, err := r.q(ctx).UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:       id,
		Name:     name,
		UsedName: toPgtypeText(usedName),
		Company:  toPgtypeText(company),
		Birth:    toPgtypeDate(birth),
	})
	if err != nil {
		return sqlc.User{}, err
	}
	return toUserFromUpdate(row)
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

func toUser(row sqlc.GetUserByIDRow) (sqlc.User, error) {
	return sqlc.User{
		ID:        row.ID,
		Uid:       row.Uid,
		Email:     row.Email,
		Name:      row.Name,
		UsedName:  row.UsedName,
		Company:   row.Company,
		Birth:     row.Birth,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func toUserFromEmail(row sqlc.GetUserByEmailRow) (sqlc.User, error) {
	return sqlc.User{
		ID:        row.ID,
		Uid:       row.Uid,
		Email:     row.Email,
		Name:      row.Name,
		UsedName:  row.UsedName,
		Company:   row.Company,
		Birth:     row.Birth,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func toUserFromUID(row sqlc.GetUserByUIDRow) (sqlc.User, error) {
	return sqlc.User{
		ID:        row.ID,
		Uid:       row.Uid,
		Email:     row.Email,
		Name:      row.Name,
		UsedName:  row.UsedName,
		Company:   row.Company,
		Birth:     row.Birth,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func toUserFromCreate(row sqlc.CreateUserRow) (sqlc.User, error) {
	return sqlc.User{
		ID:        row.ID,
		Uid:       row.Uid,
		Email:     row.Email,
		Name:      row.Name,
		UsedName:  row.UsedName,
		Company:   row.Company,
		Birth:     row.Birth,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func toUserFromUpdate(row sqlc.UpdateUserRow) (sqlc.User, error) {
	return sqlc.User{
		ID:        row.ID,
		Uid:       row.Uid,
		Email:     row.Email,
		Name:      row.Name,
		UsedName:  row.UsedName,
		Company:   row.Company,
		Birth:     row.Birth,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
