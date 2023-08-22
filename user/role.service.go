package user

import "context"

type IRoleService interface {
	CreateRole(ctx context.Context, role Role) error
	GetRoles(ctx context.Context) ([]Role, error)
}
type RoleService struct {
	repository IRoleRepository
}

func NewRoleService(repository IRoleRepository) IRoleService {
	return &RoleService{
		repository: repository,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, role Role) error {
	validationErr := ValidateRole(role)
	if validationErr != nil {
		return validationErr
	}
	err := s.repository.CreateRole(ctx, role)
	if err != nil {
		return err
	}
	return nil
}

func (s *RoleService) GetRoles(ctx context.Context) ([]Role, error) {
	return s.repository.GetRoles(ctx)
}
