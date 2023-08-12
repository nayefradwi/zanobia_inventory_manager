package user

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
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
	tx, err := r.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return common.NewInternalServerError()
	}
	defer tx.Rollback(ctx)
	id, creatErr := r._createUser(ctx, tx, user)
	if creatErr != nil {
		return creatErr
	}
	permissionError := r._addPermissionsToUser(ctx, tx, id, user.PermissionHandles)
	if permissionError != nil {
		return permissionError
	}
	tx.Commit(ctx)
	return nil
}

func (r *UserRepository) _createUser(ctx context.Context, tx pgx.Tx, user UserInput) (int, error) {
	sql := "INSERT INTO users (email, password, first_name, last_name, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var id int
	err := tx.QueryRow(ctx, sql, user.Email, user.Password, user.FirstName, user.LastName, true).Scan(&id)
	if err != nil {
		log.Printf("failed to create user: %s", err.Error())
		return 0, common.NewBadRequestError("Failed to create user", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *UserRepository) _addPermissionsToUser(ctx context.Context, tx pgx.Tx, userId int, handles []string) error {
	sql := "INSERT INTO user_permissions (user_id, permission_handle) VALUES ($1, $2)"
	for _, permission := range handles {
		_, err := tx.Exec(ctx, sql, userId, permission)
		if err != nil {
			log.Printf("failed to add permission to user: %s", err.Error())
			return common.NewBadRequestError("Failed to add permissions to user", zimutils.GetErrorCodeFromError(err))
		}
	}
	return nil
}
