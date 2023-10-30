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
		{Name: "system admin", IsSecret: true, Handle: sysAdminPermissionHandle},
	}
}

func generateHandle(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	return name
}
