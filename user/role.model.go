package user

type Role struct {
	Id                *int     `json:"id,omitempty"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	PermissionHandles []string `json:"permissionHandles"`
}
