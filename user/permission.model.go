package user

import "strings"

type Permission struct {
	Id          *int   `json:"id,omitempty"`
	Handle      string `json:"handle"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSecret    bool   `json:"isSecret"`
}

type PermissionClaim struct {
	Handle      string `json:"handle"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func generateInitialPermissions() []Permission {
	return []Permission{
		{Name: "system admin", IsSecret: true, Handle: SysAdminPermissionHandle},
		{Name: "has user control", Handle: HasUserControlPermission},
		{Name: "has inventory control", Handle: HasInventoryControlPermission},
		{Name: "has product control", Handle: HasProductControlPermission},
		{Name: "has batch control", Handle: HasBatchControlPermission},
		{Name: "can delete product", Handle: CanDeleteProductPermission},
		{Name: "can delete batch", Handle: CanDeleteBatchPermission},
	}
}

func generateHandle(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	return name
}
