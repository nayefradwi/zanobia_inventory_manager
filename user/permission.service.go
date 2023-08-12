package user

import "context"

type IPermissionService interface {
	InitiateInitialPermissions(ctx context.Context) error
	CreatePermission(ctx context.Context, permission Permission) error
}

type PermissionService struct {
	repository IPermissionRepository
}

func NewPermissionService(repository IPermissionRepository) IPermissionService {
	return &PermissionService{
		repository: repository,
	}
}

func (s *PermissionService) InitiateInitialPermissions(ctx context.Context) error {
	permissions := generateInitialPermissions()
	err := s.repository.AddAll(ctx, permissions)
	if err != nil {
		return err
	}
	return nil
}

func (s *PermissionService) CreatePermission(ctx context.Context, permission Permission) error {
	validationErr := ValidatePermission(permission)
	if validationErr != nil {
		return validationErr
	}
	permission.Handle = generateHandle(permission.Name)
	err := s.repository.CreatePermssion(ctx, permission)
	if err != nil {
		return err
	}
	return nil
}
