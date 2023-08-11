package user

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IUserRepository interface {
	Create(ctx context.Context, user UserInput) error
}

type UserRepository struct {
	*pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) IUserRepository {
	return &UserRepository{Pool: dbPool}
}

func (r *UserRepository) Create(ctx context.Context, user UserInput) error {
	sql := "INSERT INTO users (email, password, first_name, last_name, is_active) VALUES ($1, $2, $3, $4, $5)"
	c, err := r.Exec(ctx, sql, user.Email, user.Password, user.FirstName, user.LastName, true)
	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		return common.NewBadRequestError(err.Error(), zimutils.GetErrorCodeFromError(pgErr))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}
