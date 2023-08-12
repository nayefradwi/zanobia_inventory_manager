package user

type Permission struct {
	Id          *int   `json:"id,omitempty"`
	Handle      string `json:"handle"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSecret    bool   `json:"isSecret"`
}

func generateInitialPermissions() []Permission {
	return []Permission{
		{Name: "system admin", IsSecret: true, Handle: sysAdminPermissionHandle},
	}
}
