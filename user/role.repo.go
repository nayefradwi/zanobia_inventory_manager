package user

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IRoleRepository interface {
	CreateRole(ctx context.Context, role RoleInput) error
}

type RoleRepository struct {
	*pgxpool.Pool
}

func NewRoleRepository(dbPool *pgxpool.Pool) IRoleRepository {
	return &RoleRepository{Pool: dbPool}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role RoleInput) error {
	tx, err := r.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return common.NewInternalServerError()
	}
	defer tx.Rollback(ctx)
	id, creatErr := r._createRole(ctx, tx, role)
	if creatErr != nil {
		return creatErr
	}
	addErr := r._addPermissionsToRole(ctx, tx, id, role.PermissionHandles)
	if addErr != nil {
		return addErr
	}
	return nil
}

func (r *RoleRepository) _createRole(ctx context.Context, tx pgx.Tx, role RoleInput) (int, error) {
	sql := "INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id"
	var id int
	err := tx.QueryRow(ctx, sql, role.Name, role.Description).Scan(&id)
	if err != nil {
		return 0, common.NewBadRequestError("Failed to create role", zimutils.GetErrorCodeFromError(err))
	}
	return id, nil
}

func (r *RoleRepository) _addPermissionsToRole(ctx context.Context, tx pgx.Tx, roleId int, handles []string) error {
	sql := "INSERT INTO role_permissions (role_id, permission_handle) VALUES ($1, $2)"
	for _, permission := range handles {
		_, err := tx.Exec(ctx, sql, roleId, permission)
		if err != nil {
			log.Printf("failed to add permission to role: %s", err.Error())
			return common.NewBadRequestError("Failed to add permissions to user", zimutils.GetErrorCodeFromError(err))
		}
	}
	return nil
}
