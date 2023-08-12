package user

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

const (
	sysAdminPermissionHandle = "sys_admin"
)

type IPermissionRepository interface {
	AddAll(ctx context.Context, permissions []Permission) error
	FindByHandle(ctx context.Context, handle string) (Permission, error)
	DoesPermissionExist(ctx context.Context, handle string) bool
	CreatePermssion(ctx context.Context, permission Permission) error
	GetAllPermissions(ctx context.Context) ([]Permission, error)
}

type PermissionRepository struct {
	*pgxpool.Pool
}

func NewPermissionRepository(dbPool *pgxpool.Pool) IPermissionRepository {
	return &PermissionRepository{Pool: dbPool}
}

func (r *PermissionRepository) DoesPermissionExist(ctx context.Context, handle string) bool {
	permission, findErr := r.FindByHandle(ctx, handle)
	return permission.Id != nil && findErr == nil
}

func (r *PermissionRepository) FindByHandle(ctx context.Context, handle string) (Permission, error) {
	sql := "SELECT id, handle, name, description, is_secret FROM permissions WHERE handle = $1"
	row := r.QueryRow(ctx, sql, handle)
	var permission Permission
	err := row.Scan(&permission.Id, &permission.Handle, &permission.Name, &permission.Description, &permission.IsSecret)
	if err != nil {
		log.Printf("failed to find permission by handle: %s", err.Error())
		return Permission{}, common.NewNotFoundError("permission not found")
	}
	return permission, nil
}

func (r *PermissionRepository) AddAll(ctx context.Context, permissions []Permission) error {
	tx, err := r.Begin(ctx)
	if err != nil {
		log.Printf("failed to start transaction: %s", err.Error())
		return common.NewInternalServerError()
	}
	defer tx.Rollback(ctx)
	sql := "INSERT INTO permissions (handle, name, description, is_secret) VALUES ($1, $2, $3, $4)"
	for _, p := range permissions {
		_, err := tx.Exec(ctx, sql, p.Handle, p.Name, p.Description, p.IsSecret)
		if err != nil {
			log.Printf("failed to add permission: %s", err.Error())
			return common.NewBadRequestError("failed to add permissions", zimutils.GetErrorCodeFromError(err))
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return common.NewInternalServerError()
	}
	return nil
}

func (r *PermissionRepository) CreatePermssion(ctx context.Context, permission Permission) error {
	sql := "INSERT INTO permissions (handle, name, description, is_secret) VALUES ($1, $2, $3, $4)"
	c, err := r.Exec(ctx, sql, permission.Handle, permission.Name, permission.Description, permission.IsSecret)
	if err != nil {
		log.Printf("failed to add permission: %s", err.Error())
		return common.NewBadRequestError("failed to add permissions", zimutils.GetErrorCodeFromError(err))
	}
	if c.RowsAffected() == 0 {
		return common.NewInternalServerError()
	}
	return nil
}

func (r *PermissionRepository) GetAllPermissions(ctx context.Context) ([]Permission, error) {
	sql := "SELECT id, handle, name, description, is_secret FROM permissions where is_secret = false"
	rows, err := r.Query(ctx, sql)
	if err != nil {
		log.Printf("failed to get all permissions: %s", err.Error())
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	var permissions []Permission
	for rows.Next() {
		var permission Permission
		err := rows.Scan(&permission.Id, &permission.Handle, &permission.Name, &permission.Description, &permission.IsSecret)
		if err != nil {
			log.Printf("failed to get all permissions: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
