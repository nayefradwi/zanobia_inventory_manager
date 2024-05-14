package user

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
	"go.uber.org/zap"
)

const (
	SysAdminPermissionHandle    = "sys_admin"
	HasUserControlPermission    = "has_user_control"
	HasProductControlPermission = "has_product_control"
	HasBatchControlPermission   = "has_batch_control"
	CanDeleteProductPermission  = "can_delete_product"
	CanDeleteBatchPermission    = "can_delete_batch"
)

type IPermissionRepository interface {
	InitiateAll(ctx context.Context, permissions []Permission) error
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
		common.LoggerFromCtx(ctx).Error("failed to find permission by handle", zap.Error(err))
		return Permission{}, common.NewNotFoundError("permission not found")
	}
	return permission, nil
}

func (r *PermissionRepository) InitiateAll(ctx context.Context, permissions []Permission) error {
	sql := "INSERT INTO permissions (handle, name, description, is_secret) VALUES ($1, $2, $3, $4)"
	for _, p := range permissions {
		r.Exec(ctx, sql, p.Handle, p.Name, p.Description, p.IsSecret)
	}
	return nil
}

func (r *PermissionRepository) CreatePermssion(ctx context.Context, permission Permission) error {
	sql := "INSERT INTO permissions (handle, name, description, is_secret) VALUES ($1, $2, $3, $4)"
	c, err := r.Exec(ctx, sql, permission.Handle, permission.Name, permission.Description, permission.IsSecret)
	if err != nil {
		common.LoggerFromCtx(ctx).Error("failed to add permission", zap.Error(err))
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
		common.LoggerFromCtx(ctx).Error("failed to get all permissions", zap.Error(err))
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	var permissions []Permission
	for rows.Next() {
		var permission Permission
		err := rows.Scan(&permission.Id, &permission.Handle, &permission.Name, &permission.Description, &permission.IsSecret)
		if err != nil {
			common.LoggerFromCtx(ctx).Error("failed to get all permissions", zap.Error(err))
			return nil, common.NewInternalServerError()
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}
