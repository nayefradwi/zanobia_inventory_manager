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
	CreateRole(ctx context.Context, role Role) error
	GetRoles(ctx context.Context) ([]Role, error)
}

type RoleRepository struct {
	*pgxpool.Pool
}

func NewRoleRepository(dbPool *pgxpool.Pool) IRoleRepository {
	return &RoleRepository{Pool: dbPool}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role Role) error {
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
	tx.Commit(ctx)
	return nil
}

func (r *RoleRepository) _createRole(ctx context.Context, tx pgx.Tx, role Role) (int, error) {
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

func (r *RoleRepository) GetRoles(ctx context.Context) ([]Role, error) {
	sql := "SELECT roles.id, roles.name, roles.description, role_permissions.permission_handle FROM roles LEFT JOIN role_permissions ON roles.id = role_permissions.role_id"
	rows, err := r.Query(ctx, sql)
	if err != nil {
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	rolesMap := make(map[int]Role, 0)
	for rows.Next() {
		var role Role
		var permissionHandle string
		err := rows.Scan(&role.Id, &role.Name, &role.Description, &permissionHandle)
		if err != nil {
			return nil, common.NewInternalServerError()
		}
		if _, ok := rolesMap[*role.Id]; !ok {
			role.PermissionHandles = []string{permissionHandle}
			rolesMap[*role.Id] = role
			continue
		}
		role = rolesMap[*role.Id]
		role.PermissionHandles = append(role.PermissionHandles, permissionHandle)
		rolesMap[*role.Id] = role
	}
	roles := make([]Role, 0)
	for _, role := range rolesMap {
		roles = append(roles, role)
	}
	return roles, nil
}
