package user

import "context"

type IRoleService interface {
	CreateRole(ctx context.Context, role RoleInput) error
}
type RoleService struct {
	repository IRoleRepository
}

func NewRoleService(repository IRoleRepository) IRoleService {
	return &RoleService{
		repository: repository,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, role RoleInput) error {
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
