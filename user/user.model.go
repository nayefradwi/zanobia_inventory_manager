package user

import "github.com/nayefradwi/zanobia_inventory_manager/warehouse"

type UserInput struct {
	Email             string   `json:"email"`
	Password          string   `json:"password"`
	FirstName         string   `json:"firstName"`
	LastName          string   `json:"lastName"`
	PermissionHandles []string `json:"permissionHandles"`
}

type User struct {
	Id          int                        `json:"id"`
	Email       *string                    `json:"email,omitempty"`
	FirstName   string                     `json:"firstName"`
	LastName    string                     `json:"lastName"`
	IsActive    bool                       `json:"isActive"`
	Hash        *string                    `json:"hash,omitempty"`
	Warehouses  []warehouse.Warehouse      `json:"warehouses,omitempty"`
	Permissions map[string]PermissionClaim `json:"permissions,omitempty"`
}

type UserLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u User) HasPermission(permissionHandle string) bool {
	_, ok := u.Permissions[permissionHandle]
	return ok
}
