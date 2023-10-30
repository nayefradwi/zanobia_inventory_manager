package user

import (
	"context"
	"log"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nayefradwi/zanobia_inventory_manager/common"
	"github.com/nayefradwi/zanobia_inventory_manager/warehouse"
	zimutils "github.com/nayefradwi/zanobia_inventory_manager/zim_utils"
)

type IUserRepository interface {
	Create(ctx context.Context, user UserInput) error
	GetAllUsers(ctx context.Context) ([]User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserById(ctx context.Context, id int) (User, error)
	BanUser(ctx context.Context, id int) error
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

func (s *UserRepository) GetAllUsers(ctx context.Context) ([]User, error) {
	sql := "SELECT id, email, first_name, last_name, is_active FROM users"
	rows, err := s.Query(ctx, sql)
	if err != nil {
		log.Printf("failed to get users: %s", err.Error())
		return nil, common.NewInternalServerError()
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.IsActive)
		if err != nil {
			log.Printf("failed to scan user: %s", err.Error())
			return nil, common.NewInternalServerError()
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *UserRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	sql := `
	select users.id, users.email, users.first_name, users.last_name, users.password, 
	permissions.handle, permissions.name as permission_name, permissions.description,
	warehouses.id, warehouses.name as warehouse_name, warehouses.lat, warehouses.lng from users
	left join user_permissions ON user_permissions.user_id = users.id
	left join permissions ON user_permissions.permission_handle = permissions.handle
	left join user_warehouses on user_warehouses.user_id = users.id
	left join warehouses on warehouses.id = user_warehouses.warehouse_id
	where email = $1 and users.is_active = true;
	`
	rows, err := s.Query(ctx, sql, email)
	if err != nil {
		log.Printf("failed to get user: %s", err.Error())
		return User{}, common.NewInternalServerError()
	}
	defer rows.Close()
	return createUserFromRows(rows)
}

func (s *UserRepository) GetUserById(ctx context.Context, id int) (User, error) {
	sql := `
	select users.id, users.email, users.first_name, users.last_name, users.password, 
	permissions.handle, permissions.name as permission_name, permissions.description,
	warehouses.id, warehouses.name as warehouse_name, warehouses.lat, warehouses.lng from users
	left join user_permissions ON user_permissions.user_id = users.id
	left join permissions ON user_permissions.permission_handle = permissions.handle
	left join user_warehouses on user_warehouses.user_id = users.id
	left join warehouses on warehouses.id = user_warehouses.warehouse_id
	where users.id = $1 and users.is_active = true;
	`
	rows, err := s.Query(ctx, sql, id)
	if err != nil {
		log.Printf("failed to get user: %s", err.Error())
		return User{}, common.NewInternalServerError()
	}
	defer rows.Close()
	return createUserFromRows(rows)
}

func createUserFromRows(rows pgx.Rows) (User, error) {
	permissionsMap := make(map[string]PermissionClaim)
	warehousesSlice := make([]warehouse.Warehouse, 0)
	var user User
	var hash *string
	for rows.Next() {
		var warehouseId pgtype.Int4
		var warehouseName, permissionName, permissionDesc, permissionHandle pgtype.Varchar
		var warehouseLat pgtype.Float8
		var warehouseLng pgtype.Float8
		err := rows.Scan(
			&user.Id, &user.Email, &user.FirstName, &user.LastName, &hash,
			&permissionHandle, &permissionName, &permissionDesc,
			&warehouseId, &warehouseName, &warehouseLat, &warehouseLng,
		)
		if err != nil {
			log.Printf("failed to scan user: %s", err.Error())
			return User{}, common.NewInternalServerError()
		}
		if permissionHandle.Status != pgtype.Null {
			permission := PermissionClaim{
				Handle:      permissionHandle.String,
				Name:        permissionName.String,
				Description: permissionDesc.String,
			}
			permissionsMap[permission.Handle] = permission
		}

		if warehouseId.Status != pgtype.Null {
			id := int(warehouseId.Int)
			warehouse := warehouse.Warehouse{
				Id:   &id,
				Name: warehouseName.String,
				Lat:  &warehouseLat.Float,
				Lng:  &warehouseLng.Float,
			}
			warehousesSlice = append(warehousesSlice, warehouse)
		}
	}
	user.Warehouses = warehousesSlice
	user.Permissions = permissionsMap
	user.Hash = hash
	user.IsActive = true
	return user, nil
}

func (r *UserRepository) BanUser(ctx context.Context, id int) error {
	sql := "UPDATE users SET is_active = false WHERE id = $1"
	_, err := r.Exec(ctx, sql, id)
	if err != nil {
		log.Printf("failed to ban user: %s", err.Error())
		return common.NewInternalServerError()
	}
	return nil
}
